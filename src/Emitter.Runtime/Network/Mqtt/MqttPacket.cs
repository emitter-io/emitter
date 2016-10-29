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
    /// Represents a MQTT packet base.
    /// </summary>
    public abstract class MqttPacket : Packet
    {
        // The size of the MQTT header
        public const int HeaderSize = 2;

        // This is the maximum message size we can send/receive.
        public const int MaxPayloadSize = 64 * 1024;

        #region Constructors

        /// <summary>
        /// Gets or sets the protocol version.
        /// </summary>
        public MqttProtocolVersion ProtocolVersion;

        /// <summary>
        /// Gets whether the client is implemented by us.
        /// </summary>
        public bool IsEmitter;

        /// <summary>
        /// Duplicate message flag
        /// </summary>
        public bool DupFlag;

        /// <summary>
        /// Quality of Service level
        /// </summary>
        public QoS QoS;

        /// <summary>
        /// Retain message flag
        /// </summary>
        public bool Retain;

        /// <summary>
        /// Message identifier for the message
        /// </summary>
        public ushort MessageId;

        /// <summary>
        /// Recycles the packet.
        /// </summary>
        public override void Recycle()
        {
            this.Origin = null;
            this.Lifetime = PacketLifetime.Automatic;
            this.ProtocolVersion = MqttProtocolVersion.Unknown;
            this.DupFlag = false;
            this.QoS = QoS.AtMostOnce;
            this.Retain = false;
        }

        #endregion Constructors

        #region Constants

        // mask, offset and size for fixed header fields
        internal const byte MSG_TYPE_MASK = 0xF0;

        internal const byte MSG_TYPE_OFFSET = 0x04;
        internal const byte MSG_TYPE_SIZE = 0x04;
        internal const byte MSG_FLAG_BITS_MASK = 0x0F;      // [v3.1.1]
        internal const byte MSG_FLAG_BITS_OFFSET = 0x00;    // [v3.1.1]
        internal const byte MSG_FLAG_BITS_SIZE = 0x04;      // [v3.1.1]
        internal const byte DUP_FLAG_MASK = 0x08;
        internal const byte DUP_FLAG_OFFSET = 0x03;
        internal const byte DUP_FLAG_SIZE = 0x01;
        internal const byte QOS_LEVEL_MASK = 0x06;
        internal const byte QOS_LEVEL_OFFSET = 0x01;
        internal const byte QOS_LEVEL_SIZE = 0x02;
        internal const byte RETAIN_FLAG_MASK = 0x01;
        internal const byte RETAIN_FLAG_OFFSET = 0x00;
        internal const byte RETAIN_FLAG_SIZE = 0x01;

        // MQTT message types
        internal const byte MQTT_MSG_CONNECT_TYPE = 0x01;

        internal const byte MQTT_MSG_CONNACK_TYPE = 0x02;
        internal const byte MQTT_MSG_PUBLISH_TYPE = 0x03;
        internal const byte MQTT_MSG_PUBACK_TYPE = 0x04;
        internal const byte MQTT_MSG_PUBREC_TYPE = 0x05;
        internal const byte MQTT_MSG_PUBREL_TYPE = 0x06;
        internal const byte MQTT_MSG_PUBCOMP_TYPE = 0x07;
        internal const byte MQTT_MSG_SUBSCRIBE_TYPE = 0x08;
        internal const byte MQTT_MSG_SUBACK_TYPE = 0x09;
        internal const byte MQTT_MSG_UNSUBSCRIBE_TYPE = 0x0A;
        internal const byte MQTT_MSG_UNSUBACK_TYPE = 0x0B;
        internal const byte MQTT_MSG_PINGREQ_TYPE = 0x0C;
        internal const byte MQTT_MSG_PINGRESP_TYPE = 0x0D;
        internal const byte MQTT_MSG_DISCONNECT_TYPE = 0x0E;

        // [v3.1.1] MQTT flag bits
        internal const byte MQTT_MSG_CONNECT_FLAG_BITS = 0x00;

        internal const byte MQTT_MSG_CONNACK_FLAG_BITS = 0x00;
        internal const byte MQTT_MSG_PUBLISH_FLAG_BITS = 0x00; // just defined as 0x00 but depends on publish props (dup, qos, retain)
        internal const byte MQTT_MSG_PUBACK_FLAG_BITS = 0x00;
        internal const byte MQTT_MSG_PUBREC_FLAG_BITS = 0x00;
        internal const byte MQTT_MSG_PUBREL_FLAG_BITS = 0x02;
        internal const byte MQTT_MSG_PUBCOMP_FLAG_BITS = 0x00;
        internal const byte MQTT_MSG_SUBSCRIBE_FLAG_BITS = 0x02;
        internal const byte MQTT_MSG_SUBACK_FLAG_BITS = 0x00;
        internal const byte MQTT_MSG_UNSUBSCRIBE_FLAG_BITS = 0x02;
        internal const byte MQTT_MSG_UNSUBACK_FLAG_BITS = 0x00;
        internal const byte MQTT_MSG_PINGREQ_FLAG_BITS = 0x00;
        internal const byte MQTT_MSG_PINGRESP_FLAG_BITS = 0x00;
        internal const byte MQTT_MSG_DISCONNECT_FLAG_BITS = 0x00;

        // QOS levels
        public const byte QOS_LEVEL_AT_MOST_ONCE = 0x00;

        public const byte QOS_LEVEL_AT_LEAST_ONCE = 0x01;
        public const byte QOS_LEVEL_EXACTLY_ONCE = 0x02;

        // SUBSCRIBE QoS level granted failure [v3.1.1]
        public const byte QOS_LEVEL_GRANTED_FAILURE = 0x80;

        internal const ushort MAX_TOPIC_LENGTH = 65535;
        internal const ushort MIN_TOPIC_LENGTH = 1;
        internal const byte MESSAGE_ID_SIZE = 2;
        #endregion Constants

        #region Abstract Members

        /// <summary>
        /// Reads the packet from the underlying stream.
        /// </summary>
        /// <param name="buffer">The pointer to start reading at.</param>
        /// <param name="offset">The offset to start at.</param>
        /// <param name="length">The remaining length.</param>
        /// <returns>Whether the packet was parsed successfully or not.</returns>
        public abstract unsafe MqttStatus TryRead(BufferSegment buffer, int offset, int length);

        /// <summary>
        /// Writes the packet into the provided <see cref="BufferSegment"/>.
        /// </summary>
        /// <param name="segment">The buffer to write into.</param>
        /// <returns>The length written.</returns>
        public abstract unsafe int TryWrite(ArraySegment<byte> segment);

        #endregion Abstract Members

        #region Read Members

        /// <summary>
        /// Read a byte which conforms with the MQTT protocol.
        /// </summary>
        /// <param name="pBuffer">The buffer to read at.</param>
        /// <param name="offset">The current offset.</param>
        /// <returns>The byte read.</returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        protected static unsafe int ReadByte(byte* pBuffer, ref int offset)
        {
            return *(pBuffer + offset++);
        }

        /// <summary>
        /// Read a number which conforms with the MQTT protocol.
        /// </summary>
        /// <param name="pBuffer">The buffer to read at.</param>
        /// <param name="offset">The current offset.</param>
        /// <returns>The number read.</returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        protected static unsafe int ReadUInt16(byte* pBuffer, ref int offset)
        {
            return ((*(pBuffer + offset++) << 8) & 0xFF00) | *(pBuffer + offset++);
        }

        /// <summary>
        /// Read a string which conforms with the MQTT protocol.
        /// </summary>
        /// <param name="pBuffer">The buffer to read at.</param>
        /// <param name="offset">The current offset.</param>
        /// <returns>The string read.</returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        protected unsafe string ReadString(byte* pBuffer, ref int offset)
        {
            int length = ((*(pBuffer + offset++) << 8) & 0xFF00) | *(pBuffer + offset++);
            //var result = new string((sbyte*)(pBuffer), offset, length);
            var result = Memory.CopyString(pBuffer + offset, length);
            offset += length;
            return result;
        }

        #endregion Read Members

        #region Write Members

        /// <summary>
        /// Gets the size of the fixed header based on the length of the message.
        /// </summary>
        /// <param name="length">The length to calculate for.</param>
        /// <returns></returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        protected static unsafe int GetFixedHeaderSize(int length)
        {
            var fixedHeaderSize = 1;
            do
            {
                ++fixedHeaderSize;
                length = length / 128;
            } while (length > 0);
            return fixedHeaderSize;
        }

        /// <summary>
        /// Writes the remaining length into the message buffer.
        /// </summary>
        /// <param name="remainingLength">Remaining length value to encode</param>
        /// <param name="buffer">Message buffer for inserting encoded value</param>
        /// <param name="offset">Index from which insert encoded value into buffer</param>
        /// <returns>Index updated</returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        protected static void WriteLength(int remainingLength, byte[] buffer, ref int offset)
        {
            int digit = 0;
            do
            {
                digit = remainingLength % 128;
                remainingLength /= 128;
                if (remainingLength > 0)
                    digit = digit | 0x80;
                buffer[offset++] = (byte)digit;
            } while (remainingLength > 0);
        }

        /// <summary>
        /// Writes a string to the specified buffer.
        /// </summary>
        /// <param name="value">The string value to write.</param>
        /// <param name="buffer">The buffer to write to.</param>
        /// <param name="offset">The offset within the buffer.</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        protected static unsafe void WriteString(string value, byte[] buffer, ref int offset)
        {
            var length = value.Length;
            buffer[offset++] = (byte)((length >> 8) & 0x00FF); // MSB
            buffer[offset++] = (byte)(length & 0x00FF); // LSB
            fixed (char* pChannel = value)
            {
                for (int i = 0; i < length; ++i)
                    buffer[offset++] = (byte)*(pChannel + i);
            }
        }

        #endregion Write Members
    }
}