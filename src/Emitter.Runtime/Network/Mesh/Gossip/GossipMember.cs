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
using System.Net;
using Emitter.Replication;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a serialized gossip member.
    /// </summary>
    public sealed class GossipMember : ReplicatedObject, IComparable<GossipMember>
    {
        /// <summary>
        /// Gets the endpoint of the member.
        /// </summary>
        public IPEndPoint EndPoint;

        /// <summary>
        /// Deserialization constructor.
        /// </summary>
        public GossipMember()
        {
        }

        /// <summary>
        /// Creates an instance of an object.
        /// </summary>
        /// <param name="endpoint"></param>
        public GossipMember(IPEndPoint endpoint)
        {
            this.EndPoint = endpoint;
        }

        /// <summary>
        /// Compares the gossip member to another one.
        /// </summary>
        /// <param name="other"></param>
        /// <returns></returns>
        public int CompareTo(GossipMember other)
        {
            return this.EndPoint.ToString().CompareTo(other.EndPoint.ToString());
        }

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public override void Read(PacketReader reader)
        {
            MeshMember.TryParseEndpoint(reader.ReadString(), out this.EndPoint);
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        public override void Write(PacketWriter writer, ReplicatedVersion since)
        {
            writer.Write(this.EndPoint.ToString());
        }
    }

    /// <summary>
    /// Represents a serialized gossip member.
    /// </summary>
    public sealed class GossipMembership : ReplicatedDictionary<GossipMember>
    {
    }
}