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
using Emitter.Threading;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a gossip communication thread.
    /// </summary>
    internal sealed class GossipThread : ThreadBase
    {
        /// <summary>
        /// Creates a gossip thread with a specific gossip interval.
        /// </summary>
        /// <param name="interval"></param>
        public GossipThread(TimeSpan interval) : base(interval)
        {
        }

        /// <summary>
        /// Gets the message to publish to emitter channeL.
        /// </summary>
        /// <returns></returns>
        protected override void OnExecute()
        {
            // Make sure we maintain our TCP connection up and running by
            // pinging inactive nodes
            PeerHeartbeat();

            // Check the peers we are tracking actively and reconnect to offline
            // peers if needed.
            PeerTrack();

            // Choose a random node to gossip with and send our full state through
            // tcp to that node.
            PeerGossip();
        }

        /// <summary>
        /// Make sure we maintain our TCP connection up and running by
        /// pinging inactive nodes.
        /// </summary>
        private static void PeerHeartbeat()
        {
            try
            {
                // Iterate through members
                var pingAfter = Timer.UtcNow - Service.Mesh.PingAfter;
                foreach (var member in Service.Mesh.Members)
                {
                    // We're out of touch?
                    if (member.Identifier != Service.Mesh.Identifier && member.LastTouchUtc < pingAfter)
                        member.SendMeshPing();
                }
            }
            catch (Exception ex)
            {
                // Something went wrong, log it.
                Service.Logger.Log(ex);
            }
        }

        /// <summary>
        /// Check the peers we are tracking actively and reconnect to offline peers if needed.
        /// </summary>
        private static async void PeerTrack()
        {
            try
            {
                // Iterate through all the peers we're maintaning
                foreach (var peer in Service.Mesh.Members.Tracked)
                {
                    // Attempt to connect to the peer
                    await Service.Mesh.ConnectToAsync(peer, Service.Mesh.Binding);
                }
            }
            catch (Exception ex)
            {
                // Something went wrong, log it.
                Service.Logger.Log(ex);
            }
        }

        /// <summary>
        /// Chose a random node and send the gossip state to it.
        /// </summary>
        private static void PeerGossip()
        {
            try
            {
                // Get a random node to gossip with
                var node = Service.Mesh.Members.GetRandomNode();
                if (node == null)
                    return;

                // Send a gossip digest to the node
                node.SendMeshGossipDigest();
            }
            catch (Exception ex)
            {
                // Something went wrong, log it.
                Service.Logger.Log(ex);
            }
        }
    }
}