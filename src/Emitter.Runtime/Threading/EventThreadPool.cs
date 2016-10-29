// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Threading;
using System.Threading.Tasks;

namespace Emitter.Network.Threading
{
    public class EventThreadPool : IThreadPool
    {
        private readonly WaitCallback _runAction;
        private readonly WaitCallback _cancelTcs;
        private readonly WaitCallback _completeTcs;

        public EventThreadPool()
        {
            // Curry and capture log in closures once
            _runAction = (o) =>
            {
                try
                {
                    ((Action)o)();
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                }
            };

            _completeTcs = (o) =>
            {
                try
                {
                    ((TaskCompletionSource<object>)o).TrySetResult(null);
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                }
            };

            _cancelTcs = (o) =>
            {
                try
                {
                    ((TaskCompletionSource<object>)o).TrySetCanceled();
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                }
            };
        }

        public void Run(Action action)
        {
            ThreadPool.QueueUserWorkItem(_runAction, action);
        }

        public void Complete(TaskCompletionSource<object> tcs)
        {
            ThreadPool.QueueUserWorkItem(_completeTcs, tcs);
        }

        public void Cancel(TaskCompletionSource<object> tcs)
        {
            ThreadPool.QueueUserWorkItem(_cancelTcs, tcs);
        }

        public void Error(TaskCompletionSource<object> tcs, Exception ex)
        {
            // ex and _log are closure captured
            ThreadPool.QueueUserWorkItem((o) =>
            {
                try
                {
                    ((TaskCompletionSource<object>)o).TrySetException(ex);
                }
                catch (Exception e)
                {
                    Service.Logger.Log(e);
                }
            }, tcs);
        }
    }

    public interface IThreadPool
    {
        void Complete(TaskCompletionSource<object> tcs);

        void Cancel(TaskCompletionSource<object> tcs);

        void Error(TaskCompletionSource<object> tcs, Exception ex);

        void Run(Action action);
    }
}