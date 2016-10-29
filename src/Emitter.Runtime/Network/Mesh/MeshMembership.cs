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
using System.Collections;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using Emitter.Diagnostics;
using Emitter.Replication;
using TNodeKey = System.Int32;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents an mesh registry.
    /// </summary>
    public sealed class MeshMembership : IEnumerable<MeshMember>
    {
        /// <summary>
        /// Gets the registry that allows us to maintain 1-1 correspondance between mesh nodes.
        /// </summary>
        internal readonly ConcurrentDictionary<TNodeKey, MeshMember> Registry
            = new ConcurrentDictionary<TNodeKey, MeshMember>();

        /// <summary>
        /// Gets the list of tracked peers.
        /// </summary>
        private readonly ConcurrentDictionary<string, IPEndPoint> TrackedPeers
            = new ConcurrentDictionary<string, IPEndPoint>();

        /// <summary>
        /// Gets the randomizer.
        /// </summary>
        private readonly Random Random = new Random();

        /// <summary>
        /// The gossip thread.
        /// </summary>
        private readonly GossipThread Thread;

        /// <summary>
        /// Constructs a new instance of the object.
        /// </summary>
        public MeshMembership()
        {
            this.Thread = new GossipThread(TimeSpan.FromMilliseconds(250));
            Service.Started += OnServiceStarted;
        }

        /// <summary>
        /// Occurs when the service is started.
        /// </summary>
        private void OnServiceStarted()
        {
            // Register membership changes
            Service.Registry.GetMembership().Change += OnMembershipChange;

            // Register ourselves
            Registry.TryAdd(Service.Mesh.Identifier, MeshMember.Create());

            // Start the thread now that the service is started
            this.Thread.Start();
        }

        /// <summary>
        /// Checks whether we are already connected to the specified endpoint.
        /// </summary>
        /// <param name="id">The remote identifier.</param>
        public bool Contains(TNodeKey id)
        {
            return Registry.ContainsKey(id);
        }

        /// <summary>
        /// Registers a channel.
        /// </summary>
        /// <param name="channel">The channel to register.</param>
        /// <returns>Whether the channel already exists or not.</returns>
        public bool TryRegister(IPEndPoint endpoint, Connection channel, out MeshMember node)
        {
            node = null;
            if (channel == null)
                throw new ArgumentNullException("Attempting to register a peer with null channel.");

            lock (Registry)
            {
                // Try to retrieve the node first
                var id = endpoint.ToIdentifier();
                if (Registry.TryGetValue(id, out node))
                {
                    // We already have it in our registry, is the channel still OK?
                    if (node.State == ServerState.Online)
                        return false;
                }
                else
                {
                    // Create the node if not found and add it.
                    node = MeshMember.Create(endpoint, channel);
                    if (!Registry.TryAdd(id, node))
                        return false;
                }

                // Configure
                node.Channel = channel;
                channel.MeshIdentifier = id;

                // Do not timeout the connection
                channel.Timeout = TimeSpan.FromDays(10 * 365);

                // Make sure we're tracking the peer
                TrackPeer(node.EndPoint);

                // Log it to the console
                Service.Logger.Log(LogLevel.Info, "Mesh: Connected to " + endpoint.ToString());

                // Send the gossip once we've registered
                node.SendMeshGossipDigest();
                return true;
            }
        }

        /// <summary>
        /// Gets the mesh node, if available.
        /// </summary>
        /// <param name="id">The id of the node to get.</param>
        /// <returns></returns>
        public MeshMember Get(TNodeKey id)
        {
            // We could query without id, in which case we simply return null value.
            if (id == 0)
                return null;

            // Actually do the query if we have an id set.
            MeshMember info;
            if (Registry.TryGetValue(id, out info))
                return info;
            return null;
        }

        /// <summary>
        /// Unregisters an endpoint from the registry.
        /// </summary>
        /// <param name="id">The identifier to unregister.</param>
        /// <returns>Whether it was unregistered or not.</returns>
        public bool Unregister(TNodeKey id)
        {
            lock (Registry)
            {
                // Keep the registration
                MeshMember info;
                if (Registry.TryRemove(id, out info))
                {
                    // Send an event since we're connected to the node
                    Service.InvokeNodeDisconnect(new ClusterEventArgs(info));

                    //info.Disconnect -= OnDisconnect;
                    //NetTrace.WriteLine("Unregister: " + info.Channel, info.Channel, NetTraceCategory.Mesh);
                    Console.WriteLine("Unregister: " + info);
                    return true;
                }
                return false;
            }
        }

        /// <summary>
        /// Gets the tcp channel or creates a new client connection.
        /// </summary>
        /// <param name="endpoint">The endpoint to connect to.</param>
        /// <param name="binding">The mesh binding to use.</param>
        /// <returns>The channel representing the connection</returns>
        internal async Task TryConnect(IPEndPoint endpoint, IBinding binding)
        {
            try
            {
                // The binding must be set prior to the connect
                if (binding == null)
                    throw new ArgumentNullException("binding");

                // Get the peer
                IPEndPoint peer;
                if (!TrackedPeers.TryGetValue(endpoint.ToString(), out peer))
                    return;

                // Do we have the node in our mesh?
                MeshMember node;
                if (Registry.TryGetValue(peer.ToIdentifier(), out node))
                {
                    // If the connection is up and running already, skip the connection.
                    if (node.State == ServerState.Online)
                        return;
                }

                // Create a new connection
                var channel = await Connection.ConnectAsync(Service.Mesh.Binding.Context, endpoint);

                // If it's a new channel, start it and handshake
                if (channel == null)
                    return;

                // Dead? Attempt to reconnect
                NetTrace.WriteLine("Connected to peer " + peer + "...", channel, NetTraceCategory.Mesh);

                // Send the mesh handshake once we're connected
                channel.SendMeshHandshake();

                // Wait for the handshake for a few seconds
                await Task.Delay(TimeSpan.FromSeconds(5));

                // If we haven't received an ack yet, shutdown the connection
                MeshMember registered;
                if (Registry.TryGetValue(peer.ToIdentifier(), out registered) && registered.Channel == channel)
                    return;
                channel.Close();
            }
            catch (Exception)
            {
                // Service.Logger.Log(LogLevel.Warning, "Unable to connect to Mesh node: " + endpoint);
            }
        }

        /// <summary>
        /// Occurs when a new entry is being merged in the gossip.
        /// </summary>
        /// <param name="entry"></param>
        private void OnMembershipChange(ref ReplicatedDictionary<GossipMember>.Entry entry, int source, bool isMerging)
        {
            // We're only interested in merging changes
            if (!isMerging)
                return;

            // Skip ourselves
            var node = entry.Value;
            var key = node.EndPoint.ToIdentifier();
            if (key == Service.Mesh.Identifier)
                return;

            // If the node was deleted, untrack it.
            if (entry.Deleted)
            {
                ForgetPeer(node.EndPoint);
            }
            else
            {
                // Do we have the peer in our registry?
                MeshMember item;
                if (Registry.TryGetValue(key, out item))
                {
                    // Parse the endpoint and touch the entry
                    if (item.Identifier == source)
                        item.LastTouchUtc = Timer.UtcNow;
                }
                else
                {
                    try
                    {
                        // We've discovered a new node, track it
                        TrackPeer(node.EndPoint);
                    }
                    catch (Exception ex)
                    {
                        // Log the parse exception
                        Service.Logger.Log(ex);
                    }
                }
            }
        }

        #region IEnumerable Members

        /// <summary>
        /// Gets the enumerator that enumerates each element in the collection.
        /// </summary>
        /// <returns>The enumerator.</returns>
        public IEnumerator<MeshMember> GetEnumerator()
        {
            return Registry.Values.GetEnumerator();
        }

        /// <summary>
        /// Gets the enumerator that enumerates each element in the collection.
        /// </summary>
        /// <returns>The enumerator.</returns>
        IEnumerator IEnumerable.GetEnumerator()
        {
            return this.GetEnumerator();
        }

        #endregion IEnumerable Members

        #region Nested TrackedPeer

        /// <summary>
        /// Gets the tracked peers safe enumerable collection
        /// </summary>
        public IEnumerable<IPEndPoint> Tracked
        {
            get { return this.TrackedPeers.Values; }
        }

        /// <summary>
        /// Adds a peer to the tracking mechanism.
        /// </summary>
        /// <param name="ep"></param>
        public bool TrackPeer(IPEndPoint ep)
        {
            if (TrackedPeers.TryAdd(ep.ToString(), ep))
            {
                NetTrace.WriteLine("Tracking Peer: " + ep.ToString(), null, NetTraceCategory.Mesh);
                return true;
            }

            return false;
        }

        /// <summary>
        /// Removes a peer from the tracking mechanism.
        /// </summary>
        /// <param name="ep"></param>
        public bool ForgetPeer(IPEndPoint ep)
        {
            IPEndPoint peer;
            if (TrackedPeers.TryRemove(ep.ToString(), out peer))
            {
                // Forget the peer
                NetTrace.WriteLine("Forgetting Peer: " + ep.ToString(), null, NetTraceCategory.Mesh);
                return true;
            }

            return false;
        }

        /// <summary>
        /// Picks a random node, excluding onselves. Can be null.
        /// </summary>
        /// <returns></returns>
        public MeshMember GetRandomNode()
        {
            try
            {
                // Choose a random node from the ring to gossip with.
                var neighbours = Registry.Keys
                    .Where(k => k != Service.Mesh.Identifier)
                    .ToArray();

                // If we didn't find any neighbours, ignore
                if (neighbours.Length == 0)
                    return null;

                // Get a random neighbour
                var nodeIdx = Random.Next(0, neighbours.Length);
                return Get(neighbours[nodeIdx]);
            }
            catch (Exception ex)
            {
                // Something bad happened
                Service.Logger.Log(ex);
                return null;
            }
        }

        #endregion Nested TrackedPeer
    }
}