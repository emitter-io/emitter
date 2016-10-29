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
    /// Represents a HTTP cross-origin request sharing (CORS) prehandler.
    /// </summary>
    internal static class CorsPolicy
    {
        /// <summary>
        /// Applies CORS if necessary for a particular request/response pair.
        /// </summary>
        /// <param name="context">The context to apply to.</param>
        /// <returns>The type of the request.</returns>
        public static CorsType OnRequest(HttpContext context)
        {
            var request = context.Request;
            var response = context.Response;

            // Only prehandles OPTIONS requests
            if (request.Verb != HttpVerb.Options)
                return CorsType.Invalid;

            // The CORS should be enabled for this to work
            if (!Service.Cors.Enabled)
                return CorsType.Invalid;

            // Apply the policy we've defined
            return Service.Cors.ApplyPolicy(request, response);
        }
    }
}