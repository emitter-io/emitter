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
    /// Represents a pool of packet readers.
    /// </summary>
    public sealed class PacketReaderPool : ConcurrentPool<PacketReader>
    {
        /// <summary>
        /// Gets the default PacketReaderPool.
        /// </summary>
        public static readonly PacketReaderPool Default = new PacketReaderPool();

        /// <summary>
        /// Default constructor for the PacketReaderPool
        /// </summary>
        public PacketReaderPool() : base("PacketReaders", _ => new PacketReader()) { }

        /// <summary>
        /// Acquires a <see cref="PacketReader"/> instance.
        /// </summary>
        /// <param name="data">The array of the bytes to read.</param>
        /// <param name="length">The length to read.</param>
        /// <param name="offset">The starting offset.</param>
        /// <returns>Returns the acquired instance.</returns>
        public PacketReader Acquire(byte[] data, int offset, int length)
        {
            // Aquire and set the properties
            PacketReader instance = base.Acquire();
            instance.Data = data;
            instance.IndexMax = offset + length;
            instance.Index = offset;
            return instance;
        }

        /// <summary>
        /// Acquires a <see cref="PacketReader"/> instance.
        /// </summary>
        /// <param name="segment">The segment to read.</param>
        /// <returns>Returns the acquired instance.</returns>
        public PacketReader Acquire(BufferSegment segment)
        {
            // Aquire and set the properties
            PacketReader instance = base.Acquire();
            instance.Data = segment.Array;
            instance.IndexMax = segment.Offset + segment.Length;
            instance.Index = segment.Offset;
            return instance;
        }

        /// <summary>
        /// Acquires a <see cref="PacketReader"/> instance.
        /// </summary>
        /// <returns>Returns the acquired instance.</returns>
        public override PacketReader Acquire()
        {
            throw new InvalidOperationException("Parameterless Acquire() is not available for this pool. Use the override instead.");
        }
    }
}