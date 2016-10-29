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
using Emitter.Replication;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a handler of various MESH requests.
    /// </summary>
    internal static unsafe class MeshHandler
    {
        /// <summary>
        /// Handles a command.
        /// </summary>
        public static ProcessingState Process(Connection channel, MeshEvent ev)
        {
            // Trace the operation number and the length of the packet
            //if (ev.Type != MeshEventType.Gossip)
            //    NetTrace.WriteLine("Incoming " + ev.ToString(), channel, NetTraceCategory.Mesh);

            try
            {
                switch (ev.Type)
                {
                    // Acknowledge the Heartbeat/Ping
                    case MeshEventType.Ping:
                        OnPing(channel);
                        break;

                    // Acknowledge the handshake
                    case MeshEventType.Handshake:
                        OnHandshake(channel, ev as MeshHandshake);
                        break;

                    // On the ack, the remote node sends its identifier to us
                    case MeshEventType.HandshakeAck:
                        OnHandshakeAck(channel, ev);
                        break;

                    // When a node receives a gossip digest
                    case MeshEventType.GossipDigest:
                        OnGossipDigest(channel, ev as MeshGossip);
                        break;

                    // When a node receives a gossip since request
                    case MeshEventType.GossipSince:
                        OnGossipSince(channel, ev as MeshGossip);
                        break;

                    // When a node receives a gossip update
                    case MeshEventType.GossipUpdate:
                        OnGossipUpdate(channel, ev as MeshGossip);
                        break;

                    // When a custom command is received
                    case MeshEventType.Custom:
                    default:
                        var node = Service.Mesh.Members.Get(channel.MeshIdentifier);
                        if (node == null)
                            break;

                        // Only invoke if we have a properly established connection
                        Service.Mesh.OnEvent(node, ev);
                        break;
                }
            }
            catch (Exception ex)
            {
                // Send the error back
                //Service.Logger.Log("Error On " + ev.ToString() + ": " + ex.Message);
                Service.Logger.Log(ex);
                channel.SendMeshError(ex);
            }

            // We do not need to process further, as we've handled the command
            return ProcessingState.Success;
        }

        /// <summary>
        /// Handles the command.
        /// </summary>
        /// <param name="channel">The channel sending the command.</param>
        /// <param name="command">The command received.</param>
        private static void OnGossipDigest(Connection channel, MeshGossip gossip)
        {
            // this message contains a digest
            var digest = (ReplicatedVersionDigest)gossip.State;

            // Do we have the same digest?
            var ourVersion = Service.Registry.Version;
            var ourDigest = ourVersion.Digest();

            if (ourDigest != digest.Value)
            {
                // Digests didn't match, we should ask for some gossip!
                //Console.WriteLine("Digest: {0} vs {1}", ourDigest, digest.Value);
                channel.Send(
                    MeshGossip.Acquire(ourVersion)
                    );
            }
        }

        /// <summary>
        /// Handles the command.
        /// </summary>
        /// <param name="channel">The channel sending the command.</param>
        /// <param name="command">The command received.</param>
        private static void OnGossipSince(Connection channel, MeshGossip gossip)
        {
            // We received a gossip request, send them a gossip update since the requested version
            channel.Send(
                MeshGossip.Acquire(Service.Registry.Collection, (ReplicatedVersion)gossip.State)
                );
        }

        /// <summary>
        /// Handles the command.
        /// </summary>
        /// <param name="channel">The channel sending the command.</param>
        /// <param name="command">The command received.</param>
        private static void OnGossipUpdate(Connection channel, MeshGossip gossip)
        {
            // Gossip state is merged during the packet read. Nothing to do here.
            var state = gossip.State as ReplicatedHybridDictionary;
            if (state == null)
                return;

            // Check if we need to update
            //if (Service.State.Version.TryUpdate(state.Version))
            {
                // Merge the other state in our current state
                Service.Registry.Collection.MergeIn(channel.MeshIdentifier, state);
                Service.Registry.Version.TryUpdate(state.Version);
            }
        }

        /// <summary>
        /// Handles the command.
        /// </summary>
        /// <param name="channel">The channel sending the command.</param>
        /// <param name="command">The command received.</param>
        private static void OnHandshake(Connection channel, MeshHandshake handshake)
        {
            // Deserialize the handshake
            var endpoint = handshake.Identity.EndPoint;
            var id = endpoint.ToIdentifier();

            // Set the identifier of the channel
            channel.MeshIdentifier = id;

            // Validate the credentials
            if (!(Murmur32.GetHash(Service.Mesh.Cluster) == handshake.Key))
                throw new UnauthorizedAccessException("Cluster access was not authorized.");

            // Check if we're trying to connect to ourselves
            if (Service.Mesh.Identifier == id)
            {
                // It's ourselves, remove from tracking
                //var ep = channel.RemoteEndPoint as IPEndPoint;
                var port = ((IPEndPoint)Service.Mesh.Binding.EndPoint).Port;
                var ep = new IPEndPoint(channel.RemoteEndPoint.Address, port);
                if (ep != null)
                    Service.Mesh.Members.ForgetPeer(ep);

                channel.Close();
                return;
            }

            // Attempt to register
            MeshMember node;
            if (Service.Mesh.Members.TryRegister(endpoint, channel, out node))
            {
                // Send ACK first
                channel.SendMeshHandshakeAck();

                // Send an event since we're connected to the node
                Service.InvokeNodeConnect(new ClusterEventArgs(node));
            }
            else
            {
                channel.Close();
            }
        }

        /// <summary>
        /// Handles the command.
        /// </summary>
        /// <param name="channel">The channel sending the command.</param>
        /// <param name="command">The command received.</param>
        private static void OnHandshakeAck(Connection channel, MeshEvent command)
        {
            // Attempt to register
            MeshMember node;
            IPEndPoint ep;
            if (!MeshMember.TryParseEndpoint(command.Value, out ep))
            {
                Service.Logger.Log(LogLevel.Error, "Unable to parse endpoint: " + command.Value);
                channel.Close();
                return;
            }

            // Register the endpoint
            if (Service.Mesh.Members.TryRegister(ep, channel, out node))
            {
                // Send an event since we're connected to the node
                Service.InvokeNodeConnect(new ClusterEventArgs(node));
            }
            else
            {
                channel.Close();
            }
        }

        /// <summary>
        /// Handles the command.
        /// </summary>
        /// <param name="channel">The channel sending the command.</param>
        private static void OnPing(Connection channel)
        {
            // Send the ack
            channel.SendMeshPingAck();
        }
    }
}