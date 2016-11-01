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
using System.Reflection;
using Emitter.Configuration;
using Emitter.Diagnostics;
using Emitter.Providers;
using Emitter.Security;
using Emitter.Storage;

namespace Emitter
{
    internal class Program
    {
        /// <summary>
        /// Main entry point.
        /// </summary>
        private static void Main(string[] args)
        {
            // Load the configuration and check the license
            EmitterConfig.Initialize(args);
            var license = SecurityLicense.LoadAndVerify(EmitterConfig.Default.License);
            if (license == null)
            {
                // Generate new license
                var newLicense = EmitterLicenseGenerator.Default.GenerateLicense();
                Service.Logger.Log(LogLevel.Warning, "New license: " + newLicense.Sign());

                // Generate new secret key
                var newSecret = EmitterLicenseGenerator.Default.GenerateSecretKey(newLicense);
                Service.Logger.Log(LogLevel.Warning, "New secret key: " + newSecret.Value);

                // Note to user
                Service.Logger.Log(LogLevel.Warning, "New license and secret key were generated, please store them in a secure location and restart the server with the license provided.");
                SecurityLicense.Current = newLicense;
            }

            // Setup providers
            EmitterConfig.Default.Provider.LoggingProviderName = nameof(EmitterLoggingProvider);

            try
            {
                // Register assemblies
                Service.MetadataProvider.RegisterAssembly(typeof(Program).GetTypeInfo().Assembly);
                Service.MetadataProvider.RegisterAssembly(typeof(ContractProvider).GetTypeInfo().Assembly);
                Service.MetadataProvider.RegisterAssembly(typeof(ObjectStorageProvider).GetTypeInfo().Assembly);

                // Configure the mesh to the external ip address with the specified port
                Service.Mesh.BroadcastEndpoint = EmitterConfig.Default.Cluster.BroadcastEndpoint;
                Service.Mesh.Cluster = EmitterConfig.Default.Cluster.ClusterKey;

                // Setup the providers
                Service.Providers.Register<ContractProvider>(EmitterConfig.Default.Provider.ContractProviderName);
                Service.Providers.Register<LoggingProvider>(EmitterConfig.Default.Provider.LoggingProviderName);
                Service.Providers.Register<StorageProvider>(EmitterConfig.Default.Provider.StorageProviderName);
                Service.Providers.Register<CertificateProvider>(EmitterConfig.Default.Provider.CertificateProviderName);
                Service.Providers.Register<MonitoringProvider>(new EmitterMonitoringProvider());

                // Initialize the monitoring provider
                Service.Monitoring.Initialize();

                // Start listening to the endpoints
                Service.Listen(
                    new TcpBinding(EmitterConfig.Default.TcpPort),
                    new TlsBinding(EmitterConfig.Default.TlsPort),
                    new MeshBinding(EmitterConfig.Default.Cluster.Port)
                    );
            }
            catch (Exception ex)
            {
                // Unable to start, simply exit and let it restart
                Service.Logger.Log(ex);
            }
        }
    }
}