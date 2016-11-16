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

namespace Emitter.Network
{
    /// <summary>
    /// Represents a MQTT context containing some basic info.
    /// </summary>
    public sealed class MqttContext
    {
        /// <summary>
        /// Constructs a context from the connect packet.
        /// </summary>
        /// <param name="packet">The packet to use.</param>
        internal MqttContext(MqttConnectPacket packet)
        {
            this.Version = packet.ProtocolVersion;
            this.QoS = packet.QoS;
            this.IsEmitter = packet.IsEmitter;
            this.ClientId = packet.ClientId;
            this.Username = packet.Username;
        }

        /// <summary>
        /// Creates a new context.
        /// </summary>
        /// <param name="version">The version of MQTT.</param>
        /// <param name="isEmitter">Whether this is our special implementation.</param>
        /// <param name="id">The client id specified in the MQTT connect packet.</param>
        public MqttContext(MqttProtocolVersion version, QoS qos, bool isEmitter, string id, string username)
        {
            this.Version = version;
            this.QoS = qos;
            this.IsEmitter = isEmitter;
            this.ClientId = id;
            this.Username = username;
        }

        /// <summary>
        /// The version of MQTT.
        /// </summary>
        public readonly MqttProtocolVersion Version;

        /// <summary>
        /// Whether this is our special implementation.
        /// </summary>
        public readonly bool IsEmitter;

        /// <summary>
        /// Gets the quality of service requested
        /// </summary>
        public readonly QoS QoS;

        /// <summary>
        /// Gets the MQTT client id passed during the connect.
        /// </summary>
        public readonly string ClientId;

        /// <summary>
        /// Gets the MQTT username.
        /// </summary>
        public readonly string Username;
    }
}