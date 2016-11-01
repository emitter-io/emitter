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
        public ConfigOfCluster Cluster = new ConfigOfCluster();

        /// <summary>
        /// Gets or sets the vault config.
        /// </summary>
        [JsonProperty("vault")]
        public ConfigOfVault Vault = new ConfigOfVault();

        /// <summary>
        /// Gets or sets the storage config.
        /// </summary>
        [JsonProperty("storage")]
        public ConfigOfStorage Storage = new ConfigOfStorage();

        /// <summary>
        /// Gets or sets the providers config
        /// </summary>
        [JsonProperty("providers")]
        public ConfigOfProviders Provider = new ConfigOfProviders();
    }
}