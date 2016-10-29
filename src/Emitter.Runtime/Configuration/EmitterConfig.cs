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
using System.IO;
using System.Linq;
using System.Net;
using Emitter.Collections;
using Emitter.Providers;
using Emitter.Text.Json;

namespace Emitter.Configuration
{
    /// <summary>
    /// Contals all of the constants we use in this project.
    /// </summary>
    public class EmitterConfig
    {
        #region Static Members

        /// <summary>
        /// Gets or sets the default configuration.
        /// </summary>
        public static EmitterConfig Default
        {
            get;
            private set;
        }

        /// <summary>
        /// Loads the configuration by loading the file and swapping values provided through the environment variables.
        /// </summary>
        /// <param name="args">The arguments to main.</param>
        public static void Initialize(string[] args)
        {
            try
            {
                // Default configuration file
                var configFile = Path.Combine(Service.ConfigDirectory, "emitter.conf");
                //Service.Logger.Log("Configuring: " + configFile);

                // Load any other if needed
                var arguments = new Arguments(args);
                if (arguments["config"] != null)
                    configFile = Path.GetFullPath(arguments["config"]);

                // Do we have a config file?
                if (!File.Exists(configFile))
                {
                    // Create a new file
                    Service.Logger.Log(LogLevel.Warning, "Configuration file 'emitter.conf' does not exist, using default configuration.");
                    Default = new EmitterConfig();

                    // Write to disk
                    File.WriteAllText(
                        configFile,
                        JsonConvert.SerializeObject(Default, Formatting.Indented)
                        );
                }
                else
                {
                    // Deserialize
                    Default = JsonConvert.DeserializeObject<EmitterConfig>(File.ReadAllText(configFile));
                }

                // Check the environment variables for some things such as seed and keys
                Service.MetadataProvider.PopulateFromSecurityProvider("emitter", Default);

                // Now, we might have configured the vault, overwrite the secrets from the vault
                if (Default.Vault.HasVault)
                {
                    var vault = Default.Vault.Address;
                    var vaultApp = Default.Vault.Application;
                    var vaultUser = Service.Providers
                        .Resolve<AddressProvider>()
                        .GetFingerprint();

                    // Configure vault first, so we can fetch secrets from it
                    Service.Providers.Register<SecurityProvider>(new VaultSecurityProvider(vault, vaultApp, vaultUser));

                    // Now reload the configuration from the vault
                    Service.MetadataProvider.PopulateFromSecurityProvider("emitter", Default);
                }
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                Default = new EmitterConfig();
            }
        }

        #endregion Static Members

        /// <summary>
        /// Gets or sets the api port.
        /// </summary>
        [JsonProperty("tcp")]
        public int TcpPort = 8080;

        /// <summary>
        /// Gets or sets the api port.
        /// </summary>
        [JsonProperty("tls")]
        public int TlsPort = 8443;

        /// <summary>
        /// Gets or sets the secret key.
        /// </summary>
        [JsonProperty("license")]
        public string License = "";

        /// <summary>
        /// Gets or sets the cluster config.
        /// </summary>
        [JsonProperty("cluster")]
        public ClusterConfig Cluster = new ClusterConfig();

        /// <summary>
        /// Gets or sets the vault config.
        /// </summary>
        [JsonProperty("vault")]
        public VaultConfig Vault = new VaultConfig();

        /// <summary>
        /// Gets or sets the providers config
        /// </summary>
        [JsonProperty("providers")]
        public ProviderConfig Provider = new ProviderConfig();
    }

    public class ClusterConfig
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

    public class VaultConfig
    {
        /// <summary>
        /// Gets or sets the vault address.
        /// </summary>
        [JsonProperty("address")]
        public string Address = "";

        /// <summary>
        /// Gets or sets the vault app-id.
        /// </summary>
        [JsonProperty("app")]
        public string Application = "";

        /// <summary>
        /// Checks whether the vault is configured
        /// </summary>
        [JsonIgnore]
        public bool HasVault
        {
            get { return !string.IsNullOrWhiteSpace(this.Address) && !string.IsNullOrWhiteSpace(this.Application); }
        }
    }

    public class ProviderConfig
    {
        /// <summary>
        /// Gets or sets the cluster key.
        /// </summary>
        [JsonProperty("contract")]
        public string ContractProviderName = nameof(SingleContractProvider);

        /// <summary>
        /// Gets or sets the logging.
        /// </summary>
        [JsonProperty("logging")]
        public string LoggingProviderName = nameof(MultiTextLoggingProvider);

        /// <summary>
        /// Gets or sets the certificate provider.
        /// </summary>
        [JsonProperty("certificate")]
        public string CertificateProviderName = nameof(FileCertificateProvider);

        /// <summary>
        /// Gets or sets the storage.
        /// </summary>
        [JsonProperty("storage")]
        public string StorageProviderName = null;
    }
}