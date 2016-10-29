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
using Emitter.Diagnostics;
using ITransportType = Emitter.Network.TransportType;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a client that is bound to a single connection.
    /// </summary>
    public sealed class Client : IClient
    {
        #region Private Fields
        private ITransport fTransport = null;
        private ITransportType fType = ITransportType.Tcp;
        private object Lock = new object();
        private bool ConnectInvoked = false;
        private ConnectionId ConnectionId;
        #endregion Private Fields

        #region IClient Members

        /// <summary>
        /// Event that is issued when a client has a first connected channel.
        /// </summary>
        public event ClientEvent Connect;

        /// <summary>
        /// Event that is issued when a client was disconnected and about to be disposed.
        /// </summary>
        public event ClientEvent Disconnect;

        /// <summary>
        /// Binds a channel to this <see cref="IClient"/> object instance.
        /// </summary>
        /// <param name="channel">The channel to bind to this client.</param>
        public void BindChannel(Emitter.Connection channel)
        {
            lock (this.Lock)
            {
                // Should we dispatch the event?
                var dispatchEvent = (this.Channel == null && !this.ConnectInvoked);

                // If we don't have a transport yet, mark this as TCP
                if (this.Transport == null)
                    this.Transport = new TcpTransport();

                // Set the channel within the transport
                this.Transport.Channel = channel;
                this.ConnectionId = channel.ConnectionId;
                NetTrace.WriteLine("Binding to client " + this, channel, NetTraceCategory.Channel);

                // Dispatch the event once we have the transport bound
                if (dispatchEvent)
                {
                    // Flag as done
                    this.ConnectInvoked = true;

                    // Invoke connected, on first connection only
                    Service.InvokeClientConnect(new ClientConnectEventArgs(this));

                    // Fire the connected event locally
                    this.Connect?.Invoke(this);
                }
            }
        }

        /// <summary>
        /// Unbinds a channel from this <see cref="IClient"/> object instance.
        /// </summary>
        /// <param name="channel">The channel to unbind from this client.</param>
        public void UnbindChannel(Emitter.Connection channel)
        {
            // If the channel is not bound, ignore the unbind
            if (this.Transport.Channel != channel)
                return;

            NetTrace.WriteLine("Unbinding from client " + this, channel, NetTraceCategory.Channel);
            lock (this.Lock)
            {
                // Set the channel
                if (this.Transport != null)
                    this.Transport.Channel = null;
            }

            // As this is a single connection client, we simply dispose it when the connection is unbound
            switch (this.TransportType)
            {
                case ITransportType.Tcp:
                    this.Dispose();
                    break;
            }
        }

        #endregion IClient Members

        #region Send Members

        /// <summary>
        /// Sends the packet to the remote client end-point.
        /// </summary>
        /// <param name="packet">The packet to send to the remote end-point.</param>
        public void Send(Packet packet)
        {
            // If there is an active channel, forwards the send to the underlying connection.
            // Otherwise we just ignore because the connection is probably closed.
            var channel = this.Channel;
            if (channel != null)
            {
                channel.Send(packet);
                return;
            }

            // Make sure we have a session attached
            NetTrace.WriteLine("Attempting a send through an invalid connection.", channel, NetTraceCategory.Channel);
        }

        /// <summary>
        /// Gets the currently active channel for this client. This channel can be used for
        /// sending outgoing messages only.
        /// </summary>
        public Connection Channel
        {
            get
            {
                // Gets the channel from the transport
                if (this.fTransport == null)
                    return null;

                // Get the channel
                var channel = this.fTransport.Channel;
                if (channel == null || !channel.IsRunning)
                    return null;

                return channel;
            }
        }

        #endregion Send Members

        #region IFabricSession Properties

        /// <summary>
        /// Gets or sets a session token.
        /// </summary>
        public string Token
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets or sets the transport used for this session.
        /// </summary>
        public ITransport Transport
        {
            get { return this.fTransport; }
            set
            {
                // Set the transport
                this.fTransport = value;

                // Cache the type, we want to pay this only once
                if (value == null)
                    this.fType = ITransportType.Tcp;
                if (value is TcpTransport)
                    this.fType = ITransportType.Tcp;
            }
        }

        /// <summary>
        /// Gets the transport type currently used in the session.
        /// </summary>
        public ITransportType TransportType
        {
            get { return fType; }
        }

        /// <summary>
        /// Returns an IP Address of the remote transport channel (if any).
        /// </summary>
        public string Address
        {
            get
            {
                try
                {
                    return this.Transport.Channel.RemoteEndPoint.ToString();
                }
                catch
                {
                    return null;
                }
            }
        }

        /// <summary>
        /// Converts the value of the current object to its equivalent string representation.
        /// </summary>
        /// <returns>A string representation of the current object.</returns>
        public override string ToString()
        {
            if (this.Transport == null)
                return "[null]";

            if (this.TransportType == ITransportType.Tcp)
                return "[" + this.TransportType + ", " + this.Address + "]";

            return "[" + this.TransportType + ", " + this.Token + "]";
        }

        #endregion IFabricSession Properties

        #region IMqttSender Members

        /// <summary>
        /// Gets the connection id of the sender.
        /// </summary>
        public ConnectionId Id
        {
            get { return this.ConnectionId; }
        }

        /// <summary>
        /// Gets or sets the MQTT context associated with this sender.
        /// </summary>
        public MqttContext Context
        {
            get;
            set;
        }

        /// <summary>
        /// Sends the message to the remote client, remote server or a function.
        /// </summary>
        /// <param name="contract">The contract for this message.</param>
        /// <param name="channel">The channel name to send.</param>
        /// <param name="message">The message to send.</param>
        public void Send(int contract, string channel, ArraySegment<byte> message)
        {
            // Send the message out to the client.
            var msg = MqttPublishPacket.Acquire();
            msg.Channel = channel;
            msg.Message = message;
            this.Send(msg);
        }

        #endregion IMqttSender Members

        #region IDisposable Members

        /// <summary>
        /// Performs application-defined tasks associated with freeing, releasing, or
        /// resetting unmanaged resources.
        /// </summary>
        private void Dispose(bool disposing)
        {
            // Invokes a disconnected event
            Service.InvokeClientDisconnect(new ClientDisconnectEventArgs(this));

            // Fire the disconnected event locally
            this.Disconnect?.Invoke(this);

            try
            {
                // We should also dispose the transports attached
                this.fTransport?.Dispose();
            }
            catch { } // Never throw in a dispose
        }

        /// <summary>
        /// Performs application-defined tasks associated with freeing, releasing, or
        /// resetting unmanaged resources.
        /// </summary>
        public void Dispose()
        {
            this.Dispose(true);
            GC.SuppressFinalize(this);
        }

        /// <summary>
        /// Finalizer, cleans up the connection if Dispose was never called.
        /// </summary>
        ~Client()
        {
            this.Dispose(false);
        }

        #endregion IDisposable Members
    }
}