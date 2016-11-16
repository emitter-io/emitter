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
    /// Represents a handler of various MQTT requests.
    /// </summary>
    internal static unsafe class MqttHandler
    {
        #region OnConnect

        /// <summary>
        /// Invoked when a connect packet is received.
        /// </summary>
        /// <param name="client">The client sending the packet.</param>
        /// <param name="packet">The packet received from the client.</param>
        public static ProcessingState OnConnect(IClient client, MqttConnectPacket packet)
        {
            // Set the protocol version to the client
            client.Context = new MqttContext(packet);
            switch (packet.ProtocolVersion)
            {
                case MqttProtocolVersion.V3_1: break;
                case MqttProtocolVersion.V3_1_1: break;
                default:

                    // We don't have the protocol version
                    client.SendMqttConack(MqttConnackStatus.RefusedProtocolVersion, false);
                    return ProcessingState.Stop;
            }

            // Successfully connected
            client.SendMqttConack(MqttConnackStatus.Accepted, false);

            // Stop the processing
            return ProcessingState.Stop;
        }

        #endregion OnConnect

        #region OnSubscribe

        /// <summary>
        /// Invoked when a subscribe packet is received.
        /// </summary>
        /// <param name="client">The client sending the packet.</param>
        /// <param name="packet">The packet received from the client.</param>
        public static ProcessingState OnSubscribe(IClient client, MqttSubscribePacket packet)
        {
            // Prepare the ACK
            var ack = MqttSubackPacket.Acquire(packet.MessageId);
            var codes = ack.Codes;

            // We might have several channels to subscribe at once, iterate through
            for (int i = 0; i < packet.Channels.Count; ++i)
            {
                // Subscribe the channel
                var status = HandleSubscribe.Process(client, packet.Channels[i]);
                if (status != EmitterEventCode.Success)
                    Service.Logger.Log("Subscribe failed with code: " + status.ToString());

                // Add the QoS code to the packet
                codes.Add(status == EmitterEventCode.Success
                    ? QoS.AtMostOnce
                    : QoS.Failure);
            }

            // Send the ack and stop the processing
            client.Send(ack);
            return ProcessingState.Stop;
        }

        #endregion OnSubscribe

        #region OnUnsubscribe

        /// <summary>
        /// Invoked when an unsubscribe packet is received.
        /// </summary>
        /// <param name="client">The client sending the packet.</param>
        /// <param name="packet">The packet received from the client.</param>
        public static ProcessingState OnUnsubscribe(IClient client, MqttUnsubscribePacket packet)
        {
            // We might have several channels to unsubscribe at once, iterate through
            for (int i = 0; i < packet.Channels.Count; ++i)
            {
                // Subscribe the channel
                HandleUnsubscribe.Process(client, packet.Channels[i]);
            }

            // Send the ack and stop the processing
            client.SendMqttUnsuback(packet.MessageId);
            return ProcessingState.Stop;
        }

        #endregion OnUnsubscribe

        #region OnPublish

        /// <summary>
        /// Invoked when a publish packet is received.
        /// </summary>
        /// <param name="client">The client sending the packet.</param>
        /// <param name="packet">The packet received from the client.</param>
        public static ProcessingState OnPublish(IClient client, MqttPublishPacket packet)
        {
            // Publish through emitter
            HandlePublish.Process(client, packet.Channel, packet.Message);

            // Send the ack and stop the processing
            if (client.Context.QoS > QoS.AtMostOnce)
                client.SendMqttPuback(packet.MessageId);

            return ProcessingState.Stop;
        }

        #endregion OnPublish

        #region OnPublishAck

        /// <summary>
        /// Invoked when a publish ack packet is received.
        /// </summary>
        /// <param name="client">The client sending the packet.</param>
        /// <param name="packet">The packet received from the client.</param>
        public static ProcessingState OnPublishAck(IClient client, MqttPubackPacket packet)
        {
            // Stop the processing
            return ProcessingState.Stop;
        }

        #endregion OnPublishAck

        #region OnPing

        /// <summary>
        /// Invoked when a ping packet is received.
        /// </summary>
        /// <param name="client">The client sending the packet.</param>
        /// <param name="packet">The packet received from the client.</param>
        public static ProcessingState OnPing(IClient client, MqttPingReqPacket packet)
        {
            // Send ping response
            client.SendMqttPingResp();

            // Stop the processing
            return ProcessingState.Stop;
        }

        #endregion OnPing

        #region OnDisconnect

        /// <summary>
        /// Invoked when a disconnect packet is received.
        /// </summary>
        /// <param name="client">The client sending the packet.</param>
        /// <param name="packet">The packet received from the client.</param>
        public static ProcessingState OnDisconnect(IClient client, MqttDisconnectPacket packet)
        {
            // Close the connection
            client.Dispose();

            // Stop the processing
            return ProcessingState.Stop;
        }

        #endregion OnDisconnect
    }
}