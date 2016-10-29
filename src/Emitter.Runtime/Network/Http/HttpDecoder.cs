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

using System.Runtime.CompilerServices;
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a decoder of various HTTP requests.
    /// </summary>
    public static unsafe class HttpDecoder
    {
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

            // Check if it looks like an http request but only with a single byte
            var pBuffer = buffer.AsBytePointer();
            var length = context.Buffer.Length;
            var first = (char)*pBuffer;
            if (length == 1 &&
                   (first == 'P' ||
                    first == 'G' ||
                    first == 'H' ||
                    first == 'D' ||
                    first == 'O' ||
                    first == 'T' ||
                    first == 'C'))
                return ProcessingState.InsufficientData;

            // Check if we can handle it
            if (!CanHandle(buffer))
                return ProcessingState.Failure;

            // Acquire a HTTP context and save it in the session
            HttpContext httpContext = HttpContextPool.Default.Acquire(channel);
            context.Session = httpContext;

            // Make sure we have HTTP timeout set
            if (channel.Timeout != httpContext.ScriptTimeout)
                channel.Timeout = httpContext.ScriptTimeout;

            // Check if we have a http request
            if (buffer.Length < 14) // "GET / HTTP/1.1".Length;
                return ProcessingState.InsufficientData;

            // Parse http request
            var errorOccured = false;
            var insufficientData = false;
            var bytesParsed = httpContext.ParseRequest(context.Buffer.AsBytePointer(), length,
                out errorOccured, out insufficientData);

            // If there was en error during the parse, fail it
            if (errorOccured)
                return ProcessingState.Failure;

            // Parsing of http headers was not complete, need to read more bytes
            if (insufficientData)
                return ProcessingState.InsufficientData;

            // Trace HTTP request
            var request = httpContext.Request;
            NetTrace.WriteLine(request.HttpVerb.ToString().ToUpper() + " " + request.Path, channel, NetTraceCategory.Http);

            // Attempt to read the body
            if (buffer.Length < bytesParsed + request.ContentLength)
                return ProcessingState.InsufficientData;

            // We got enough data, map the read data to the request
            if (request.ContentLength > 0)
            {
                // Switch to the current buffer, release the old one
                context.Throttle(bytesParsed + request.ContentLength);
                context.SwitchBuffer(
                    buffer.Split(bytesParsed)
                    );

                // Parse the body to the request
                request.Body = context.Buffer.AsArray();
            }
            else
            {
                // Make a special case for draft websockets, as they send a body with a
                // GET request. Bastards!
                if (bytesParsed + 8 <= buffer.Length &&
                    request.ContainsHeader("Sec-WebSocket-Key1"))
                {
                    // Get the 8 bytes of websocket challenge
                    context.Throttle(bytesParsed + 8);
                    context.SwitchBuffer(
                        buffer.Split(bytesParsed)
                        );

                    // Parse the body to the request
                    request.Body = context.Buffer.AsArray();
                }
                else
                {
                    // Switch to the current buffer, release the old one
                    context.Throttle(bytesParsed);
                    context.SwitchBuffer(null);
                }
            }

            // Redirect to the various built-in handlers
            context.Redirect(
                Handle.WebSocketUpgrade,
                Handle.Http
                );

            return ProcessingState.Success;
        }

        #region Parse Check

        /// <summary>
        /// Checks whether the incoming request is HTTP one.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static bool CanHandle(BufferSegment buffer)
        {
            if (buffer.Length < 6)
                return false;

            byte* pBuffer = buffer.AsBytePointer();
            char first = (char)*pBuffer;

            if (first == 'P' ||
                first == 'G' ||
                first == 'H' ||
                first == 'D' ||
                first == 'O' ||
                first == 'T' ||
                first == 'C')
            {
                // Next, check if it contains a HTTP Verb

                if (*(pBuffer) == 'P' &&
                    *(pBuffer + 1) == 'O' &&
                    *(pBuffer + 2) == 'S' &&
                    *(pBuffer + 3) == 'T')
                    return true;
                if (*(pBuffer) == 'G' &&
                    *(pBuffer + 1) == 'E' &&
                    *(pBuffer + 2) == 'T')
                    return true;
                if (*(pBuffer) == 'P' &&
                    *(pBuffer + 1) == 'U' &&
                    *(pBuffer + 2) == 'T')
                    return true;
                if (*(pBuffer) == 'H' &&
                    *(pBuffer + 1) == 'E' &&
                    *(pBuffer + 2) == 'A' &&
                    *(pBuffer + 3) == 'D')
                    return true;
                if (*(pBuffer) == 'D' &&
                    *(pBuffer + 1) == 'E' &&
                    *(pBuffer + 2) == 'L' &&
                    *(pBuffer + 3) == 'E' &&
                    *(pBuffer + 4) == 'T' &&
                    *(pBuffer + 5) == 'E')
                    return true;
                if (*(pBuffer) == 'O' &&
                    *(pBuffer + 1) == 'P' &&
                    *(pBuffer + 2) == 'T' &&
                    *(pBuffer + 3) == 'I' &&
                    *(pBuffer + 4) == 'O' &&
                    *(pBuffer + 5) == 'N' &&
                    *(pBuffer + 6) == 'S')
                    return true;
                if (*(pBuffer) == 'T' &&
                    *(pBuffer + 1) == 'R' &&
                    *(pBuffer + 2) == 'A' &&
                    *(pBuffer + 3) == 'C' &&
                    *(pBuffer + 4) == 'E')
                    return true;
                if (*(pBuffer) == 'C' &&
                    *(pBuffer + 1) == 'O' &&
                    *(pBuffer + 2) == 'N' &&
                    *(pBuffer + 3) == 'N' &&
                    *(pBuffer + 4) == 'E' &&
                    *(pBuffer + 5) == 'C' &&
                    *(pBuffer + 6) == 'T')
                    return true;
            }
            return false;
        }

        #endregion Parse Check
    }
}