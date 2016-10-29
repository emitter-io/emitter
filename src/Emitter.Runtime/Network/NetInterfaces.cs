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
using System.Threading;
using System.Threading.Tasks;
using Emitter.Network;
using Emitter.Network.Threading;

namespace Emitter
{
    /// <summary>
    /// Defines the contract that represents a remote client.
    /// </summary>
    public interface IClient : IPacketSender, IMqttSender, IDisposable
    {
        /// <summary>
        /// Event that is issued when a client has a first connected channel.
        /// </summary>
        event ClientEvent Connect;

        /// <summary>
        /// Event that is issued when a client was disconnected and about to be disposed.
        /// </summary>
        event ClientEvent Disconnect;

        /// <summary>
        /// Binds a channel to this <see cref="IClient"/> object instance.
        /// </summary>
        /// <param name="channel">The channel to bind to this client.</param>
        void BindChannel(Connection channel);

        /// <summary>
        /// Unbinds a channel from this <see cref="IClient"/> object instance.
        /// </summary>
        /// <param name="channel">The channel to unbind from this client.</param>
        void UnbindChannel(Connection channel);

        /// <summary>
        /// Gets a fabric session token.
        /// </summary>
        string Token { get; }

        /// <summary>
        /// Gets or sets the fabric transport used for this session.
        /// </summary>
        ITransport Transport { get; set; }

        /// <summary>
        /// Gets the transport type currently used in the session.
        /// </summary>
        Network.TransportType TransportType { get; }
    }

    /// <summary>
    /// Represents a contract for a packet sender.
    /// </summary>
    public interface IPacketSender
    {
        /// <summary>
        /// Sends the packet to the remote client end-point.
        /// </summary>
        /// <param name="packet">The packet to send to the remote end-point.</param>
        void Send(Packet packet);

        /// <summary>
        /// Gets the currently active channel for this client. This channel can be used for
        /// sending outgoing messages only.
        /// </summary>
        Connection Channel { get; }
    }

    /// <summary>
    /// Defines a network transport layer listener
    /// </summary>
    public interface IListener
    {
        /// <summary>
        /// Sets the listener in listening state on a particular end-point.
        /// </summary>
        /// <param name="address">End-point to listen to.</param>
        /// <param name="thread">The thread to use for listening.</param>
        Task ListenAsync(ServiceAddress address, EventThread thread);

        /// <summary>
        /// Gets the binding for the listener.
        /// </summary>
        IBinding Binding { get; set; }
    }

    /// <summary>
    /// Defines a binding of an end point, a network listener and a message processor
    /// </summary>
    public interface IBinding
    {
        /// <summary>
        /// Gets the EndPoint used for listening
        /// </summary>
        EndPoint EndPoint { get; }

        /// <summary>
        /// Gets the decoding pipeline.
        /// </summary>
        Processor[] Decoding { get; }

        /// <summary>
        /// Gets the encoding pipeline.
        /// </summary>
        Processor[] Encoding { get; }

        /// <summary>
        /// Gets the schema prefix for the binding.
        /// </summary>
        string Schema { get; }

        /// <summary>
        /// Gets or sets the primary listener context for the binding.
        /// </summary>
        ListenerContext Context { get; set; }
    }

    /// <summary>
    /// Represents a connection contract.
    /// </summary>
    public interface IConnection
    {
        /// <summary>
        /// Pauses the connection.
        /// </summary>
        void Pause();

        /// <summary>
        /// Resumes the connection.
        /// </summary>
        void Resume();

        /// <summary>
        /// Terminates the connection.
        /// </summary>
        void Close();

        /// <summary>
        /// Writes some data to the connection.
        /// </summary>
        void Write(ArraySegment<byte> data);

        /// <summary>
        /// Writes some data to the connection.
        /// </summary>
        Task WriteAsync(ArraySegment<byte> data, CancellationToken cancellationToken);

        /// <summary>
        /// Flushes the connection.
        /// </summary>
        void Flush();

        /// <summary>
        /// Flushes the connection.
        /// </summary>
        Task FlushAsync(CancellationToken cancellationToken);

        /// <summary>
        /// Event that is issued when a channel was disconnected and about to be disposed.
        /// </summary>
        event ChannelEvent Disconnect;
    }
}