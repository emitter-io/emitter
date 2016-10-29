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
using Emitter.Collections;

namespace Emitter.Network
{
    /// <summary>
    /// Represents the SUBSCRIBE message from client to broker.
    /// </summary>
    public sealed class MqttPublishPacket : MqttPacket
    {
        #region Static Members

        private static readonly ConcurrentPool<MqttPublishPacket> Pool =
            new ConcurrentPool<MqttPublishPacket>("MQTT Publish Packets", (c) => new MqttPublishPacket());

        /// <summary>
        /// Acquires a new MQTT packet from the pool.
        /// </summary>
        /// <returns></returns>
        public static MqttPublishPacket Acquire()
        {
            return Pool.Acquire();
        }

        /// <summary>
        /// Acquires a new MQTT packet from the pool.
        /// </summary>
        /// <returns></returns>
        public static MqttPublishPacket Acquire(string channel, ArraySegment<byte> message)
        {
            var packet = Pool.Acquire();
            packet.Channel = channel;
            packet.Message = message;
            return packet;
        }

        #endregion Static Members

        #region Public Properties

        /// <summary>
        /// Gets or sets the channel.
        /// </summary>
        public string Channel;

        /// <summary>
        /// Gets or sets the message payload.
        /// </summary>
        public ArraySegment<byte> Message;

        /// <summary>
        /// Recycles the packet.
        /// </summary>
        public override void Recycle()
        {
            // Call the base
            base.Recycle();

            // Recycle the properties
            this.Channel = null;
            this.Message = default(ArraySegment<byte>);
        }

        #endregion Public Properties

        #region Read/Write Members

        /// <summary>
        /// Reads the packet from the underlying stream.
        /// </summary>
        /// <param name="buffer">The pointer to start reading at.</param>
        /// <param name="offset">The offset to start at.</param>
        /// <param name="length">The remaining length.</param>
        /// <returns>Whether the packet was parsed successfully or not.</returns>
        public override unsafe MqttStatus TryRead(BufferSegment buffer, int offset, int length)
        {
            // Get the pointer
            var pBuffer = buffer.AsBytePointer();
            var header = offset;

            // Read the topic name
            this.Channel = ReadString(pBuffer, ref offset);

            // Read the QoS level, dup and retain from the header
            this.QoS = (QoS)((*pBuffer & QOS_LEVEL_MASK) >> QOS_LEVEL_OFFSET);
            this.DupFlag = (((*pBuffer & DUP_FLAG_MASK) >> DUP_FLAG_OFFSET) == 0x01);
            this.Retain = (((*pBuffer & RETAIN_FLAG_MASK) >> RETAIN_FLAG_OFFSET) == 0x01);

            // message id is valid only with QOS level 1 or QOS level 2
            if (this.IsEmitter || this.QoS != QoS.AtMostOnce)
                this.MessageId = (ushort)ReadUInt16(pBuffer, ref offset);

            // Copy the message in the segment
            var segment = buffer.AsSegment();
            this.Message = new ArraySegment<byte>(segment.Array, segment.Offset + offset, length - offset + header);

            // => TEST
            //var message = new byte[length - offset + header];
            //Memory.Copy(segment.Array, segment.Offset + offset, message, 0, length - offset + header);
            //this.Message = new ArraySegment<byte>(message);

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
            // The offset
            var buffer = segment.Array;
            int offset = segment.Offset;

            // Calculate the variable header size
            int headerLength = this.Channel.Length + MqttPacket.HeaderSize;

            // message id is valid only with QOS level 1 or QOS level 2
            if ((this.QoS == QoS.AtLeastOnce) || (this.QoS == QoS.ExactlyOnce))
                headerLength += MESSAGE_ID_SIZE;

            // Calculate the payload size and the remaining length
            int payloadLength = this.Message.Count;

            // Write first fixed header byte
            buffer[offset] = (byte)((MQTT_MSG_PUBLISH_TYPE << MSG_TYPE_OFFSET) | ((byte)this.QoS << QOS_LEVEL_OFFSET));
            buffer[offset] |= this.DupFlag ? (byte)(1 << DUP_FLAG_OFFSET) : (byte)0x00;
            buffer[offset] |= this.Retain ? (byte)(1 << RETAIN_FLAG_OFFSET) : (byte)0x00;
            ++offset;

            // Write the length
            WriteLength(headerLength + payloadLength, buffer, ref offset);

            // Write the channel name
            WriteString(this.Channel, buffer, ref offset);

            // Write message id, valid only with QOS level 1 or QOS level 2
            if (this.QoS != QoS.AtMostOnce)
            {
                buffer[offset++] = (byte)((this.MessageId >> 8) & 0x00FF); // MSB
                buffer[offset++] = (byte)(this.MessageId & 0x00FF); // LSB
            }

            // Write the payload, if we have any
            if (this.Message.Count > 0)
                Memory.Copy(this.Message.Array, this.Message.Offset, buffer, offset, this.Message.Count);

            // Calculate the size of the buffer
            return GetFixedHeaderSize(headerLength + payloadLength)
                + headerLength
                + payloadLength;
        }

        #endregion Read/Write Members
    }
}