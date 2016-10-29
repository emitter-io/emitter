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
using System.Net;
using Emitter.Network.Http;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider of various IP address information.
    /// </summary>
    public abstract class AddressProvider : Provider
    {
        /// <summary>
        /// Gets the external ip address through various means.
        /// </summary>
        /// <returns>The external IP address.</returns>
        public abstract IPAddress GetExternal();

        /// <summary>
        /// Gets the external ip address from HTTP.
        /// </summary>
        /// <param name="url">The HTTP provider to query.</param>
        /// <param name="timeout">The timeout for the query, in milliseconds.</param>
        /// <returns>The external IP address.</returns>
        public abstract IPAddress GetExternalFrom(string url, int timeout);

        /// <summary>
        /// Gets the fingerprint address, which uniquely identifies this machine.
        /// </summary>
        /// <returns>The hardware fingerprint.</returns>
        public abstract string GetFingerprint();
    }

    /// <summary>
    /// Represents a provider of health status.
    /// </summary>
    public class DefaultAddressProvider : AddressProvider
    {
        private IPAddress CachedAddress = IPAddress.None;

        private readonly static string[] ExternalProviders = new string[]{
            "http://ipv4.icanhazip.com",
            "http://www.trackip.net/ip",
            "http://automation.whatismyip.com/n09230945.asp",
            "http://api.ipify.org/"
        };

        /// <summary>
        /// Gets the external ip address through various means.
        /// </summary>
        /// <returns>The external IP address.</returns>
        public override IPAddress GetExternal()
        {
            if (this.CachedAddress != IPAddress.None)
                return this.CachedAddress;

            foreach (var url in ExternalProviders)
            {
                // Return the first found
                var addr = GetExternalFrom(url, 30000);
                if (addr != IPAddress.None)
                {
                    this.CachedAddress = addr;
                    return addr;
                }
            }

            // Didn't find any
            return IPAddress.None;
        }

        /// <summary>
        /// Gets the external ip address from HTTP.
        /// </summary>
        /// <param name="url">The HTTP provider to query.</param>
        /// <param name="timeout">The timeout for the query, in milliseconds.</param>
        /// <returns>The external IP address.</returns>
        public override IPAddress GetExternalFrom(string url, int timeout)
        {
            try
            {
                var response = HttpUtility.Get(url, timeout);
                if (response.Failure)
                    return IPAddress.None;

                if (response.HasValue)
                {
                    var value = response.Value;
                    value = value.Replace("\n", String.Empty);
                    value = value.Trim();

                    IPAddress address;
                    if (IPAddress.TryParse(value, out address))
                        return address;
                    return IPAddress.None;
                }
                return IPAddress.None;
            }
            catch (Exception)
            {
                return IPAddress.None;
            }
        }

        /// <summary>
        /// Gets the fingerprint address, which uniquely identifies this machine.
        /// </summary>
        /// <returns>The hardware fingerprint.</returns>
        public override string GetFingerprint()
        {
            return GetExternal()
                .ToString()
                .Replace(".", "-");
        }
    }
}