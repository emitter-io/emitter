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

using Emitter.Collections;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a encoder of various MQTT events.
    /// </summary>
    public static unsafe class MqttEncoder
    {
        /// <summary>
        /// The buffers for write.
        /// </summary>
        private readonly static ConcurrentPool<BufferSegment> Buffers =
            new ConcurrentPool<BufferSegment>("MQTT Buffers", (c) => new BufferSegment(MqttPacket.MaxPayloadSize), 128);

        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Connection channel, ProcessingContext context)
        {
            // It can only process packets
            var packet = context.Packet as MqttPacket;
            if (packet == null)
                return ProcessingState.Failure;

            // Check the transport type currently used for sending our packet
            if (channel.Client.TransportType != TransportType.Tcp)
                return ProcessingState.Failure;

            // Acquire a buffer which will be released once everything is sent
            var buffer = Buffers.Acquire();

            // If we have protocol version set, add it to the packet
            var mqttCtx = channel.Client.Context;
            if (mqttCtx != null)
            {
                packet.ProtocolVersion = mqttCtx.Version;
                packet.IsEmitter = mqttCtx.IsEmitter;
            }

            // Write the packet to the buffer
            buffer.Size = packet.TryWrite(buffer.AsSegment());

            // Set the buffer to send out
            context.SwitchBuffer(buffer);
            return ProcessingState.Success;
        }
    }
}