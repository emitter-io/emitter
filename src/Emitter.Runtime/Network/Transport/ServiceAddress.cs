// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Globalization;
using System.Net;
using System.Net.NetworkInformation;

namespace Emitter.Network
{
    public class ServiceAddress
    {
        /// <summary>
        /// Gets the host part of the address.
        /// </summary>
        public string Host { get; private set; }

        /// <summary>
        /// Gets the path base of the address.
        /// </summary>
        public string PathBase { get; private set; }

        /// <summary>
        /// Gets the port of the address.
        /// </summary>
        public int Port { get; private set; }

        /// <summary>
        /// Gets the scheme of the address.
        /// </summary>
        public string Scheme { get; private set; }

        /// <summary>
        /// Returns an <see cref="IPEndPoint"/> for this address.
        /// </summary>
        public IPEndPoint EndPoint
        {
            get
            {
                // TODO: IPv6 support
                IPAddress ip;
                if (!IPAddress.TryParse(this.Host, out ip))
                {
                    if (string.Equals(this.Host, "localhost", StringComparison.OrdinalIgnoreCase))
                    {
                        ip = IPAddress.Loopback;
                    }
                    else
                    {
                        ip = IPAddress.IPv6Any;
                    }
                }

                return new IPEndPoint(ip, this.Port);
            }
        }

        /// <summary>
        /// Gets whether the address is a unix pipe.
        /// </summary>
        public bool IsUnixPipe
        {
            get
            {
                return Host.StartsWith(Constants.UnixPipeHostPrefix);
            }
        }

        /// <summary>
        /// Gets the the unix pipe path.
        /// </summary>
        public string UnixPipePath
        {
            get
            {
                Debug.Assert(IsUnixPipe);

                return Host.Substring(Constants.UnixPipeHostPrefix.Length - 1);
            }
        }

        /// <summary>
        /// Converts to a string representation.
        /// </summary>
        public override string ToString()
        {
            return /*Scheme.ToLowerInvariant() + "://" + */Host.ToLowerInvariant() + ":" + Port.ToString(CultureInfo.InvariantCulture) + PathBase.ToLowerInvariant();
        }

        /// <summary>
        /// Gets the hash code.
        /// </summary>
        public override int GetHashCode()
        {
            return ToString().GetHashCode();
        }

        /// <summary>
        /// Checks for equality/
        /// </summary>
        public override bool Equals(object obj)
        {
            var other = obj as ServiceAddress;
            if (other == null)
            {
                return false;
            }
            return string.Equals(Scheme, other.Scheme, StringComparison.OrdinalIgnoreCase)
                && string.Equals(Host, other.Host, StringComparison.OrdinalIgnoreCase)
                && Port == other.Port
                && string.Equals(PathBase, other.PathBase, StringComparison.OrdinalIgnoreCase);
        }

        /// <summary>
        /// Gets the list of addresses for a binding.
        /// </summary>
        /// <param name="binding">The binding to retrieve the addresses from.</param>
        public static IEnumerable<ServiceAddress> FromBinding(IBinding binding)
        {
            var addresses = new List<ServiceAddress>();
            try
            {
                var ipep = binding.EndPoint as IPEndPoint;
                if (ipep.Address.Equals(IPAddress.Any))
                {
                    var adapters = NetworkInterface.GetAllNetworkInterfaces();
                    foreach (NetworkInterface adapter in adapters)
                    {
                        var properties = adapter.GetIPProperties();
                        foreach (IPAddressInformation unicast in properties.UnicastAddresses)
                        {
                            // Only IPv4 for now, as Libuv gives us an error on ubuntu
                            if (unicast.Address.AddressFamily == System.Net.Sockets.AddressFamily.InterNetwork)
                                addresses.Add(FromUrl(String.Format("{0}://{1}:{2}", binding.Schema, unicast.Address, ipep.Port)));
                        }
                    }
                }
                else
                {
                    addresses.Add(FromUrl(String.Format("{0}://{1}:{2}", binding.Schema, ipep.Address, ipep.Port)));
                }
            }
            catch (Exception ex)
            {
                // Something bad happened during the bind
                throw ex;
            }
            return addresses;
        }

        /// <summary>
        /// Creates a <see cref="ServiceAddress"/> from a url.
        /// </summary>
        /// <param name="url">The url to create from.</param>
        /// <returns></returns>
        public static ServiceAddress FromUrl(string url)
        {
            url = url ?? string.Empty;

            int schemeDelimiterStart = url.IndexOf("://", StringComparison.Ordinal);
            if (schemeDelimiterStart < 0)
            {
                int port;
                if (int.TryParse(url, NumberStyles.None, CultureInfo.InvariantCulture, out port))
                {
                    return new ServiceAddress()
                    {
                        Scheme = "http",
                        Host = "+",
                        Port = port,
                        PathBase = "/"
                    };
                }
                return null;
            }
            int schemeDelimiterEnd = schemeDelimiterStart + "://".Length;

            var isUnixPipe = url.IndexOf(Constants.UnixPipeHostPrefix, schemeDelimiterEnd, StringComparison.Ordinal) == schemeDelimiterEnd;

            int pathDelimiterStart;
            int pathDelimiterEnd;
            if (!isUnixPipe)
            {
                pathDelimiterStart = url.IndexOf("/", schemeDelimiterEnd, StringComparison.Ordinal);
                pathDelimiterEnd = pathDelimiterStart;
            }
            else
            {
                pathDelimiterStart = url.IndexOf(":", schemeDelimiterEnd + Constants.UnixPipeHostPrefix.Length, StringComparison.Ordinal);
                pathDelimiterEnd = pathDelimiterStart + ":".Length;
            }

            if (pathDelimiterStart < 0)
            {
                pathDelimiterStart = pathDelimiterEnd = url.Length;
            }

            var serverAddress = new ServiceAddress();
            serverAddress.Scheme = url.Substring(0, schemeDelimiterStart);

            var hasSpecifiedPort = false;
            if (!isUnixPipe)
            {
                int portDelimiterStart = url.LastIndexOf(":", pathDelimiterStart - 1, pathDelimiterStart - schemeDelimiterEnd, StringComparison.Ordinal);
                if (portDelimiterStart >= 0)
                {
                    int portDelimiterEnd = portDelimiterStart + ":".Length;

                    string portString = url.Substring(portDelimiterEnd, pathDelimiterStart - portDelimiterEnd);
                    int portNumber;
                    if (int.TryParse(portString, NumberStyles.Integer, CultureInfo.InvariantCulture, out portNumber))
                    {
                        hasSpecifiedPort = true;
                        serverAddress.Host = url.Substring(schemeDelimiterEnd, portDelimiterStart - schemeDelimiterEnd);
                        serverAddress.Port = portNumber;
                    }
                }

                if (!hasSpecifiedPort)
                {
                    if (string.Equals(serverAddress.Scheme, "http", StringComparison.OrdinalIgnoreCase))
                    {
                        serverAddress.Port = 80;
                    }
                    else if (string.Equals(serverAddress.Scheme, "https", StringComparison.OrdinalIgnoreCase))
                    {
                        serverAddress.Port = 443;
                    }
                }
            }

            if (!hasSpecifiedPort)
            {
                serverAddress.Host = url.Substring(schemeDelimiterEnd, pathDelimiterStart - schemeDelimiterEnd);
            }

            // Path should not end with a / since it will be used as PathBase later
            if (url[url.Length - 1] == '/')
            {
                serverAddress.PathBase = url.Substring(pathDelimiterEnd, url.Length - pathDelimiterEnd - 1);
            }
            else
            {
                serverAddress.PathBase = url.Substring(pathDelimiterEnd);
            }

            //serverAddress.PathBase = PathNormalizer.NormalizeToNFC(serverAddress.PathBase);

            return serverAddress;
        }

        /// <summary>
        /// Creates a <see cref="ServiceAddress"/> from an <see cref="IPEndPoint"/>.
        /// </summary>
        /// <param name="endpoint">The endpoint to create from.</param>
        /// <returns></returns>
        public static ServiceAddress FromIPEndPoint(IPEndPoint endpoint)
        {
            return FromUrl(String.Format("tcp://{0}:{1}", endpoint.Address, endpoint.Port));
        }
    }
}