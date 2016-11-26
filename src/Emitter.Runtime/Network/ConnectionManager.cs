// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Emitter.Collections;
using Emitter.Network.Native;
using Emitter.Network.Threading;

namespace Emitter.Network
{
    /// <summary>
    /// Represents LibUv connection manager which closes the connections.
    /// </summary>
    internal class ConnectionManager
    {
        private EventThread EventThread;
        private List<Task> ConnectionStopTasks = new List<Task>();

        private WeakReferenceList<Connection> Connections = new WeakReferenceList<Connection>();

        public ConnectionManager(EventThread thread)
        {
            this.EventThread = thread;
            Timer.PeriodicCall(TimeSpan.FromSeconds(5), this.CheckAllAlive);
        }

        /// <summary>
        /// Walks through all connections and expires if necessary.
        /// </summary>
        internal void CheckAllAlive()
        {
            lock (Connections)
            {
                try
                {
                    // Purge and iterate
                    var connections = this.Connections;
                    foreach (var connection in connections)
                    {
                        if (!connection.CheckAlive())
                        {
                            // Unbind first
                            connection.Client.UnbindChannel(connection);

                            // Request a stop
                            var stopTask = connection.StopAsync();

                            // Enqueue the stop task
                            lock (this.ConnectionStopTasks)
                                ConnectionStopTasks.Add(stopTask);

                            //connection.Close(CloseType.SocketDisconnect);
                        }
                    }
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                }
            }
        }

        /// <summary>
        /// Registers a conections within the manager.
        /// </summary>
        /// <param name="connection">The connection to register.</param>
        public void Register(Connection connection)
        {
            this.Connections.Add(connection);
        }

        /// <summary>
        /// Unregisters a conections from the manager.
        /// </summary>
        /// <param name="connection">The connection to unregister.</param>
        public void Unregister(Connection connection)
        {
            this.Connections.Remove(connection);
        }

        /// <summary>
        /// Closes the connections.
        /// This must be called on the libuv event loop
        /// </summary>
        public void WalkConnectionsAndClose()
        {
            lock (this.ConnectionStopTasks)
            {
                EventThread.Walk(ptr =>
                {
                    var handle = UvMemory.FromIntPtr<UvHandle>(ptr);
                    var connection = (handle as UvStreamHandle)?.Connection;
                    if (connection != null)
                    {
                        ConnectionStopTasks.Add(connection.StopAsync());
                    }
                });
            }
        }

        public Task WaitForConnectionCloseAsync()
        {
            if (ConnectionStopTasks == null)
                throw new InvalidOperationException($"{nameof(WalkConnectionsAndClose)} must be called first.");

            return Task.WhenAll(ConnectionStopTasks);
        }
    }
}