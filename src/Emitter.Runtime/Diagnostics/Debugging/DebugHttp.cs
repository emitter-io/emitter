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
using Emitter.Network.Http;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a handle which can be used for inspecting.
    /// </summary>
    public sealed class DebugHttp : IHttpHandler
    {
        public bool CanHandle(HttpContext context, HttpVerb verb, string url)
        {
            return verb == HttpVerb.Get && url.StartsWith("/debug");
        }

        public void ProcessRequest(HttpContext context)
        {
            var path = context.Request.Path;
            if (path.EndsWith("/debug"))
            {
                context.Response.Write(this.Content.Value);
                return;
            }

            var id = context.Request.Path.Replace("/debug/", "");
            Debugger.Default.Inspect(id);
            context.Response.Write("OK");
        }

        /// <summary>
        /// Load the content lazily.
        /// </summary>
        private Lazy<byte[]> Content = new Lazy<byte[]>(() =>
        {
            var response = HttpUtility.Get("http://s3-eu-west-1.amazonaws.com/cdn.emitter.io/web/debug.html", 30000);
            if (response.Success && response.HasValue)
                return response.Value.AsUTF8();
            return ArrayUtils<byte>.Empty;
        });
    }
}