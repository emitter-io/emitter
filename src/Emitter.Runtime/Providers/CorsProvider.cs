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
using Emitter.Network.Http;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider that provides a cross-origin resource sharing policy.
    /// </summary>
    public abstract class CorsProvider : Provider
    {
        /// <summary>
        /// Gets or sets whether cross-origin resource sharing policy should be enabled and
        /// whether all HTTP request should have the policy applied automatically.
        /// </summary>
        public bool Enabled
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the the Access-Control-Allow-Methods header which indicates, as part of
        /// the response to a preflight request, which methods can be used during the actual request.
        /// </summary>
        public string AllowMethods
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the Access-Control-Allow-Headers header which indicates, as part of the
        /// response to a preflight request, which header field names can be used during the actual
        /// request.
        /// </summary>
        public string AllowHeaders
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the Access-Control-Allow-Origin header which indicates whether a resource
        /// can be shared based by returning the value of the Origin request header, "*", or "null"
        /// in the response.
        /// </summary>
        public string AllowOrigin
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the Access-Control-Expose-Headers header which indicates which headers are
        /// safe to expose to the API of a CORS API specification.
        /// </summary>
        public string ExposeHeaders
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the Access-Control-Allow-Credentials header which indicates whether the
        /// response to request can be exposed when the omit credentials flag is unset. When part of
        /// the response to a preflight request it indicates that the actual request can include user
        /// credentials.
        /// </summary>
        public bool AllowCredentials
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the Access-Control-Max-Age header which indicates how long the results of a
        /// preflight request can be cached in a preflight result cache.
        /// </summary>
        public int MaxAge
        {
            get;
            set;
        }

        /// <summary>
        /// Gets the type of cross-origin resource sharing request.
        /// </summary>
        /// <param name="request">The request to check.</param>
        /// <returns>Whether the request is a preflight request, actual or invalid.</returns>
        public CorsType GetType(HttpRequest request)
        {
            // The request must have the origin
            if (String.IsNullOrWhiteSpace(request.Origin))
                return CorsType.Invalid;

            // The preflight request must be OPTIONS
            if (request.Verb == HttpVerb.Options)
            {
                // The request must have a method that it requests
                string method;
                if (!request.TryGetHeader("Access-Control-Request-Method", out method))
                    return CorsType.Request;

                // TODO: In future, we should add validation for method & header
                // (cf: http://www.html5rocks.com/static/images/cors_server_flowchart.png)

                // We have a valid CORS preflight
                return CorsType.Preflight;
            }
            else
            {
                // We have a normal CORS request
                return CorsType.Request;
            }
        }

        /// <summary>
        /// Applies the default cross-origin sharing policy policy to the response.
        /// </summary>
        /// <param name="request">The request for the resource.</param>
        /// <param name="response">The response with the applied.</param>
        /// <returns>
        /// If the CORS type is anything but invalid, the response will contain the applied
        /// headers. If a preflight message have been sent, this should not proceed further as
        /// the message must be sent to the client without a body.
        /// </returns>
        public abstract CorsType ApplyPolicy(HttpRequest request, HttpResponse response);
    }

    /// <summary>
    /// Represents a provider that provides the default cross-origin resource sharing policy.
    /// </summary>
    public sealed class DefaultCorsProvider : CorsProvider
    {
        /// <summary>
        /// This is a default options for the cors profider.
        /// </summary>
        public DefaultCorsProvider()
        {
            this.Enabled = true;
            this.AllowMethods = "POST,GET,PUT,DELETE,OPTIONS";
            this.AllowHeaders = "X-Requested-With,Content-Type";
            this.AllowOrigin = "*";
            this.AllowCredentials = true;
            this.MaxAge = -1;
            this.ExposeHeaders = null;
        }

        /// <summary>
        /// Applies the default cross-origin sharing policy policy to the response.
        /// </summary>
        /// <param name="request">The request for the resource.</param>
        /// <param name="response">The response with the applied.</param>
        /// <returns>
        /// If the CORS type is anything but invalid, the response will contain the applied
        /// headers. If a preflight message have been sent, this should not proceed further as
        /// the message must be sent to the client without a body.
        /// </returns>
        public override CorsType ApplyPolicy(HttpRequest request, HttpResponse response)
        {
            // First we get the type of the request
            var type = GetType(request);
            if (type == CorsType.Invalid)
                return CorsType.Invalid;

            if (type == CorsType.Request)
            {
                // This is a normal request
                if (this.ExposeHeaders != null)
                    response.Headers.Set("Access-Control-Expose-Headers", this.ExposeHeaders);
            }
            else
            {
                // This is a preflight request
                response.Headers.Set("Access-Control-Allow-Methods", this.AllowMethods);
                response.Headers.Set("Access-Control-Allow-Headers", this.AllowHeaders);

                if (this.MaxAge != -1)
                    response.Headers.Set("Access-Control-Max-Age", this.MaxAge.ToString());

                response.Status = "200";
            }

            // Remove if we have origin
            response.HeaderCache.Remove("Access-Control-Allow-Origin");

            // Set the origin, if we specified everything it is safer to set the same
            response.SetHeader("Access-Control-Allow-Origin", this.AllowOrigin == "*"
                ? (request.Origin == null ? "*" : request.Origin)
                : this.AllowOrigin
                );

            // If we accept cookies say so
            response.SetHeader("Access-Control-Allow-Credentials", this.AllowCredentials
                ? "true"
                : "false"
                );

            // The type we've processed
            return type;
        }
    }
}