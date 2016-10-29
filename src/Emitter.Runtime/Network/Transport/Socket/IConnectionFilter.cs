// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System.IO;
using System.Threading.Tasks;

namespace Emitter.Network
{
    public interface IConnectionFilter
    {
        Task OnConnectionAsync(ConnectionFilterContext context);
    }

    public class ConnectionFilterContext
    {
        public ServiceAddress Address { get; set; }
        public Stream Connection { get; set; }
    }

    public interface IBufferSizeControl
    {
        void Add(int count);

        void Subtract(int count);
    }
}