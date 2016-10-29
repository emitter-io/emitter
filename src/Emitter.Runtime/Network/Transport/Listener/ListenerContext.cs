// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System.Collections.Generic;

using Emitter.Network.Native;
using Emitter.Network.Threading;

namespace Emitter.Network
{
    public class ListenerContext : ServiceContext
    {
        public ListenerContext()
        {
        }

        public ListenerContext(ServiceContext serviceContext)
            : base(serviceContext)
        {
            // Set memory pool
            Memory = new MemoryPool();
            WriteReqPool = new Queue<UvWriteReq>(SocketOutput.MaxPooledWriteReqs);
        }

        public ListenerContext(ListenerContext listenerContext)
            : base(listenerContext)
        {
            ServerAddress = listenerContext.ServerAddress;
            Thread = listenerContext.Thread;
            Memory = listenerContext.Memory;
            ConnectionManager = listenerContext.ConnectionManager;
            WriteReqPool = listenerContext.WriteReqPool;
        }

        public ServiceAddress ServerAddress { get; set; }

        public EventThread Thread { get; set; }

        public MemoryPool Memory { get; set; }

        internal ConnectionManager ConnectionManager { get; set; }

        internal Queue<UvWriteReq> WriteReqPool { get; set; }
    }
}