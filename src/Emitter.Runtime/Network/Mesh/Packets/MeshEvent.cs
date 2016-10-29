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
    /// Represents mesh command packet.
    /// </summary>
    public class MeshEvent : MeshPacket
    {
        #region Static Members

        /// <summary>
        /// The pool of transport packets.
        /// </summary>
        private readonly static PacketPool<MeshEvent> Pool =
            new PacketPool<MeshEvent>("Mesh Event Packets", (p) => new MeshEvent());

        /// <summary>
        /// Acquires a <see cref="MeshEvent"/> instance.
        /// </summary>
        /// <param name="type">The event type.</param>
        /// <param name="value">The value of the event.</param>
        /// <returns> A <see cref="MeshEvent"/> that can be sent to the remote client.</returns>
        public static MeshEvent Acquire(MeshEventType type, string value = "")
        {
            // Set the type
            var packet = Pool.Acquire();
            packet.Type = type;
            packet.Value = value;
            return packet;
        }

        /// <summary>
        /// Acquires a <see cref="MeshEvent"/> instance.
        /// </summary>
        /// <returns> A <see cref="MeshEvent"/> that can be sent to the remote client.</returns>
        public static MeshEvent Acquire()
        {
            return Pool.Acquire();
        }

        #endregion Static Members

        /// <summary>
        /// Gets or sets the command type.
        /// </summary>
        public MeshEventType Type;

        /// <summary>
        /// Gets or sets the command value.
        /// </summary>
        public string Value;

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public override void Read(PacketReader reader)
        {
            // Read the type
            this.Type = (MeshEventType)reader.ReadByte();

            // Read the version and the payload
            var version = reader.ReadUInt16();
            switch (version)
            {
                case 1:
                    this.Value = reader.ReadString();
                    break;
            }
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        public override void Write(PacketWriter writer)
        {
            // Write the type first
            writer.Write((byte)Type);

            // Write the version
            writer.Write((ushort)1);
            writer.Write(this.Value);
        }

        /// <summary>
        /// Converts the command to a string representation.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            return "{" + Enum.GetName(typeof(MeshEventType), this.Type) + (string.IsNullOrEmpty(this.Value) ? "" : this.Value) + "}";
        }

        /// <summary>
        /// Recycles the packet object.
        /// </summary>
        public override void Recycle()
        {
            this.Type = default(MeshEventType);
            this.Value = null;
            base.Recycle();
        }
    }
}