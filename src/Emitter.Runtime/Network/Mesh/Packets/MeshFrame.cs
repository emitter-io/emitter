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
    /// Represents mesh inform packet.
    /// </summary>
    public sealed class MeshFrame : MeshPacket
    {
        #region Static Members

        /// <summary>
        /// The pool of transport packets.
        /// </summary>
        private readonly static PacketPool<MeshFrame> Pool =
            new PacketPool<MeshFrame>("Mesh Frame Packets", (p) => new MeshFrame());

        /// <summary>
        /// Acquires a <see cref="MeshFrame"/> instance.
        /// </summary>
        /// <param name="frame">The data frame</param>
        /// <returns> A <see cref="MeshFrame"/> that can be sent to the remote client.</returns>
        public static MeshFrame Acquire(ArraySegment<byte> frame)
        {
            var packet = Pool.Acquire();
            packet.Buffer = frame;
            return packet;
        }

        #endregion Static Members

        /// <summary>
        /// Gets or sets the internal buffer.
        /// </summary>
        private ArraySegment<byte> Buffer;

        /// <summary>
        /// Gets buffer payload associated with this frame.
        /// </summary>
        public ArraySegment<byte> Data
        {
            get { return this.Buffer; }
        }

        /// <summary>
        /// Gets the payload length, in bytes.
        /// </summary>
        public int Length
        {
            get { return this.Buffer.Count; }
        }

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public override void Read(PacketReader reader)
        {
            throw new InvalidOperationException();
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        public override void Write(PacketWriter writer)
        {
            writer.Write(this.Buffer.Array, this.Buffer.Offset, this.Buffer.Count);
        }

        /// <summary>
        /// Converts the command to a string representation.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            return "{Frame: byte[" + this.Buffer.Count + "]}";
        }

        /// <summary>
        /// Recycles the packet object.
        /// </summary>
        public override void Recycle()
        {
            this.Buffer = default(ArraySegment<byte>);
            base.Recycle();
        }
    }
}