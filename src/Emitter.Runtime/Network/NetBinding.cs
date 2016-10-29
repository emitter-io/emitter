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
using System.Linq;
using System.Net;
using Emitter.Network;

namespace Emitter
{
    /// <summary>
    /// Defines a network binding
    /// </summary>
    public abstract class NetBinding : IBinding
    {
        #region Default Pipelines

        /// <summary>
        /// Default encoders to apply for every newly acquired context.
        /// </summary>
        public static readonly Processor[] DefaultEncode = new[]{
            Encode.Http,
            Encode.String,
            Encode.Byte,
            Encode.Mqtt
        };

        /// <summary>
        /// Default decoders to apply for every newly acquired context.
        /// </summary>
        public static readonly Processor[] DefaultDecode = new[]{
            Decode.Http,
            Decode.Mqtt
        };

        /// <summary>
        /// Default mesh encoders to apply for every newly acquired context.
        /// </summary>
        public static readonly Processor[] MeshEncode = new[]{
            Encode.Mesh
        };

        /// <summary>
        /// Default mesh decoders to apply for every newly acquired context.
        /// </summary>
        public static readonly Processor[] MeshDecode = new[]{
            Decode.Mesh
        };

        #endregion Default Pipelines

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="endPoint">The local end-point to create the binding for.</param>
        /// <param name="listener">The listener used to accept incoming connections.</param>
        /// <param name="decoding">The decoding pipeline to set.</param>
        /// <param name="encoding">The encoding pipeline to set.</param>
        public NetBinding(EndPoint endPoint, Processor[] decoding = null, Processor[] encoding = null)
        {
            this.EndPoint = endPoint;
            this.Decoding = decoding ?? DefaultDecode;
            this.Encoding = encoding ?? DefaultEncode;
        }

        /// <summary>
        /// Gets or sets the service address
        /// </summary>
        public ServiceAddress[] Addresses
        {
            get;
            set;
        }

        /// <summary>
        /// Gets the EndPoint used for listening
        /// </summary>
        public EndPoint EndPoint
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets the decoding pipeline.
        /// </summary>
        public Processor[] Decoding
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets the encoding pipeline.
        /// </summary>
        public Processor[] Encoding
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets or sets the primary listener context for the binding.
        /// </summary>
        public ListenerContext Context
        {
            get;
            set;
        }

        /// <summary>
        /// Gets the schema prefix for the binding.
        /// </summary>
        public abstract string Schema
        {
            get;
        }

        /// <summary>
        /// Convers the <see cref="NetBinding"/> to a string representation.
        /// </summary>
        /// <returns>A string representation of <see cref="NetBinding"/> instance.</returns>
        public override string ToString()
        {
            return String.Format("Binding: {0}", this.EndPoint);
        }
    }

    /// <summary>
    /// Represents a TCP/IP Binding with a custom message processor.
    /// </summary>
    public class TcpBinding : NetBinding
    {
        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="endPoint">The local end-point to create the binding for.</param>
        public TcpBinding(EndPoint endPoint)
            : base(endPoint) { }

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="endPoint">The local end-point to create the binding for.</param>
        /// <param name="decoding">The decoding pipeline to set.</param>
        /// <param name="encoding">The encoding pipeline to set.</param>
        public TcpBinding(EndPoint endPoint, Processor[] decoding, Processor[] encoding)
            : base(endPoint, decoding, encoding)
        { }

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="address">The local address to create the binding for.</param>
        /// <param name="port">The local port to create the binding for.</param>
        public TcpBinding(IPAddress address, int port)
            : base(new IPEndPoint(address, port)) { }

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="port">The local port to create the binding for.</param>
        public TcpBinding(int port)
            : base(new IPEndPoint(IPAddress.Any, port)) { }

        /// <summary>
        /// Gets the schema prefix for the binding.
        /// </summary>
        public override string Schema
        {
            get { return "tcp"; }
        }
    }

    /// <summary>
    /// Represents a Secure TCP/IP Binding with a custom message processor.
    /// </summary>
    public class TlsBinding : NetBinding
    {
        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="endPoint">The local end-point to create the binding for.</param>
        public TlsBinding(EndPoint endPoint)
            : base(endPoint) { }

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="endPoint">The local end-point to create the binding for.</param>
        /// <param name="decoding">The decoding pipeline to set.</param>
        /// <param name="encoding">The encoding pipeline to set.</param>
        public TlsBinding(EndPoint endPoint, Processor[] decoding, Processor[] encoding)
            : base(endPoint, decoding, encoding)
        { }

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="address">The local address to create the binding for.</param>
        /// <param name="port">The local port to create the binding for.</param>
        public TlsBinding(IPAddress address, int port)
            : base(new IPEndPoint(address, port)) { }

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="port">The local port to create the binding for.</param>
        public TlsBinding(int port)
            : base(new IPEndPoint(IPAddress.Any, port)) { }

        /// <summary>
        /// Gets the schema prefix for the binding.
        /// </summary>
        public override string Schema
        {
            get { return "https"; }
        }
    }

    /// <summary>
    /// Represents a TCP/IP Mesh Binding.
    /// </summary>
    public sealed class MeshBinding : TcpBinding
    {
        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="endPoint">The local end-point to create the binding for.</param>
        public MeshBinding(EndPoint endPoint)
            : base(endPoint, MeshDecode, MeshEncode) { }

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="address">The local address to create the binding for.</param>
        /// <param name="port">The local port to create the binding for.</param>
        public MeshBinding(IPAddress address, int port)
            : base(new IPEndPoint(address, port), MeshDecode, MeshEncode) { }

        /// <summary>
        /// Creates a new network binding.
        /// </summary>
        /// <param name="port">The local port to create the binding for.</param>
        public MeshBinding(int port)
            : base(new IPEndPoint(IPAddress.Any, port), MeshDecode, MeshEncode) { }

        /// <summary>
        /// Gets the schema prefix for the binding.
        /// </summary>
        public override string Schema
        {
            get { return "tcp"; }
        }
    }
}