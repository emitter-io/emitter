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
using System.Threading;

namespace Emitter.Threading
{
    internal interface IThreadPool : IDisposable
    {
        void QueueUserWorkItem(WaitCallback work, object obj);
    }

    public class ThreadPool : IThreadPool
    {
        public ThreadPool() :
            this(Environment.ProcessorCount)
        { }

        public ThreadPool(int concurrencyLevel)
        {
            if (concurrencyLevel <= 0)
                throw new ArgumentOutOfRangeException("concurrencyLevel");
            this.ConcurrencyLevel = concurrencyLevel;
        }

        // Each work item consists of a closure: work + (optional) state obj + context.
        private struct WorkItem
        {
            internal WaitCallback m_work;
            internal object m_obj;

            internal WorkItem(WaitCallback work, object obj)
            {
                m_work = work;
                m_obj = obj;
            }

            internal void Invoke()
            {
                // Run normally (delegate invoke) or under context, as appropriate.
                m_work(m_obj);
            }
        }

        private readonly int ConcurrencyLevel;
        private readonly Queue<WorkItem> WorkQueue = new Queue<WorkItem>();
        private Thread[] m_threads;
        private int m_threadsWaiting;
        private bool m_shutdown;

        // Methods to queue work.

        public void QueueUserWorkItem(WaitCallback work)
        {
            QueueUserWorkItem(work, null);
        }

        public void QueueUserWorkItem(WaitCallback work, object obj)
        {
            var wi = new WorkItem(work, obj);

            // Make sure the pool is started (threads created, etc).
            EnsureStarted();

            // Now insert the work item into the queue, possibly waking a thread.
            lock (WorkQueue)
            {
                WorkQueue.Enqueue(wi);
                if (m_threadsWaiting > 0)
                    Monitor.Pulse(WorkQueue);
            }
        }

        // Ensures that threads have begun executing.
        private void EnsureStarted()
        {
            if (m_threads == null)
            {
                lock (WorkQueue)
                {
                    if (m_threads == null)
                    {
                        m_threads = new Thread[ConcurrencyLevel];
                        for (int i = 0; i < m_threads.Length; i++)
                        {
                            m_threads[i] = new Thread(DispatchLoop);
                            m_threads[i].Start();
                        }
                    }
                }
            }
        }

        // Each thread runs the dispatch loop.
        private void DispatchLoop()
        {
            while (true)
            {
                var wi = default(WorkItem);
                lock (WorkQueue)
                {
                    // If shutdown was requested, exit the thread.
                    if (m_shutdown)
                        return;

                    // Find a new work item to execute.
                    while (WorkQueue.Count == 0)
                    {
                        m_threadsWaiting++;
                        try { Monitor.Wait(WorkQueue); }
                        finally { m_threadsWaiting--; }

                        // If we were signaled due to shutdown, exit the thread.
                        if (m_shutdown)
                            return;
                    }

                    // We found a work item! Grab it ...
                    wi = WorkQueue.Dequeue();
                }

                // ...and Invoke it. Note: exceptions will go unhandled (and crash).
                wi.Invoke();
            }
        }

        // Disposing will signal shutdown, and then wait for all threads to finish.
        public void Dispose()
        {
            m_shutdown = true;
            lock (WorkQueue)
            {
                Monitor.PulseAll(WorkQueue);
            }

            for (int i = 0; i < m_threads.Length; i++)
                m_threads[i].Join();
        }
    }
}