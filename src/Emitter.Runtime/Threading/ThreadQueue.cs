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
using System.Threading;

namespace Emitter.Threading
{
    /// <summary>
    /// Represents a worker with a work queue and a separate thread.
    /// </summary>
    public abstract class ThreadQueue<TWork> : DisposableObject
    {
        /// <summary>
        /// The thread performing the work
        /// </summary>
        private Thread Thread;

        /// <summary>
        /// The work queue.
        /// </summary>
        protected readonly ConcurrentQueue<TWork> WorkQueue = new ConcurrentQueue<TWork>();

        /// <summary>
        /// Creates a new work scheduler.
        /// </summary>
        public ThreadQueue(int interval = 1000)
        {
            this.Interval = TimeSpan.FromMilliseconds(interval);
            this.Thread = new Thread(this.Execute);
            this.Thread.Start();
        }

        /// <summary>
        /// Gets or sets the interval for the worker thread.
        /// </summary>
        public TimeSpan Interval
        {
            get;
            set;
        }

        /// <summary>
        /// Processes the index queue and streams it to the database.
        /// </summary>
        private void Execute()
        {
            // Invoke on start
            this.OnStart();

            // Wait until we're starting
            while (!Service.IsRunning)
                Thread.Sleep(1000);

            // Run while the service is running
            while (Service.IsRunning)
            {
                try
                {
                    this.OnProcess();
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                }

                // Deschedule the thread
                Thread.Sleep((int)this.Interval.TotalMilliseconds);
            }
        }

        /// <summary>
        /// Occurs when the scheduler is starting.
        /// </summary>
        protected virtual void OnStart()
        {
        }

        /// <summary>
        /// Occurs when work needs to be done.
        /// </summary>
        protected abstract void OnProcess();
    }
}