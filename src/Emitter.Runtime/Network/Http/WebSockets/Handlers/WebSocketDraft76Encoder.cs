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
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a decoder for WebSocket packets.
    /// </summary>
    public static class WebSocketDraft76Encoder
    {
        private const byte Start = 0;
        private const byte End = 255;
        private const int MaxSize = 1024 * 1024 * 5;

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

            // Get the length & reserve new buffer
            var length = inputBuffer.Length;
            var buffer = context.BufferReserve(length + 2);

            // Wrap in between 0 and 255
            buffer.Array[buffer.Offset] = Start;
            Memory.Copy(inputBuffer.Array, inputBuffer.Offset, buffer.Array, buffer.Offset + 1, length);
            buffer.Array[buffer.Offset + length + 1] = End;

            // Switch to the new buffer
            context.SwitchBuffer(buffer);

            // Trace a websocket event
            NetTrace.WriteLine("Draft76 frame [" + buffer.Length + "] encoded", channel, NetTraceCategory.WebSocket);

            return ProcessingState.Success;
        }
    }
}