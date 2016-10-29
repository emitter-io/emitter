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
    public struct HttpRequestContent
    {
        /// <summary>
        /// The actual bytes of the content.
        /// </summary>
        private readonly byte[] Bytes;

        /// <summary>
        /// The encoding type
        /// </summary>
        private readonly Encoding Encoding;

        /// <summary>
        /// Constructs a new <see cref="HttpRequestContent"/>.
        /// </summary>
        /// <param name="encoding">HttpRequest encoding type</param>
        /// <param name="bytes">Raw bytes of the body</param>
        public HttpRequestContent(Encoding encoding, byte[] bytes)
        {
            this.Encoding = encoding;
            this.Bytes = bytes == null ? ArrayUtils<byte>.Empty : bytes;
        }

        /// <summary>
        /// Gets the raw bytes representation of the request body.
        /// </summary>
        /// <returns>Raw bytes representation of the request body.</returns>
        public byte[] AsBytes()
        {
            return this.Bytes;
        }

        /// <summary>
        /// Gets a readable <see cref="ByteStream"/> representation of the request body.
        /// </summary>
        /// <returns>A readable <see cref="ByteStream"/> representation of the request body.</returns>
        public Stream AsStream()
        {
            var stream = ByteStreamPool.Default.Acquire();
            stream.Write(this.Bytes);
            return stream;
        }

        /// <summary>
        /// Gets the string representation of the request body.
        /// </summary>
        /// <returns>String representation of the request body.</returns>
        public string AsString()
        {
            if (this.Encoding == null)
                throw new HttpEncodingException("Unable to represent the content as a string due to unknown Content-Encoding");
            if (this.Bytes == ArrayUtils<byte>.Empty)
                return String.Empty;

            return this.Encoding.GetString(this.Bytes);
        }
    }

    /// <summary>
    /// Represents a exception relation to HTTP Encoding.
    /// </summary>
    public sealed class HttpEncodingException : Exception
    {
        /// <summary>
        /// Constructs a new instance of <see cref="HttpEncodingException"/>.
        /// </summary>
        /// <param name="message">The message of the exception.</param>
        public HttpEncodingException(string message) : base(message)
        {
        }
    }
}