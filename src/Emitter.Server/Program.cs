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
            // Print out fingerprint
            // Service.Logger.Log("Machine: " + Service.Providers.Resolve<AddressProvider>().GetFingerprint());
            NetTrace.Enabled = true;
            NetTrace.Listeners.Add(new ConsoleTraceListener());
            NetTrace.TraceMesh = true;
            //NetTrace.TraceChannel = true;

            // Load the configuration and check the license
            EmitterConfig.Initialize(args);
            var license = SecurityLicense.LoadAndVerify(EmitterConfig.Default.License);
            if (license == null)
            {
                Service.Logger.Log("Creating a new dummy license, do not use this for production.");
                license = new SecurityLicense();
                license.EncryptionKey = "0000000000000000000000";
                SecurityLicense.Current = license;
            }

            if (license.Type == SecurityLicenseType.Cloud)
            {
                // Setup providers used by our cloud
                EmitterConfig.Default.Provider.ContractProviderName = nameof(HttpContractProvider);
                EmitterConfig.Default.Provider.LoggingProviderName = nameof(EmitterLoggingProvider);
                EmitterConfig.Default.Provider.CertificateProviderName = nameof(VaultCertificateProvider);
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