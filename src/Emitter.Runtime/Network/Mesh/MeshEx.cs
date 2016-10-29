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
using Emitter.Providers;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a set of Mesh extension methods.
    /// </summary>
    public static class MeshEx
    {
        /// <summary>
        /// Gets the registry value for the gossip membership.
        /// </summary>
        /// <param name="registry">The registry to extend.</param>
        /// <returns></returns>
        public static GossipMembership GetMembership(this RegistryProvider registry)
        {
            return registry.Get<GossipMembership>("gossip");
        }

        /// <summary>
        /// Sends a mesh gossip through the specified channel.
        /// </summary>
        /// <param name="server">The channel to send the mesh gossip to.</param>
        internal static void SendMeshGossipDigest(this IServer server)
        {
            var provider = Service.Providers.Resolve<MeshProvider>();
            if (provider == null)
                return;

            // Compute the digest first
            var digest = Service.Registry.GetMembership().Version.Digest();

            // Send the digest around
            server.Send(MeshGossip.Acquire(digest));
        }

        /// <summary>
        /// Sends a mesh ping through the specified channel.
        /// </summary>
        /// <param name="server">The channel to send the packet to.</param>
        internal static void SendMeshPing(this IServer server)
        {
            server.Send(MeshEvent.Acquire(MeshEventType.Ping));
        }

        /// <summary>
        /// Sends a mesh PingAck through the specified channel.
        /// </summary>
        /// <param name="channel">The channel to send the packet to.</param>
        internal static void SendMeshPingAck(this Connection channel)
        {
            // Send the ack
            channel.Send(MeshEvent.Acquire(MeshEventType.PingAck));
        }

        /// <summary>
        /// Sends a mesh Handshake through the specified channel.
        /// </summary>
        /// <param name="channel">The channel to send the packet to.</param>
        internal static void SendMeshHandshake(this Connection channel)
        {
            // Acquire and set the handhshake packet
            var packet = MeshHandshake.Acquire();
            packet.Key = Murmur32.GetHash(Service.Mesh.Cluster);
            packet.Identity = new GossipMember(Service.Mesh.BroadcastEndpoint);

            // Send the handshake packet
            channel.Send(packet);
        }

        /// <summary>
        /// Sends a mesh Handshake through the specified channel.
        /// </summary>
        /// <param name="channel">The channel to send the packet to.</param>
        internal static void SendMeshHandshakeAck(this Connection channel)
        {
            var provider = Service.Providers.Resolve<MeshProvider>();
            if (provider == null)
                return;

            // Send the ack
            channel.Send(MeshEvent.Acquire(MeshEventType.HandshakeAck, provider.BroadcastEndpoint.ToString()));
        }

        /// <summary>
        /// Sends a mesh PingAck through the specified channel.
        /// </summary>
        /// <param name="channel">The channel to send the packet to.</param>
        /// <param name="ex">The exception to wrap.</param>
        internal static void SendMeshError(this Connection channel, Exception ex)
        {
            // Send the error back
            channel.Send(MeshEvent.Acquire(MeshEventType.Error, ex.Message));
        }
    }
}