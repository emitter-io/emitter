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
using System.IO;
using System.Linq;
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a encoder of various <see cref="HttpResponse"/> packets.
    /// </summary>
    public static unsafe class HttpEncoder
    {
        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Emitter.Connection channel, ProcessingContext context)
        {
            // It can only process StringPacket objects
            var response = context.Packet as HttpResponse;
            if (response == null)
                return ProcessingState.Failure;

            // If we explicitely disabled the response, don't send it
            if (!response.ShouldSend)
                return ProcessingState.Failure;

            // Get the client
            var client = context.Client;
            var stream = response.GetUnderlyingStream();

            // Flush the writer, ensuring we have the correct compiled buffer here.
            response.Flush();

            // Compile the headers
            var headers = HttpHeaders.Compile(response);

            // Create the buffer
            var buffer = context.BufferReserve(headers.Length + (int)stream.Length);
            Memory.Copy(headers, 0, buffer.Array, buffer.Offset, headers.Length);
            Memory.Copy(stream.GetBuffer(), 0, buffer.Array, buffer.Offset + headers.Length, (int)stream.Length);

            // Trace HTTP request
            NetTrace.WriteLine(response.Status + " " + response.ContentType + " Length: " + stream.Length, channel, NetTraceCategory.Http);

            // Write to the buffer
            context.SwitchBuffer(buffer);

            return ProcessingState.Success;
        }
    }
}