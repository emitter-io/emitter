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
using System.Linq;
using Emitter.Collections;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a pool of <see cref="ByteStream"/> instances.
    /// </summary>
    public sealed class ByteStreamPool : ConcurrentPool<ByteStream>
    {
        /// <summary>
        /// Gets the default ByteStreamPool.
        /// </summary>
        public static readonly ByteStreamPool Default = new ByteStreamPool();

        /// <summary>
        /// Default constructor for the ByteStreamPool. We use large initial capacity in order
        /// to reduce memory fragmentation.
        /// </summary>
        public ByteStreamPool() : base("ByteStreams", _ => new ByteStream(16 * 1024)) { }

        /// <summary>
        /// Acquires an instance of the byte stream from the internal pool.
        /// </summary>
        /// <returns>The acquired bytestream instance.</returns>
        public override ByteStream Acquire()
        {
            return Acquire(0, 0);
        }

        /// <summary>
        /// Acquires an instance of the byte stream from the internal pool with a specified
        /// left padding. To the specified number of bytes, the value 0x00 is written.
        /// </summary>
        /// <param name="padding">Number of bytes to skip in the beginning of the stream.</param>
        /// <returns>The acquired bytestream instance.</returns>
        public ByteStream Acquire(int padding)
        {
            return Acquire(padding, 0);
        }

        /// <summary>
        /// Acquires an instance of the byte stream from the internal pool with a specified
        /// left padding. To the specified number of bytes, the value 0x00 is written.
        /// </summary>
        /// <param name="padding">Number of bytes to skip in the beginning of the stream.</param>
        /// <param name="capacity">The maximum capacity of the buffer.</param>
        /// <returns>The acquired bytestream instance.</returns>
        public ByteStream Acquire(int padding, int capacity)
        {
            // Acquire from the pool
            var stream = base.Acquire();
            if (capacity > 0)
                stream.Capacity = capacity;

            // Append bytes in the beginning
            for (int i = 0; i < padding; ++i)
                stream.WriteByte(0);

            return stream;
        }
    }
}