//
// System.Threading.ReaderWriterLock.cs
//
// Author:
//   Dick Porter (dick@ximian.com)
//   Jackson Harper (jackson@ximian.com)
//   Lluis Sanchez Gual (lluis@ximian.com)
//
// (C) Ximian, Inc.  http://www.ximian.com
// (C) 2004 Novell, Inc (http://www.novell.com)
//

//
// Copyright (C) 2004 Novell, Inc (http://www.novell.com)
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
#if DOTNET

using System.Collections;
using System.Runtime.InteropServices;

namespace System.Threading
{
    [ComVisible(true)]
    public sealed class ReaderWriterLock //: CriticalFinalizerObject
    {
        private int seq_num = 1;
        private int state;
        private int readers;
        private int writer_lock_owner;
        private LockQueue writer_queue;
        private Hashtable reader_locks;

        public ReaderWriterLock()
        {
            writer_queue = new LockQueue(this);
            reader_locks = new Hashtable();

            GC.SuppressFinalize(this);
        }

        ~ReaderWriterLock()
        {
        }

        public bool IsReaderLockHeld
        {
            //[ReliabilityContract (Consistency.WillNotCorruptState, Cer.Success)]
            get
            {
                lock (this) return reader_locks.ContainsKey(Thread.CurrentThread.ManagedThreadId);
            }
        }

        public bool IsWriterLockHeld
        {
            //[ReliabilityContract (Consistency.WillNotCorruptState, Cer.Success)]
            get
            {
                lock (this) return (state < 0 && Thread.CurrentThread.ManagedThreadId == writer_lock_owner);
            }
        }

        public int WriterSeqNum
        {
            get
            {
                lock (this) return seq_num;
            }
        }

        public void AcquireReaderLock(int millisecondsTimeout)
        {
            AcquireReaderLock(millisecondsTimeout, 1);
        }

        private void AcquireReaderLock(int millisecondsTimeout, int initialLockCount)
        {
            lock (this)
            {
                if (HasWriterLock())
                {
                    AcquireWriterLock(millisecondsTimeout, initialLockCount);
                    return;
                }

                object nlocks = reader_locks[Thread.CurrentThread.ManagedThreadId];
                if (nlocks == null)
                {
                    // Not currently holding a reader lock
                    // Wait if there is a write lock
                    readers++;
                    try
                    {
                        if (state < 0 || !writer_queue.IsEmpty)
                        {
                            do
                            {
                                if (!Monitor.Wait(this, millisecondsTimeout))
                                    throw new Exception("Timeout expired");
                            } while (state < 0);
                        }
                    }
                    finally
                    {
                        readers--;
                    }

                    reader_locks[Thread.CurrentThread.ManagedThreadId] = initialLockCount;
                    state += initialLockCount;
                }
                else
                {
                    reader_locks[Thread.CurrentThread.ManagedThreadId] = ((int)nlocks) + 1;
                    state++;
                }
            }
        }

        public void AcquireReaderLock(TimeSpan timeout)
        {
            int ms = CheckTimeout(timeout);
            AcquireReaderLock(ms, 1);
        }

        public void AcquireWriterLock(int millisecondsTimeout)
        {
            AcquireWriterLock(millisecondsTimeout, 1);
        }

        private void AcquireWriterLock(int millisecondsTimeout, int initialLockCount)
        {
            lock (this)
            {
                if (HasWriterLock())
                {
                    state--;
                    return;
                }

                // wait while there are reader locks or another writer lock, or
                // other threads waiting for the writer lock
                if (state != 0 || !writer_queue.IsEmpty)
                {
                    do
                    {
                        if (!writer_queue.Wait(millisecondsTimeout))
                            throw new Exception("Timeout expired");
                    } while (state != 0);
                }

                state = -initialLockCount;
                writer_lock_owner = Thread.CurrentThread.ManagedThreadId;
                seq_num++;
            }
        }

        public void AcquireWriterLock(TimeSpan timeout)
        {
            int ms = CheckTimeout(timeout);
            AcquireWriterLock(ms, 1);
        }

        public bool AnyWritersSince(int seqNum)
        {
            lock (this)
            {
                return (this.seq_num > seqNum);
            }
        }

        public void DowngradeFromWriterLock(ref LockCookie lockCookie)
        {
            lock (this)
            {
                if (!HasWriterLock())
                    throw new Exception("The thread does not have the writer lock.");

                if (lockCookie.WriterLocks != 0)
                    state++;
                else
                {
                    state = lockCookie.ReaderLocks;
                    reader_locks[Thread.CurrentThread.ManagedThreadId] = state;
                    if (readers > 0)
                    {
                        Monitor.PulseAll(this);
                    }
                }

                // MSDN: A thread does not block when downgrading from the writer lock,
                // even if other threads are waiting for the writer lock
            }
        }

        public LockCookie ReleaseLock()
        {
            LockCookie cookie;
            lock (this)
            {
                cookie = GetLockCookie();
                if (cookie.WriterLocks != 0)
                    ReleaseWriterLock(cookie.WriterLocks);
                else if (cookie.ReaderLocks != 0)
                {
                    ReleaseReaderLock(cookie.ReaderLocks, cookie.ReaderLocks);
                }
            }
            return cookie;
        }

        //[ReliabilityContract (Consistency.WillNotCorruptState, Cer.Success)]
        public void ReleaseReaderLock()
        {
            lock (this)
            {
                if (HasWriterLock())
                {
                    ReleaseWriterLock();
                    return;
                }
                else if (state > 0)
                {
                    object read_lock_count = reader_locks[Thread.CurrentThread.ManagedThreadId];
                    if (read_lock_count != null)
                    {
                        ReleaseReaderLock((int)read_lock_count, 1);
                        return;
                    }
                }

                throw new Exception("The thread does not have any reader or writer locks.");
            }
        }

        private void ReleaseReaderLock(int currentCount, int releaseCount)
        {
            int new_count = currentCount - releaseCount;

            if (new_count == 0)
                reader_locks.Remove(Thread.CurrentThread.ManagedThreadId);
            else
                reader_locks[Thread.CurrentThread.ManagedThreadId] = new_count;

            state -= releaseCount;
            if (state == 0 && !writer_queue.IsEmpty)
                writer_queue.Pulse();
        }

        //[ReliabilityContract (Consistency.WillNotCorruptState, Cer.Success)]
        public void ReleaseWriterLock()
        {
            lock (this)
            {
                if (!HasWriterLock())
                    throw new Exception("The thread does not have the writer lock.");

                ReleaseWriterLock(1);
            }
        }

        private void ReleaseWriterLock(int releaseCount)
        {
            state += releaseCount;
            if (state == 0)
            {
                if (readers > 0)
                {
                    Monitor.PulseAll(this);
                }
                else if (!writer_queue.IsEmpty)
                    writer_queue.Pulse();
            }
        }

        public void RestoreLock(ref LockCookie lockCookie)
        {
            lock (this)
            {
                if (lockCookie.WriterLocks != 0)
                    AcquireWriterLock(-1, lockCookie.WriterLocks);
                else if (lockCookie.ReaderLocks != 0)
                    AcquireReaderLock(-1, lockCookie.ReaderLocks);
            }
        }

        public LockCookie UpgradeToWriterLock(int millisecondsTimeout)
        {
            LockCookie cookie;
            lock (this)
            {
                cookie = GetLockCookie();
                if (cookie.WriterLocks != 0)
                {
                    state--;
                    return cookie;
                }

                if (cookie.ReaderLocks != 0)
                    ReleaseReaderLock(cookie.ReaderLocks, cookie.ReaderLocks);
            }

            // Don't do this inside the lock, since it can cause a deadlock.
            AcquireWriterLock(millisecondsTimeout);
            return cookie;
        }

        public LockCookie UpgradeToWriterLock(TimeSpan timeout)
        {
            int ms = CheckTimeout(timeout);
            return UpgradeToWriterLock(ms);
        }

        private LockCookie GetLockCookie()
        {
            LockCookie cookie = new LockCookie(Thread.CurrentThread.ManagedThreadId);

            if (HasWriterLock())
                cookie.WriterLocks = -state;
            else
            {
                object locks = reader_locks[Thread.CurrentThread.ManagedThreadId];
                if (locks != null) cookie.ReaderLocks = (int)locks;
            }
            return cookie;
        }

        private bool HasWriterLock()
        {
            return (state < 0 && Thread.CurrentThread.ManagedThreadId == writer_lock_owner);
        }

        private int CheckTimeout(TimeSpan timeout)
        {
            int ms = (int)timeout.TotalMilliseconds;

            if (ms < -1)
                throw new ArgumentOutOfRangeException("timeout",
                        "Number must be either non-negative or -1");
            return ms;
        }
    }

    public struct LockCookie
    {
        internal int ThreadId;
        internal int ReaderLocks;
        internal int WriterLocks;

        internal LockCookie(int thread_id)
        {
            ThreadId = thread_id;
            ReaderLocks = 0;
            WriterLocks = 0;
        }

        internal LockCookie(int thread_id, int reader_locks, int writer_locks)
        {
            ThreadId = thread_id;
            ReaderLocks = reader_locks;
            WriterLocks = writer_locks;
        }

        public override int GetHashCode()
        {
            return (base.GetHashCode());
        }

        public bool Equals(LockCookie obj)
        {
            if (this.ThreadId == obj.ThreadId &&
                this.ReaderLocks == obj.ReaderLocks &&
                this.WriterLocks == obj.WriterLocks)
            {
                return (true);
            }
            else
            {
                return (false);
            }
        }

        public override bool Equals(Object obj)
        {
            if (!(obj is LockCookie))
            {
                return (false);
            }

            return (obj.Equals(this));
        }

        public static bool operator ==(LockCookie a, LockCookie b)
        {
            return a.Equals(b);
        }

        public static bool operator !=(LockCookie a, LockCookie b)
        {
            return !a.Equals(b);
        }
    }

    // Used for queueing on a reader writer lock
    internal class LockQueue
    {
        private ReaderWriterLock rwlock;
        private int lockCount = 0;

        public LockQueue(ReaderWriterLock rwlock)
        {
            this.rwlock = rwlock;
        }

        public bool Wait(int timeout)
        {
            bool _lock = false;

            try
            {
                lock (this)
                {
                    lockCount++;
                    Monitor.Exit(rwlock);
                    _lock = true;
                    return Monitor.Wait(this, timeout);
                }
            }
            finally
            {
                if (_lock)
                {
                    Monitor.Enter(rwlock);
                    lockCount--;
                }
            }
        }

        public bool IsEmpty
        {
            get { lock (this) return (lockCount == 0); }
        }

        public void Pulse()
        {
            lock (this) Monitor.Pulse(this);
        }
    }
}

#endif