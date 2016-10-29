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
    public static unsafe class WebSocketDraft76Decoder
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
            // Get the buffer
            var buffer = context.Buffer;
            if (buffer == null)
                return ProcessingState.Failure;

            // Get the info
            var pBuffer = buffer.AsBytePointer();
            int length = buffer.Length;

            // To close the connection cleanly, a frame consisting of just a 0xFF
            // byte followed by a 0x00 byte is sent from one peer to ask that the
            // other peer close the connection.
            if (*pBuffer == 0xFF && *(pBuffer + 1) == 0x00)
            {
                // Trace a websocket event
                NetTrace.WriteLine("Draft76 connection termination requested", channel, NetTraceCategory.WebSocket);

                length = 0;
                channel.Close();
                return ProcessingState.Stop;
            }

            if (*pBuffer != Start)
                throw new WebSocketException("WebSocket frame is invalid.");

            // Search for the next end byte
            int endIndex = -1;
            for (int i = 0; i < length; ++i)
            {
                if (*(pBuffer + i) == End)
                {
                    endIndex = i;
                    break;
                }
            }

            if (endIndex == -1)
            {
                // Trace a websocket event
                NetTrace.WriteLine("Draft76 frame is not complete", channel, NetTraceCategory.WebSocket);
                return ProcessingState.InsufficientData;
            }

            if (endIndex > MaxSize)
                throw new WebSocketException("WebSocket frame is too large.");

            // Throttle the rest
            context.Throttle(endIndex + 1);

            // Decode the body
            var decoded = context.BufferWrite(buffer.Array, buffer.Offset + 1, endIndex - 1);
            context.SwitchBuffer(decoded);

            // Trace a websocket event
            NetTrace.WriteLine("Draft76 frame decoded", channel, NetTraceCategory.WebSocket);
            return ProcessingState.Success;
        }
    }
}