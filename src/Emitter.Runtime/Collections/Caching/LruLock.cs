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
using System.Threading;

namespace Emitter.Collections
{
    internal enum LruLockStatus
    {
        Unlocked,
        ReadLock,
        WriteLock
    }

    internal class LruLock : IDisposable
    {
        public delegate ResultType DoWorkFunc<ResultType>();

        public static int defaultTimeout = 30000;
        private LruLockStatus status = LruLockStatus.Unlocked;
        private ReaderWriterLock lockObj;
        private int timeout;
        private LockCookie cookie;
        private bool upgraded = false;

        #region delegate based methods

        public static ResultType GetWriteLock<ResultType>(ReaderWriterLock lockObj, int timeout, DoWorkFunc<ResultType> doWork)
        {
            var status = (lockObj.IsWriterLockHeld
                ? LruLockStatus.WriteLock
                : (lockObj.IsReaderLockHeld ? LruLockStatus.ReadLock : LruLockStatus.Unlocked));

            LockCookie writeLock = default(LockCookie);
            if (status == LruLockStatus.ReadLock)
                writeLock = lockObj.UpgradeToWriterLock(timeout);
            else if (status == LruLockStatus.Unlocked)
                lockObj.AcquireWriterLock(timeout);
            try
            {
                return doWork();
            }
            finally
            {
                if (status == LruLockStatus.ReadLock)
                    lockObj.DowngradeFromWriterLock(ref writeLock);
                else if (status == LruLockStatus.Unlocked)
                    lockObj.ReleaseWriterLock();
            }
        }

        public static ResultType GetReadLock<ResultType>(ReaderWriterLock lockObj, int timeout, DoWorkFunc<ResultType> doWork)
        {
            bool releaseLock = false;
            if (!lockObj.IsWriterLockHeld && !lockObj.IsReaderLockHeld)
            {
                lockObj.AcquireReaderLock(timeout);
                releaseLock = true;
            }
            try
            {
                return doWork();
            }
            finally
            {
                if (releaseLock)
                    lockObj.ReleaseReaderLock();
            }
        }

        #endregion delegate based methods

        #region disposable based methods

        public static LruLock GetReadLock(ReaderWriterLock lockObj)
        {
            return new LruLock(lockObj, LruLockStatus.ReadLock, defaultTimeout);
        }

        public static LruLock GetWriteLock(ReaderWriterLock lockObj)
        {
            return new LruLock(lockObj, LruLockStatus.WriteLock, defaultTimeout);
        }

        public LruLock(ReaderWriterLock lockObj, LruLockStatus status, int timeoutMS)
        {
            this.lockObj = lockObj;
            this.timeout = timeoutMS;
            this.Status = status;
        }

        public void Dispose()
        {
            Status = LruLockStatus.Unlocked;
        }

        public LruLockStatus Status
        {
            get
            {
                return status;
            }
            set
            {
                if (status != value)
                {
                    if (status == LruLockStatus.Unlocked)
                    {
                        upgraded = false;
                        if (value == LruLockStatus.ReadLock)
                            lockObj.AcquireReaderLock(timeout);
                        else if (value == LruLockStatus.WriteLock)
                            lockObj.AcquireWriterLock(timeout);
                    }
                    else if (value == LruLockStatus.Unlocked)
                        lockObj.ReleaseLock();
                    else if (value == LruLockStatus.WriteLock) // && status==RWLockStatus.READ_LOCK
                    {
                        cookie = lockObj.UpgradeToWriterLock(timeout);
                        upgraded = true;
                    }
                    else if (upgraded) // value==RWLockStatus.READ_LOCK && status==RWLockStatus.WRITE_LOCK
                    {
                        lockObj.DowngradeFromWriterLock(ref cookie);
                        upgraded = false;
                    }
                    else
                    {
                        lockObj.ReleaseLock();
                        status = LruLockStatus.Unlocked;
                        lockObj.AcquireReaderLock(timeout);
                    }
                    status = value;
                }
            }
        }

        #endregion disposable based methods
    }
}