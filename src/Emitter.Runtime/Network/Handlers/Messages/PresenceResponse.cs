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
using Emitter.Text.Json;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a key generation response.
    /// </summary>
    internal class PresenceResponse : EmitterResponse
    {
        /// <summary>
        /// Creates a response packet.
        /// </summary>
        /// <param name="channel">The target channel for this notification.</param>
        /// <param name="occupancy">The occupancy.</param>
        public PresenceResponse(string channel, int occupancy, List<PresenceInfo> clients)
        {
            this.Event = "status";
            this.Time = Timer.UtcNow.ToUnixTimestamp();
            this.Channel = channel;
            this.Who = clients;
            this.Occupancy = occupancy;
        }

        /// <summary>
        /// The action, must be "status".
        /// </summary>
        [JsonProperty("event")]
        public readonly string Event;

        /// <summary>
        /// The target channel for the key.
        /// </summary>
        [JsonProperty("channel")]
        public string Channel;

        /// <summary>
        /// The amount of connections present.
        /// </summary>
        [JsonProperty("occupancy")]
        public int Occupancy;

        /// <summary>
        /// The UNIX timestamp.
        /// </summary>
        [JsonProperty("time")]
        public readonly long Time;

        /// <summary>
        /// The list of subscribers.
        /// </summary>
        [JsonProperty("who")]
        public List<PresenceInfo> Who;
    }

    /// <summary>
    /// Represents a state notification.
    /// </summary>
    internal class PresenceNotification
    {
        /// <summary>
        /// Creates a notification packet.
        /// </summary>
        /// <param name="type">The event type.</param>
        /// <param name="channel">The target channel for this notification.</param>
        /// <param name="occupancy">The occupancy.</param>
        public PresenceNotification(PresenceEvent type, string channel, PresenceInfo who, int occupancy)
        {
            switch (type)
            {
                case PresenceEvent.Status: this.Event = "status"; break;
                case PresenceEvent.Subscribe: this.Event = "subscribe"; break;
                case PresenceEvent.Unsubscribe: this.Event = "unsubscribe"; break;
            }

            this.Time = Timer.UtcNow.ToUnixTimestamp();
            this.Channel = channel;
            this.Who = who;
            this.Occupancy = occupancy;
        }

        /// <summary>
        /// The event, must be "status", "join" or "leave".
        /// </summary>
        [JsonProperty("event")]
        public readonly string Event;

        /// <summary>
        /// The target channel for the key.
        /// </summary>
        [JsonProperty("channel")]
        public string Channel;

        /// <summary>
        /// The amount of connections present.
        /// </summary>
        [JsonProperty("occupancy")]
        public readonly int Occupancy;

        /// <summary>
        /// The UNIX timestamp.
        /// </summary>
        [JsonProperty("time")]
        public readonly long Time;

        /// <summary>
        /// The target subscriber for the key.
        /// </summary>
        [JsonProperty("who")]
        public PresenceInfo Who;
    }

    /// <summary>
    /// Represents a client info used for the presence.
    /// </summary>
    internal class PresenceInfo
    {
        /// <summary>
        /// The internal client id.
        /// </summary>
        [JsonProperty("id")]
        public string Id;

        /// <summary>
        /// The username exposed by the client.
        /// </summary>
        [JsonProperty("username")]
        public string Username;
    }

    /// <summary>
    /// Represents a presence action.
    /// </summary>
    internal enum PresenceEvent
    {
        /// <summary>
        /// The status of the presence.
        /// </summary>
        Status = 0,

        /// <summary>
        /// The client joins a channel.
        /// </summary>
        Subscribe = 1,

        /// <summary>
        /// The client leaves a channel.
        /// </summary>
        Unsubscribe = 2
    }
}