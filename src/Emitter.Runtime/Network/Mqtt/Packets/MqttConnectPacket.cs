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
    /// Represents the CONNECT message from client to broker.
    /// </summary>
    internal sealed class MqttConnectPacket : MqttPacket
    {
        #region Static Members

        private static readonly ConcurrentPool<MqttConnectPacket> Pool =
            new ConcurrentPool<MqttConnectPacket>("MQTT Connect Packets", (c) => new MqttConnectPacket());

        /// <summary>
        /// Acquires a new MQTT packet from the pool.
        /// </summary>
        /// <returns></returns>
        public static MqttConnectPacket Acquire()
        {
            return Pool.Acquire();
        }

        #endregion Static Members

        #region Constants

        // protocol name supported
        internal const string PROTOCOL_NAME_V31 = "MQIsdp"; // [v.3.1]

        internal const string PROTOCOL_NAME_V311 = "MQTT";   // [v.3.1.1]

        // max length for client id (removed in 3.1.1)
        internal const int CLIENT_ID_MAX_LENGTH = 23;

        // variable header fields
        internal const byte PROTOCOL_NAME_LEN_SIZE = 2;

        internal const byte PROTOCOL_NAME_V3_1_SIZE = 6;
        internal const byte PROTOCOL_NAME_V3_1_1_SIZE = 4; // [v.3.1.1]
        internal const byte PROTOCOL_VERSION_SIZE = 1;
        internal const byte CONNECT_FLAGS_SIZE = 1;
        internal const byte KEEP_ALIVE_TIME_SIZE = 2;

        internal const ushort KEEP_ALIVE_PERIOD_DEFAULT = 60; // seconds
        internal const ushort MAX_KEEP_ALIVE = 65535; // 16 bit

        // connect flags
        internal const byte USERNAME_FLAG_MASK = 0x80;

        internal const byte USERNAME_FLAG_OFFSET = 0x07;
        internal const byte USERNAME_FLAG_SIZE = 0x01;
        internal const byte PASSWORD_FLAG_MASK = 0x40;
        internal const byte PASSWORD_FLAG_OFFSET = 0x06;
        internal const byte PASSWORD_FLAG_SIZE = 0x01;
        internal const byte WILL_RETAIN_FLAG_MASK = 0x20;
        internal const byte WILL_RETAIN_FLAG_OFFSET = 0x05;
        internal const byte WILL_RETAIN_FLAG_SIZE = 0x01;
        internal const byte WILL_QOS_FLAG_MASK = 0x18;
        internal const byte WILL_QOS_FLAG_OFFSET = 0x03;
        internal const byte WILL_QOS_FLAG_SIZE = 0x02;
        internal const byte WILL_FLAG_MASK = 0x04;
        internal const byte WILL_FLAG_OFFSET = 0x02;
        internal const byte WILL_FLAG_SIZE = 0x01;
        internal const byte CLEAN_SESSION_FLAG_MASK = 0x02;
        internal const byte CLEAN_SESSION_FLAG_OFFSET = 0x01;
        internal const byte CLEAN_SESSION_FLAG_SIZE = 0x01;

        // [v.3.1.1] lsb (reserved) must be now 0
        internal const byte RESERVED_FLAG_MASK = 0x01;

        internal const byte RESERVED_FLAG_OFFSET = 0x00;
        internal const byte RESERVED_FLAG_SIZE = 0x01;
        #endregion Constants

        #region Public Properties

        /// <summary>
        /// Gets or sets the name of the protocol.
        /// </summary>
        public string ProtocolName;

        /// <summary>
        /// Gets or sets the client identifier.
        /// </summary>
        public string ClientId;

        /// <summary>
        /// Gets or sets the Will Topic.
        /// </summary>
        public string WillTopic;

        /// <summary>
        /// Gets or sets the Will Message.
        /// </summary>
        public string WillMessage;

        /// <summary>
        /// Gets or sets the username.
        /// </summary>
        public string Username;

        /// <summary>
        /// Gets or sets the password.
        /// </summary>
        public string Password;

        /// <summary>
        /// Recycles the packet.
        /// </summary>
        public override void Recycle()
        {
            // Call the base
            base.Recycle();

            // Recycle the properties
            this.ProtocolName = null;
            this.ProtocolVersion = MqttProtocolVersion.Unknown;
            this.ClientId = null;
            this.WillTopic = null;
            this.WillMessage = null;
            this.Username = null;
            this.Password = null;
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
            var pBuffer = buffer.AsBytePointer();

            // Read the protocol name
            this.ProtocolName = ReadString(pBuffer, ref offset);

            // Validate the protocol name
            if (!this.ProtocolName.Equals(PROTOCOL_NAME_V31) && !this.ProtocolName.Equals(PROTOCOL_NAME_V311))
                return MqttStatus.InvalidProtocolName;

            // Read the protocol version
            this.ProtocolVersion = (MqttProtocolVersion)(*(pBuffer + offset++));

            // [v3.1.1] check lsb (reserved) must be 0
            var flags = 0;
            flags = ReadByte(pBuffer, ref offset);

            //if ((this.ProtocolVersion == MqttProtocolVersion.V3_1_1) && ((flags & RESERVED_FLAG_MASK) != 0x00))
            //    return MqttStatus.InvalidConnectFlags;

            // Get whether this is our special implementation of MQTT
            this.IsEmitter = (flags & RESERVED_FLAG_MASK) != 0x00;

            // Read MQTT flags
            var isUsernameFlag = (flags & USERNAME_FLAG_MASK) != 0x00;
            var isPasswordFlag = (flags & PASSWORD_FLAG_MASK) != 0x00;
            var willRetain = (flags & WILL_RETAIN_FLAG_MASK) != 0x00;
            var willQosLevel = (byte)((flags & WILL_QOS_FLAG_MASK) >> WILL_QOS_FLAG_OFFSET);
            var willFlag = (flags & WILL_FLAG_MASK) != 0x00;
            var cleanSession = (flags & CLEAN_SESSION_FLAG_MASK) != 0x00;

            // Read the keep alive period
            var keepAlivePeriod = (ushort)ReadUInt16(pBuffer, ref offset);

            // client identifier [v3.1.1] it may be zero bytes long (empty string)
            this.ClientId = ReadString(pBuffer, ref offset);

            // [v3.1.1] if client identifier is zero bytes long, clean session must be true
            if ((this.ProtocolVersion == MqttProtocolVersion.V3_1_1) && (ClientId.Length == 0) && (!cleanSession))
                return MqttStatus.InvalidClientId;

            // Read will topic and will message
            if (willFlag)
            {
                this.WillTopic = ReadString(pBuffer, ref offset);
                this.WillMessage = ReadString(pBuffer, ref offset);
            }

            // Read the username
            if (isUsernameFlag)
                this.Username = ReadString(pBuffer, ref offset);

            // Read the password
            if (isPasswordFlag)
                this.Password = ReadString(pBuffer, ref offset);

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