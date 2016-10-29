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
using System.Linq;
using System.Threading;

namespace Emitter.Threading
{
    /// <summary>
    /// Represents a thread with an attached FSM.
    /// </summary>
    public abstract class ThreadBase
    {
        #region Constructors
        private object Lock = new object();
        private Thread CurrentThread;
        private ThreadState CurrentState;
        private AutoResetEvent SampleEvent = new AutoResetEvent(false);

        /// <summary>
        /// Constructs a new instance of a profiler, bound to a particular thread.
        /// </summary>
        /// <param name="interval">The interval to publish measurements info.</param>
        public ThreadBase(TimeSpan interval)
        {
            this.Interval = interval;
        }

        #endregion Constructors

        #region Public Properties

        /// <summary>
        /// Gets the current state of the instrument.
        /// </summary>
        public ThreadState State
        {
            get { return this.CurrentState; }
        }

        /// <summary>
        /// Gets or sets the sampling interval. Default is 1 second.
        /// </summary>
        public TimeSpan Interval
        {
            get;
            set;
        }

        #endregion Public Properties

        #region Start/Stop Members

        /// <summary>
        /// Starts the instrument thread.
        /// </summary>
        /// <returns>Whether the instrument has started or not.</returns>
        public bool Start()
        {
            lock (this.Lock)
            {
                if (this.State != ThreadState.Stopped)
                    return false;

                this.OnStart();
                this.CurrentState = ThreadState.Starting;
                this.CurrentThread = new Thread(this.Run);
                this.CurrentThread.Start();
                return true;
            }
        }

        /// <summary>
        /// Stops the instrument thread.
        /// </summary>
        /// <returns>Whether the instrument has stopped or not.</returns>
        public bool Stop()
        {
            lock (this.Lock)
            {
                if (this.State == ThreadState.Stopped)
                    return false;

                this.OnStop();
                this.CurrentState = ThreadState.Stopping;
                this.CurrentThread = null;
                return true;
            }
        }

        /// <summary>
        /// Occurs when the instrument is starting.
        /// </summary>
        protected virtual void OnStart()
        {
        }

        /// <summary>
        /// Occurs when the instrument is stopping.
        /// </summary>
        protected virtual void OnStop()
        {
        }

        #endregion Start/Stop Members

        #region Threading Members

        /// <summary>
        /// Runs the instrument thread.
        /// </summary>
        private void Run()
        {
            // Mark as running
            this.CurrentState = ThreadState.Running;

            // Run the instrument loop
            while (this.CurrentState == ThreadState.Running && Service.IsRunning)
            {
                // Execute the instrument
                this.OnExecute();

                // Deschedule the thread
                this.SampleEvent.WaitOne(this.Interval);
            }

            // Once we exit this loop, we're done
            this.CurrentState = ThreadState.Stopped;
        }

        /// <summary>
        /// Gets the message to publish to emitter channeL.
        /// </summary>
        /// <returns></returns>
        protected abstract void OnExecute();

        #endregion Threading Members
    }

    /// <summary>
    /// Represents a profiler state.
    /// </summary>
    public enum ThreadState
    {
        Stopped,
        Running,
        Stopping,
        Starting
    }
}