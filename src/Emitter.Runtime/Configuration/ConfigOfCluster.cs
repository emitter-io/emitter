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
using Emitter.Providers;
using Emitter.Text.Json;

namespace Emitter.Configuration
{
    /// <summary>
    /// Represents a configuration of the cluster.
    /// </summary>
    public class ConfigOfCluster
    {
        /// <summary>
        /// The address to broadcast
        /// </summary>
        [JsonProperty("broadcast")]
        public string BroadcastAddress = "public";

        /// <summary>
        /// Gets or sets the mesh port.
        /// </summary>
        [JsonProperty("port")]
        public int Port = 4000;

        /// <summary>
        /// Gets or sets the seed hostname address.
        /// </summary>
        [JsonProperty("seed")]
        public string Seed = "127.0.0.1:4000";

        /// <summary>
        /// Gets or sets the cluster key.
        /// </summary>
        [JsonProperty("key")]
        public string ClusterKey = "emitter-io";

        /// <summary>
        /// Resolve the seed endpoints.
        /// </summary>
        [JsonIgnore]
        public IPEndPoint SeedEndpoint
        {
            get
            {
                // Parse as URI
                Uri uri;
                if (!Uri.TryCreate("tcp://" + Seed, UriKind.Absolute, out uri))
                    return null;

                // Resolve DNS
                var task = Dns.GetHostAddressesAsync(uri.DnsSafeHost);
                task.Wait();

                // Get the port
                var port = uri.IsDefaultPort ? Port : uri.Port;
                var addr = uri.HostNameType == UriHostNameType.Dns
                    ? task.Result.LastOrDefault()
                    : IPAddress.Parse(uri.Host);

                return new IPEndPoint(addr, port);
            }
        }

        /// <summary>
        /// Gets the endpoint to broadcast within the mesh.
        /// </summary>
        [JsonIgnore]
        public IPEndPoint BroadcastEndpoint
        {
            get
            {
                IPAddress address = null;
                var addressString = this.BroadcastAddress.ToLower();
                if (addressString == "public")
                    address = Service.Providers.Resolve<AddressProvider>().GetExternal();
                if (addressString == "local")
                    address = IPAddress.Parse("127.0.0.1");

                if (address == null && !IPAddress.TryParse(addressString, out address))
                {
                    Service.Logger.Log(LogLevel.Error, "Unable to parse " + addressString + " as a valid IP Address. Using 127.0.0.1 instead.");
                    address = IPAddress.Parse("127.0.0.1");
                }

                return new IPEndPoint(address, Port);
            }
        }
    }
}