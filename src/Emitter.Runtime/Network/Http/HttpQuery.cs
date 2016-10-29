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
    /// Represents a fault-tolerant HTTP endpoint service. The focus here is on
    /// fault tolerance and not specifically performance.
    /// </summary>
    public class HttpQuery
    {
        #region Constructor

        /// <summary>
        /// The list of hostnames to resolve periodically.
        /// </summary>
        private string Hostname;

        /// <summary>
        /// Creates a new <see cref="HttpQuery"/> with the hostnames provided.
        /// </summary>
        /// <param name="hostname">The hostnames to use for IP resolution.</param>
        public HttpQuery(string hostname)
        {
            this.Hostname = hostname;
        }

        #endregion Constructor

        /// <summary>
        /// Issues a GET http request and returns the answer string.
        /// </summary>
        /// <param name="url">The url of the http request.</param>
        /// <param name="timeout">Timeout for operation.</param>
        /// <returns>The parsed string.</returns>
        public virtual string Get(string url, int timeout)
        {
            // HTTP GET
            var response = HttpUtility.Get(this.Hostname + url, timeout);
            if (response.Failure || !response.HasValue)
            {
                Service.Logger.Log(response.Error);
            }

            // If we've failed again, return null
            if (response.Failure)
                return null;

            // Return the response
            return response.Value;
        }
    }
}