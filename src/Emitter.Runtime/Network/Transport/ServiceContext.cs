// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using Emitter.Network.Threading;
using Emitter.Providers;

namespace Emitter.Network
{
    public class ServiceContext
    {
        public ServiceContext()
        {
        }

        public ServiceContext(ServiceContext context)
        {
            this.ThreadPool = context.ThreadPool;
            this.Binding = context.Binding;
        }

        /// <summary>
        /// Gets or sets the threadpool used by the service.
        /// </summary>
        public IThreadPool ThreadPool { get; set; }

        /// <summary>
        /// Gets the binding configured for this context.
        /// </summary>
        public IBinding Binding { get; set; }

        /// <summary>
        /// Gets or sets the options for the service
        /// </summary>
        public TransportProvider Options
        {
            get { return Service.Transport; }
        }
    }
}