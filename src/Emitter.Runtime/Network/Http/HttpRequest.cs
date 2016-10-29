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
using System.Collections.Specialized;
using System.Globalization;
using System.Linq;
using System.Text;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a HTTP Protocol Version
    /// </summary>
    public enum HttpVersion
    {
        /// <summary>
        /// HTTP Version 1.0
        /// </summary>
        V1_0,

        /// <summary>
        /// HTTP Version 1.1
        /// </summary>
        V1_1
    }

    /// <summary>
    /// Represents an incoming HTTP Request
    /// </summary>
    public class HttpRequest : RecyclableObject
    {
        internal TextRange HttpHeader;
        internal TextRange URI;

        internal HttpVersion Version;
        internal HttpVerb Verb;

        internal TextRange[] Keys = new TextRange[16];
        internal TextRange[] Values = new TextRange[16];
        internal int HeaderCount;

        internal NameValueCollection HeaderCache;
        internal HttpCookieCollection CookieCache;
        internal byte[] Body = ArrayUtils<byte>.Empty;

        // Private Members
        private Encoding EncodingCache;

        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        public override void Recycle()
        {
            this.HttpHeader = TextRange.Empty;
            this.URI = TextRange.Empty;

            this.Version = HttpVersion.V1_0;
            this.Verb = HttpVerb.Unknown;

            this.HeaderCount = 0;

            this.Keys = new TextRange[16];
            this.Values = new TextRange[16];
            this.Body = ArrayUtils<byte>.Empty;

            if (this.HeaderCache != null)
                this.HeaderCache.Clear();

            if (this.CookieCache != null)
                this.CookieCache.Clear();

            this.EncodingCache = null;
        }

        /// <summary>
        /// Attempts to get a specific header.
        /// </summary>
        /// <param name="headerName">The name of the header to check</param>
        /// <param name="headerValue">The value of the header to check</param>
        /// <returns>Returns whether the request contains the header or not</returns>
        public bool TryGetHeader(string headerName, out string headerValue)
        {
            headerValue = null;
            int count = this.HeaderCount;
            for (int i = 0; i < count; i++)
            {
                if (this.Keys[i].Is(headerName))
                {
                    headerValue = this.Values[i].ToString();
                    return true;
                }
            }

            return false;
        }

        /// <summary>
        /// Checks whether the request contains the header or not.
        /// </summary>
        /// <param name="headerName">The name of the header to check</param>
        /// <returns>Returns whether the request contains the header or not</returns>
        public bool ContainsHeader(string headerName)
        {
            int count = this.HeaderCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.Keys[i].Is(headerName))
                    return true;
            }
            return false;
        }

        #region Public Properties

        /// <summary>
        /// Gets a collection of HTTP headers.
        /// </summary>
        public NameValueCollection Headers
        {
            get
            {
                int count = this.HeaderCount;

                if (this.HeaderCache == null)
                    this.HeaderCache = new NameValueCollection(count);

                if (this.HeaderCache.Count == 0)
                {
                    for (int i = 0; i < count; i++)
                        this.HeaderCache[this.Keys[i].ToString()] = this.Values[i].ToString();
                }

                return this.HeaderCache;
            }
        }

        /// <summary>
        /// Gets a collection of cookies sent by the client.
        /// </summary>
        public HttpCookieCollection Cookies
        {
            get
            {
                if (this.CookieCache == null)
                    this.CookieCache = new HttpCookieCollection(false, false);

                if (this.CookieCache.Count == 0)
                {
                    var cookies = this.SafeGetHeader("Cookie");
                    if (!String.IsNullOrEmpty(cookies))
                    {
                        string[] cookie_components = cookies.Split(';');
                        foreach (string kv in cookie_components)
                        {
                            int pos = kv.IndexOf('=');
                            if (pos == -1)
                            {
                                continue;
                            }
                            else
                            {
                                string key = kv.Substring(0, pos);
                                string val = kv.Substring(pos + 1);

                                this.CookieCache.Add(new HttpCookie(key.Trim(), val.Trim()));
                            }
                        }
                    }
                }

                return this.CookieCache;
            }
        }

        /// <summary>
        /// Gets the Http Verb of the request
        /// </summary>
        public HttpVerb HttpVerb
        {
            get { return Verb; }
        }

        /// <summary>
        /// Gets the Accept of the incoming request.
        /// </summary>
        public string Accept
        {
            get { return this.SafeGetHeader("Accept"); }
        }

        /// <summary>
        /// Gets the Accept-Encoding of the incoming request.
        /// </summary>
        public string AcceptEncoding
        {
            get { return this.SafeGetHeader("Accept-Encoding"); }
        }

        /// <summary>
        /// Gets the Origin of the incoming request.
        /// </summary>
        public string Origin
        {
            get
            {
                var origin = this.SafeGetHeader("Origin");
                if (origin == null)
                    return origin;
                if (origin.Length == 0)
                    return null;
                return origin;
            }
        }

        /// <summary>
        /// Gets the Accept-Language of the incoming request.
        /// </summary>
        public string AcceptLanguage
        {
            get { return this.SafeGetHeader("Accept-Language"); }
        }

        /// <summary>
        /// Gets the Host of the incoming request.
        /// </summary>
        public string Host
        {
            get { return this.SafeGetHeader("Host"); }
        }

        /// <summary>
        /// Gets the Accept-Charset of the incoming request.
        /// </summary>
        public string AcceptCharset
        {
            get { return this.SafeGetHeader("Accept-Charset"); }
        }

        /// <summary>
        /// Gets information about the URL of the client's previous request that linked to the current URL.
        /// </summary>
        public Uri UrlReferrer
        {
            get
            {
                try
                {
                    // Attempt to get a URI
                    return new Uri(this.SafeGetHeader("Referer"));
                }
                catch
                {
                    throw new UriFormatException("The HTTP Referer request header is malformed and cannot be converted to a Uri object.");
                }
            }
        }

        /// <summary>
        /// Gets the character set of the entity-body.
        /// </summary>
        public Encoding ContentEncoding
        {
            get
            {
                if (this.EncodingCache == null)
                {
                    this.EncodingCache = this.GetEncodingFromHeaders();
                    if (this.EncodingCache == null)
                        this.EncodingCache = Encoding.UTF8;
                }
                return this.EncodingCache;
            }
            set
            {
                this.EncodingCache = value;
            }
        }

        /// <summary>
        /// Gets the MIME content type of the incoming request.
        /// </summary>
        public string ContentType
        {
            get { return this.SafeGetHeader("Content-Type"); }
        }

        /// <summary>
        /// Gets the type of Emitter Ray command.
        /// </summary>
        internal string RayType
        {
            get { return this.SafeGetHeader("Ray-Type"); }
        }

        /// <summary>
        /// Gets the type of Emitter Comet identity.
        /// </summary>
        internal string CometIdentity
        {
            get { return this.SafeGetHeader("Comet-Identity"); }
        }

        /// <summary>
        /// Specifies the length, in bytes, of content sent by the client.
        /// </summary>
        public int ContentLength
        {
            get
            {
                string s = this.GetHeader("Content-Length");

                if (string.IsNullOrEmpty(s))
                    return 0;

                int n;
                int.TryParse(s, out n);

                return n;
            }
        }

        /// <summary>
        /// Gets the virtual path of the current request.
        /// </summary>
        public string Path
        {
            get { return this.URI.ToString(); }
        }

        /// <summary>
        /// Gets the raw user agent string of the client browser.
        /// </summary>
        public string UserAgent
        {
            get { return this.SafeGetHeader("User-Agent"); }
        }

        /// <summary>
        /// Gets the raw referer string of the client browser.
        /// </summary>
        public string Referer
        {
            get { return this.SafeGetHeader("Referer"); }
        }

        /// <summary>
        /// Gets the IP host address of the remote client.
        /// </summary>
        public string UserHostAddress
        {
            get { return this.SafeGetHeader("HOST"); }
        }

        /// <summary>
        /// Gets the body content.
        /// </summary>
        public HttpRequestContent Content
        {
            get { return new HttpRequestContent(this.ContentEncoding, this.Body); }
        }

        /// <summary>
        /// Prints out the headers for debug view.
        /// </summary>
        internal string ViewAsString
        {
            get
            {
                var output = new StringBuilder();
                output.AppendLine("HTTP 1.1 " + this.Verb.ToString().ToUpper() + " " + this.Path);
                foreach (var header in this.Headers.AllKeys)
                {
                    string value;
                    if (this.TryGetHeader(header, out value))
                        output.AppendLine(header + " = " + value);
                }

                output.AppendLine();
                output.AppendLine();
                if (this.Body != null && this.Body.Length > 0)
                    output.AppendLine(this.Content.AsString());

                return output.ToString();
            }
        }

        #endregion Public Properties

        #region Internal Properties

        /// <summary>
        /// Gets whether the request is a fabric request.
        /// </summary>
        internal unsafe bool IsFabric
        {
            get
            {
                //const string UrlPatternA = "/socket.io/";
                //const string UrlPatternB = "/engine.io/";
                const int minLength = 16; // sizeof("/socket.io/?eio=")

                // Size check
                var url = this.URI;
                if (url.Length <= minLength)
                    return false;

                var charAt = url.pStart;
                return charAt[0] == '/' &&
                        (charAt[1] == 's' || charAt[1] == 'e') &&
                        (charAt[2] == 'o' || charAt[2] == 'n') &&
                        (charAt[3] == 'c' || charAt[3] == 'g') &&
                        (charAt[4] == 'k' || charAt[4] == 'i') &&
                        (charAt[5] == 'e' || charAt[5] == 'n') &&
                        (charAt[6] == 't' || charAt[6] == 'e') &&
                        charAt[7] == '.' &&
                        charAt[8] == 'i' &&
                        charAt[9] == 'o' &&
                        charAt[10] == '/';
            }
        }

        /// <summary>
        /// Views the headers of this request, for debugging purpose.
        /// </summary>
        internal string[] ViewHeaders
        {
            get
            {
                var view = new List<string>();
                foreach (string key in this.Headers.Keys)
                    view.Add(key + ": " + this.Headers.Get(key));
                return view.ToArray();
            }
        }

        /// <summary>
        /// Views the content of this request, for debugging purpose.
        /// </summary>
        internal string ViewContent
        {
            get
            {
                return this.Content.AsString();
            }
        }

        #endregion Internal Properties

        #region Private Members

        private string SafeGetHeader(string headerName)
        {
            int count = this.HeaderCount;
            for (int i = 0; i < count; i++)
            {
                if (this.Keys[i].Is(headerName))
                    return this.Values[i].ToString();
            }

            return string.Empty;
        }

        private string GetHeader(string headerName)
        {
            int count = this.HeaderCount;
            for (int i = 0; i < count; i++)
            {
                if (this.Keys[i].Is(headerName))
                    return this.Values[i].ToString();
            }

            return null;
        }

        private Encoding GetEncodingFromHeaders()
        {
            if ((this.UserAgent != null) && CultureInfo.InvariantCulture.CompareInfo.IsPrefix(this.UserAgent, "UP"))
            {
                string str = this.Headers["x-up-devcap-post-charset"];
                if (!string.IsNullOrEmpty(str))
                {
                    // Get encoding from string
                    try { return Encoding.GetEncoding(str); }
                    catch { }
                }
            }

            // If there's no body, return null
            if (this.Body == null)
                return null;
            string contentType = this.ContentType;
            if (contentType == null)
                return null;

            string attributeFromHeader = GetAttributeFromHeader(contentType, "charset");
            if (attributeFromHeader == null)
                return null;

            Encoding encoding = null;
            try { encoding = Encoding.GetEncoding(attributeFromHeader); }
            catch { }
            return encoding;
        }

        private static string GetAttributeFromHeader(string headerValue, string attrName)
        {
            int index;
            if (headerValue == null)
            {
                return null;
            }
            int length = headerValue.Length;
            int num2 = attrName.Length;
            int startIndex = 1;
            while (startIndex < length)
            {
                startIndex = CultureInfo.InvariantCulture.CompareInfo.IndexOf(headerValue, attrName, startIndex, CompareOptions.IgnoreCase);
                if ((startIndex < 0) || ((startIndex + num2) >= length))
                {
                    break;
                }
                char c = headerValue[startIndex - 1];
                char ch2 = headerValue[startIndex + num2];
                if ((((c == ';') || (c == ',')) || char.IsWhiteSpace(c)) && ((ch2 == '=') || char.IsWhiteSpace(ch2)))
                {
                    break;
                }
                startIndex += num2;
            }
            if ((startIndex < 0) || (startIndex >= length))
            {
                return null;
            }
            startIndex += num2;
            while ((startIndex < length) && char.IsWhiteSpace(headerValue[startIndex]))
            {
                startIndex++;
            }
            if ((startIndex >= length) || (headerValue[startIndex] != '='))
            {
                return null;
            }
            startIndex++;
            while ((startIndex < length) && char.IsWhiteSpace(headerValue[startIndex]))
            {
                startIndex++;
            }
            if (startIndex >= length)
            {
                return null;
            }
            if ((startIndex < length) && (headerValue[startIndex] == '"'))
            {
                if (startIndex == (length - 1))
                {
                    return null;
                }
                index = headerValue.IndexOf('"', startIndex + 1);
                if ((index < 0) || (index == (startIndex + 1)))
                {
                    return null;
                }
                return headerValue.Substring(startIndex + 1, (index - startIndex) - 1).Trim();
            }
            index = startIndex;
            while (index < length)
            {
                if ((headerValue[index] == ' ') || (headerValue[index] == ','))
                {
                    break;
                }
                index++;
            }
            if (index == startIndex)
            {
                return null;
            }
            return headerValue.Substring(startIndex, index - startIndex).Trim();
        }

        #endregion Private Members
    }
}