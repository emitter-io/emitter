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
using System.Runtime.CompilerServices;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a decoder of various MQTT requests.
    /// </summary>
    public static unsafe class MqttDecoder
    {
        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Connection channel, ProcessingContext context)
        {
            // Get the buffer
            var buffer = context.Buffer;

            // Extract the message type first
            var pBuffer = buffer.AsBytePointer();
            if (!ValidMqtt(pBuffer))
                return ProcessingState.Failure;

            // Do we have a full header?
            if (buffer.Length < MqttPacket.HeaderSize)
                return ProcessingState.InsufficientData;

            // Get the length of the packet
            int headerSize;
            var length = PeekLength(pBuffer + 1, buffer.Length, out headerSize);
            if (length > MqttPacket.MaxPayloadSize)
                return ProcessingState.Failure;

            // Do we have enough data?
            if (length == -1 || (headerSize + length) > buffer.Length)
                return ProcessingState.InsufficientData;

            // There might be several packets in the same segment. We need to specify
            // that one is decoded and forward only that one to the next decoder.
            // However, we must not discard the segment completely as we might loose data!
            context.Throttle(headerSize + length);

            // Acquire the appropriate packet
            var msgType = (MqttPacketType)((*pBuffer & MqttPacket.MSG_TYPE_MASK) >> MqttPacket.MSG_TYPE_OFFSET);
            var packet = Acquire(msgType, channel.Client);
            if (packet == null)
                return ProcessingState.Stop;

            // If we don't have enough data...
            //if (length > buffer.Length)
            //    return ProcessingState.InsufficientData;

            // If we have protocol version set, add it to the packet
            var mqttCtx = channel.Client.Context;
            if (mqttCtx != null)
            {
                packet.ProtocolVersion = mqttCtx.Version;
                packet.IsEmitter = mqttCtx.IsEmitter;
            }

            try
            {
                // Read the packet
                packet.TryRead(context.Buffer, headerSize, length);

                // Call the appropriate handler
                switch (msgType)
                {
                    case MqttPacketType.Connect:
                        return MqttHandler.OnConnect(channel.Client, packet as MqttConnectPacket);

                    case MqttPacketType.Subscribe:
                        return MqttHandler.OnSubscribe(channel.Client, packet as MqttSubscribePacket);

                    case MqttPacketType.Unsubscribe:
                        return MqttHandler.OnUnsubscribe(channel.Client, packet as MqttUnsubscribePacket);

                    case MqttPacketType.PingReq:
                        return MqttHandler.OnPing(channel.Client, packet as MqttPingReqPacket);

                    case MqttPacketType.Disconnect:
                        return MqttHandler.OnDisconnect(channel.Client, packet as MqttDisconnectPacket);

                    case MqttPacketType.Publish:
                        return MqttHandler.OnPublish(channel.Client, packet as MqttPublishPacket);

                    case MqttPacketType.PubAck:
                        return MqttHandler.OnPublishAck(channel.Client, packet as MqttPubackPacket);

                    default: break;
                }
            }
            catch (Exception ex)
            {
                // Log the exception
                Service.Logger.Log(ex);
            }
            finally
            {
                // Release the packet now
                if (packet.Lifetime == PacketLifetime.Automatic)
                    packet.TryRelease();
            }

            // We're really done handling this, do not queue additional processors
            return ProcessingState.Stop;
        }

        #region Private Members

        /// <summary>
        /// Gets the <see cref="MqttPacket"/> for the provided type.
        /// </summary>
        /// <param name="type">The type to acquire.</param>
        /// <returns></returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static MqttPacket Acquire(MqttPacketType type, IClient client)
        {
            switch (type)
            {
                case MqttPacketType.Connect: return MqttConnectPacket.Acquire();
                case MqttPacketType.Subscribe: return MqttSubscribePacket.Acquire();
                case MqttPacketType.Unsubscribe: return MqttUnsubscribePacket.Acquire();
                case MqttPacketType.PingReq: return MqttPingReqPacket.Acquire();
                case MqttPacketType.Disconnect: return MqttDisconnectPacket.Acquire();
                case MqttPacketType.Publish: return MqttPublishPacket.Acquire();
                case MqttPacketType.PubAck: return MqttPubackPacket.Acquire();

                default:
                    Service.Logger.Log("Unknown MQTT Type: " + type);
                    return null;
            }
        }

        /// <summary>
        /// Checks whether the packet is a valid MQTT packet.
        /// </summary>
        /// <param name="pBuffer">The packet pointer.</param>
        /// <returns></returns>
        private static bool ValidMqtt(byte* pBuffer)
        {
            var type = (MqttPacketType)((*pBuffer & MqttPacket.MSG_TYPE_MASK) >> MqttPacket.MSG_TYPE_OFFSET);
            var validType = false;
            switch (type)
            {
                case MqttPacketType.Connect:
                case MqttPacketType.Subscribe:
                case MqttPacketType.Unsubscribe:
                case MqttPacketType.PingReq:
                case MqttPacketType.Disconnect:
                case MqttPacketType.Publish:
                case MqttPacketType.PubAck:
                    validType = true;
                    break;

                default: return false;
            }

            return validType;
        }

        /// <summary>
        /// Decode remaining length reading bytes from socket.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static int PeekLength(byte* pBuffer, int capacity, out int headerSize)
        {
            int multiplier = 1;
            int value = 0;
            int digit = 0;
            headerSize = 1;
            do
            {
                if (capacity < 2)
                    return -1;

                digit = *(pBuffer++);
                value += ((digit & 127) * multiplier);
                multiplier *= 128;
                ++headerSize;
                --capacity;
            } while ((digit & 128) != 0);
            return value;
        }

        #endregion Private Members
    }
}