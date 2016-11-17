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
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Runtime.InteropServices;
using System.Threading;
using Emitter.Configuration;
using Emitter.Diagnostics;
using Emitter.Network;
using Emitter.Network.Http;
using Emitter.Providers;

namespace Emitter
{
    /// <summary>
    /// The <see cref="Service"/> class represents the main entry point to Emitter Engine.
    /// </summary>
    public static partial class Service
    {
        #region Private Fields

        private delegate bool ConsoleEventHandler(ConsoleEventType type);

        private static AutoResetEvent fSliceEvent = new AutoResetEvent(true);
        private static Thread fTimerExecutor;
        private static Assembly fRuntimeAssembly;
        private static Process fProcess;
        private static Thread fThread;
        private static string fBaseDirectory;
        private static string fDataDirectory;
        private static bool fMultiProcessor;
        private static bool fStopping;
        private static bool fRunning;
        private static bool fCrashed;
        private static int fProcessorCount;

#if !DOTNET
        private static ConsoleEventHandler fConsoleEventHandler;
#endif

        private static ClientRegistry fClients = new ClientRegistry();
        private static List<MethodInfo> fTerminators = new List<MethodInfo>();
        private static MetadataProvider fMetadata = new DefaultMetadataProvider();
        private static RegistryProvider fRegistry;
        private static ProvidersContainer fProviders = new ProvidersContainer(DefaultProviders.Associations);

        [DllImport("Kernel32")]
        private static extern bool SetConsoleCtrlHandler(ConsoleEventHandler callback, bool add);

        #endregion Private Fields

        #region Public Properties

        /// <summary>
        /// Gets the Emitter Runtime assembly.
        /// </summary>
        public static Assembly RuntimeAssembly
        {
            get { return fRuntimeAssembly; }
        }

        /// <summary>
        /// Gets the process where the kernel is running.
        /// </summary>
        public static Process Process
        {
            get { return fProcess; }
        }

        /// <summary>
        /// Gets the thread where the kernel is running.
        /// </summary>
        public static Thread Thread
        {
            get { return fThread; }
        }

        /// <summary>
        /// Gets current processor count of the machine.
        /// </summary>
        public static int ProcessorCount
        {
            get { return fProcessorCount; }
        }

        /// <summary>
        /// Gets whether the operating system is 64bit or not.
        /// </summary>
        public static bool Is64Bit
        {
            get { return IntPtr.Size == 8; }
        }

        /// <summary>
        /// Gets whether the service is shutting down or not.
        /// </summary>
        public static bool IsStopping
        {
            get { return fStopping; }
        }

        /// <summary>
        /// Gets whether the service is running or not.
        /// </summary>
        public static bool IsRunning
        {
            get { return fRunning && !fStopping; }
        }

        /// <summary>
        /// Gets whether the host machine is a multi-procesor one or not.
        /// </summary>
        public static bool IsMultiProcessor
        {
            get { return fMultiProcessor; }
        }

        /// <summary>
        /// Gets the log provider (default: multi console output) can be used as error output.
        /// </summary>
        public static LoggingProvider Logger
        {
            get { return fProviders.Resolve<LoggingProvider>(); }
        }

        /// <summary>
		/// Gets the http handlers provider which is used to handle different http requests.
		/// </summary>
		public static HttpProvider Http
        {
            get { return fProviders.Resolve<HttpProvider>(); }
        }

        /// <summary>
        /// Gets the monitoring provider.
        /// </summary>
        public static MonitoringProvider Monitoring
        {
            get { return fProviders.Resolve<MonitoringProvider>(); }
        }

        /// <summary>
        /// Gets the mesh provider for inter-cluster communication.
        /// </summary>
        public static MeshProvider Mesh
        {
            get
            {
                // Return the mesh provider
                var provider = fProviders.Resolve<MeshProvider>();
                if (provider != null)
                    return provider;

                // Register a new mesh provider if none is available yet
                Providers.RegisterInstance<MeshProvider>(new DefaultMeshProvider());
                return fProviders.Resolve<MeshProvider>();
            }
        }

        /// <summary>
        /// Gets the cross-origin resource sharing policy provider.
        /// </summary>
        public static CorsProvider Cors
        {
            get { return fProviders.Resolve<CorsProvider>(); }
        }

        /// <summary>
        /// Gets providers container (an IoC container) for different providers of the server.
        /// </summary>
        public static ProvidersContainer Providers
        {
            get { return fProviders; }
        }

        /// <summary>
        /// Gets the collection of currently connected clients.
        /// </summary>
        public static ClientRegistry Clients
        {
            get { return fClients; }
        }

        /// <summary>
        /// Gets the provider allowing to configure transport security layer.
        /// </summary>
        public static CertificateProvider Tls
        {
            get { return fProviders.Resolve<CertificateProvider>(); }
        }

        /// <summary>
        /// Gets the provider for the internal and replicated registry.
        /// </summary>
        public static RegistryProvider Registry
        {
            get { return fRegistry; }
        }

        /// <summary>
        /// Gets or sets the metadata provider containing the registered module assemblies and
        /// various metadata information for those assemblies. This provider is handled separately
        /// from the all other providers as it is used for bootstrapping the Service.
        /// </summary>
        public static MetadataProvider MetadataProvider
        {
            get { return fMetadata; }
            set
            {
                if (value == null)
                    throw new ArgumentNullException("value");
                fMetadata = value;
            }
        }

        /// <summary>
        /// Gets or the transport provider.
        /// </summary>
        public static TransportProvider Transport
        {
            get
            {
                // Return the mesh provider
                var provider = fProviders.Resolve<TransportProvider>();
                if (provider != null)
                    return provider;

                // Register a new mesh provider if none is available yet
                Providers.RegisterInstance<TransportProvider>(new UvTransportProvider());
                return fProviders.Resolve<TransportProvider>();
            }
        }

        /// <summary>
        /// Gets or sets the base working directory of the executable
        /// </summary>
        public static string BaseDirectory
        {
            get
            {
                if (fBaseDirectory == null)
                {
                    try
                    {
#if DOTNET
                        fBaseDirectory = AppContext.BaseDirectory;
#else
                        fBaseDirectory = Assembly.GetEntryAssembly().Location;
#endif
                        if (fBaseDirectory.Length > 0)
                            fBaseDirectory = Path.GetDirectoryName(fBaseDirectory);
                    }
                    catch
                    {
                        fBaseDirectory = "";
                    }
                }

                return fBaseDirectory;
            }
            set
            {
                fBaseDirectory = value;
            }
        }

        /// <summary>
        /// Gets or sets the data directory for the server
        /// </summary>
        public static string DataDirectory
        {
            get
            {
                if (fDataDirectory == null)
                {
                    fDataDirectory = Path.Combine(BaseDirectory, "data");
                    if (!Directory.Exists(fDataDirectory))
                        Directory.CreateDirectory(fDataDirectory);
                }
                return fDataDirectory;
            }
            set
            {
                fDataDirectory = value;
            }
        }

        /// <summary>
        /// Gets or sets the configuration directory for the server
        /// </summary>
        public static string ConfigDirectory
        {
            get
            {
                if (fDataDirectory == null)
                {
                    fDataDirectory = Path.Combine(BaseDirectory, "config");
                    if (!Directory.Exists(fDataDirectory))
                        Directory.CreateDirectory(fDataDirectory);
                }
                return fDataDirectory;
            }
            set
            {
                fDataDirectory = value;
            }
        }

        /// <summary>
        /// Gets or sets the idle interval on which a new cycle should be forced
        /// </summary>
        public static TimeSpan Interval
        {
            get { return Timer.IdleCycleInterval; }
            set { Timer.IdleCycleInterval = value; }
        }

        #endregion Public Properties

        #region Main Method

        /// <summary>
        /// Begins running the main Emitter Engine loop on the current thread.
        /// </summary>
        /// <param name="bindings">Network binding configuration to use</param>
        public static void Listen(params IBinding[] bindings)
        {
            if (bindings == null || bindings.Length == 0)
                throw new ArgumentNullException("bindings");

            // Set the mesh binding
            Mesh.Binding = bindings
                .Where(b => b is MeshBinding)
                .FirstOrDefault();

            // Set the registry and set our own identifier
            fRegistry = new ReplicatedRegistryProvider(Mesh.Identifier);

            // Register the the mesh handler
            Service.Mesh.Register(new MessageHandler());

            // Tell that we're running
            fRunning = true;

#if !DOTNET
            // Set the priority
            Thread.CurrentThread.Priority = ThreadPriority.Highest;

            // Hook the exceptions and exit
            AppDomain.CurrentDomain.UnhandledException += OnUnhandledException;
			AppDomain.CurrentDomain.ProcessExit += CurrentDomain_ProcessExit;
#endif

            // Initialize Thread, Process and Assemblies
            fThread = Thread.CurrentThread;
            fProcess = Process.GetCurrentProcess();
            fRuntimeAssembly = typeof(Service).GetTypeInfo().Assembly;

            // Add Primary assemblies
            MetadataProvider.RegisterAssembly(fRuntimeAssembly);

            // Name the thread
            if (fThread != null)
            {
                // To avoid an exception in case of relisten, check if the thread is not already named
                if (fThread.Name != "Emitter Service")
                    fThread.Name = "Emitter Service";
            }

            if (BaseDirectory.Length > 0)
                Directory.SetCurrentDirectory(BaseDirectory);

            // Create a timer thread scheduler if necessary
            if (fTimerExecutor == null)
            {
                fTimerExecutor = new Thread(Timer.Executor.Run);
                if (fTimerExecutor.Name != "Timer Executor")
                    fTimerExecutor.Name = "Timer Executor";
            }

            // Check whether we are running on a multiprocessor hardware
            fProcessorCount = Environment.ProcessorCount;
            if (fProcessorCount > 1)
                fMultiProcessor = true;

            // Registers default providers (if they are not already there)
            DefaultProviders.RegisterAllDefault();

            // Register a new mesh provider if none is available yet
            if (Mesh.Cluster == null)
                Mesh.Cluster = Providers.Resolve<SecurityProvider>().CreateSessionToken();

            var configures = new List<MethodInfo>();
            var initializes = new List<MethodInfo>();
            MetadataProvider.Types.ForEach((type) =>
            {
                // Perform InvokeAt
                configures.AddRange(type.GetInvokeAt(InvokeAtType.Configure));
                initializes.AddRange(type.GetInvokeAt(InvokeAtType.Initialize));
                fTerminators.AddRange(type.GetInvokeAt(InvokeAtType.Terminate));
            });

            // Sorts the Configure() by priority and invokes them
            configures.Sort(new InvokePriorityComparer());
            foreach (var configure in configures)
            {
                try
                {
                    configure.Invoke(null, null);
                }
                catch (Exception ex) { Logger.Log(ex); }
            }
            configures.Clear();

            // Create sockeds and bind the socket pump
            GC.Collect();

            // Invoke server startup event
            InvokeServerStartup();

            // Start the timer executor now
            fTimerExecutor.Start();

            // Start listening
            Transport.Listen(bindings);

            // Check X509
            if (bindings.Where(b => b is TlsBinding).Any() && Tls.Certificate == null)
                Logger.Log(LogLevel.Warning, "Unable to detect TLS/SSL certificate.");

            // Sorts the Initialize() by priority and invokes them
            initializes.Sort(new InvokePriorityComparer());
            foreach (var initialize in initializes)
            {
                try
                {
                    initialize.Invoke(null, null);
                }
                catch (Exception ex) { Logger.Log(ex); }
            }
            initializes.Clear();

            // Join the cluster through the seed
            Mesh.ConnectTo(EmitterConfig.Default.Cluster.SeedEndpoint);

            // Register http handlers
            Http.Register(new HttpHealthHandler());
            Http.Register(new HandleKeyGen());
            Http.Register(new DebugHttp());

            // Send analytics about this server launch. This is to gather global statistics about the number
            // of unique machines that use emitter.
            Providers.Resolve<AnalyticsProvider>().Track("emitter", "launch");

            // Timer scheduler slice loop
            while (IsRunning)
            {
                Timer.Scheduler.Slice();
                Thread.Sleep(20);
            }

            // We're done
            fRunning = false;
        }

        #endregion Main Method

        #region Console Hooks

        private static bool OnConsoleEvent(ConsoleEventType type)
        {
            Shutdown();
            return true;
        }

        private enum ConsoleEventType
        {
            CTRL_C_EVENT,
            CTRL_BREAK_EVENT,
            CTRL_CLOSE_EVENT,
            CTRL_LOGOFF_EVENT = 5,
            CTRL_SHUTDOWN_EVENT
        }

        #endregion Console Hooks

        #region Process Exit Handling
#if !DOTNET
        /// <summary>
        /// Handles an unhandled exception from the AppDomain.
        /// </summary>
        /// <param name="sender">The sender.</param>
        /// <param name="e">The event arguments.</param>
        private static void OnUnhandledException(object sender, UnhandledExceptionEventArgs e)
        {
            OnUnhandledException(e.ExceptionObject as Exception, e.IsTerminating);
        }

        /// <summary>
        /// Occurs when the current AppDomain is exiting.
        /// </summary>
        /// <param name="sender">The sender.</param>
        /// <param name="e">The event arguments.</param>
        private static void CurrentDomain_ProcessExit(object sender, EventArgs e)
        {
            Shutdown();
        }
#endif

        /// <summary>
        /// Occurs when an unhandled exception is thrown.
        /// </summary>
        /// <param name="ex">The exception thrown.</param>
        /// <param name="fatal">Whether this is fatal exception or not.</param>
        private static void OnUnhandledException(Exception ex, bool fatal = false)
        {
            try
            {
                // Log it
                Logger.Log(ex);

                // If it's fatal error
                if (fatal)
                {
                    fCrashed = true;

                    bool close = false;

                    try
                    {
                        ServerShutdownEventArgs args = new ServerShutdownEventArgs(ex);
                        Service.InvokeServerShutdown(args);
                        close = args.Close;
                    }
                    catch { }

                    // Dispose the pump
                    Transport.Dispose();
                    fStopping = true;
                }
            }
            catch (Exception e)
            {
                // Something went wrong
                Logger.Log(e);
            }
        }

        /// <summary>
        /// Stops the server without restarting
        /// </summary>
        public static void Shutdown()
        {
            if (fStopping)
                return;

            fStopping = true;
            fRunning = false;

            Service.Logger.Log(LogLevel.Info, "The service is shutting down"
                + (fCrashed
                    ? " due to a crash."
                    : "."
                ));

            if (!fCrashed)
            {
                Service.InvokeServerShutdown(new ServerShutdownEventArgs());

                // Call the finalizers as the shutdown process is normal
                fTerminators.Sort(new InvokePriorityComparer());
                foreach (var terminator in fTerminators)
                    terminator.Invoke(null, null);
            }

            // Terminate the timer thread
            Timer.Scheduler.Set();
        }

        #endregion Process Exit Handling
    }

    /// <summary>
    /// Represents the event log granularity
    /// </summary>
    public enum LogLevel
    {
        /// <summary>
        /// Information only
        /// </summary>
        Info,

        /// <summary>
        /// Warnings, non-critical errors
        /// </summary>
        Warning,

        /// <summary>
        /// Critical errors
        /// </summary>
        Error
    }
}