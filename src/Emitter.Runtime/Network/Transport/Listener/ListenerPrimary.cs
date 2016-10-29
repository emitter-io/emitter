// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Collections.Generic;
using System.Threading.Tasks;

using Emitter.Network.Native;
using Emitter.Network.Threading;

namespace Emitter.Network
{
    /// <summary>
    /// A primary listener waits for incoming connections on a specified socket. Incoming
    /// connections may be passed to a secondary listener to handle.
    /// </summary>
    internal abstract class ListenerPrimary : Listener
    {
        private readonly List<UvPipeHandle> _dispatchPipes = new List<UvPipeHandle>();
        private int _dispatchIndex;
        private string _pipeName;

        // this message is passed to write2 because it must be non-zero-length,
        // but it has no other functional significance
        private readonly ArraySegment<ArraySegment<byte>> _dummyMessage = new ArraySegment<ArraySegment<byte>>(new[] { new ArraySegment<byte>(new byte[] { 1, 2, 3, 4 }) });

        protected ListenerPrimary(ServiceContext serviceContext) : base(serviceContext)
        {
        }

        private UvPipeHandle ListenPipe { get; set; }

        public async Task StartAsync(
            string pipeName,
            ServiceAddress address,
            EventThread thread)
        {
            _pipeName = pipeName;

            await ListenAsync(address, thread).ConfigureAwait(false);

            await Thread.PostAsync(state => ((ListenerPrimary)state).PostCallback(),
                                   this).ConfigureAwait(false);
        }

        private void PostCallback()
        {
            ListenPipe = new UvPipeHandle();
            ListenPipe.Init(Thread.Loop, Thread.QueueCloseHandle, false);
            ListenPipe.Bind(_pipeName);
            ListenPipe.Listen(Constants.ListenBacklog,
                (pipe, status, error, state) => ((ListenerPrimary)state).OnListenPipe(pipe, status, error), this);
        }

        private void OnListenPipe(UvStreamHandle pipe, int status, Exception error)
        {
            if (status < 0)
            {
                return;
            }

            var dispatchPipe = new UvPipeHandle();
            dispatchPipe.Init(Thread.Loop, Thread.QueueCloseHandle, true);

            try
            {
                pipe.Accept(dispatchPipe);
            }
            catch (UvException ex)
            {
                dispatchPipe.Dispose();
                Service.Logger.Log(ex);
                return;
            }

            _dispatchPipes.Add(dispatchPipe);
        }

        protected override void DispatchConnection(UvStreamHandle socket)
        {
            var index = _dispatchIndex++ % (_dispatchPipes.Count + 1);
            if (index == _dispatchPipes.Count)
            {
                base.DispatchConnection(socket);
            }
            else
            {
                var dispatchPipe = _dispatchPipes[index];
                var write = new UvWriteReq();
                write.Init(Thread.Loop);
                write.Write2(
                    dispatchPipe,
                    _dummyMessage,
                    socket,
                    (write2, status, error, state) =>
                    {
                        write2.Dispose();
                        ((UvStreamHandle)state).Dispose();
                    },
                    socket);
            }
        }

        public override async Task DisposeAsync()
        {
            // Call base first so the ListenSocket gets closed and doesn't
            // try to dispatch connections to closed pipes.
            await base.DisposeAsync().ConfigureAwait(false);

            if (Thread.FatalError == null && ListenPipe != null)
            {
                await Thread.PostAsync(state =>
                {
                    var listener = (ListenerPrimary)state;
                    listener.ListenPipe.Dispose();

                    foreach (var dispatchPipe in listener._dispatchPipes)
                    {
                        dispatchPipe.Dispose();
                    }
                }, this).ConfigureAwait(false);
            }
        }
    }
}