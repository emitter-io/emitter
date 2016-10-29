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
using System.Runtime.CompilerServices;
using Emitter.Collections;
using Emitter.Network;

namespace Emitter.Compression
{
    /// <summary>
    /// The snappy tag type.
    /// </summary>
    internal class SnappyTag
    {
        public const byte Literal = 0x00;
        public const byte Copy1 = 0x01;
        public const byte Copy2 = 0x02;
        public const byte Copy4 = 0x03;
    }

    /// <summary>
    /// The snappy compressor.
    /// </summary>
    public sealed class Snappy : RecyclableObject
    {
        #region Pool Members

        /// <summary>
        /// Pool of Snappy Compressors
        /// </summary>
        internal static readonly ConcurrentPool<Snappy> Compressors
            = new ConcurrentPool<Snappy>("Snappy Compressors", _ => new Snappy());

        /// <summary>
        /// Acquires a compressor.
        /// </summary>
        /// <returns>Returns a Snappy instance.</returns>
        public static Snappy Acquire()
        {
            return Compressors.Acquire();
        }

        #endregion Pool Members

        #region Instance Members

        /// <summary>
        /// The internal lookup table.
        /// </summary>
        private readonly int[] Table = new int[MaxTableSize];

        /// <summary>
        /// The index offset.
        /// </summary>
        private int Index;

        /// <summary>
        /// Recycles the compressor.
        /// </summary>
        public override void Recycle()
        {
            this.Index = 0;
        }

        #endregion Instance Members

        public static byte[] Decode(byte[] src)
        {
            return Decode(src, 0, src.Length);
        }

        public static byte[] Decode(byte[] src, int srcOff, int srcLen)
        {
            var len = VarInt.UvarInt(src, srcOff);
            byte[] ret = new byte[len.Value];

            Decode(src, srcOff, srcLen, ret, 0, ret.Length);
            return ret;
        }

        public static int Decode(byte[] source, int sourceOffset, int sourceLength, byte[] destination, int destinationOffset, int destinationLength)
        {
            // Get the uncompressed size
            var len = VarInt.UvarInt(source, sourceOffset);

            // Destination array doesn't have enough room for the decoded data.
            if (destinationLength < (int)len.Value)
                throw new IndexOutOfRangeException("snappy: destination array too short");

            // Adjust the lengths
            sourceLength += sourceOffset;
            destinationLength += destinationOffset;

            // Current offsets
            int s = sourceOffset + len.VarIntLength;
            int d = destinationOffset;
            int offset = 0;
            int length = 0;

            while (s < sourceLength)
            {
                byte tag = (byte)(source[s] & 0x03);
                if (tag == SnappyTag.Literal)
                {
                    uint x = (uint)(source[s] >> 2);
                    if (x < 60)
                    {
                        s += 1;
                    }
                    else if (x == 60)
                    {
                        s += 2;
                        if (s > sourceLength)
                        {
                            throw new IndexOutOfRangeException("snappy: corrupt input");
                        }
                        x = (uint)source[s - 1];
                    }
                    else if (x == 61)
                    {
                        s += 3;
                        if (s > sourceLength)
                        {
                            throw new IndexOutOfRangeException("snappy: corrupt input");
                        }
                        x = (uint)source[s - 2] | (uint)source[s - 1] << 8;
                    }
                    else if (x == 62)
                    {
                        s += 4;
                        if (s > sourceLength)
                        {
                            throw new IndexOutOfRangeException("snappy: corrupt input");
                        }
                        x = (uint)source[s - 3] | (uint)source[s - 2] << 8 | (uint)source[s - 1] << 16;
                    }
                    else if (x == 63)
                    {
                        s += 5;
                        if (s > sourceLength)
                        {
                            throw new IndexOutOfRangeException("snappy: corrupt input");
                        }
                        x = (uint)source[s - 4] | (uint)source[s - 3] << 8 | (uint)source[s - 2] << 16 | (uint)source[s - 1] << 24;
                    }
                    length = (int)(x + 1);
                    if (length <= 0)
                    {
                        throw new IndexOutOfRangeException("snappy: unsupported literal length");
                    }
                    if (length > destinationLength - d || length > sourceOffset + sourceLength - s)
                    {
                        throw new IndexOutOfRangeException("snappy: corrupt input");
                    }
                    Array.Copy(source, s, destination, d, length);
                    d += length;
                    s += length;
                    continue;
                }
                else if (tag == SnappyTag.Copy1)
                {
                    s += 2;
                    if (s > sourceOffset + sourceLength)
                    {
                        throw new IndexOutOfRangeException("snappy: corrupt input");
                    }
                    length = 4 + (((int)source[s - 2] >> 2) & 0x7);
                    offset = ((int)source[s - 2] & 0xe0) << 3 | (int)source[s - 1];
                }
                else if (tag == SnappyTag.Copy2)
                {
                    s += 3;
                    if (s > sourceOffset + sourceLength)
                    {
                        throw new IndexOutOfRangeException("snappy: corrupt input");
                    }
                    length = 1 + ((int)source[s - 3] >> 2);
                    offset = (int)source[s - 2] | (int)source[s - 1] << 8;
                }
                else if (tag == SnappyTag.Copy4)
                {
                    throw new NotSupportedException("snappy: unsupported COPY_4 tag");
                }

                int end = d + length;
                if (offset > d || end > destinationLength)
                {
                    throw new IndexOutOfRangeException("snappy: corrupt input");
                }

                for (; d < end; d++)
                {
                    destination[d] = destination[d - offset];
                }
            }

            return d;
        }

        // Limit how far copy back-references can go, the same as the C++ code.
        private const int MaxOffset = 1 << 15;

        private const int MaxTableSize = 1 << 14;

        public static int MaxEncodedLen(int srcLen)
        {
            return 32 + srcLen + srcLen / 6;
        }

        /// <summary>
        /// Encodes the buffer into the stream provided.
        /// </summary>
        /// <param name="source">The source buffer.</param>
        /// <param name="sourceOffset">The offset to start encoding.</param>
        /// <param name="sourceLength">The number of bytes to encode.</param>
        /// <param name="destination">The destination to write into.</param>
        /// <param name="destinationOffset">The offset to start writing into.</param>
        /// <returns></returns>
        public int Encode(byte[] source, int sourceOffset, int sourceLength, BufferSegment destination, int destinationOffset)
        {
            if (destination.Size - destinationOffset < MaxEncodedLen(sourceLength))
                throw new IndexOutOfRangeException("Snappy destination array too short.");

            // The block starts with the varint-encoded length of the decompressed bytes.
            var start = this.Index = destination.Offset + destinationOffset;
            this.Index += VarInt.WriteTo(destination, this.Index, (ulong)sourceLength);

            // Return early if src is short.
            if (sourceLength <= 4 && sourceLength != 0)
            {
                EmitLiteral(destination, source, sourceOffset, sourceLength);
                return this.Index - start;
            }

            // Initialize the hash table. Its size ranges from 1<<8 to 1<<14 inclusive.
            int shift = 32 - 8;
            uint tableSize = 1 << 8;
            while (tableSize < MaxTableSize && tableSize < sourceLength)
            {
                shift--;
                tableSize *= 2;
            }

            // Clear the table up until we're going to use it, to save time
            Array.Clear(Table, 0, (int)tableSize);

            // Iterate over the source bytes.
            int s = sourceOffset;
            int t = 0;
            int lit = sourceOffset;

            while (s + 3 < sourceLength)
            {
                // Update the hash table.
                byte b0 = source[s];
                byte b1 = source[s + 1];
                byte b2 = source[s + 2];
                byte b3 = source[s + 3];
                uint h = (uint)b0 | ((uint)b1) << 8 | ((uint)b2) << 16 | ((uint)b3) << 24;
                var p = (h * 0x1e35a7bd) >> shift;

                // We need to to store values in [-1, inf) in table. To save
                // some initialization time, (re)use the table's zero value
                // and shift the values against this zero: add 1 on writes,
                // subtract 1 on reads.
                t = this.Table[p] - 1;
                this.Table[p] = s + 1;

                // If t is invalid or src[s:s+4] differs from src[t:t+4],
                // accumulate a literal byte.
                if (t < 0 || s - t >= MaxOffset || b0 != source[t] || b1 != source[t + 1] || b2 != source[t + 2] || b3 != source[t + 3])
                {
                    s++;
                    continue;
                }

                // Otherwise, we have a match. First, emit any pending literal bytes.
                if (lit != s)
                {
                    EmitLiteral(destination, source, lit, s - lit);
                }

                // Extend the match to be as long as possible.
                int s0 = s;
                s = s + 4;
                t = t + 4;
                while (s < sourceLength && source[s] == source[t])
                {
                    s++;
                    t++;
                }

                // Emit the copied bytes.
                EmitCopy(destination, s - t, s - s0 + sourceOffset);
                lit = s;
            }

            // Emit any final pending literal bytes and return.
            if (lit != sourceLength)
            {
                EmitLiteral(destination, source, lit, sourceLength - lit);
            }

            return this.Index - start;
        }

        private void EmitLiteral(BufferSegment dst, byte[] lit, int litOff, int litLen)
        {
            uint n = (uint)(litLen - 1);
            if (n < 60)
            {
                dst.Array[this.Index++] = (byte)(n << 2 | SnappyTag.Literal);
            }
            else if (n < (1 << 8))
            {
                dst.Array[this.Index++] = (60 << 2 | SnappyTag.Literal);
                dst.Array[this.Index++] = (byte)n;
            }
            else if (n < (1 << 16))
            {
                dst.Array[this.Index++] = (61 << 2 | SnappyTag.Literal);
                dst.Array[this.Index++] = (byte)n;
                dst.Array[this.Index++] = (byte)(n >> 8);
            }
            else if (n < (1 << 24))
            {
                dst.Array[this.Index++] = (62 << 2 | SnappyTag.Literal);
                dst.Array[this.Index++] = (byte)n;
                dst.Array[this.Index++] = (byte)(n >> 8);
                dst.Array[this.Index++] = (byte)(n >> 16);
            }
            else if ((Int64)n < 1 << 32)
            {
                dst.Array[this.Index++] = (63 << 2 | SnappyTag.Literal);
                dst.Array[this.Index++] = (byte)n;
                dst.Array[this.Index++] = (byte)(n >> 8);
                dst.Array[this.Index++] = (byte)(n >> 16);
                dst.Array[this.Index++] = (byte)(n >> 24);
            }
            else
            {
                throw new IndexOutOfRangeException("snappy: source buffer is too long");
            }

            Memory.Copy(lit, litOff, dst.Array, this.Index, litLen);
            this.Index += litLen;
        }

        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private void EmitCopy(BufferSegment dst, int offset, int length)
        {
            while (length > 0)
            {
                int x = length - 4;
                if ((0 <= x) && (x < 1 << 3) && (offset < 1 << 11))
                {
                    dst.Array[this.Index++] = ((byte)((((offset >> 8) & 0x07) << 5) | (byte)(x << 2) | SnappyTag.Copy1));
                    dst.Array[this.Index++] = (byte)offset;
                    break;
                }

                x = length;
                if (x > 1 << 6)
                    x = 1 << 6;

                dst.Array[this.Index++] = (byte)((byte)(x - 1) << 2 | SnappyTag.Copy2);
                dst.Array[this.Index++] = (byte)offset;
                dst.Array[this.Index++] = (byte)(offset >> 8);
                length -= x;
            }
        }
    }

    internal static class VarInt
    {
        public const int MaxVarIntLen16 = 3;
        public const int MaxVarIntLen32 = 5;
        public const int MaxVarIntLen64 = 10;

        public class UvarIntRet
        {
            public ulong Value;
            public int VarIntLength;

            public UvarIntRet(ulong val, int valLen)
            {
                this.Value = val;
                this.VarIntLength = valLen;
            }
        }

        public static UvarIntRet UvarInt(byte[] src, int srcOff)
        {
            ulong x = 0;
            int s = 0;
            for (int i = 0; i < src.Length - srcOff; i++)
            {
                byte b = src[srcOff + i];
                if (b < 0x80)
                {
                    if (i > 9 || i == 9 && b > 1)
                    {
                        return new UvarIntRet(0, -(i + 1));
                    }
                    return new UvarIntRet(x | ((ulong)b) << s, i + 1);
                }
                x |= (ulong)(b & 0x7f) << s;
                s += 7;
            }

            return new UvarIntRet(0, 0);
        }

        public static int WriteTo(BufferSegment buffer, int offset, ulong x)
        {
            var i = 0;
            while (x >= 0x80)
            {
                buffer.Array[offset + i] = (byte)(x | 0x80);
                x >>= 7;
                i++;
            }

            buffer.Array[offset + i] = (byte)x;
            return i + 1;
        }
    }
}