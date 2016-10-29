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
using System.Threading.Tasks;
using Emitter.Collections;
using Emitter.Network;
using Emitter.Network.Mesh;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider that provides a socket.io and engine.io api.
    /// </summary>
    public abstract class MeshProvider : Provider
    {
        #region Properties

        /// <summary>
        /// Gets the address to broadcast in the mesh.
        /// </summary>
        public virtual IPEndPoint BroadcastEndpoint
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the cluster key to check during the mesh handshake.
        /// </summary>
        public virtual string Cluster
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the identifier of the local node.
        /// </summary>
        public virtual int Identifier
        {
            get
            {
                return this.BroadcastEndpoint.ToIdentifier();
            }
        }

        /// <summary>
        /// Gets or sets the duration of inactivity to wait after which a ping should be sent.
        /// </summary>
        public TimeSpan PingAfter
        {
            get;
            set;
        }

        /// <summary>
        /// Gets the <see cref="MeshBinding"/> configured for mesh interconnect.
        /// </summary>
        public IBinding Binding
        {
            get;
            internal set;
        }

        /// <summary>
        /// Gets the list of members nodes of the cluster.
        /// </summary>
        public abstract MeshMembership Members
        {
            get;
        }

        #endregion Properties

        #region Connection Members

        /// <summary>
        /// Connect to other nodes.
        /// </summary>
        /// <param name="peers">The peer nodes to connect the current mesh to.</param>
        public abstract void ConnectTo(params IPEndPoint[] peers);

        /// <summary>
        /// Gets the tcp channel or creates a new client connection.
        /// </summary>
        /// <param name="endpoint">The endpoint to connect to.</param>
        /// <param name="binding">The mesh binding to use.</param>
        /// <returns>The channel representing the connection</returns>
        public abstract Task ConnectToAsync(IPEndPoint endpoint, IBinding binding);

        #endregion Connection Members

        #region Handler Members

        private readonly ArrayList<IMeshHandler> Registry
            = new ArrayList<IMeshHandler>();

        /// <summary>
        /// Handles 'message' events of the Fabric layer.
        /// </summary>
        /// <param name="server">The server which is sending the frame.</param>
        /// <param name="buffer">The buffer to process.</param>
        /// <returns>The processing state of the event.</returns>
        internal ProcessingState OnFrame(IServer server, BufferSegment buffer)
        {
            if (server == null || buffer == null)
                return ProcessingState.Failure;

            int count = this.Registry.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.Registry.HasElementAt(i))
                {
                    var handler = this.Registry.Get(i);
                    var result = handler.ProcessFrame(server, buffer.AsSegment());
                    if (result == ProcessingState.Failure)
                        continue;
                    return result;
                }
            }

            // We failed processing
            return ProcessingState.Failure;
        }

        /// <summary>
        /// Handles 'command' events of the Fabric layer.
        /// </summary>
        /// <param name="server">The server which is sending the command.</param>
        /// <param name="packet">The packet containing the data.</param>
        /// <returns>The processing state of the event.</returns>
        internal ProcessingState OnEvent(IServer server, MeshEvent packet)
        {
            int count = this.Registry.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.Registry.HasElementAt(i))
                {
                    var handler = this.Registry.Get(i);
                    var result = handler.ProcessEvent(server, packet);
                    if (result == ProcessingState.Failure)
                        continue;

                    // On anything else, we're done here
                    // We might as well release the packet to reduce memory pressure
                    if (packet.Lifetime == PacketLifetime.Automatic)
                        packet.TryRelease();
                    return result;
                }
            }

            // We failed processing
            return ProcessingState.Failure;
        }

        /// <summary>
        /// Registers a handler to the provider.
        /// </summary>
        /// <param name="handler">The handler to register</param>
        public void Register(IMeshHandler handler)
        {
            lock (this.Registry)
            {
                this.Registry.Add(handler);
            }
        }

        /// <summary>
        /// Unregisters a handler from the provider.
        /// </summary>
        /// <param name="handler">The handler to unregister</param>
        public void Unregister(IMeshHandler handler)
        {
            lock (this.Registry)
            {
                this.Registry.Remove(handler);
            }
        }

        #endregion Handler Members
    }

    /// <summary>
    /// Represents a default provider that provides a socket.io and engine.io api.
    /// </summary>
    public sealed class DefaultMeshProvider : MeshProvider
    {
        private readonly MeshMembership Membership;

        /// <summary>
        /// Constructs a <see cref="DefaultMeshProvider"/> instance.
        /// </summary>
        public DefaultMeshProvider() : base()
        {
            this.Membership = new MeshMembership();
            this.PingAfter = TimeSpan.FromSeconds(60);
        }

        /// <summary>
        /// Gets the list of members nodes of the cluster.
        /// </summary>
        public override MeshMembership Members
        {
            get { return this.Membership; }
        }

        /// <summary>
        /// Connect to other nodes.
        /// </summary>
        /// <param name="peers">The peer nodes to connect the current mesh to.</param>
        public override void ConnectTo(params IPEndPoint[] peers)
        {
            if (this.Binding == null)
                throw new InvalidOperationException("No valid MeshBinding was set. Use Listen( new MeshBinding(...) ).");

            // Connect to every peer
            foreach (var endpoint in peers)
            {
                // Register the peer to track
                this.Membership.TrackPeer(endpoint);

                // Attempt to connect to
                this.Membership.TryConnect(endpoint, this.Binding).Forget();
            }
        }

        /// <summary>
        /// Gets the tcp channel or creates a new client connection.
        /// </summary>
        /// <param name="endpoint">The endpoint to connect to.</param>
        /// <param name="binding">The mesh binding to use.</param>
        /// <returns>The channel representing the connection</returns>
        public override Task ConnectToAsync(IPEndPoint endpoint, IBinding binding)
        {
            return this.Membership.TryConnect(endpoint, binding);
        }
    }
}