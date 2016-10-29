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
using System.Net;

namespace Emitter
{
    internal static class IPAddressExtensions
    {
        /// <summary>
        /// Internal IP Address Pool
        /// </summary>
        private static readonly Dictionary<IPAddress, IPAddress> AddressPool
            = new Dictionary<IPAddress, IPAddress>();

        /// <summary>
        /// Ensures that the IP Address is contained in the static internal pool.
        /// </summary>
        internal static IPAddress Intern(this IPAddress ipAddress)
        {
            IPAddress interned;
            if (!AddressPool.TryGetValue(ipAddress, out interned))
            {
                interned = ipAddress;
                AddressPool[ipAddress] = interned;
            }
            return interned;
        }

        /// <summary>
        /// Gets the node identifier from the endpoint.
        /// </summary>
        /// <param name="ep">The endpoint to convert.</param>
        /// <returns></returns>
        internal static int ToIdentifier(this IPEndPoint ep)
        {
            return Murmur32.GetHash(ep.ToString());
        }
    }
}