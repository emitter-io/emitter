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
    /// Represents a mesh subscribe/unsubscribe packet.
    /// </summary>
    public sealed class MeshEmitterEvent : MeshEvent
    {
        #region Static Members

        /// <summary>
        /// The pool of transport packets.
        /// </summary>
        private readonly static PacketPool<MeshEmitterEvent> Pool =
            new PacketPool<MeshEmitterEvent>("Mesh Emitter Packets", (p) => new MeshEmitterEvent());

        /// <summary>
        /// Acquires a <see cref="MeshEmitterEvent"/> instance.
        /// </summary>
        /// <param name="nodes">The list of nodes to populate the gossip packet with.</param>
        public static new MeshEmitterEvent Acquire()
        {
            // Acquires a new emitter event packet
            return Pool.Acquire();
        }

        /// <summary>
        /// Acquires a <see cref="MeshEmitterEvent"/> instance.
        /// </summary>
        /// <param name="nodes">The list of nodes to populate the gossip packet with.</param>
        public static MeshEmitterEvent Acquire(MeshEventType type, int contract, string channel, string value = "")
        {
            // Acquires a new emitter event packet
            var packet = Pool.Acquire();
            packet.Type = type;
            packet.Value = value;
            packet.Contract = contract;
            packet.Channel = channel;
            return packet;
        }

        #endregion Static Members

        /// <summary>
        /// Gets or sets the contract id.
        /// </summary>
        public int Contract;

        /// <summary>
        /// Gets or setst the channel for the command.
        /// </summary>
        public string Channel;

        /// <summary>
        /// Recycles the packet object.
        /// </summary>
        public override void Recycle()
        {
            base.Recycle();
            this.Type = MeshEventType.Error;
            this.Contract = 0;
            this.Channel = null;
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
                    // Read the list of members
                    this.Contract = reader.ReadInt32();
                    this.Channel = reader.ReadString();
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

            // Write the version
            writer.Write((ushort)1);

            // Write the values
            writer.Write(this.Contract);
            writer.Write(this.Channel);
        }

        /// <summary>
        /// Converts to a string representation.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            switch (this.Type)
            {
                default:
                    return "{" + Enum.GetName(typeof(MeshEventType), this.Type) + ": " + this.Channel + "}";
            }
        }
    }
}