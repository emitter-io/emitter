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
using System.Linq;
using Emitter.Network;
using Emitter.Replication;

namespace Emitter
{
    /// <summary>
    /// Represents a presence information state.
    /// </summary>
    public sealed class SubscriptionPresenceInfo : ReplicatedObject
    {
        /// <summary>
        /// Creates a new presence info for deseiralization.
        /// </summary>
        public SubscriptionPresenceInfo()
        {
        }

        /// <summary>
        /// Creates a new presence info from the client.
        /// </summary>
        /// <param name="client">The client just subscribed.</param>
        public SubscriptionPresenceInfo(IMqttSender client)
        {
            this.ClientId = client.Context.ClientId;
            this.Username = client.Context.Username;
        }

        /// <summary>
        /// Gets or sets the client identifier that identifies the connection.
        /// </summary>
        public string ClientId;

        /// <summary>
        /// Gets or sets some user data attached.
        /// </summary>
        public string Username;

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public override void Read(PacketReader reader)
        {
            var version = reader.ReadUInt16();
            this.ClientId = reader.ReadString();
            this.Username = reader.ReadString();
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        /// <param name="since">The minimum version of updates to pack.</param>
        public override void Write(PacketWriter writer, ReplicatedVersion since)
        {
            writer.Write((ushort)1);
            writer.Write(this.ClientId);
            writer.Write(this.Username);
        }

        /// <summary>
        /// Converts repliccated state to a presence info.
        /// </summary>
        /// <returns></returns>
        internal PresenceInfo AsInfo()
        {
            var info = new PresenceInfo();
            info.Id = this.Key;
            info.Username = this.Username;
            return info;
        }
    }
}