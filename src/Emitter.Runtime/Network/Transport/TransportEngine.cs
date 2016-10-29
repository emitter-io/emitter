// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Emitter.Network.Native;
using Emitter.Network.Threading;

namespace Emitter.Network
{
    public class TransportEngine : ServiceContext, IDisposable
    {
        public TransportEngine(ServiceContext context)
            : this(new Libuv(), context)
        { }

        // For testing
        internal TransportEngine(Libuv uv, ServiceContext context)
           : base(context)
        {
            Libuv = uv;
            Threads = new List<EventThread>();
        }

        internal Libuv Libuv { get; private set; }
        public List<EventThread> Threads { get; private set; }

        public void Start(int count)
        {
            for (var index = 0; index < count; index++)
            {
                Threads.Add(new EventThread(this));
            }

            foreach (var thread in Threads)
            {
                thread.StartAsync().Wait();
            }
        }

        public void Dispose()
        {
            foreach (var thread in Threads)
            {
                thread.Stop(TimeSpan.FromSeconds(2.5));
            }
            Threads.Clear();
        }

        public IDisposable CreateServer(ServiceAddress address, IBinding binding)
        {
            var listeners = new List<IAsyncDisposable>();
            var usingPipes = address.IsUnixPipe;

            try
            {
                var pipeName = (Libuv.IsWindows ? @"\\.\pipe\emitter_" : "/tmp/emitter_") + Guid.NewGuid().ToString("n");
                var single = Threads.Count == 1;
                var first = true;

                foreach (var thread in Threads)
                {
                    if (single)
                    {
                        var listener = new TcpListener(this);
                        listeners.Add(listener);
                        listener.Binding = binding;
                        listener.ListenAsync(address, thread).Wait();
                        binding.Context = listener;
                    }
                    else if (first)
                    {
                        var listener = new TcpListenerPrimary(this);
                        listeners.Add(listener);
                        listener.Binding = binding;
                        listener.StartAsync(pipeName, address, thread).Wait();
                        binding.Context = listener;
                    }
                    else
                    {
                        var listener = new TcpListenerSecondary(this);
                        listeners.Add(listener);
                        listener.Binding = binding;
                        listener.StartAsync(pipeName, address, thread).Wait();
                    }

                    first = false;
                }

                return new Disposable(() =>
                {
                    DisposeListeners(listeners);
                });
            }
            catch
            {
                DisposeListeners(listeners);

                throw;
            }
        }

        private void DisposeListeners(List<IAsyncDisposable> listeners)
        {
            var disposeTasks = new List<Task>();

            foreach (var listener in listeners)
            {
                disposeTasks.Add(listener.DisposeAsync());
            }

            if (!Task.WhenAll(disposeTasks).Wait(Options.ShutdownTimeout))
            {
                //Log.NotAllConnectionsClosedGracefully();
            }
        }
    }
}