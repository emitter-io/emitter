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
using System.IO;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using Emitter.Replication;

namespace Emitter.Network
{
    /// <summary>
    /// Provides functionality for writing primitive binary data.
    /// </summary>
    public sealed class PacketWriter : RecyclableObject
    {
        #region Ctor & Properties

        /// <summary>
        /// Internal stream which holds the entire packet.
        /// </summary>
        private ByteStream fStream;

        /// <summary>
        /// Internal format buffer.
        /// </summary>
        private byte[] fBuffer = new byte[32];

        /// <summary>
        /// Instantiates a new PacketWriter instance with a given capacity.
        /// </summary>
        public PacketWriter()
        {
            // Note: No need to pool ByteStreams, since the PacketWriter is pooled already
        }

        /// <summary>
        /// Gets the total stream length.
        /// </summary>
        public long Length
        {
            get { return fStream.Length; }
        }

        /// <summary>
        /// Gets or sets the current stream position.
        /// </summary>
        public long Position
        {
            get { return fStream.Position; }
            set { fStream.Position = value; }
        }

        /// <summary>
        /// Gets or sets the target stream to use this PacketWriter on.
        /// </summary>
        public ByteStream Stream
        {
            get { return fStream; }
            set { fStream = value; }
        }

        /// <summary>
        /// Offsets the current position from an origin.
        /// </summary>
        public long Seek(long offset, SeekOrigin origin)
        {
            return fStream.Seek(offset, origin);
        }

        #endregion Ctor & Properties

        #region Write - Primitives

        /// <summary>
        /// Writes a 1-byte boolean value to the underlying stream. False is represented by 0, true by 1.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(bool value)
        {
            fStream.WriteByte((byte)(value ? 1 : 0));
        }

        /// <summary>
        /// Writes a 1-byte unsigned integer value to the underlying stream.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(byte value)
        {
            fStream.WriteByte(value);
        }

        /// <summary>
        /// Writes a 1-byte signed integer value to the underlying stream.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(sbyte value)
        {
            fStream.WriteByte((byte)value);
        }

        /// <summary>
        /// Writes a 2-byte signed integer value to the underlying stream.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(short value)
        {
            fBuffer[0] = (byte)(value >> 8);
            fBuffer[1] = (byte)value;

            fStream.Write(fBuffer, 0, 2);
        }

        /// <summary>
        /// Writes a 2-byte unsigned integer value to the underlying stream.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(ushort value)
        {
            fBuffer[0] = (byte)(value >> 8);
            fBuffer[1] = (byte)value;

            fStream.Write(fBuffer, 0, 2);
        }

        /// <summary>
        /// Writes a 4-byte signed integer value to the underlying stream.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(int value)
        {
            fBuffer[0] = (byte)(value >> 24);
            fBuffer[1] = (byte)(value >> 16);
            fBuffer[2] = (byte)(value >> 8);
            fBuffer[3] = (byte)value;

            fStream.Write(fBuffer, 0, 4);
        }

        /// <summary>
        /// Writes a 4-byte unsigned integer value to the underlying stream.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(uint value)
        {
            fBuffer[0] = (byte)(value >> 24);
            fBuffer[1] = (byte)(value >> 16);
            fBuffer[2] = (byte)(value >> 8);
            fBuffer[3] = (byte)value;

            fStream.Write(fBuffer, 0, 4);
        }

        /// <summary>
        /// Writes a 8-byte signed integer value to the underlying stream.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(long value)
        {
            fBuffer[0] = (byte)(value >> 56);
            fBuffer[1] = (byte)(value >> 48);
            fBuffer[2] = (byte)(value >> 40);
            fBuffer[3] = (byte)(value >> 32);
            fBuffer[4] = (byte)(value >> 24);
            fBuffer[5] = (byte)(value >> 16);
            fBuffer[6] = (byte)(value >> 8);
            fBuffer[7] = (byte)value;

            fStream.Write(fBuffer, 0, 8);
        }

        /// <summary>
        /// Writes a 8-byte signed integer value to the underlying stream.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(ulong value)
        {
            fBuffer[0] = (byte)(value >> 56);
            fBuffer[1] = (byte)(value >> 48);
            fBuffer[2] = (byte)(value >> 40);
            fBuffer[3] = (byte)(value >> 32);
            fBuffer[4] = (byte)(value >> 24);
            fBuffer[5] = (byte)(value >> 16);
            fBuffer[6] = (byte)(value >> 8);
            fBuffer[7] = (byte)value;

            fStream.Write(fBuffer, 0, 8);
        }

        /// <summary>
        /// Writes a DateTime to a sequence of bytes to the underlying stream
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(DateTime value)
        {
            Write(value.Ticks);
        }

        /// <summary>
        /// Writes an IEEE 754 double-precision (64-bit) floating-point number to the buffer
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public unsafe void Write(double value)
        {
            // Write unsafely directly
            fixed (byte* pBuffer = fBuffer)
                *((double*)(pBuffer)) = value;

            // Which way should we write?
            if (BitConverter.IsLittleEndian)
            {
                // We have to reverse
                fStream.WriteByte(fBuffer[7]);
                fStream.WriteByte(fBuffer[6]);
                fStream.WriteByte(fBuffer[5]);
                fStream.WriteByte(fBuffer[4]);
                fStream.WriteByte(fBuffer[3]);
                fStream.WriteByte(fBuffer[2]);
                fStream.WriteByte(fBuffer[1]);
                fStream.WriteByte(fBuffer[0]);
            }
            else
            {
                // We have it in order
                fStream.Write(fBuffer, 0, 8);
            }
        }

        /// <summary>
        /// Writes an IEEE 754 single-precision (32-bit) floating-point number to the buffer
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public unsafe void Write(float value)
        {
            // Write unsafely directly
            fixed (byte* pBuffer = fBuffer)
                *((float*)(pBuffer)) = value;

            // Which way should we write?
            if (BitConverter.IsLittleEndian)
            {
                // We have to reverse
                fStream.WriteByte(fBuffer[3]);
                fStream.WriteByte(fBuffer[2]);
                fStream.WriteByte(fBuffer[1]);
                fStream.WriteByte(fBuffer[0]);
            }
            else
            {
                // We have it in order
                fStream.Write(fBuffer, 0, 4);
            }
        }

        /// <summary>
        /// Writes a sequence of bytes to the underlying stream
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(byte[] buffer, int offset, int size)
        {
            fStream.Write(buffer, offset, size);
        }

        /// <summary>
        /// Writes a fixed-length big-endian unicode string value to the underlying stream. To fit (size), the string content is either truncated or padded with null characters.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public unsafe void Write(string value)
        {
            if (String.IsNullOrEmpty(value))
            {
                Write((int)0);
            }
            else
            {
                var stringBuffer = Encoding.UTF8.GetBytes(value);
                var size = stringBuffer.Length;

                fBuffer[0] = (byte)(size >> 24);
                fBuffer[1] = (byte)(size >> 16);
                fBuffer[2] = (byte)(size >> 8);
                fBuffer[3] = (byte)size;

                if (size > 0)
                {
                    fStream.Write(fBuffer, 0, 4);
                    fStream.Write(stringBuffer, 0, size);
                }

                return;

                /*var size = value.Length;
                fBuffer[0] = (byte)(size >> 24);
                fBuffer[1] = (byte)(size >> 16);
                fBuffer[2] = (byte)(size >> 8);
                fBuffer[3] = (byte)size;

                if (size > 0)
                {
                    fStream.Write(fBuffer, 0, 4);
                    fixed (char* pFirstChar = value)
                    {
                        char* pChar = pFirstChar;
                        for (int i = 0; i < value.Length; ++i)
                        {
                            fStream.WriteByte((byte)*pChar);
                            ++pChar;
                        }
                    }
                }*/
            }
        }

        /// <summary>
        /// Writes a IPartialEntity in the packet
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(ISerializable item)
        {
            item.Write(this);
        }

        /// <summary>
        /// Writes a IPartialEntity in the packet
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IReplicated item, ReplicatedVersion since)
        {
            item.Write(this, since);
        }

        /// <summary>
        /// Writes a byte array
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(byte[] array)
        {
            Write((int)array.Length);
            fStream.Write(array, 0, array.Length);
        }

        #endregion Write - Primitives

        #region Write ILists

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<Boolean> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<Double> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<Single> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<byte> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<sbyte> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<short> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<ushort> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<int> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<uint> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<long> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<ulong> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<DateTime> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<string> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                Write(list[i]);
        }

        /// <summary>
        /// Writes a list of packets
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write(IList<ISerializable> list)
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                list[i].Write(this);
        }

        /// <summary>
        /// Writes a list of packets
        /// </summary>
        /// <param name="list">The list instance to write</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public void Write<T>(IList<T> list)
            where T : ISerializable
        {
            if (list == null)
            {
                Write((int)0);
                return;
            }
            Write((int)list.Count);
            for (int i = 0; i < list.Count; ++i)
                list[i].Write(this);
        }

        #endregion Write ILists

        #region IRecyclable Members

        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        public override void Recycle()
        {
            fStream = null;
        }

        #endregion IRecyclable Members
    }
}