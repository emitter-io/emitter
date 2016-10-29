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

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a context for HTTP Request/Response pair.
    /// </summary>
    public unsafe class HttpContext : RecyclableObject
    {
        #region Current Context

        [ThreadStatic]
        private static HttpContext CurrentContext;

        /// <summary>
        /// Gets the current client bound to the thread.
        /// </summary>
        public static HttpContext Current
        {
            get { return CurrentContext; }
            internal set { CurrentContext = value; }
        }

        #endregion Current Context

        private const int DefaultScriptTimeout = 60;

        internal HttpContext()
        {
            fScriptTimeout = TimeSpan.FromSeconds(DefaultScriptTimeout);

            // Just new ones, since the HttpContext is pooled already.
            fRequest = new HttpRequest();
            fResponse = new HttpResponse();
            fSession = new HttpSession();
        }

        private Emitter.Connection fConnection;
        private HttpRequest fRequest;
        private HttpResponse fResponse;
        private HttpSession fSession;

        private TimeSpan fScriptTimeout;
        private int fParseStart;
        private int fParseEnd;
        private ParseState fState;
        private bool fSeenKeepAlive;
        private bool fKeepAlive;
        private TextRange fParseKey;

        /// <summary>
        /// Gets or sets the handle of the context object for pooling.
        /// </summary>
        internal int Handle = -1;

        /// <summary>
        /// Gets or sets whether keep-alive is disabled on this context or not
        /// </summary>
        public bool DisableKeepAlive { get; set; }

        /// <summary>
        /// Gets whether Connection: Keep-Alive was specified
        /// </summary>
        public bool KeepAlive { get { return fKeepAlive; } }

        /// <summary>
        /// Gets or sets the maximum amount of time a script can run before it is terminated
        /// </summary>
        public TimeSpan ScriptTimeout
        {
            get { return fScriptTimeout; }
            set
            {
                fScriptTimeout = value;
                if (fConnection != null)
                    fConnection.Timeout = fScriptTimeout;
            }
        }

        /// <summary>
        /// Gets the underlying connection
        /// </summary>
        public Connection Connection
        {
            get { return fConnection; }
            internal set
            {
                // Switch to the new client
                fConnection = value;
            }
        }

        /// <summary>
        /// Gets <see cref="HttpRequest"/> object for the current request.
        /// </summary>
        public HttpRequest Request
        {
            get { return fRequest; }
        }

        /// <summary>
        /// Gets <see cref="HttpResponse"/> object for the current request.
        /// </summary>
        public HttpResponse Response
        {
            get { return fResponse; }
        }

        /// <summary>
        /// Gets <see cref="HttpSession"/> object for the current request.
        /// </summary>
        public HttpSession Session
        {
            get { return fSession; }
            internal set { fSession = value; }
        }

        /// <summary>
        /// Resets the parsing state machine for the next request
        /// </summary>
        public override void Recycle()
        {
            fScriptTimeout = TimeSpan.FromSeconds(DefaultScriptTimeout);
            fState = ParseState.ReadingVerb;
            fParseStart = 0;
            fParseEnd = 0;
            fSeenKeepAlive = false;

            fRequest.Recycle();
            fResponse.Recycle();
        }

        /// <summary>
        /// Gets a type of encoding from an <see cref="HttpContext"/>.
        /// </summary>
        /// <returns>Whether the encoding is string or binary.</returns>
        internal TransportEncoding GetFabricEncoding()
        {
            // Get the query
            var query = HttpUtility.ParseQueryString(this.Request.Path);

            // Check if it's specifically base64 or not
            return (query["b64"] == "1")
                ? TransportEncoding.Text
                : TransportEncoding.Binary;
        }

        /// <summary>
        /// Parses the request in order to fill the http context (or alter it if necessary)
        /// </summary>
        internal unsafe int ParseRequest(byte* buffer, int bufferLength, out bool hasError, out bool insufficientData)
        {
            hasError = false;
            insufficientData = true;

            HttpRequest request = fRequest;

            byte* pBlock = buffer;
            if (pBlock == null)
                return 0;

            byte* pStart = pBlock + fParseStart;
            byte* pCurrent = pBlock + fParseEnd;
            byte* pEnd = pBlock + bufferLength;

            while (pCurrent < pEnd)
            {
                int len = (int)(pCurrent - pStart);
                byte ch = *pCurrent;

                switch (this.fState)
                {
                    case ParseState.ReadingVerb:
                        if (ch == ' ')
                        {
                            switch (len)
                            {
                                case 3: // GET or PUT
                                    if (pStart[0] == 'G' &&
                                        pStart[1] == 'E' &&
                                        pStart[2] == 'T')
                                        request.Verb = HttpVerb.Get;
                                    else if (
                                        pStart[0] == 'P' &&
                                        pStart[1] == 'U' &&
                                        pStart[2] == 'T')
                                        request.Verb = HttpVerb.Put;
                                    break;

                                case 4: // HEAD or POST
                                    if (pStart[0] == 'H' &&
                                        pStart[1] == 'E' &&
                                        pStart[2] == 'A' &&
                                        pStart[3] == 'D')
                                        request.Verb = HttpVerb.Head;
                                    else if (
                                        pStart[0] == 'P' &&
                                        pStart[1] == 'O' &&
                                        pStart[2] == 'S' &&
                                        pStart[3] == 'T')
                                        request.Verb = HttpVerb.Post;
                                    break;

                                case 5: // TRACE
                                    if (pStart[0] == 'T' &&
                                        pStart[1] == 'R' &&
                                        pStart[2] == 'A' &&
                                        pStart[3] == 'C' &&
                                        pStart[4] == 'E')
                                        request.Verb = HttpVerb.Trace;
                                    break;

                                case 6: // DELETE
                                    if (pStart[0] == 'D' &&
                                        pStart[1] == 'E' &&
                                        pStart[2] == 'L' &&
                                        pStart[3] == 'E' &&
                                        pStart[4] == 'T' &&
                                        pStart[5] == 'E')
                                        request.Verb = HttpVerb.Delete;
                                    break;

                                case 7: // OPTIONS or CONNECT
                                    if (pStart[0] == 'O' &&
                                        pStart[1] == 'P' &&
                                        pStart[2] == 'T' &&
                                        pStart[3] == 'I' &&
                                        pStart[4] == 'O' &&
                                        pStart[5] == 'N' &&
                                        pStart[6] == 'S')
                                        request.Verb = HttpVerb.Options;
                                    else if (
                                        pStart[0] == 'C' &&
                                        pStart[1] == 'O' &&
                                        pStart[2] == 'N' &&
                                        pStart[3] == 'N' &&
                                        pStart[4] == 'E' &&
                                        pStart[5] == 'C' &&
                                        pStart[6] == 'T')
                                        request.Verb = HttpVerb.Connect;
                                    break;
                            }

                            pStart = pCurrent + 1;
                            this.fParseStart = (int)(pStart - pBlock);
                            this.fState = ParseState.ReadingPath;
                        }

                        break;

                    case ParseState.ReadingPath:
                        if (ch == ' ')
                        {
                            request.URI = new TextRange(pStart, (int)(pCurrent - pStart));
                            pStart = pCurrent + 1;
                            this.fParseStart = (int)(pStart - pBlock);
                            this.fState = ParseState.ReadingHTTP;
                        }

                        break;

                    case ParseState.ReadingHTTP:
                        if (ch == '/')
                        {
                            // we don't care about actually generating the string
                            pStart = pCurrent + 1;
                            this.fParseStart = (int)(pStart - pBlock);
                            this.fState = ParseState.ReadingHTTPVersion;
                        }

                        break;

                    case ParseState.ReadingHTTPVersion:
                        if (ch == 13 ||
                            ch == 10)
                        {
                            // optimize for known cases to avoid costly string alloc
                            // for the sake of a string compare
                            if (len == 3 &&
                                pStart[0] == '1' &&
                                pStart[1] == '.' &&
                                pStart[2] == '1')
                            {
                                request.Version = HttpVersion.V1_1;

                                if (this.DisableKeepAlive)
                                    this.fKeepAlive = false;
                                else
                                    this.fKeepAlive = true;
                            }
                            else
                            {
                                request.Version = HttpVersion.V1_0;
                                this.fKeepAlive = false;
                            }

                            // deal with clients sending either
                            // CRLF or just LF as per HTTP spec
                            if (ch == 13 && pCurrent[1] == 10)
                                pCurrent++;

                            pStart = pCurrent + 1;
                            this.fParseStart = (int)(pStart - pBlock);
                            this.fState = ParseState.ReadingKey;
                        }

                        break;

                    case ParseState.ReadingKey:
                        if (pCurrent == pStart)
                        {
                            // check for CRLF at beginning of line
                            // signifying the end of the HTTP header
                            if (ch == 13 ||
                                ch == 10)
                            {
                                // end of HTTP header detected
                                insufficientData = false;
                                //this.fRequestCompleted = true;
                                this.fParseStart = (int)(pStart - pBlock);
                                this.fParseEnd = (int)(pCurrent - pBlock);

                                return fParseEnd + 2;
                            }
                        }

                        if (ch == ':')
                        {
                            fParseKey = new TextRange(pStart, (int)(pCurrent - pStart));
                            pStart = pCurrent + 1;
                            this.fParseStart = (int)(pStart - pBlock);
                            this.fState = ParseState.EatWhite;
                        }

                        break;

                    case ParseState.EatWhite:
                        if (ch == ' ' ||
                            ch == '\t')
                            pStart++;
                        else
                        {
                            this.fState = ParseState.ReadingValue;
                            continue;
                        }

                        break;

                    case ParseState.ReadingValue:
                        if (ch == 13 ||
                            ch == 10)
                        {
                            request.Keys[request.HeaderCount] = fParseKey;
                            request.Values[request.HeaderCount] = new TextRange(pStart, len);

                            if (!fSeenKeepAlive)
                            {
                                if (fParseKey.Is("CONNECTION"))
                                {
                                    fSeenKeepAlive = true;

                                    if (!this.DisableKeepAlive)
                                    {
                                        if (request.Version == HttpVersion.V1_0)
                                        {
                                            if (request.Values[request.HeaderCount].Contains("KEEP-ALIVE"))
                                                this.fKeepAlive = true;
                                        }
                                        else
                                        {
                                            if (request.Values[request.HeaderCount].Contains("CLOSE"))
                                                this.fKeepAlive = false;
                                        }
                                    }
                                }
                            }

                            request.HeaderCount++;

                            // deal with clients sending either
                            // CRLF or just LF as per HTTP spec
                            if (ch == 13 && pCurrent[1] == 10)
                                pCurrent++;

                            pStart = pCurrent + 1;
                            this.fParseStart = (int)(pStart - pBlock);
                            this.fState = ParseState.ReadingKey;
                        }

                        break;
                }

                pCurrent++;
            }

            // Set indices
            this.fParseStart = (int)(pStart - pBlock);
            this.fParseEnd = (int)(pCurrent - pBlock);

            // Return the amount of bytes parsed
            return fParseEnd;
        }

        private enum ParseState
        {
            ReadingVerb,
            ReadingPath,
            ReadingHTTP,
            ReadingHTTPVersion,
            ReadingKey,
            EatWhite,
            ReadingValue,
        }
    }
}