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
using System.Linq;
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a decoder for WebSocket packets.
    /// </summary>
    public static unsafe class WebSocketHybi13Decoder
    {
        [ThreadStatic]
        private static byte[] MaskBuffer;

        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Emitter.Connection channel, ProcessingContext context)
        {
            var buffer = context.Buffer;
            if (buffer == null)
                return ProcessingState.Failure;

            // Get the buffer pointer
            var pBuffer = buffer.AsBytePointer();
            var offset = 0;
            var length = buffer.Length;

            // Decode the frame
            var queue = new Queue<BufferSegment>();
            int index; bool isFinal; BufferSegment decoded;
            do
            {
                //var trace = (context.Buffer.ViewAsHybi13 == "(Invalid: Hybi13)" ? context.Buffer.ViewAsBytes : context.Buffer.ViewAsHybi13);
                var result = DecodeFrame(context, pBuffer, length, out index, out isFinal, out decoded);
                //Console.WriteLine(trace + " -> " + result);
                if (result != ProcessingState.Success)
                    return result;

                pBuffer += index;
                offset += index;
                length -= index;

                queue.Enqueue(decoded);
            } while (!isFinal);

            // Merge the queue
            decoded = queue.Dequeue();
            while (queue.Count > 0)
                decoded = decoded.Join(queue.Dequeue());

            // Throttle the rest
            if (offset < buffer.Length)
                context.Throttle(offset);

            // Swap the buffer
            context.SwitchBuffer(decoded);

            NetTrace.WriteLine("Hybi13 frame decoded", channel, NetTraceCategory.WebSocket);
            return ProcessingState.Success;
        }

        /// <summary>
        /// Decodes a single hybi13 frame.
        /// </summary>
        /// <param name="context">The context for decoding.</param>
        /// <param name="pBuffer">The byte pointer to start decoding process.</param>
        /// <param name="length">The length of the buffer, starting at pBuffer offset.</param>
        /// <param name="index">The amount of procesed bytes.</param>
        /// <param name="isFinal">Whether the frame is final or not.</param>
        /// <param name="decoded">The frame decoded.</param>
        /// <returns>Processing state.</returns>
        private static ProcessingState DecodeFrame(ProcessingContext context, byte* pBuffer, int length, out int index, out bool isFinal, out BufferSegment decoded)
        {
            // Default values
            index = 0;
            decoded = null;

            // Check if we have a buffer for the mask
            if (MaskBuffer == null)
                MaskBuffer = new byte[4];

            // Get the first part of the header
            var frameType = (WebSocketFrameType)(*pBuffer & 15);
            isFinal = (*pBuffer & 128) != 0;

            // Must have at least some bytes
            if (length == 0)
                return ProcessingState.InsufficientData;

            // We only got one byte, kthx iOS
            if (length == 1 && (frameType == WebSocketFrameType.Text || frameType == WebSocketFrameType.Binary))
                return ProcessingState.InsufficientData;

            // Get the second part of the header
            var isMasked = (*(pBuffer + 1) & 128) != 0;
            var dataLength = (*(pBuffer + 1) & 127);

            // Close the connection if requested
            if (frameType == WebSocketFrameType.Close)
            {
                // Client sent us a close frame, we should close the connection
                context.Channel.Close();
                return ProcessingState.Stop;
            }

            // We got some weird shit
            if (!isMasked)
                throw new WebSocketException("Incorrect incoming Hybi13 data format.");

            // Validate the frame type
            switch (frameType)
            {
                // Frames with a payload
                case WebSocketFrameType.Text:
                case WebSocketFrameType.Binary:
                case WebSocketFrameType.Continuation:
                case WebSocketFrameType.Ping:
                case WebSocketFrameType.Pong:
                    break;

                // Unrecognized frame type
                default: throw new WebSocketException("Unsupported Huby13 frame (" + frameType + ") received.");
            }

            index = 2;
            int payloadLength;

            if (dataLength == 127)
            {
                if (length < index + 8)
                    return ProcessingState.InsufficientData;

                // TODO: Support little endian too
                payloadLength = *(pBuffer + 2) << 56
                              | *(pBuffer + 3) << 48
                              | *(pBuffer + 4) << 40
                              | *(pBuffer + 5) << 32
                              | *(pBuffer + 6) << 24
                              | *(pBuffer + 7) << 16
                              | *(pBuffer + 8) << 8
                              | *(pBuffer + 9);

                index += 8;
            }
            else if (dataLength == 126)
            {
                if (length < index + 2)
                    return ProcessingState.InsufficientData;

                // TODO: Support little endian too
                payloadLength = *(pBuffer + 2) << 8
                              | *(pBuffer + 3);

                index += 2;
            }
            else
            {
                payloadLength = dataLength;
            }

            // Check if we have the whole frame here
            if (length < index + payloadLength + 4)
                return ProcessingState.InsufficientData;

            // Get the mask bytes
            MaskBuffer[0] = *(pBuffer + index);
            MaskBuffer[1] = *(pBuffer + index + 1);
            MaskBuffer[2] = *(pBuffer + index + 2);
            MaskBuffer[3] = *(pBuffer + index + 3);
            index += 4;

            // Reserve a decode buffer to decode into
            decoded = context.BufferReserve(payloadLength);

            // Decode the body
            byte* pTo = decoded.AsBytePointer();
            byte* pFrom = pBuffer + index;
            for (int i = 0; i < payloadLength; ++i, ++pTo, ++pFrom)
            {
                *pTo = (byte)((*pFrom) ^ MaskBuffer[i % 4]);
            }

            // Set the new index
            index += payloadLength;
            switch (frameType)
            {
                // We need to reply to the ping
                case WebSocketFrameType.Ping:
                    {
                        // Send the websocket with the copy of the body (as per RFC)
                        context.Channel.Send(WebSocketPong.Acquire(decoded));
                        decoded.TryRelease();
                        return ProcessingState.Stop;
                    }

                // Pong is just used for the heartbeat
                case WebSocketFrameType.Pong: return ProcessingState.Stop;
            }

            // Trace a websocket event
            return ProcessingState.Success;
        }
    }
}