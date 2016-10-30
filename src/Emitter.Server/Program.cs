using System;
using System.Reflection;
using Emitter.Configuration;
using Emitter.Diagnostics;
using Emitter.Providers;
using Emitter.Security;

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

            try
            {
                // Register assemblies
                Service.MetadataProvider.RegisterAssembly(typeof(Program).GetTypeInfo().Assembly);
                Service.MetadataProvider.RegisterAssembly(typeof(ContractProvider).GetTypeInfo().Assembly);
                Service.MetadataProvider.RegisterAssembly(typeof(StorageProvider).GetTypeInfo().Assembly);

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