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
using Emitter.Collections;

namespace Emitter.Network.Http
{
    internal sealed class HttpContextPool : ConcurrentPool<HttpContext>
    {
        /// <summary>
        /// Default constructor for the HttpContextPool
        /// </summary>
        public HttpContextPool() : base("HttpContexts", _ => new HttpContext()) { }

        /// <summary>
        /// A poof of HttpContexts used for requests data-containment
        /// </summary>
        public static readonly HttpContextPool Default = new HttpContextPool();

        /// <summary>
        /// Acquires an instance of HttpContext and binds it to a connection.
        /// </summary>
        /// <param name="connection">The connection for this HttpContext.</param>
        /// <returns>The acquired instance.</returns>
        public HttpContext Acquire(Emitter.Connection connection)
        {
            HttpContext context = base.Acquire();
            context.Connection = connection;
            return context;
        }
    }
}