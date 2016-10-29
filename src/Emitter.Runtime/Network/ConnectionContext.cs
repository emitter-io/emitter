// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Net;

namespace Emitter.Network
{
    /// <summary>
    /// Represents the connection context.
    /// </summary>
    public class ConnectionContext : ListenerContext
    {
        protected ProcessingSettings DecodingSettings;
        protected ProcessingSettings EncodingSettings;

        public ConnectionContext()
        {
        }

        public ConnectionContext(ListenerContext context) : base(context)
        {
            this.DecodingSettings = new ProcessingSettings(ProcessingType.Decoding, context.Binding.Decoding);
            this.EncodingSettings = new ProcessingSettings(ProcessingType.Encoding, context.Binding.Encoding);
        }

        public ConnectionContext(ConnectionContext context) : base(context)
        {
            this.SocketInput = context.SocketInput;
            this.SocketOutput = context.SocketOutput;
            this.RemoteEndPoint = context.RemoteEndPoint;
            this.LocalEndPoint = context.LocalEndPoint;
            this.ConnectionId = context.ConnectionId;
            this.DecodingSettings = new ProcessingSettings(ProcessingType.Decoding, context.Binding.Decoding);
            this.EncodingSettings = new ProcessingSettings(ProcessingType.Encoding, context.Binding.Encoding);
        }

        public SocketInput SocketInput { get; set; }

        public ISocketOutput SocketOutput { get; set; }

        public IPEndPoint RemoteEndPoint { get; set; }

        public IPEndPoint LocalEndPoint { get; set; }

        public int MeshIdentifier { get; set; }

        public ConnectionId ConnectionId { get; set; }

        public IClient Client { get; set; }

        /// <summary>
        /// Gets the <see cref="ProcessingSettings"/> used to encode outgoing messages.
        /// </summary>
        public ProcessingSettings Encoding
        {
            get { return this.EncodingSettings; }
        }

        /// <summary>
        /// Gets the <see cref="ProcessingSettings"/> used to decode incoming messages.
        /// </summary>
        public ProcessingSettings Decoding
        {
            get { return this.DecodingSettings; }
        }
    }
}