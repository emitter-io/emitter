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
using Emitter.Compression;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a decoder for Snappy compression.
    /// </summary>
    public static unsafe class SnappyDecoder
    {
        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Emitter.Connection channel, ProcessingContext context)
        {
            // Get the buffer to decompress
            var input = context.Buffer.AsSegment();
            try
            {
                // Reserve the buffer, we know exactly how many bytes to decompress
                var output = context.BufferReserve(
                    (int)VarInt.UvarInt(input.Array, input.Offset).Value
                    );

                // Decompress
                var length = Snappy.Decode(
                    input.Array, input.Offset, input.Count,
                    output.Array, output.Offset, output.Size
                    );

                // Switch the buffer to the decompressed one
                context.SwitchBuffer(output);
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }

            return ProcessingState.Success;
        }
    }
}