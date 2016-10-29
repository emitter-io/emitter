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
using Emitter.Diagnostics;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a encoder of various MESH informs.
    /// </summary>
    public static unsafe class MeshEncoder
    {
        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Emitter.Connection channel, ProcessingContext context)
        {
            // It can only process packets
            var packet = context.Packet as MeshPacket;
            if (packet == null)
                return ProcessingState.Failure;

            if (channel.Client.TransportType != TransportType.Tcp)
                return ProcessingState.Failure;

            var command = packet as MeshEvent;
            if (command != null && command.Type == MeshEventType.Custom)
                NetTrace.WriteLine("Sending " + command.Value, channel, NetTraceCategory.Mesh);

            // Prepare a stream variable
            var stream = ByteStreamPool.Default.Acquire();
            var writer = PacketWriterPool.Default.Acquire(stream);
            try
            {
                // Write and compile the packet
                packet.Write(writer);

                // Compress the packet
                var encoded = SnappyEncoder.Process(stream, PacketHeader.TotalSize, context);
                int messageLength = encoded.Size - PacketHeader.TotalSize;

                // Add a sample to the average
                NetStat.Compression.Sample(1 - ((double)messageLength / stream.Length));

                // Write mesh prefix
                encoded.Array[encoded.Offset + 0] = 77; // 'M'
                encoded.Array[encoded.Offset + 1] = 83; // 'S'
                encoded.Array[encoded.Offset + 2] = 72; // 'H'

                // Write message type
                if (packet is MeshFrame)
                    encoded.Array[encoded.Offset + 3] = 70; // 'F'
                else if (packet is MeshEvent)
                    encoded.Array[encoded.Offset + 3] = 67; // 'C'

                // Write message length
                encoded.Array[encoded.Offset + 4] = (byte)(messageLength >> 24);
                encoded.Array[encoded.Offset + 5] = (byte)(messageLength >> 16);
                encoded.Array[encoded.Offset + 6] = (byte)(messageLength >> 8);
                encoded.Array[encoded.Offset + 7] = (byte)messageLength;

                // Copy to a buffer segment
                context.SwitchBuffer(encoded);
            }
            finally
            {
                // Make sure we release the stream and the writer
                stream.TryRelease();
                writer.TryRelease();
            }

            return ProcessingState.Success;
        }
    }
}