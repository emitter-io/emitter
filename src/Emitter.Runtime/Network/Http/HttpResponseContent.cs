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
using System.Text;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents an HTTP Request content body.
    /// </summary>
    public struct HttpResponseContent
    {
        /// <summary>
        /// The actual bytes of the content.
        /// </summary>
        private readonly MemoryStream Stream;

        /// <summary>
        /// The encoding type
        /// </summary>
        private readonly Encoding Encoding;

        /// <summary>
        /// Constructs a new <see cref="HttpRequestContent"/>.
        /// </summary>
        /// <param name="owner"><see cref="HttpRequest"/> object.</param>
        internal HttpResponseContent(HttpResponse owner)
        {
            // Flush the writer
            owner.Flush();

            // Set the fields
            this.Encoding = owner.ContentEncoding;
            this.Stream = owner.GetUnderlyingStream();
        }

        /// <summary>
        /// Gets the raw bytes representation of the response body.
        /// </summary>
        /// <returns>Raw bytes representation of the response body.</returns>
        public byte[] AsBytes()
        {
            var length = (int)this.Stream.Length;
            var source = this.Stream.GetBuffer();
            var target = new byte[length];
            Memory.Copy(source, 0, target, 0, length);
            return target;
        }

        /// <summary>
        /// Gets a readable <see cref="ByteStream"/> representation of the request body.
        /// </summary>
        /// <returns>A readable <see cref="ByteStream"/> representation of the request body.</returns>
        public Stream AsStream()
        {
            return this.Stream;
        }

        /// <summary>
        /// Gets the string representation of the request body.
        /// </summary>
        /// <returns>String representation of the request body.</returns>
        public string AsString()
        {
            if (this.Encoding == null)
                throw new HttpEncodingException("Unable to represent the content as a string due to unknown Content-Encoding");
            if (this.Stream == null || this.Stream.Length == 0)
                return String.Empty;

            return this.Encoding.GetString(this.Stream.GetBuffer(), 0, (int)this.Stream.Length);
        }
    }
}