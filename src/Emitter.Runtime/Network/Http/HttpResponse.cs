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
using System.Collections.Specialized;
using System.IO;
using System.Linq;
using System.Text;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents an outgoing HTTP Response
    /// </summary>
    public class HttpResponse : Packet
    {
        private MemoryStream fStream;
        private StreamWriter fWriter;
        private string fStatus = "200";
        private bool fShouldSend = true;
        private bool fKeepAlive = true;
        private bool fCookieless = false;
        private bool fIsBeingRedirected = false;

        internal NameValueCollection HeaderCache = new NameValueCollection();
        internal HttpCookieCollection CookieCache;
        private bool fPureBinary = true;

        internal HttpResponse() : base()
        {
            // Acquire a new buffer that will be used for sending this packet
            this.fStream = new MemoryStream();

            // Recycle to reset everything
            Recycle();
        }

        /// <summary>
        /// Recycles this <see cref="HttpResponse"/> instance, reinitializing everything.
        /// </summary>
        public override void Recycle()
        {
            // Recycle the base packet
            base.Recycle();

            // Lifetime is managed by the HttpContext
            this.Lifetime = PacketLifetime.Manual;

            // Reset the buffer
            this.fStream.Position = 0;
            this.fStream.SetLength(0);
            this.fWriter = new StreamWriter(this.fStream, Encoding.UTF8);

            // Set the flags
            this.fShouldSend = true;
            this.fKeepAlive = true;
            this.fPureBinary = true;
            this.fCookieless = false;
            this.fIsBeingRedirected = false;

            // Clear the header cache
            if (this.HeaderCache != null)
                this.HeaderCache.Clear();

            // Clear the cookie cache
            if (this.CookieCache != null)
                this.CookieCache.Clear();

            // Default values
            this.HeaderCache.Set("Content-Type", "text/html");
            this.HeaderCache.Set("Location", null);
            this.HeaderCache.Set("Server", "Emitter");
            this.HeaderCache.Set("Access-Control-Allow-Origin", "*");

            // Default status is 200 (OK)
            fStatus = "200";
        }

        #region Properties

        /// <summary>
        /// Gets or sets the status code causes a string describing the status of the HTTP output to be returned to the client. The default value is 200 (OK).
        /// </summary>
        public string Status
        {
            get { return fStatus; }
            set { fStatus = value; }
        }

        /// <summary>
        /// Gets or sets the HTTP MIME type of the output stream. The default value is "text/html".
        /// </summary>
        public string ContentType
        {
            get { return this.HeaderCache.Get("Content-Type"); }
            set { this.HeaderCache.Set("Content-Type", value); }
        }

        /// <summary>
        /// Gets or sets the content encoding type for this request. Be aware that it resets the writer to
        /// the beginning of the stream.
        /// </summary>
        public Encoding ContentEncoding
        {
            get { return fWriter.Encoding; }
            set
            {
                // TODO: Pool StreamWriters
                fWriter = new StreamWriter(this.fStream, value);
            }
        }

        /// <summary>
        /// Gets or sets the redirect location (to be used in conjuction with http 300)
        /// </summary>
        public string Location
        {
            get { return this.HeaderCache.Get("Location"); }
            set { this.HeaderCache.Set("Location", value); }
        }

        /// <summary>
        /// Gets or sets whether this Http response should be sent or not.
        /// </summary>
        public bool ShouldSend
        {
            get { return fShouldSend; }
            set { fShouldSend = value; }
        }

        /// <summary>
        /// Gets or sets whether this response maintains alive the connection or not
        /// </summary>
        public bool KeepAlive
        {
            get { return fKeepAlive; }
            set { fKeepAlive = value; }
        }

        /// <summary>
        /// Gets or sets whether this response should not include any cookies.
        /// </summary>
        public bool Cookieless
        {
            get { return fCookieless; }
            set { fCookieless = value; }
        }

        /// <summary>
        /// Gets a collection of HTTP headers.
        /// </summary>
        public NameValueCollection Headers
        {
            get { return this.HeaderCache; }
        }

        /// <summary>
        /// Gets the response cookie collection.
        /// </summary>
        public HttpCookieCollection Cookies
        {
            get
            {
                if (this.CookieCache == null)
                    this.CookieCache = new HttpCookieCollection(true, false);
                return this.CookieCache;
            }
        }

        /// <summary>
        /// Gets whether the request is being redirected or not.
        /// </summary>
        public bool IsRequestBeingRedirected
        {
            get { return fIsBeingRedirected; }
        }

        #endregion Properties

        #region Public Methods

        /// <summary>
        /// Gets the underlying stream buffer.
        /// </summary>
        /// <returns>The underlying stream buffer.</returns>
        public MemoryStream GetUnderlyingStream()
        {
            return this.fStream;
        }

        /// <summary>
        /// Gets the underlying stream writer.
        /// </summary>
        /// <returns>The underlying stream writer.</returns>
        public StreamWriter GetUnderlyingWriter()
        {
            return this.fWriter;
        }

        /// <summary>
        /// Gets the content written so far in the http response.
        /// </summary>
        public HttpResponseContent Content
        {
            get { return new HttpResponseContent(this); }
        }

        /// <summary>
        /// Clears the headers written so far to this <see cref="HttpResponse"/>.
        /// </summary>
        public void ClearHeaders()
        {
            // Clear the cache
            if (this.HeaderCache != null)
                this.HeaderCache.Clear();

            // Set the default headers back
            this.HeaderCache.Set("Content-Type", "text/html");
            this.HeaderCache.Set("Location", null);
            this.HeaderCache.Set("Server", "Emitter");
        }

        /// <summary>
        /// Clears the buffer written so far to this <see cref="HttpResponse"/>.
        /// </summary>
        public void ClearContent()
        {
            this.fPureBinary = true;
            this.fStream.Position = 0;
            this.fStream.SetLength(0);
            this.fWriter = new StreamWriter(this.fStream, Encoding.UTF8);
        }

        /// <summary>
        /// Writes a character to this <see cref="HttpResponse"/>.
        /// </summary>
        /// <param name="ch">The character to write.</param>
        public void Write(char ch)
        {
            fPureBinary = false;
            fWriter.Write(ch);
        }

        /// <summary>
        /// Writes the object's string representation to this <see cref="HttpResponse"/>.
        /// </summary>
        /// <param name="o">The object's string representation to write.</param>
        public void Write(object o)
        {
            fPureBinary = false;
            fWriter.Write(o);
        }

        /// <summary>
        /// Writes the string to this <see cref="HttpResponse"/>.
        /// </summary>
        /// <param name="s">The string to write.</param>
        public void Write(string s)
        {
            if (s == null || s == String.Empty)
                return;
            fPureBinary = false;
            fWriter.Write(s);
        }

        /// <summary>
        /// Writes an array of character to this <see cref="HttpResponse"/>.
        /// </summary>
        /// <param name="array">The array to write.</param>
        /// <param name="offset">The offset in the input array. Specifies where to begin.</param>
        /// <param name="length">The number of characters to write starting from the offset.</param>
        public void Write(char[] array, int offset, int length)
        {
            fPureBinary = false;
            fWriter.Write(array, offset, length);
        }

        /// <summary>
        /// Writes an byte array to this <see cref="HttpResponse"/>.
        /// </summary>
        /// <param name="array">The array to write.</param>
        /// <param name="offset">The offset in the input array. Specifies where to begin.</param>
        /// <param name="length">The number of bytes to write starting from the offset.</param>
        public void Write(byte[] array, int offset, int length)
        {
            this.fStream.Write(array, offset, length);
        }

        /// <summary>
        /// Writes an byte array to this <see cref="HttpResponse"/>.
        /// </summary>
        public void Write(byte[] array)
        {
            this.fStream.Write(array, 0, array.Length);
        }

        /// <summary>
        /// Overwrites the contents of the response with the specified bytes.
        /// </summary>
        /// <param name="bytes">The byte array to write into the response.</param>
        public void Overwrite(byte[] bytes)
        {
            this.ClearContent();
            this.Write(bytes);
        }

        /// <summary>
        /// Overwrites an byte array to this <see cref="HttpResponse"/>.
        /// </summary>
        /// <param name="array">The array to write.</param>
        /// <param name="offset">The offset in the input array. Specifies where to begin.</param>
        /// <param name="length">The number of bytes to write starting from the offset.</param>
        public void Overwrite(byte[] array, int offset, int length)
        {
            this.ClearContent();
            this.Write(array, offset, length);
        }

        /// <summary>
        /// Flushes the underlying buffer stream.
        /// </summary>
        public void Flush()
        {
            // If no body expected, do nothing
            if (fStatus == "304")
                return;

            // Pure binary, avoid flushing the writer (otherwise it adds noise)
            if (fPureBinary)
                return;

            fWriter.Flush();
        }

        /// <summary>
        /// Sets an HTTP header to the target value.
        /// </summary>
        /// <param name="name">The name/key of the HTTP header to set.</param>
        /// <param name="value">The value for this HTTP header.</param>
        public void SetHeader(string name, string value)
        {
            this.HeaderCache.Add(name, value);
        }

        /// <summary>
        ///  Adds an HTTP cookie to the intrinsic cookie collection.
        /// </summary>
        /// <param name="cookie">A cookie is appended after the HTTP headers have been sent.</param>
        public void AppendCookie(HttpCookie cookie)
        {
            Cookies.Add(cookie);
        }

        #endregion Public Methods

        #region Redirect

        /// <summary>
        /// Redirects the response to a specific url.
        /// </summary>
        /// <param name="url">The url the redirection should be fowarded to.</param>
        /// <param name="endResponse">Whether the response should be ended or not.</param>
        /// <param name="code">The HTTP code of the response.</param>
        private void Redirect(string url, bool endResponse, int code)
        {
            if (url == null)
                throw new ArgumentNullException("url");

            if (url.IndexOf('\n') != -1)
                throw new ArgumentException("Redirect URI cannot contain newline characters.", "url");

            // Mark as redirected and clear contents & headers
            fIsBeingRedirected = true;
            ClearHeaders();
            ClearContent();

            // Set the status code
            this.Status = code.ToString();

            // ???
            //url = ApplyAppPathModifier(url);

            bool isFullyQualified = (
                 url.StartsWith("http:", StringComparison.OrdinalIgnoreCase) ||
                 url.StartsWith("https:", StringComparison.OrdinalIgnoreCase) ||
                 url.StartsWith("file:", StringComparison.OrdinalIgnoreCase) ||
                 url.StartsWith("ftp:", StringComparison.OrdinalIgnoreCase)
                 );

            /*if (!isFullyQualified)
            {
                HttpRuntimeSection config = HttpRuntime.Section;
                if (config != null && config.UseFullyQualifiedRedirectUrl)
                {
                    var ub = new UriBuilder(context.Request.Url);
                    int qpos = url.IndexOf('?');
                    if (qpos == -1)
                    {
                        ub.Path = url;
                        ub.Query = null;
                    }
                    else
                    {
                        ub.Path = url.Substring(0, qpos);
                        ub.Query = url.Substring(qpos + 1);
                    }
                    ub.Fragment = null;
                    ub.Password = null;
                    ub.UserName = null;
                    url = ub.Uri.ToString();
                }
            }*/

            // Setting the redirect location
            this.Location = url;

            // Text for browsers that can't handle location header
            Write("<html><head><title>Object moved</title></head><body>\r\n");
            Write("<h2>Object moved to <a href=\"" + url + "\">here</a></h2>\r\n");
            Write("</body><html>\r\n");

            // If we should end the response, simply throw the redirect to move the callstack up.
            if (endResponse)
                throw new HttpRedirectException();

            fIsBeingRedirected = false;
        }

        /// <summary>
        /// Temporarily redirects the request to the specified url.
        /// </summary>
        /// <param name="url">The url to which the request should be redirected.</param>
        public void Redirect(string url)
        {
            Redirect(url, true);
        }

        /// <summary>
        /// Temporarily redirects the request to the specified url.
        /// </summary>
        /// <param name="url">The url to which the request should be redirected.</param>
        /// <param name="endResponse">Whether the response should be terminated or not.</param>
        public void Redirect(string url, bool endResponse)
        {
            Redirect(url, endResponse, 302);
        }

        /// <summary>
        /// Permanently redirects the request to the specified url.
        /// </summary>
        /// <param name="url">The url to which the request should be redirected.</param>
		public void RedirectPermanent(string url)
        {
            RedirectPermanent(url, true);
        }

        /// <summary>
        /// Permanently redirects the request to the specified url.
        /// </summary>
        /// <param name="url">The url to which the request should be redirected.</param>
        /// <param name="endResponse">Whether the response should be terminated or not.</param>
		public void RedirectPermanent(string url, bool endResponse)
        {
            Redirect(url, endResponse, 301);
        }

        #endregion Redirect

        #region Static Constructors

        /// <summary>
        /// Creates and sends an HttpResponse packet for an exception, wraps it to an error 500
        /// </summary>
        internal static void Send500(Emitter.Connection client, Exception ex)
        {
            HttpResponse response = new HttpResponse();
            response.Status = "500";
            response.Write("<html><head><title>500 Internal Server Error</title></head><body>");
            response.Write("<h3>Internal Server Error (500)</h3>");
            response.Write("&nbsp;&nbsp;&nbsp;Exception : " + ex.GetType().FullName.ToString());
            response.Write("<br />&nbsp;&nbsp;&nbsp;Message   : " + ex.Message.ToString());
            response.Write("<br />&nbsp;&nbsp;&nbsp;Source    : " + ex.Source.ToString());
            if (!String.IsNullOrEmpty(ex.StackTrace))
            {
                response.Write("<br /><br />&nbsp;&nbsp;&nbsp;Stack     : ");
                response.Write(ex.StackTrace.Replace(Environment.NewLine, "<br />&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"));
            }
            response.Write("</body></html>");
            client.Send(response);
        }

        /// <summary>
        /// Creates and sends an HttpResponse packet for an exception, wraps it to an error 401
        /// </summary>
        internal static void Send401(Emitter.Connection client, Exception ex)
        {
            var response = new HttpResponse();
            response.Status = "401";
            response.Write("<html><head><title>401 Authorization Required</title></head><body>");
            response.Write("<h3>Authorization Required (401)</h3>");
            response.Write(ex.Message);
            response.Write("</body></html>");
            client.Send(response);
        }

        /// <summary>
        /// Creates and sends an HttpResponse packet for an not-found handler as error 404
        /// </summary>
        internal static void Send404(Emitter.Connection client)
        {
            var response = new HttpResponse();
            response.Status = "404";
            response.Write("<html><head><title>404 Not found</title></head><body>");
            response.Write("<h3>Not found (404)</h3>");
            response.Write("The resource you requested was not found on this server");
            response.Write("</body></html>");
            client.Send(response);
        }

        #endregion Static Constructors
    }
}