using System;
using System.Diagnostics;
using System.Runtime.CompilerServices;
using System.Text;
using Emitter.Configuration;
using Emitter.Diagnostics;
using Emitter.Security;
using Emitter.Text.Json;

namespace Emitter
{
    /// <summary>
    /// Represents the statistics for the emitter.io node, published inside emitter.io itself.
    /// </summary>
    internal sealed partial class EmitterStatus
    {
        #region Static Members

        /// <summary>
        /// The node identifier, generated each time the process restarts.
        /// </summary>
        public static readonly int NodeIdentifier;

        /// <summary>
        /// Gets the machine name.
        /// </summary>
        private static readonly string MachineName = Process.GetCurrentProcess().MachineName;

        /// <summary>
        /// Gets the user name.
        /// </summary>
        internal static readonly string Address;

        /// <summary>
        /// Gets the default instance of the status.
        /// </summary>
        public static readonly EmitterStatus Default;

        /// Sampling queues
        private static readonly PerfCrossSection SamplerMPSIn = new PerfCrossSection(10);

        private static readonly PerfCrossSection SamplerMPSOut = new PerfCrossSection(10);

        /// <summary>
        /// Gets the current process id.  This method exists because of how CAS operates on the call stack, checking
        /// for permissions before executing the method.  Hence, if we inlined this call, the calling method would not execute
        /// before throwing an exception requiring the try/catch at an even higher level that we don't necessarily control.
        /// </summary>
        [MethodImpl(MethodImplOptions.NoInlining)]
        private static int GetCurrentProcessId()
        {
            return Process.GetCurrentProcess().Id;
        }

        /// <summary>
        /// Constructs the state object.
        /// </summary>
        static EmitterStatus()
        {
            try
            {
                // Prepare the datastructure
                Default = new EmitterStatus();
                Default.Network = new EmitterStatusNetwork();

                // Use the provider and get the external ip
                Address = EmitterConfig.Default.Cluster.BroadcastEndpoint.ToString();

                // Hash both values into one integer
                NodeIdentifier = Murmur32.GetHash(Address);
            }
            catch (Exception ex)
            {
                // Log the exception here
                Service.Logger.Log(ex);
            }
        }

        /// <summary>
        /// Starts the monitoring
        /// </summary>
        [InvokeAt(InvokeAtType.Initialize)]
        public static void Initialize()
        {
            // Post 10 times a second.
            Timer.PeriodicCall(TimeSpan.FromMilliseconds(100), EmitterStatus.Default.Publish);
        }

        #endregion Static Members

        /// <summary>
        /// Publishes the status to the cluster.
        /// </summary>
        public void Publish()
        {
            // Enqueue into samplers
            SamplerMPSIn.CumulativeDelta(Service.Monitoring.MessagesIncoming);
            SamplerMPSOut.CumulativeDelta(Service.Monitoring.MessagesOutgoing);

            // Push the time and static values
            this.Time = DateTime.UtcNow;
            this.Node = EmitterStatus.NodeIdentifier.ToString();
            this.Host = EmitterStatus.Address;
            this.Machine = EmitterStatus.MachineName;
            this.CPU = Math.Round(NetStat.AverageCpuUsage.Value, 2);
            this.Memory = Math.Round(NetStat.AverageWorkingSet.Value, 2);
            this.Uptime = NetStat.Uptime;

            // Set the network info
            this.Network.Connections = Service.Clients.Count;
            this.Network.AveragePPSIncoming = Math.Round(NetStat.AveragePPSIncoming.Value, 2);
            this.Network.AveragePPSOutgoing = Math.Round(NetStat.AveragePPSOutgoing.Value, 2);
            this.Network.AverageBPSIncoming = Math.Round(NetStat.AverageBPSIncoming.Value, 2);
            this.Network.AverageBPSOutgoing = Math.Round(NetStat.AverageBPSOutgoing.Value, 2);
            this.Network.AverageMPSIncoming = Math.Round(SamplerMPSIn.Sum(), 2);
            this.Network.AverageMPSOutgoing = Math.Round(SamplerMPSOut.Sum(), 2);
            this.Network.Compression = Math.Round(NetStat.Compression.Value, 2);

            // Create the message
            var message = Encoding.UTF8.GetBytes(
                JsonConvert.SerializeObject(this)
                ).AsSegment();

            // Prepare options
            var contract = SecurityLicense.Current.Contract;
            var channel = "cluster/" + this.Host + "/";

            // Get the channel hash for query
            EmitterChannel info;
            if (!EmitterChannel.TryParse(channel, false, out info))
                return;

            // Publish the message in the cluster, don't need to store it as it might slow everything down
            Dispatcher.Publish(contract, info.Target, channel, message, 60);
        }
    }
}