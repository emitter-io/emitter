// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

namespace System.Threading.Tasks
{
    public static class TaskUtilities
    {
#if DOTNET
        public static Task CompletedTask = Task.CompletedTask;
#else
        public static Task CompletedTask = Task.FromResult<object>(null);
#endif
        public static Task<int> ZeroTask = Task.FromResult(0);

        public static Task GetCancelledTask(CancellationToken cancellationToken)
        {
#if DOTNET
            return Task.FromCanceled(cancellationToken);
#else
            var tcs = new TaskCompletionSource<object>();
            tcs.TrySetCanceled();
            return tcs.Task;
#endif
        }

        public static Task<int> GetCancelledZeroTask()
        {
            // Task<int>.FromCanceled doesn't return Task<int>
            var tcs = new TaskCompletionSource<int>();
            tcs.TrySetCanceled();
            return tcs.Task;
        }
    }
}