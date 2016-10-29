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
using Emitter.Compression;

namespace Emitter.Network
{
    /// <summary>
    /// Represents an encoder for Snappy compression.
    /// </summary>
    public static unsafe class SnappyEncoder
    {
        /// <summary>
        /// Applies a compressin to the buffer and returns a result within a stream.
        /// </summary>
        /// <param name="buffer">The buffer to compress.</param>
        /// <param name="padding">The padding offset to pre-allocate.</param>
        /// <param name="context">The context of the processing.</param>
        /// <returns>The length segment</returns>
        public static BufferSegment Process(ByteStream stream, int padding, ProcessingContext context)
        {
            stream.Flush();

            // Calculate the maximum compressed length
            var outSize = Snappy.MaxEncodedLen((int)stream.Length);
            var output = context.BufferReserve(outSize + padding);

            // Acquire a snappy encoder which comes in with a table for state
            using (var snappy = Snappy.Acquire())
            {
                output.Size = snappy.Encode(stream.GetBuffer(), 0, (int)stream.Length, output, padding) + padding;
                return output;
            }
        }
    }
}