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
using System.Net.Sockets;
using Emitter.Network.Mesh;
using Emitter.Security;

namespace Emitter
{
    /// <summary>
    /// Represents a set of event codes used for emitter.
    /// </summary>
    public enum EmitterEventCode
    {
        Message = 100,
        Success = 200,
        BadRequest = 400,
        Unauthorized = 401,
        PaymentRequired = 402,
        Forbidden = 403,
        NotFound = 404,
        ServerError = 500,
        NotImplemented = 501,
    }

    /// <summary>
    /// Represents the method that will handle the cluster event of a <see cref="Service"/>.
    /// </summary>
    public delegate void ClusterEventHandler(ClusterEventArgs e);

    /// <summary>
    /// Represents the method that will handle the Started event of a <see cref="Service"/>.
    /// </summary>
    public delegate void ServerStartupEventHandler();

    /// <summary>
    /// Represents the method that will handle the Stopped event of a <see cref="Service"/>.
    /// </summary>
    public delegate void ServerShutdownEventHandler(ServerShutdownEventArgs e);

    /// <summary>
    /// Represents the method that will handle the ClientConnect event of a <see cref="Service"/>.
    /// </summary>
    public delegate void ClientConnectEventHandler(ClientConnectEventArgs e);

    /// <summary>
    /// Represents the method that will handle the ClientDisconnect event of a <see cref="Service"/>.
    /// </summary>
    public delegate void ClientDisconnectEventHandler(ClientDisconnectEventArgs e);

    /// <summary>
    /// Represents the method that will handle the SocketConnect event of a <see cref="Service"/>.
    /// </summary>
    public delegate void SocketConnectEventHandler(SocketConnectEventArgs e);

    /// <summary>
    /// Represents the method that will handle the ProviderRegistered event of a <see cref="Service"/>.
    /// </summary>
    public delegate void ProviderRegisteredEventHandler(ProviderRegisteredEventArgs e);

    /// <summary>
    /// Represents the method that will handle the TimeSlice event of a <see cref="Service"/>.
    /// </summary>
    public delegate void TimerSliceEventHandler();

    /// <summary>
    /// Represents the method that will handle the UnhandledTimerException event of a <see cref="Service"/>.
    /// </summary>
    public delegate void TimerExceptionHandler(TimerExceptionEventArgs e);

    /// <summary>
    /// Represehts a handler invoked when a new message is published/sent on the channel.
    /// </summary>
    /// <param name="contract">The contract for this message.</param>
    /// <param name="channel">The channel for this message.</param>
    /// <param name="length">The length of the message, in bytes.</param>
    public delegate void MonitoringHandler(IContract contract, string channel, int length);

    #region Event Args

    /// <summary>
    /// Provides data for the node join/leave events.
    /// </summary>
    public sealed class ClusterEventArgs : EventArgs
    {
        /// <summary>
        /// Gets the node that joins or leaves the cluster.
        /// </summary>
        public MeshMember Node { get; private set; }

        /// <summary>
        /// Creates a new instance of the <see cref="ClusterEventArgs"/> object.
        /// </summary>
        /// <param name="n">Node that joins or leaves the cluster.</param>
        public ClusterEventArgs(MeshMember n)
        {
            Node = n;
        }
    }

    /// <summary>
    /// Provides data for the SocketConnect event.
    /// </summary>
    public sealed class SocketConnectEventArgs : EventArgs
    {
        private Socket fSocket;
        private bool fAllowConnection;

        /// <summary>
        /// Gets the socket that have just been connected.
        /// </summary>
        public Socket Socket { get { return fSocket; } }

        /// <summary>
        /// Gets or sets whether the connection for the socket should be allowed or not.
        /// </summary>
        public bool AllowConnection { get { return fAllowConnection; } set { fAllowConnection = value; } }

        /// <summary>
        /// Creates a new instance of the <see cref="SocketConnectEventArgs"/> object.
        /// </summary>
        /// <param name="s">Socket that have just been connected.</param>
        public SocketConnectEventArgs(Socket s)
        {
            fSocket = s;
            fAllowConnection = true;
        }
    }

    /// <summary>
    /// Provides data for the ClientConnect event.
    /// </summary>
    public sealed class ClientConnectEventArgs : EventArgs
    {
        /// <summary>
        /// Gets the <see cref="IClient"/> with whom the connection was established.
        /// </summary>
        public IClient Client { get; private set; }

        /// <summary>
        /// Creates a new instance of the <see cref="ClientConnectEventArgs"/> object.
        /// </summary>
        /// <param name="client">The <see cref="IClient"/> with whom the connection was established.</param>
        public ClientConnectEventArgs(IClient client)
        {
            Client = client;
        }
    }

    /// <summary>
    /// Provides data for the ClientDisconnect event.
    /// </summary>
    public sealed class ClientDisconnectEventArgs : EventArgs
    {
        /// <summary>
        /// Gets the <see cref="IClient"/> with whom the connection was broken.
        /// </summary>
        public IClient Client { get; private set; }

        /// <summary>
        /// Creates a new instance of the <see cref="ClientDisconnectEventArgs"/> object.
        /// </summary>
        /// <param name="client">The <see cref="IClient"/> with whom the connection was broken.</param>
        public ClientDisconnectEventArgs(IClient client)
        {
            this.Client = client;
        }
    }

    /// <summary>
    /// Provides data for the Stopped event.
    /// </summary>
    public sealed class ServerShutdownEventArgs : EventArgs
    {
        /// <summary>
        /// Creates a new instance of the <see cref="ServerShutdownEventArgs"/> object.
        /// </summary>
        public ServerShutdownEventArgs()
        {
            this.IsCrash = false;
        }

        /// <summary>
        /// Creates a new instance of the <see cref="ServerShutdownEventArgs"/> object.
        /// </summary>
        /// <param name="e">The exception that caused the crash.</param>
        public ServerShutdownEventArgs(Exception e)
        {
            this.Exception = e;
            this.IsCrash = true;
        }

        /// <summary>
        /// Gets the exception that caused the shutdown. Can be null.
        /// </summary>
        public Exception Exception
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets or sets whether the service must be closed.
        /// </summary>
        public bool Close
        {
            get;
            set;
        }

        /// <summary>
        /// Gets whether the shutdown was due to a crash or not.
        /// </summary>
        public bool IsCrash
        {
            get;
            private set;
        }
    }

    /// <summary>
    /// Provides data for the ProviderRegistered event.
    /// </summary>
    public sealed class ProviderRegisteredEventArgs : EventArgs
    {
        /// <summary>
        /// Creates a new instance of the <see cref="ProviderRegisteredEventArgs"/> object.
        /// </summary>
        /// <param name="providerBaseType">The type of the provider that was registered.</param>
        /// <param name="newProvider">The provider that performs the registration.</param>
        public ProviderRegisteredEventArgs(Type providerBaseType, Provider newProvider)
        {
            this.ProviderBaseType = providerBaseType;
            this.NewProvider = newProvider;
        }

        /// <summary>
        /// Gets the type of the provider that was registered
        /// </summary>
        public Type ProviderBaseType
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets the provider that performs the registration.
        /// </summary>
        public Provider NewProvider
        {
            get;
            private set;
        }
    }

    /// <summary>
    /// Provides data for the UnhandledTimerException event.
    /// </summary>
    public sealed class TimerExceptionEventArgs : EventArgs
    {
        private Timer fTimer;
        private bool fStopTimer;
        private bool fHandled;
        private Exception fException;

        /// <summary>
        /// Creates a new instance of the <see cref="TimerExceptionEventArgs"/> object.
        /// </summary>
        /// <param name="t">The timer that has thrown the exception.</param>
        /// <param name="ex">The exception that occured on the timer.</param>
        public TimerExceptionEventArgs(Timer t, Exception ex)
        {
            fTimer = t;
            fException = ex;
            fStopTimer = true;
            fHandled = false;
        }

        /// <summary>
        /// Gets the timer object that have thrown the exception
        /// </summary>
        public Timer Timer
        {
            get { return fTimer; }
        }

        /// <summary>
        /// Gets os sets whether the timer should be stopped or not.
        /// </summary>
        public bool StopTimer
        {
            get { return fStopTimer; }
            set { fStopTimer = value; }
        }

        /// <summary>
        /// Gets or sets whether the exception was handled or not
        /// </summary>
        public bool Handled
        {
            get { return fHandled; }
            set { fHandled = value; }
        }

        /// <summary>
        /// Gets the exception that have been catched.
        /// </summary>
        public Exception Exception
        {
            get { return fException; }
        }
    }

    #endregion Event Args

    public static partial class Service
    {
        /// <summary>
        /// Occurs when the service was configured and started.
        /// </summary>
        public static event ServerStartupEventHandler Started;

        /// <summary>
        /// Occurs when the service was stopped.
        /// </summary>
        public static event ServerShutdownEventHandler Stopped;

        /// <summary>
        /// Occurs when a socket connection was established.
        /// </summary>
        public static event SocketConnectEventHandler SocketConnect;

        /// <summary>
        /// Occurs when the service have established a connection with a remote <see cref="IClient"/>.
        /// </summary>
        public static event ClientConnectEventHandler ClientConnect;

        /// <summary>
        /// Occurs when the service have broken an established connection with a remote <see cref="IClient"/>.
        /// </summary>
        public static event ClientDisconnectEventHandler ClientDisconnect;

        /// <summary>
        /// Occurs when a provider have been registered (added to the service IoC container).
        /// </summary>
        public static event ProviderRegisteredEventHandler ProviderRegistered;

        /// <summary>
        /// Occurs when the service have encountered an unhandled timer exception.
        /// </summary>
        public static event TimerExceptionHandler UnhandledTimerException;

        /// <summary>
        /// Occurs when the gossip provider detects a node joining the cluster.
        /// </summary>
        public static event ClusterEventHandler NodeConnect;

        /// <summary>
        /// Occurs when the gossip provider detects a node leaving the cluster.
        /// </summary>
        public static event ClusterEventHandler NodeDisconnect;

        /// <summary>
        /// Occurs when a message was received.
        /// </summary>
        public static MonitoringHandler MessageReceived;

        /// <summary>
        /// Occurs when a message was sent.
        /// </summary>
        public static MonitoringHandler MessageSent;

        #region Invokes

        internal static void InvokeServerStartup()
        {
            Started?.Invoke();
        }

        internal static void InvokeSocketConnect(SocketConnectEventArgs e)
        {
            SocketConnect?.Invoke(e);
        }

        internal static void InvokeClientConnect(ClientConnectEventArgs e)
        {
            ClientConnect?.Invoke(e);
        }

        internal static void InvokeClientDisconnect(ClientDisconnectEventArgs e)
        {
            ClientDisconnect?.Invoke(e);
        }

        internal static void InvokeServerShutdown(ServerShutdownEventArgs e)
        {
            // Notify all clients
            try
            {
                foreach (var client in Service.Clients)
                    InvokeClientDisconnect(new ClientDisconnectEventArgs(client));
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }

            // Invoke shutdown
            Stopped?.Invoke(e);
        }

        internal static void InvokeProviderRegistered(ProviderRegisteredEventArgs e)
        {
            ProviderRegistered?.Invoke(e);
        }

        internal static void InvokeUnhandledTimerException(TimerExceptionEventArgs e)
        {
            UnhandledTimerException?.Invoke(e);
        }

        internal static void InvokeNodeConnect(ClusterEventArgs e)
        {
            NodeConnect?.Invoke(e);
        }

        internal static void InvokeNodeDisconnect(ClusterEventArgs e)
        {
            NodeDisconnect?.Invoke(e);
        }

        #endregion Invokes

        internal static void Reset()
        {
            Started = null;
            Stopped = null;
            SocketConnect = null;
            ClientConnect = null;
            ClientDisconnect = null;
            ProviderRegistered = null;
            UnhandledTimerException = null;
            NodeConnect = null;
            NodeDisconnect = null;
        }
    }
}