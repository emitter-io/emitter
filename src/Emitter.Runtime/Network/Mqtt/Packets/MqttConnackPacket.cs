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
    /// Represents the CONNACK message from broker to client.
    /// </summary>
    internal sealed class MqttConnackPacket : MqttPacket
    {
        #region Static Members

        private static readonly ConcurrentPool<MqttConnackPacket> Pool =
            new ConcurrentPool<MqttConnackPacket>("MQTT ConnAck Packets", (c) => new MqttConnackPacket());

        /// <summary>
        /// Acquires a new MQTT packet from the pool.
        /// </summary>
        /// <returns></returns>
        public static MqttConnackPacket Acquire(MqttConnackStatus type, bool sessionPresent)
        {
            var packet = Pool.Acquire();
            packet.ReturnCode = type;
            packet.SessionPresent = sessionPresent;
            return packet;
        }

        #endregion Static Members

        #region Constants
        private const byte TOPIC_NAME_COMP_RESP_BYTE_OFFSET = 0;
        private const byte TOPIC_NAME_COMP_RESP_BYTE_SIZE = 1;

        // [v3.1.1] connect acknowledge flags replace "old" topic name compression respone (not used in 3.1)
        private const byte CONN_ACK_FLAGS_BYTE_OFFSET = 0;

        private const byte CONN_ACK_FLAGS_BYTE_SIZE = 1;

        // [v3.1.1] session present flag
        private const byte SESSION_PRESENT_FLAG_MASK = 0x01;

        private const byte SESSION_PRESENT_FLAG_OFFSET = 0x00;
        private const byte SESSION_PRESENT_FLAG_SIZE = 0x01;
        private const byte CONN_RETURN_CODE_BYTE_OFFSET = 1;
        private const byte CONN_RETURN_CODE_BYTE_SIZE = 1;
        #endregion Constants

        #region Public Properties

        /// <summary>
        /// Gets or sets the ack type.
        /// </summary>
        public MqttConnackStatus ReturnCode;

        /// <summary>
        /// The session present flag [v3.1.1]
        /// </summary>
        public bool SessionPresent;

        /// <summary>
        /// Recycles the packet.
        /// </summary>
        public override void Recycle()
        {
            // Call the base
            base.Recycle();

            // Recycle the properties
            this.ReturnCode = MqttConnackStatus.RefusedNotAuthorized;
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
            throw new NotImplementedException();
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
            int headerLength = this.ProtocolVersion == MqttProtocolVersion.V3_1_1
                ? (CONN_ACK_FLAGS_BYTE_SIZE + CONN_RETURN_CODE_BYTE_SIZE)
                : (TOPIC_NAME_COMP_RESP_BYTE_SIZE + CONN_RETURN_CODE_BYTE_SIZE);

            // Calculate the payload size and the remaining length
            const int payloadLength = 0;

            // Write the type
            buffer[offset++] = ProtocolVersion == MqttProtocolVersion.V3_1_1
                ? (byte)((MQTT_MSG_CONNACK_TYPE << MSG_TYPE_OFFSET) | MQTT_MSG_CONNACK_FLAG_BITS)
                : (byte)(MQTT_MSG_CONNACK_TYPE << MSG_TYPE_OFFSET);

            // Write the length
            WriteLength(headerLength + payloadLength, buffer, ref offset);

            // Write the session present flag
            buffer[offset++] = ProtocolVersion == MqttProtocolVersion.V3_1_1
                ? (this.SessionPresent ? (byte)(1 << SESSION_PRESENT_FLAG_OFFSET) : (byte)0x00)
                : (byte)0x00;

            // Write the return code
            buffer[offset++] = (byte)this.ReturnCode;

            // Calculate the size of the buffer
            return GetFixedHeaderSize(headerLength + payloadLength)
                + headerLength
                + payloadLength;
        }

        #endregion Read/Write Members
    }

    /// <summary>
    /// Represents various return codes for the ack.
    /// </summary>
    public enum MqttConnackStatus : byte
    {
        Accepted = 0x00,
        RefusedProtocolVersion = 0x01,
        RefusedIdentityReject = 0x02,
        RefusedServerUnavalable = 0x03,
        RefusedUsernamePassword = 0x04,
        RefusedNotAuthorized = 0x05
    }
}