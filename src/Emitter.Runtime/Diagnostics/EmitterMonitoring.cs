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
using System.Text;
using System.Threading;
using Emitter.Providers;
using Emitter.Security;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a monitoring provider for Prometheus.
    /// </summary>
    public sealed class EmitterMonitoringProvider : MonitoringProvider
    {
        private const int TickInterval = 1;

        /// <summary>
        /// Constructs the instance of the provider
        /// </summary>
        public EmitterMonitoringProvider()
        {
            Service.MessageReceived = OnMessageIn;
            Service.MessageSent = OnMessageOut;
        }

        /// <summary>
        /// Initializes the monitoring system. This should setup the initial state of the counters and things.
        /// </summary>
        public override void Initialize()
        {
            // Post every x seconds
            Timer.PeriodicCall(TimeSpan.FromSeconds(TickInterval), Publish);
        }

        /// <summary>
        /// Occurs when a new message is published on the channel.
        /// </summary>
        /// <param name="contract">The contract for this message.</param>
        /// <param name="channel">The channel for this message.</param>
        /// <param name="length">The length of the message, in bytes.</param>
        private void OnMessageIn(IContract contract, string channel, int length)
        {
            // Get the usage from the contract
            var usage = ((EmitterContract)contract).Usage;

            usage.IncomingMessages.Increment();
            usage.IncomingTraffic.IncrementBy(length);

            Interlocked.Increment(ref MessagesIncoming);
        }

        /// <summary>
        /// Occurs when a new message is sent for the channel.
        /// </summary>
        /// <param name="contract">The contract for this message.</param>
        /// <param name="channel">The channel for this message.</param>
        /// <param name="length">The length of the message, in bytes.</param>
        private void OnMessageOut(IContract contract, string channel, int length)
        {
            // Get the usage from the contract
            var usage = ((EmitterContract)contract).Usage;

            usage.OutgoingMessages.Increment();
            usage.OutgoingTraffic.IncrementBy(length);

            Interlocked.Increment(ref MessagesOutgoing);
        }

        /// <summary>
        /// Periodically publishes the usage monitors.
        /// </summary>
        private void Publish()
        {
            try
            {
                // Prepare options
                var ourContract = SecurityLicense.Current.Contract;
                var builder = new StringBuilder();

                // Publish every monitor as a separate message
                foreach (var c in Services.Contract)
                {
                    // Get the usage from the contract
                    var contract = c as EmitterContract;
                    if (contract == null)
                        continue;

                    // If the monitor didn't change, we don't need to publish it
                    var usage = contract.Usage;
                    if (usage.IsZero)
                        continue;

                    try
                    {
                        // Reuse the builder
                        builder.Clear();

                        // Sample to the througput limiter
                        usage.MessageFrequency.Sample(usage.IncomingMessages.Value + usage.OutgoingMessages.Value);

                        // Write and reset the usage
                        usage.Write(builder, contract.Oid);
                        usage.Reset();

                        // Convert to a message, TODO: remove unnecessary allocations
                        var message = builder
                            .ToString()
                            .AsASCII()
                            .AsSegment();

                        // The channel to publish into
                        var channel = "usage/" + contract.Oid + "/";

                        // Get the channel hash for query
                        EmitterChannel info;
                        if (!EmitterChannel.TryParse(channel, false, out info))
                            return;

                        // Publish the message in the cluster, don't need to store it as it might slow everything down
                        Dispatcher.Publish(ourContract, info.Target, channel, message, 60);
                    }
                    catch (Exception ex)
                    {
                        // Log the exception
                        Service.Logger.Log(ex);
                    }
                }
            }
            catch (Exception ex)
            {
                // Log the exception
                Service.Logger.Log(ex);
            }
        }
    }
}