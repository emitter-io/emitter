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

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a cached resource item.
    /// </summary>
    public class HttpResource
    {
        #region Mime Resolver
        private static readonly HttpMimeMap MimeMap = new HttpMimeMap();
        #endregion Mime Resolver

        #region Constructor

        // Lazy fields
        private FileInfo FilePtr = null;

        private byte[] CacheRaw = null;
        private string Mime = null;

        /// <summary>
        /// Constructs a new instance of <see cref="HttpResource"/>.
        /// </summary>
        /// <param name="lastWriteUtc">The last write time in universal format.</param>
        /// <param name="rawContent">The raw byte content of this resource.</param>
        /// <param name="mime">The content type of the resource.</param>
        public HttpResource(DateTime lastWriteUtc, byte[] rawContent, string mime)
        {
            this.LastWriteUtc = lastWriteUtc;
            this.CacheRaw = rawContent;
            this.Mime = mime;
        }

        /// <summary>
        /// Constructs a new instance of <see cref="HttpResource"/>.
        /// </summary>
        /// <param name="file">The file on disk to cache.</param>
        public HttpResource(FileInfo file)
        {
            this.LastWriteUtc = file.LastWriteTimeUtc;
            this.FilePtr = file;
            this.Mime = MimeMap.GetMime(file.Extension);
        }

        #endregion Constructor

        #region Properties

        /// <summary>
        /// Gets the last write time.
        /// </summary>
        public DateTime LastWriteUtc
        {
            get;
            private set;
        }

        /// <summary>
        /// Gest the raw representation of a cached resource.
        /// </summary>
        public byte[] Raw
        {
            get
            {
                if (this.CacheRaw == null && this.FilePtr != null)
                    this.CacheRaw = File.ReadAllBytes(this.FilePtr.FullName);

                return this.CacheRaw;
            }
        }

        #endregion Properties

        /// <summary>
        /// Writes the cached resource to the http response.
        /// </summary>
        /// <param name="context">The HttpContext to write to.</param>
        public void WriteTo(HttpContext context)
        {
            var request = context.Request;
            var response = context.Response;

            // Check if-modified-since header
            var ifModifiedSince = request.Headers.Get("If-Modified-Since");
            if (!String.IsNullOrEmpty(ifModifiedSince))
            {
                DateTime cachedDate;
                if (DateTime.TryParse(ifModifiedSince, out cachedDate))
                {
                    if (this.LastWriteUtc <= cachedDate)
                    {
                        response.Status = "304";
                        return;
                    }
                }
            }

            // Get the expiration date
            var expiresDate = DateTime.UtcNow + TimeSpan.FromDays(8);

            // Write default headers
            response.Headers.Set("Date", DateTime.UtcNow.ToString("R"));
            response.Headers.Set("Last-Modified", this.LastWriteUtc.ToString("R"));
            response.Headers.Set("Expires", expiresDate.ToString("R"));
            response.Cookieless = true;

            // Always write MIME
            response.ContentType = this.Mime;

            // Write the content
            response.Write(this.Raw);
            response.Flush();
        }
    }
}