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
using System.Collections.Concurrent;
using System.Linq;
using Emitter.Collections;

namespace Emitter.Network
{
    /// <summary>
    /// Defines a class that maintains a registry of all currently connected clients
    /// </summary>
    public class ClientRegistry : ConcurrentList<IClient>
    {
        #region Constructors

        /// <summary>
        /// A timer used to process the queue.
        /// </summary>
        private Timer QueueTimer = null;

        private ConcurrentQueue<Action> Queue = new ConcurrentQueue<Action>();

        /// <summary>
        /// Constructs a new instance of <see cref="ClientRegistry"/> class. As there can be only one instance, it can be accessed from <see cref="Service"/> class.
        /// </summary>
        internal ClientRegistry()
        {
            // Register the events
            Service.ClientConnect += OnClientConnect;
            Service.ClientDisconnect += OnClientDisconnect;

            // Setup the timer
            if (this.QueueTimer == null)
                this.QueueTimer = Timer.PeriodicCall(TimeSpan.FromSeconds(1), this.OnProcessQueue);
        }

        #endregion Constructors

        #region Private Members

        /// <summary>
        /// Invoked when a client is connected.
        /// </summary>
        private void OnClientConnect(ClientConnectEventArgs e)
        {
            this.Enqueue(() =>
            {
                this.Add(e.Client);
            });
        }

        /// <summary>
        /// Invoked when a client is disconnected.
        /// </summary>
        private void OnClientDisconnect(ClientDisconnectEventArgs e)
        {
            this.Enqueue(() =>
            {
                if (!Service.IsStopping)
                    this.Remove(e.Client);
            });
        }

        /// <summary>
        /// Enqueues an action to the action queue.
        /// </summary>
        private void Enqueue(Action action)
        {
            // Push add/remove to the queue to execute.
            this.Queue.Enqueue(action);
        }

        /// <summary>
        /// Invoked when the add/remove queue should be processed.
        /// </summary>
        private void OnProcessQueue()
        {
            if (this.Queue.Count == 0)
                return;

            Action action;
            while (this.Queue.TryDequeue(out action))
            {
                try
                {
                    // Executes the action
                    action();
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                }
            }
        }

        #endregion Private Members

        #region IDisposable Members

        /// <summary>
        /// Releases the unmanaged resources used by the ByteSTream class and optionally releases the managed resources.
        /// </summary>
        /// <param name="disposing"> If set to true, release both managed and unmanaged resources, othewise release only unmanaged resources. </param>
        protected override void Dispose(bool disposing)
        {
            // Make sure we call the base
            base.Dispose(disposing);

            // Unregister the events
            Service.ClientConnect -= OnClientConnect;
            Service.ClientDisconnect -= OnClientDisconnect;
        }

        #endregion IDisposable Members
    }
}