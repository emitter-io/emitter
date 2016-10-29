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
using System.Collections.Generic;
using Emitter.Collections;

namespace Emitter.Network
{
    /// <summary>
    /// Represents the UNSUBSCRIBE message from client to broker.
    /// </summary>
    internal sealed class MqttUnsubscribePacket : MqttPacket
    {
        #region Static Members

        private static readonly ConcurrentPool<MqttUnsubscribePacket> Pool =
            new ConcurrentPool<MqttUnsubscribePacket>("MQTT Unsubscribe Packets", (c) => new MqttUnsubscribePacket());

        /// <summary>
        /// Acquires a new MQTT packet from the pool.
        /// </summary>
        /// <returns></returns>
        public static MqttUnsubscribePacket Acquire()
        {
            return Pool.Acquire();
        }

        #endregion Static Members

        #region Public Properties

        /// <summary>
        /// Gets or sets the channels to subscribe to.
        /// </summary>
        public readonly List<string> Channels = new List<string>();

        /// <summary>
        /// Recycles the packet.
        /// </summary>
        public override void Recycle()
        {
            // Call the base
            base.Recycle();

            // Recycle the properties
            this.Channels.Clear();
        }

        #endregion Public Properties

        #region Read/Write Members

        /// <summary>
        /// Reads the packet from the underlying stream.
        /// </summary>
        /// <param name="buffer">The pointer to start reading at.</param>
        /// <param name="length">The remaining length.</param>
        /// <param name="offset">The offset to start at.</param>
        /// <returns>Whether the packet was parsed successfully or not.</returns>
        public override unsafe MqttStatus TryRead(BufferSegment buffer, int offset, int length)
        {
            // Get the pointer
            var pBuffer = buffer.AsBytePointer();

            // [v3.1.1] check flag bits
            if (this.ProtocolVersion == MqttProtocolVersion.V3_1_1
                && (*pBuffer & MSG_FLAG_BITS_MASK) != MQTT_MSG_UNSUBSCRIBE_FLAG_BITS)
                return MqttStatus.InvalidFlagBits;

            // only 3.1.0
            if (this.ProtocolVersion == MqttProtocolVersion.V3_1)
            {
                // read QoS level from fixed header
                this.QoS = (QoS)((*pBuffer & QOS_LEVEL_MASK) >> QOS_LEVEL_OFFSET);

                // read DUP flag from fixed header
                this.DupFlag = (((*pBuffer & DUP_FLAG_MASK) >> DUP_FLAG_OFFSET) == 0x01);

                // retain flag not used
                this.Retain = false;
            }

            // Read the message id
            this.MessageId = (ushort)ReadUInt16(pBuffer, ref offset);

            do
            {
                // Read the channel
                var channel = ReadString(pBuffer, ref offset);

                // Read the topic string and push it to the channels list
                this.Channels.Add(channel);
            } while (offset < length);

            // We've parsed it
            return MqttStatus.Success;
        }

        /// <summary>
        /// Writes the packet into the provided <see cref="BufferSegment"/>.
        /// </summary>
        /// <param name="segment">The buffer to write into.</param>
        /// <returns>The length written.</returns>
        public override unsafe int TryWrite(ArraySegment<byte> segment)
        {
            throw new NotImplementedException();
        }

        #endregion Read/Write Members
    }
}