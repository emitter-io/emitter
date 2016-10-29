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
using Emitter.Replication;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a mesh handshake packet.
    /// </summary>
    internal sealed class MeshGossip : MeshEvent
    {
        #region Static Members

        /// <summary>
        /// The pool of transport packets.
        /// </summary>
        private readonly static PacketPool<MeshGossip> Pool =
            new PacketPool<MeshGossip>("Mesh Gossip Packets", (p) => new MeshGossip());

        /// <summary>
        /// Acquires a <see cref="MeshGossip"/> instance.
        /// </summary>
        public static new MeshGossip Acquire()
        {
            return Pool.Acquire();
        }

        /// <summary>
        /// Acquires a <see cref="MeshGossip"/> instance.
        /// </summary>
        /// <param name="digest">The digest to send to the neigbour.</param>
        public static MeshGossip Acquire(int digest)
        {
            // Create a digest wrapper
            var state = new ReplicatedVersionDigest();
            state.Value = digest;

            // Acquires a new gossip packet
            var packet = Pool.Acquire();
            packet.Type = MeshEventType.GossipDigest;
            packet.State = state;
            return packet;
        }

        /// <summary>
        /// Acquires a <see cref="MeshGossip"/> instance.
        /// </summary>
        /// <param name="since">The version to request from the neigbour.</param>
        public static MeshGossip Acquire(ReplicatedVersion since)
        {
            // Acquires a new gossip packet
            var packet = Pool.Acquire();
            packet.Type = MeshEventType.GossipSince;
            packet.State = since;
            return packet;
        }

        /// <summary>
        /// Acquires a <see cref="MeshGossip"/> instance.
        /// </summary>
        /// <param name="update">The update to send to the neigbour.</param>
        /// <param name="since">The version for the update to write.</param>
        public static MeshGossip Acquire(ReplicatedHybridDictionary update, ReplicatedVersion since)
        {
            // Acquires a new gossip packet
            var packet = Pool.Acquire();
            packet.Since = since;
            packet.Type = MeshEventType.GossipUpdate;
            packet.State = update;
            return packet;
        }

        #endregion Static Members

        /// <summary>

        /// Gets or sets the replicated value to send.
        /// </summary>
        public ISerializable State;

        /// <summary>
        /// Gets or sets the version of the update to write.
        /// </summary>
        public ReplicatedVersion Since;

        /// <summary>
        /// Constructs an instance of the object.
        /// </summary>
        private MeshGossip()
        {
            this.Type = MeshEventType.GossipDigest;
        }

        /// <summary>
        /// Recycles the packet object.
        /// </summary>
        public override void Recycle()
        {
            base.Recycle();
            this.Type = MeshEventType.GossipDigest;
            this.State = null;
            this.Since = null;
        }

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public override void Read(PacketReader reader)
        {
            // Read the base first
            base.Read(reader);
            switch (this.Type)
            {
                // If we've received a gossip digest, our payload is the version digest
                case MeshEventType.GossipDigest:
                    this.State = reader.ReadSerializable<ReplicatedVersionDigest>();
                    break;

                // If we've received a gossip since, our payload is the version vector
                case MeshEventType.GossipSince:
                    this.State = reader.ReadSerializable<ReplicatedVersion>();
                    break;

                // If we've received a gossip update, our payload is the updated state
                case MeshEventType.GossipUpdate:
                    this.State = reader.ReadReplicated<ReplicatedHybridDictionary>();
                    break;
            }
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        public override void Write(PacketWriter writer)
        {
            // Write the base first
            base.Write(writer);

            // If we have a replicated state, write as replicated
            var replicated = this.State as IReplicated;
            if (this.Since != null && replicated != null)
            {
                // Write as replciated with a specific version
                writer.Write(replicated, this.Since);
            }
            else
            {
                // Write as a normal state
                writer.Write(this.State);
            }
        }

        /// <summary>
        /// Converts the command to a string representation.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            return "{" + Enum.GetName(typeof(MeshEventType), this.Type) + "}";
        }
    }
}