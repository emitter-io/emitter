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
    /// The different types of messages defined in the MQTT protocol.
    /// </summary>
    public enum MqttPacketType : byte
    {
        Connect = 1,
        Connack = 2,
        Publish = 3,
        PubAck = 4,
        PubRec = 5,
        PubRel = 6,
        PubComp = 7,
        Subscribe = 8,
        Suback = 9,
        Unsubscribe = 10,
        Unsuback = 11,
        PingReq = 12,
        PingResp = 13,
        Disconnect = 14
    }

    /// <summary>
    /// Represents the protocol version.
    /// </summary>
    public enum MqttProtocolVersion : byte
    {
        Unknown = 0x00,
        VEmitter = 0x02,
        V3_1 = 0x03,
        V3_1_1 = 0x04
    }

    /// <summary>
    /// Represents MQTT client error code.
    /// </summary>
    public enum MqttStatus
    {
        /// <summary>
        /// Successfully parsed MQTT.
        /// </summary>
        Success = 0,

        /// <summary>
        /// Will error (topic, message or QoS level)
        /// </summary>
        WillWrong = 1,

        /// <summary>
        /// Keep alive period too large
        /// </summary>
        KeepAliveWrong,

        /// <summary>
        /// Topic contains wildcards
        /// </summary>
        TopicWildcard,

        /// <summary>
        /// Topic length wrong
        /// </summary>
        TopicLength,

        /// <summary>
        /// QoS level not allowed
        /// </summary>
        QosNotAllowed,

        /// <summary>
        /// Topics list empty for subscribe
        /// </summary>
        TopicsEmpty,

        /// <summary>
        /// Qos levels list empty for subscribe
        /// </summary>
        QosLevelsEmpty,

        /// <summary>
        /// Topics / Qos Levels not match in subscribe
        /// </summary>
        TopicsQosLevelsNotMatch,

        /// <summary>
        /// Wrong message from broker
        /// </summary>
        WrongBrokerMessage,

        /// <summary>
        /// Wrong Message Id
        /// </summary>
        WrongMessageId,

        /// <summary>
        /// Inflight queue is full
        /// </summary>
        InflightQueueFull,

        // [v3.1.1]
        /// <summary>
        /// Invalid flag bits received
        /// </summary>
        InvalidFlagBits,

        // [v3.1.1]
        /// <summary>
        /// Invalid connect flags received
        /// </summary>
        InvalidConnectFlags,

        // [v3.1.1]
        /// <summary>
        /// Invalid client id
        /// </summary>
        InvalidClientId,

        // [v3.1.1]
        /// <summary>
        /// Invalid protocol name
        /// </summary>
        InvalidProtocolName
    }

    /// <summary>
    /// Quality of service levels
    /// </summary>
    public enum QoS : byte
    {
        /// <summary>
        /// At most once quality of service.
        /// </summary>
        AtMostOnce = 0x00,

        /// <summary>
        /// At least once message should be delivered.
        /// </summary>
        AtLeastOnce = 0x01,

        /// <summary>
        /// Exactly once semantics. Not implemented.
        /// </summary>
        ExactlyOnce = 0x02,

        /// <summary>
        /// Failure error code.
        /// </summary>
        Failure = 0x80
    }
}