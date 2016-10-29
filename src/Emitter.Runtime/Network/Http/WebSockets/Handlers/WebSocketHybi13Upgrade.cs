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
using System.Text;
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a handler for a websocket connection upgrade of hybi13 implementation.
    /// </summary>
    public class WebSocketHybi13Upgrade : IWebSocketProtocol
    {
        /// <summary>
        /// Upgrades the connection to the particular protocol. Handles the handshake.
        /// </summary>
        /// <param name="context"><see cref="ProcessingContext"/> for the current connection.</param>
        /// <param name="httpContext"><see cref="HttpContext"/> for the current connection.</param>
        /// <returns>The handlers that have been inserted in the pipeline.</returns>
        public WebSocketPipeline Upgrade(ProcessingContext context, HttpContext httpContext)
        {
            // Compute the handshake response key
            var inputKey = httpContext.Request.Headers["Sec-WebSocket-Key"];
            var responseKey = System.Convert.ToBase64String((inputKey.Trim() + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
                .GetSHA1Bytes());

            var request = httpContext.Request;

            var builder = new StringBuilder();
            builder.Append("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + responseKey + "\r\n");
            if (request.Headers["Sec-WebSocket-Protocol"] != null)
                builder.AppendFormat("Sec-WebSocket-Protocol: {0}\r\n", request.Headers["Sec-WebSocket-Protocol"]);

            builder.Append("\r\n");
            var response = new StringPacket(Encoding.ASCII, builder.ToString());

            // Prepare the response packet
            /*var response = new StringPacket(Encoding.ASCII,
                "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + responseKey + "\r\n\r\n"
                );*/

            // Get the channel
            var channel = httpContext.Connection;

            // Send the handshake response
            channel.Send(response);

            // Set the encoder & the decoder for this websocket handler
            channel.Encoding.PipelineAddLast(Encode.WebSocketHybi13);
            channel.Decoding.PipelineAddFirst(Decode.WebSocketHybi13);

            // Trace a websocket event
            NetTrace.WriteLine("Upgraded to Hybi13 ", channel, NetTraceCategory.WebSocket);
            return new WebSocketPipeline(Encode.WebSocketHybi13, Decode.WebSocketHybi13);
        }
    }
}