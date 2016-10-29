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
using Emitter.Network;
using Emitter.Network.Threading;
using Emitter.Network.Tls;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a transport provider.
    /// </summary>
    public abstract class TransportProvider : Provider
    {
        /// <summary>
        /// Gets or sets the connection filter.
        /// </summary>
        public IConnectionFilter ConnectionFilter { get; set; } = new TlsFilter();

        /// <summary>
        /// Gets or sets value that instructs <seealso cref="EmitterServer"/> whether it is safe to
        /// pool the Request and Response <seealso cref="System.IO.Stream"/> objects
        /// for another request after the Response's OnCompleted callback has fired.
        /// When this values is greater than zero, it is not safe to retain references to feature components after this event has fired.
        /// Value is zero by default.
        /// </summary>
        public int MaxPooledStreams { get; set; }

        /// <summary>
        /// Gets or sets value that instructs <seealso cref="EmitterServer"/> whether it is safe to
        /// pool the Request and Response headers
        /// for another request after the Response's OnCompleted callback has fired.
        /// When this values is greater than zero, it is not safe to retain references to feature components after this event has fired.
        /// Value is zero by default.
        /// </summary>
        public int MaxPooledHeaders { get; set; }

        /// <summary>
        /// The amount of time after the server begins shutting down before connections will be forcefully closed.
        /// By default, Emitter will wait 5 seconds for any ongoing requests to complete before terminating
        /// the connection.
        /// </summary>
        public TimeSpan ShutdownTimeout { get; set; } = TimeSpan.FromSeconds(5);

        /// <summary>
        /// Gets or sets the number of threads for the event loop.
        /// </summary>
        public int ThreadCount { get; set; } = ProcessorThreadCount;

        /// <summary>
        /// Gets the number of processor threads.
        /// </summary>
        private static int ProcessorThreadCount
        {
            get
            {
                // Actual core count would be a better number
                // rather than logical cores which includes hyper-threaded cores.
                // Divide by 2 for hyper-threading, and good defaults (still need threads to do webserving).
                var threadCount = Environment.ProcessorCount >> 1;
                if (threadCount < 1)
                {
                    // Ensure shifted value is at least one
                    return 1;
                }

                if (threadCount > 16)
                {
                    // Receive Side Scaling RSS Processor count currently maxes out at 16
                    // would be better to check the NIC's current hardware queues; but xplat...
                    return 16;
                }

                return threadCount;
            }
        }

        /// <summary>
        /// Begins running the main Emitter Engine loop on the current thread.
        /// </summary>
        /// <param name="bindings">Network binding configuration to use</param>
        public abstract void Listen(params IBinding[] bindings);
    }

    public sealed class UvTransportProvider : TransportProvider
    {
        private Stack<IDisposable> Listeners;
        private TransportEngine Engine;

        /// <summary>
        /// Begins running the main Emitter Engine loop on the current thread.
        /// </summary>
        /// <param name="bindings">Network binding configuration to use</param>
        public override void Listen(params IBinding[] bindings)
        {
            // Is it already started?
            if (this.Listeners != null)
                throw new InvalidOperationException("Server has already started.");
            this.Listeners = new Stack<IDisposable>();

            // Check the bindings
            if (bindings == null || bindings.Length == 0)
                throw new ArgumentNullException("bindings");

            // Create an engine per binding
            foreach (var binding in bindings)
            {
                try
                {
                    // Get the addresses to listen to
                    var addresses = ServiceAddress.FromBinding(binding);

                    // Start a new transport engine for the binding
                    this.Engine = new TransportEngine(new ServiceContext
                    {
                        ThreadPool = new EventThreadPool()
                    });

                    Listeners.Push(this.Engine);

                    // Use only one thread for mesh
                    var threadCount = binding is MeshBinding ? 1 : ThreadCount;
                    if (threadCount <= 0)
                    {
                        throw new ArgumentOutOfRangeException(nameof(threadCount),
                            threadCount,
                            "ThreadCount must be positive.");
                    }

                    this.Engine.Start(threadCount);
                    var atLeastOneListener = false;

                    // Add a listener per address
                    foreach (var address in addresses)
                    {
                        try
                        {
                            atLeastOneListener = true;
                            Listeners.Push(
                                this.Engine.CreateServer(address, binding)
                                );

                            Service.Logger.Log("Listening: " + address);
                        }
                        catch (AggregateException aggr)
                        {
                            var ex = aggr.InnerException;
                            if (ex.Message.Contains("EADDRINUSE"))
                                Service.Logger.Log(LogLevel.Error, "Address In Use: " + address);
                            else
                            {
                                Service.Logger.Log(LogLevel.Error, ex.Message + ": " + address);
                            }
                        }
                    }

                    if (!atLeastOneListener)
                    {
                        throw new InvalidOperationException("No recognized listening addresses were configured.");
                    }
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                    Dispose();
                    throw;
                }
            }
        }

        /// <summary>
        /// Disposes the provider.
        /// </summary>
        /// <param name="disposing">Whether disposing or finalizing</param>
        protected override void Dispose(bool disposing)
        {
            base.Dispose(disposing);

            if (Listeners != null)
            {
                while (Listeners.Count > 0)
                    Listeners.Pop().Dispose();
                Listeners = null;
            }
        }
    }
}