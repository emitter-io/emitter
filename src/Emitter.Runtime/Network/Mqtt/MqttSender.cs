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
    /// Represents a sender of MQTT messages.
    /// </summary>
    public interface IMqttSender
    {
        /// <summary>
        /// Sends the message to the remote client, remote server or a function.
        /// </summary>
        /// <param name="contract">The contract for this message.</param>
        /// <param name="channel">The channel name to send.</param>
        /// <param name="message">The message to send.</param>
        void Send(int contract, string channel, ArraySegment<byte> message);

        /// <summary>
        /// Gets the connection id of the sender.
        /// </summary>
        ConnectionId Id { get; }

        /// <summary>
        /// Gets or sets the MQTT context associated with this sender.
        /// </summary>
        MqttContext Context { get; set; }
    }

    /// <summary>
    /// Represents a sender of MQTT messages.
    /// </summary>
    public sealed class MqttForward : IMqttSender
    {
        /// <summary>
        /// Gets or sets the forwarding function.
        /// </summary>
        private Action<int, string, ArraySegment<byte>> Forward;

        /// <summary>
        /// Gets the connection id.
        /// </summary>
        private ConnectionId Identity;

        /// <summary>
        /// Constructs a new sender for a function.
        /// </summary>
        /// <param name="forward"></param>
        public MqttForward(Action<int, string, ArraySegment<byte>> forward)
        {
            this.Forward = forward;
            this.Identity = ConnectionId.NewConnectionId();
        }

        /// <summary>
        /// Sends the message to the remote client, remote server or a function.
        /// </summary>
        /// <param name="contract">The contract for this message.</param>
        /// <param name="channel">The channel name to send.</param>
        /// <param name="message">The message to send.</param>
        public void Send(int contract, string channel, ArraySegment<byte> message)
        {
            // Call the function
            Forward(contract, channel, message);
        }

        /// <summary>
        /// Gets the connection id of the sender.
        /// </summary>
        public ConnectionId Id
        {
            get { return this.Identity; }
        }

        /// <summary>
        /// Gets or sets the MQTT context associated with this sender.
        /// </summary>
        public MqttContext Context
        {
            get
            {
                return new MqttContext(MqttProtocolVersion.VEmitter, QoS.AtMostOnce, true, "broker", "broker");
            }
            set
            {
                throw new NotImplementedException();
            }
        }
    }
}