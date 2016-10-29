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
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a decoder of various HTTP requests.
    /// </summary>
    public static unsafe class HttpHandler
    {
        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Connection channel, ProcessingContext context)
        {
            // Only handle stuff with HTTP Context
            var httpContext = context.GetSession<HttpContext>();
            if (httpContext == null)
                return ProcessingState.Failure;

            try
            {
                // Get request
                var request = httpContext.Request;

                // We should apply cross-origin resource sharing first
                if (CorsPolicy.OnRequest(httpContext) == CorsType.Preflight)
                {
                    // Send pending preflight response. It should be empty.
                    context.Channel.Send(HttpHeaders.Create(httpContext.Response));
                    return ProcessingState.Stop;
                }

                // Try to handle the incoming request
                var handler = Service.Http.GetHandler(httpContext, request.Verb, request.Path);
                if (handler == null)
                {
                    // Send the error as HTTP 404
                    HttpResponse.Send404((Emitter.Connection)context.Channel);
                    return ProcessingState.Stop;
                }

                // Trace HTTP request
                NetTrace.WriteLine(request.HttpVerb.ToString().ToUpper() + " " + request.Path + " handled by " + handler.GetType().ToString(), channel, NetTraceCategory.Http);

                // Check if session id is present in the cookie. If there is no session id
                // we create a new session.
                HttpSession.OnRequest(httpContext);

                // Check the security and process the request
                if (handler is ISecureHttpHandler)
                {
                    ISecureHttpHandler secure = handler as ISecureHttpHandler;
                    if (secure.Security == null || secure.Security.Authorize(httpContext))
                    {
                        handler.ProcessRequest(httpContext);
                        context.Channel.Send(httpContext.Response); // Send pending
                    }
                }
                else
                {
                    handler.ProcessRequest(httpContext);
                    context.Channel.Send(httpContext.Response); // Send pending
                }

                // Keep-Alive or close?
                if (!httpContext.Response.KeepAlive)
                    channel.Close();

                return ProcessingState.Stop;
            }
            catch (NonAuthorizedException ex)
            {
                // Send the error as HTTP 500
                HttpResponse.Send401((Emitter.Connection)context.Channel, ex);
                return ProcessingState.Stop;
            }
            catch (HttpRedirectException)
            {
                // Send the response as it contains the redirect headers
                context.Channel.Send(httpContext.Response);
                return ProcessingState.Stop;
            }
            catch (Exception ex)
            {
                // Send the error as HTTP 500
                HttpResponse.Send500((Emitter.Connection)context.Channel, ex);
                return ProcessingState.Stop;
            }
            finally
            {
                // Release the HTTP context back to the pool
                httpContext.TryRelease();
            }
        }
    }

    /// <summary>
    /// Represents a non-authorized exception for an HTTP request.
    /// </summary>
    public class NonAuthorizedException : Exception
    {
        /// <summary>
        /// Constructs an instance of a <see cref="NonAuthorizedException"/>.
        /// </summary>
        public NonAuthorizedException() : this("This server could not verify that you are authorized to access the document requested. Either you supplied the wrong credentials (e.g., bad password), or your browser doesn't understand how to supply the credentials required.") { }

        /// <summary>
        /// Constructs an instance of a <see cref="NonAuthorizedException"/>.
        /// </summary>
        /// <param name="message">The default message.</param>
        public NonAuthorizedException(string message) : base(message) { }
    }

    /// <summary>
    /// Represents an exception redirecting an HTTP request.
    /// </summary>
    public class HttpRedirectException : Exception
    {
        /// <summary>
        /// Constructs an instance of a <see cref="HttpRedirectException"/>.
        /// </summary>
        public HttpRedirectException() : base("The request have been redirected.") { }
    }

    /// <summary>
    /// Represents a forbidden exception for an HTTP request.
    /// </summary>
    public class HttpForbiddenException : Exception
    {
        /// <summary>
        /// Constructs an instance of a <see cref="HttpForbiddenException"/>.
        /// </summary>
        public HttpForbiddenException() : this("The server declined to show the web page.") { }

        /// <summary>
        /// Constructs an instance of a <see cref="HttpForbiddenException"/>.
        /// </summary>
        /// <param name="message">The default message.</param>
        public HttpForbiddenException(string message) : base(message) { }
    }
}