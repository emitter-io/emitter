#region Copyright (c) 2009-2016 Misakai Ltd.
/*************************************************************************
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU Affero General Public License as
* published by the Free Software Foundation, either version 3 of the
* License, or(at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
*  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.See the
* GNU Affero General Public License for more details.
*
* You should have received a copy of the GNU Affero General Public License
* along with this program.If not, see<http://www.gnu.org/licenses/>.
*************************************************************************/
#endregion Copyright (c) 2009-2016 Misakai Ltd.

using System;
using System.Linq;
using System.Runtime.CompilerServices;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a decoder of various MESH requests.
    /// </summary>
    public static unsafe class MeshDecoder
    {
        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Emitter.Connection channel, ProcessingContext context)
        {
            // Get the buffer
            if (context.Buffer == null || context.Buffer.Length < PacketHeader.TotalSize)
                return ProcessingState.Failure;

            // Check for 'MSH' prefix
            var pBuffer = context.Buffer.AsBytePointer();
            if (*(pBuffer) != 'M' ||
                *(pBuffer + 1) != 'S' ||
                *(pBuffer + 2) != 'H')
                return ProcessingState.Failure;

            // Get the type of the mesh frame
            var type = (char)(*(pBuffer + 3));
            if (type != 'F' && type != 'C')
                return ProcessingState.Failure;

            // Must be a TCP channel
            var tcpChannel = channel as Connection;
            if (tcpChannel == null)
                return ProcessingState.Failure;

            // Get the packet length
            int length = PeekPacketLength(pBuffer + 4);
            if (length < 0 || length > context.MaxMemory)
                return ProcessingState.Failure;

            // Check whether we have all the bytes
            if (context.Buffer.Length < length + PacketHeader.TotalSize)
                return ProcessingState.InsufficientData;

            // There might be several packets in the same segment. We need to specify
            // that one is decoded and forward only that one to the next decoder.
            // However, we must not discard the segment completely as we might loose data!
            context.Throttle(PacketHeader.TotalSize + length);
            context.SwitchBuffer(
                context.Buffer.Split(PacketHeader.TotalSize)
                );

            // Decompress with Snappy
            if (Decode.Snappy(channel, context) != ProcessingState.Success)
                return ProcessingState.Failure;

            try
            {
                // Update the last seen, if we have a mesh identifier
                var server = Service.Mesh.Members.Get(tcpChannel.MeshIdentifier);
                if (server != null)
                    server.LastTouchUtc = Timer.UtcNow;

                // Something went wrong?
                if (context.Buffer == null)
                    return ProcessingState.Stop;

                switch (type)
                {
                    // This represents a mesh 'Command' type including commands that need to establish mesh
                    // handshake between the nodes in the cluster. In that case, we don't have an ID set
                    // on the connection and must let everything pass.
                    case 'C':

                        // Acquire a command packet and the reader
                        using (var reader = PacketReaderPool.Default.Acquire(context.Buffer))
                        using (var packet = Create((MeshEventType)reader.PeekByte()))
                        {
                            // Deserialize the packet
                            packet.Read(reader);

                            // Process the command and exit
                            MeshHandler.Process(channel as Connection, packet);
                            break;
                        }

                    // This represents a mesh 'Frame', which is a simple array segment forwarded to a set
                    // of handlers, in the way in which the Fabric is done as well.
                    case 'F':

                        // Process the frame and exit
                        Service.Mesh.OnFrame(server, context.Buffer);
                        break;

                    // We've received some weird data
                    default:
                        Service.Logger.Log("Mesh: Unknown frame type '" + type + "'");
                        break;
                }
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }

            // We're done here, stop the processing. The buffer will be freed right after
            // we've returned here, in the ProcessingContext's code.
            return ProcessingState.Stop;
        }

        /// <summary>
        /// Gets the packet data length.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static int PeekPacketLength(byte* pBuffer)
        {
            return (*(pBuffer)) << 24
                    | (*(pBuffer + 1) << 16)
                    | (*(pBuffer + 2) << 8)
                    | (*(pBuffer + 3));
        }

        /// <summary>
        /// Acquires an appropriate event envelope.
        /// </summary>
        /// <returns></returns>
        private static MeshEvent Create(MeshEventType type)
        {
            // Get the appropriate packet
            switch (type)
            {
                case MeshEventType.Error:
                case MeshEventType.Ping:
                case MeshEventType.PingAck:
                case MeshEventType.HandshakeAck:
                    return MeshEvent.Acquire();

                case MeshEventType.Handshake:
                    return MeshHandshake.Acquire();

                case MeshEventType.GossipDigest:
                case MeshEventType.GossipSince:
                case MeshEventType.GossipUpdate:
                    return MeshGossip.Acquire();

                case MeshEventType.Subscribe:
                case MeshEventType.Unsubscribe:
                case MeshEventType.Custom:
                    return MeshEmitterEvent.Acquire();

                default:
                    throw new InvalidOperationException("Unknown mesh packet");
            }
        }
    }
}