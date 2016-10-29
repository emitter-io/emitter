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
    /// Represents extension methods for IClient.
    /// </summary>
    internal static unsafe class MqttExtension
    {
        public static void SendMqttConack(this IClient client, MqttConnackStatus status, bool sessionPresent)
        {
            client.Send(
                MqttConnackPacket.Acquire(status, sessionPresent)
                );
        }

        public static void SendMqttUnsuback(this IClient client, ushort messageId)
        {
            client.Send(
                MqttUnsubackPacket.Acquire(messageId)
                );
        }

        public static void SendMqttPublish(this IClient client, string channel, ArraySegment<byte> message)
        {
            client.Send(
                MqttPublishPacket.Acquire(channel, message)
                );
        }

        public static void SendMqttPuback(this IClient client, ushort messageId)
        {
            client.Send(
                MqttPubackPacket.Acquire(messageId)
                );
        }

        public static void SendMqttPingResp(this IClient client)
        {
            client.Send(
                MqttPingRespPacket.Acquire()
                );
        }
    }
}