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
using Emitter.Providers;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a handler that handles /health endpoint.
    /// </summary>
    public sealed class HttpHealthHandler : IHttpHandler
    {
        /// <summary>
        /// Checks whether the handler can handle an incoming request or not
        /// </summary>
        /// <param name="context">An HttpContext object that provides references to the intrinsic
        /// server objects (for example, Request, Response, Session, and Server) used
        /// to service HTTP requests.</param>
        /// <param name="verb">Verb of the request</param>
        /// <param name="url">Url passed in parameter</param>
        /// <returns></returns>
        public bool CanHandle(HttpContext context, HttpVerb verb, string url)
        {
            return verb == HttpVerb.Get && url.EndsWith("/health");
        }

        /// <summary>
        /// Enables processing of HTTP Web requests by a custom HttpHandler that implements
        /// the IHttpHandler interface.
        /// </summary>
        /// <param name="context">An HttpContext object that provides references to the intrinsic
        /// server objects (for example, Request, Response, Session, and Server) used
        /// to service HTTP requests.</param>
        public void ProcessRequest(HttpContext context)
        {
            try
            {
                // If we have a provider, return whatever the provider tells us
                var provider = Service.Providers.Resolve<HealthProvider>();
                if (provider == null || provider.IsHealthy())
                {
                    context.Response.Status = "200";
                    context.Response.Write("OK");
                    return;
                }

                // Default implementation, so we don't need to add a provider
                context.Response.Status = "503";
                context.Response.Write("Service Unavailable");
            }
            catch (Exception ex)
            {
                context.Response.Status = "500";
                Service.Logger.Log(ex);
            }
        }
    }
}