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

namespace Emitter.Network.Http
{
    /// <summary>
    /// Defines the contract that Emitter implements to synchronously process HTTP
    /// Web requests using custom HTTP handlers.
    /// </summary>
    public interface IHttpHandler
    {
        /// <summary>
        /// Enables processing of HTTP Web requests by a custom HttpHandler that implements
        /// the IHttpHandler interface.
        /// </summary>
        /// <param name="context">An HttpContext object that provides references to the intrinsic
        /// server objects (for example, Request, Response, Session, and Server) used
        /// to service HTTP requests.</param>
        void ProcessRequest(HttpContext context);

        /// <summary>
        /// Checks whether the handler can handle an incoming request or not
        /// </summary>
        /// <param name="context">An HttpContext object that provides references to the intrinsic
        /// server objects (for example, Request, Response, Session, and Server) used
        /// to service HTTP requests.</param>
        /// <param name="verb">Verb of the request</param>
        /// <param name="url">Url passed in parameter</param>
        /// <returns></returns>
        bool CanHandle(HttpContext context, HttpVerb verb, string url);
    }
}