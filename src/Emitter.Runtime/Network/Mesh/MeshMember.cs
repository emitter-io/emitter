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
using System.Globalization;
using System.Net;
using TNodeKey = System.Int32;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a membership registration.
    /// </summary>
    public sealed class MeshMember : IServer
    {
        // The underlying pipe
        private Connection Connection = null;

        /// <summary>
        /// Constructor for JSON.
        /// </summary>
        internal MeshMember()
        {
            this.LastTouchUtc = DateTime.UtcNow;
        }

        /// <summary>
        /// Gets or sets the identifier.
        /// </summary>
        public TNodeKey Identifier;

        /// <summary>
        /// Gets whether this node is alive or not.
        /// </summary>
        public ServerState State
        {
            get
            {
                // Is the server online?
                if (this.Channel != null && this.Channel.IsRunning)
                    return ServerState.Online;

                // If it's ourselves, always report online
                if (this.Identifier == Service.Mesh.Identifier)
                    return ServerState.Online;

                // Offline
                return ServerState.Offline;
            }
        }

        /// <summary>
        /// Gets or sets when the node was seen the last time.
        /// </summary>
        public DateTime LastTouchUtc;

        /// <summary>
        /// Parses the endpoint from the identifier.
        /// </summary>
        public IPEndPoint EndPoint
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets or sets the channel which can be used for communication.
        /// </summary>
        public Connection Channel
        {
            get { return this.Connection; }
            set
            {
                if (this.Identifier == 0)
                    throw new InvalidOperationException("The idenfitier must be set before the channel.");

                this.Connection = value;
                this.Connection.Disconnect += (connection) => Service.Mesh.Members.Unregister(this.Identifier);
                this.LastTouchUtc = DateTime.UtcNow;
            }
        }

        /// <summary>
        /// Gets or sets the session object for the server.
        /// </summary>
        public object Session
        {
            get;
            set;
        }

        /// <summary>
        /// Sends the packet to the remote client end-point.
        /// </summary>
        /// <param name="packet">The packet to send to the remote end-point.</param>
        public void Send(Packet packet)
        {
            // We shouldn't send anything to ourselves
            if (this.Identifier == Service.Mesh.Identifier)
                return;

            // Only send to online peers
            if (this.State == ServerState.Online)
                this.Channel.Send(packet);
        }

        /// <summary>
        /// Returns a string representation of the peer.
        /// </summary>
        public override string ToString()
        {
            return string.Format("{0}: {1}", this.EndPoint, (this.State == ServerState.Online ? "UP" : "DOWN"));
        }

        #region Factory Members

        /// <summary>
        /// Constructs a new info.
        /// </summary>
        /// <param name="channel">The tcp channel to bind.</param>
        internal static MeshMember Create(IPEndPoint endpoint, Connection channel)
        {
            var node = new MeshMember();
            node.LastTouchUtc = DateTime.UtcNow;
            node.Identifier = endpoint.ToIdentifier();
            node.EndPoint = endpoint;
            return node;
        }

        /// <summary>
        /// Constructs a new info node for oneself.
        /// </summary>
        internal static MeshMember Create()
        {
            var node = new MeshMember();
            node.LastTouchUtc = DateTime.UtcNow;
            node.Identifier = Service.Mesh.Identifier;
            node.EndPoint = Service.Mesh.BroadcastEndpoint;

            // Also add to the membership
            Service.Registry.GetMembership().Add(Service.Mesh.BroadcastEndpoint.ToString(), new GossipMember(node.EndPoint));
            return node;
        }

        #endregion Factory Members

        #region Endpoint Parsing

        /// <summary>
        /// Attempts to parse an endpoint.
        /// </summary>
        /// <param name="stringEndpoint"></param>
        /// <param name="endpoint"></param>
        /// <returns></returns>
        internal static bool TryParseEndpoint(string stringEndpoint, out IPEndPoint endpoint)
        {
            endpoint = null;
            var ep = stringEndpoint.Split(':');
            if (ep.Length < 2)
                return false;

            IPAddress ip;
            if (ep.Length > 2)
            {
                if (!IPAddress.TryParse(string.Join(":", ep, 0, ep.Length - 1), out ip))
                    return false;
            }
            else
            {
                if (!IPAddress.TryParse(ep[0], out ip))
                    return false;
            }
            int port;
            if (!int.TryParse(ep[ep.Length - 1], NumberStyles.None, NumberFormatInfo.CurrentInfo, out port))
                return false;

            endpoint = new IPEndPoint(ip, port);
            return true;
        }

        #endregion Endpoint Parsing
    }

    /// <summary>
    /// Represents the state of the peer node.
    /// </summary>
    public enum ServerState : ushort
    {
        /// <summary>
        /// The node is offline.
        /// </summary>
        Offline,

        /// <summary>
        /// The node might be offline.
        /// </summary>
        Suspect,

        /// <summary>
        /// The node is online.
        /// </summary>
        Online,
    }
}