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
    /// Represents a packet containing raw bytes to send.
    /// </summary>
    public class BytePacket : Packet
    {
        private static readonly PacketPool<BytePacket> InternalPool = new PacketPool<BytePacket>("Byte Packets",
            (_) => new BytePacket());

        private byte[] fBuffer;
        private int fOffset = 0;
        private int fCount = 0;

        /// <summary>
        /// Constructs a new <see cref="BytePacket"/> instance.
        /// </summary>
        public BytePacket() : base()
        {
            fBuffer = ArrayUtils<byte>.Empty;
        }

        /// <summary>
        /// Constructs a new <see cref="BytePacket"/> instance.
        /// </summary>
        /// <param name="defaultBuffer">The default buffer to set on this packet.</param>
        public BytePacket(byte[] defaultBuffer) : base()
        {
            fBuffer = defaultBuffer;
        }

        /// <summary>
        /// Gets the byte buffer of this packet.
        /// </summary>
        public byte[] Buffer
        {
            get { return fBuffer; }
            protected set { fBuffer = value; }
        }

        /// <summary>
        /// Gets the write offset for byte byte buffer of this packet.
        /// </summary>
        public int Offset
        {
            get { return fOffset; }
            protected set { fOffset = value; }
        }

        /// <summary>
        /// Gets the write count for byte buffer of this packet.
        /// </summary>
        public int Count
        {
            get { return fCount; }
            protected set { fCount = value; }
        }

        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        public override void Recycle()
        {
            base.Recycle();
            fBuffer = null;
            fOffset = 0;
            fCount = 0;
        }

        /// <summary>
        /// Acquires a byte packet and fills it with the provided buffer.
        /// </summary>
        /// <param name="buffer">The default buffer to set to the byte packet.</param>
        public static BytePacket Acquire(byte[] buffer)
        {
            var packet = InternalPool.Acquire();
            packet.Buffer = buffer;
            packet.Offset = 0;
            packet.Count = buffer.Length;
            return packet;
        }

        /// <summary>
        /// Acquires a byte packet and fills it with the provided buffer.
        /// </summary>
        /// <param name="buffer">The default buffer to set to the byte packet.</param>
        /// <param name="count">The length of the buffer.</param>
        /// <param name="offset">The write offset of the buffer.</param>
        public static BytePacket Acquire(byte[] buffer, int offset, int count)
        {
            var packet = InternalPool.Acquire();
            packet.Buffer = buffer;
            packet.Offset = offset;
            packet.Count = count;
            return packet;
        }
    }
}