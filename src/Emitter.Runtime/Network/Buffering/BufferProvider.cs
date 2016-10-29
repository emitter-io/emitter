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

using System.IO;
using System.Runtime.CompilerServices;
using Mem = Emitter.Memory;

namespace Emitter.Network
{
    /// <summary>
    /// Defines a class that represents a circular byte queue.
    /// </summary>
    internal unsafe sealed class BufferProvider : BuddyPool
    {
        #region Properties

        /// <summary>
        /// Gets whether the buffer provider was disposed or not.
        /// </summary>
        public bool IsDisposed
        {
            get { return this.Disposed; }
        }

        #endregion Properties

        #region Public Members

        /// <summary>
        /// Writes a memory stream to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="stream">The memory stream to write to this buffer.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment Write(MemoryStream stream)
        {
            var buffer = stream.GetBuffer();
            return this.Write(buffer, 0, (int)stream.Length);
        }

        /// <summary>
        /// Writes a memory stream to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="stream">The memory stream to write to this buffer.</param>
        /// <param name="offset">The starting offset in the byte array.</param>
        /// <param name="length">The amount of bytes to write.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment Write(MemoryStream stream, int offset, int length)
        {
            var buffer = stream.GetBuffer();
            return this.Write(buffer, offset, length);
        }

        /// <summary>
        /// Writes the array of bytes to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="buffer">The array of bytes to write to this buffer.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment Write(byte[] buffer)
        {
            return this.Write(buffer, 0, buffer.Length);
        }

        /// <summary>
        /// Writes the array of bytes to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="buffer">The array of bytes to write to this buffer.</param>
        /// <param name="offset">The starting offset in the byte array.</param>
        /// <param name="length">The amount of bytes to write.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment Write(byte[] buffer, int offset, int length)
        {
            if (this.Disposed || length == 0)
                return null;

            lock (this.Lock)
            {
                // Acquire a block of memory
                var blockOffset = this.Acquire(length);

                // Acquire a segment
                var segment = BufferSegment.Acquire(this, this.Memory, blockOffset, length, blockOffset);

                // Write to the block
                Mem.Copy(buffer, offset, this.Memory, blockOffset, length);

                // Return the segment
                return segment;
            }
        }

        /// <summary>
        /// Reserves a specific segment which can be used for various operations.
        /// </summary>
        /// <param name="length">The amount of bytes to reserve.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment Reserve(int length)
        {
            if (this.Disposed)
                return null;

            lock (this.Lock)
            {
                // Acquire a block of memory
                var blockOffset = this.Acquire(length);

                // Acquire a segment
                var segment = BufferSegment.Acquire(this, this.Memory, blockOffset, length, blockOffset);

                // Return the segment
                return segment;
            }
        }

        /// <summary>
        /// Releases the memory associated with the particular <see cref="BufferSegment"/>.
        /// </summary>
        /// <param name="segment">The segment which should be released.</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        internal void Release(BufferSegment segment)
        {
            lock (this.Lock)
            {
                // Avoid calling twice
                if (segment.Handle < 0 || this.Disposed)
                    return;

                // Release the memory block back
                this.Release(segment.Handle);

                // Set the handle to negative
                segment.Handle = -1;
            }
        }

        #endregion Public Members
    }
}