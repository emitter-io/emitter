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

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a mesh handshake packet.
    /// </summary>
    internal sealed class MeshHandshake : MeshEvent
    {
        #region Static Members

        /// <summary>
        /// The pool of transport packets.
        /// </summary>
        private readonly static PacketPool<MeshHandshake> Pool =
            new PacketPool<MeshHandshake>("Mesh Handshake Packets", (p) => new MeshHandshake());

        /// <summary>
        /// Acquires a <see cref="MeshHandshake"/> instance.
        /// </summary>
        public new static MeshHandshake Acquire()
        {
            // Acquires a new gossip packet
            var packet = Pool.Acquire();
            packet.Type = MeshEventType.Handshake;
            return packet;
        }

        #endregion Static Members

        /// <summary>
        /// Gets or sets the hash of the cluster key for the handshake.
        /// </summary>
        public int Key;

        /// <summary>
        /// Gets or sets the identity information.
        /// </summary>
        public GossipMember Identity;

        /// <summary>
        /// Recycles the packet object.
        /// </summary>
        public override void Recycle()
        {
            base.Recycle();
            this.Type = MeshEventType.Handshake;
            this.Key = 0;
            this.Identity = default(GossipMember);
        }

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public override void Read(PacketReader reader)
        {
            // Read the base first
            base.Read(reader);

            // Read the values
            var version = reader.ReadUInt16();
            switch (version)
            {
                case 1:
                    this.Key = reader.ReadInt32();
                    this.Identity = reader.ReadReplicated<GossipMember>();
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

            // Write the values
            writer.Write((ushort)1);
            writer.Write(this.Key);
            this.Identity.Write(writer, null);
        }
    }
}