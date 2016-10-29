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
using System.Security.Cryptography;
using System.Text;
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Websockets Draft76 implementation.
    /// </summary>
    public class WebSocketDraft76Upgrade : IWebSocketProtocol
    {
        /// <summary>
        /// Upgrades the connection to the particular protocol. Handles the handshake.
        /// </summary>
        /// <param name="context"><see cref="ProcessingContext"/> for the current connection.</param>
        /// <param name="httpContext"><see cref="HttpContext"/> for the current connection.</param>
        /// <returns>The handlers that have been inserted in the pipeline.</returns>
        public WebSocketPipeline Upgrade(ProcessingContext context, HttpContext httpContext)
        {
            var request = httpContext.Request;
            var builder = new StringBuilder();

            builder.Append("HTTP/1.1 101 WebSocket Protocol Handshake\r\n");
            builder.Append("Upgrade: WebSocket\r\n");
            builder.Append("Connection: Upgrade\r\n");
            builder.AppendFormat("Sec-WebSocket-Origin: {0}\r\n", request.Headers["Origin"]);
            //builder.AppendFormat("Sec-WebSocket-Location: {0}://{1}{2}\r\n", secure ? "wss" : "ws", request.Headers["Host"], request.Path);
            builder.AppendFormat("Sec-WebSocket-Location: {0}://{1}{2}\r\n", "ws", request.Headers["Host"], request.Path);

            if (request.Headers["Sec-WebSocket-Protocol"] != null)
                builder.AppendFormat("Sec-WebSocket-Protocol: {0}\r\n", request.Headers["Sec-WebSocket-Protocol"]);

            builder.Append("\r\n");

            var key1 = request.Headers["Sec-WebSocket-Key1"];
            var key2 = request.Headers["Sec-WebSocket-Key2"];

            // Get last bytes
            byte[] challenge = request.Body;

            // Compile the body
            var part1 = Encoding.ASCII.GetBytes(builder.ToString());
            var part2 = CalculateAnswerBytes(key1, key2, challenge);
            var buffer = new byte[part1.Length + part2.Length];
            Memory.Copy(part1, 0, buffer, 0, part1.Length);
            Memory.Copy(part2, 0, buffer, part1.Length, part2.Length);

            // Prepare the response packet
            var response = BytePacket.Acquire(buffer);

            // Get the channel
            var channel = httpContext.Connection;

            // Send the handshake response
            channel.Send(response);

            // Set the encoder & the decoder for this websocket handler
            channel.Encoding.PipelineAddLast(Encode.WebSocketDraft76);
            channel.Decoding.PipelineAddFirst(Decode.WebSocketDraft76);

            // Trace a websocket event
            NetTrace.WriteLine("Upgraded to Draft76 ", channel, NetTraceCategory.WebSocket);
            return new WebSocketPipeline(Encode.WebSocketDraft76, Decode.WebSocketDraft76);
        }

        private static byte[] CalculateAnswerBytes(string key1, string key2, byte[] challenge)
        {
            byte[] result1Bytes = ParseKey(key1);
            byte[] result2Bytes = ParseKey(key2);

            var rawAnswer = new byte[16];
            Array.Copy(result1Bytes, 0, rawAnswer, 0, 4);
            Array.Copy(result2Bytes, 0, rawAnswer, 4, 4);
            Array.Copy(challenge, 0, rawAnswer, 8, 8);

            return MD5.Create().ComputeHash(rawAnswer);
        }

        private static byte[] ParseKey(string key)
        {
            key = key.TrimStart();

            int spaces = key.Count(x => x == ' ');
            var digits = new String(key.Where(Char.IsDigit).ToArray());

            var value = (Int32)(Int64.Parse(digits) / spaces);

            byte[] result = BitConverter.GetBytes(value);
            if (BitConverter.IsLittleEndian)
                Array.Reverse(result);
            return result;
        }
    }
}