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
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a decoder for WebSocket packets.
    /// </summary>
    public static unsafe class WebSocketHybi13Encoder
    {
        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Emitter.Connection channel, ProcessingContext context)
        {
            // Only process precompiled buffer
            var inputBuffer = context.Buffer;
            if (inputBuffer == null)
                return ProcessingState.Failure;

            // Get the length & the final length
            var length = inputBuffer.Length;
            var finalLength = length > UInt16.MaxValue
                ? length + 10
                : length > 125
                    ? length + 4
                    : length + 2;

            // Reserve the buffer
            var buffer = context.BufferReserve(finalLength);

            // Get the frame type
            /*var frameType = channel.TransportEncoding == TransportEncoding.Binary
                    ? WebSocketFrameType.Binary
                    : WebSocketFrameType.Text;*/
            var frameType = WebSocketFrameType.Binary;

            // Is it a pong request? change the frame type
            if (context.Packet is WebSocketPong)
                frameType = WebSocketFrameType.Pong;

            // Get the operation byte
            var op = (byte)((byte)frameType + 128);
            var segment = buffer.AsSegment();
            var offset = segment.Offset;

            // Write the operation first
            segment.Array[offset] = op;
            ++offset;

            if (length > UInt16.MaxValue)
            {
                // Length flag
                segment.Array[offset] = (byte)127;
                ++offset;

                // Length value
                var lengthBytes = length.ToBigEndianBytes<ulong>();
                Memory.Copy(lengthBytes, 0, segment.Array, offset, lengthBytes.Length);
                offset += lengthBytes.Length;
            }
            else if (length > 125)
            {
                // Length flag
                segment.Array[offset] = (byte)126;
                ++offset;

                // Length value
                var lengthBytes = length.ToBigEndianBytes<ushort>();
                Memory.Copy(lengthBytes, 0, segment.Array, offset, lengthBytes.Length);
                offset += lengthBytes.Length;
            }
            else
            {
                // Length value
                segment.Array[offset] = (byte)length;
                ++offset;
            }

            // Write the payload
            Memory.Copy(inputBuffer.Array, inputBuffer.Offset, segment.Array, offset, inputBuffer.Length);
            offset += length;

            // Switch to the new buffer
            context.SwitchBuffer(buffer);

            // Trace a websocket event
            NetTrace.WriteLine("Hybi13 frame [" + buffer.Length + "] encoded", channel, NetTraceCategory.WebSocket);

            return ProcessingState.Success;
        }
    }
}