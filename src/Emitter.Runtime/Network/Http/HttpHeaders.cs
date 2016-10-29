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

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a packet for HTTP headers.
    /// </summary>
    public sealed class HttpHeaders : StringPacket
    {
        /// <summary>
        /// Pool that manages request packet instances.
        /// </summary>
        static internal readonly PacketPool PacketPool = new PacketPool("HttpHeaders", (_) => new HttpHeaders());

        /// <summary>
        /// Creates an instance of a http header packet
        /// </summary>
        public static HttpHeaders Create(HttpResponse response)
        {
            HttpHeaders packet = PacketPool.Acquire() as HttpHeaders;
            HttpHeaders.CompileTo(response, packet.StringBuilder);
            return packet;
        }

        /// <summary>
        /// Constructs a new HttpHeaders object instance.
        /// </summary>
        public HttpHeaders() : base(Encoding.ASCII) { }

        /// <summary>
        /// Writes the HttpResponse to this HttpHeaders packet.
        /// </summary>
        /// <param name="response">HttpResponse containing the headers.</param>
        internal static byte[] Compile(HttpResponse response)
        {
            // Build the headers
            // TODO: improve the way it's done
            var sb = new StringBuilder();
            HttpHeaders.CompileTo(response, sb);
            return Encoding.ASCII.GetBytes(sb.ToString());
        }

        /// <summary>
        /// Writes the HttpResponse to this HttpHeaders packet.
        /// </summary>
        /// <param name="response">HttpResponse containing the headers.</param>
        /// <param name="destination">The string builder to compile into.</param>
        internal static void CompileTo(HttpResponse response, StringBuilder destination)
        {
            destination.Append("HTTP/1.1 ");
            destination.Append(response.Status);
            destination.Append("\r\nContent-Length: ");
            destination.Append(response.GetUnderlyingStream().Length);

            string[] keys = response.HeaderCache.AllKeys;
            int keyCount = keys.Length;
            for (int i = 0; i < keyCount; ++i)
            {
                string value = response.HeaderCache.Get(keys[i]);
                if (!String.IsNullOrEmpty(value))
                {
                    destination.Append("\r\n");
                    destination.Append(keys[i]);
                    destination.Append(": ");
                    destination.Append(value);
                }
            }

            if (response.CookieCache != null && response.CookieCache.Count > 0 && !response.Cookieless)
            {
                keyCount = response.CookieCache.Count;
                for (int i = 0; i < keyCount; ++i)
                {
                    destination.Append("\r\n");
                    destination.Append("Set-Cookie: ");
                    destination.Append(response.CookieCache[i].GetCookieHeader());
                }
            }

            destination.Append("\r\n\r\n");
        }
    }
}