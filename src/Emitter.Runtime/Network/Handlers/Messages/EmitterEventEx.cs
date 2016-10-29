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

using Emitter.Network.Mesh;

namespace Emitter.Network
{
    /// <summary>
    /// Represents extension methods.
    /// </summary>
    internal static class EmitterEventEx
    {
        /// <summary>
        /// Sends a subscribe event through the mesh.
        /// </summary>
        /// <param name="subscription"></param>
        /// <param name="peer">The node to notify.</param>
        public static void NotifySubscribe(this Subscription subscription, MeshMember peer)
        {
            // Send it as a command
            peer.Send(
                MeshEmitterEvent.Acquire(MeshEventType.Subscribe, subscription.ContractKey, subscription.Channel)
                );
        }

        /// <summary>
        /// Sends a subscribe event through the mesh.
        /// </summary>
        /// <param name="subscription"></param>
        public static void BroadcastSubscribe(this Subscription subscription, IMqttSender client)
        {
            // Broadcast to peers
            foreach (var peer in Service.Mesh.Members)
            {
                // Send it as a command
                peer.Send(
                    MeshEmitterEvent.Acquire(MeshEventType.Subscribe, subscription.ContractKey, subscription.Channel)
                    );
            }
        }

        /// <summary>
        /// Sends an unsubscribe event through the mesh.
        /// </summary>
        /// <param name="subscription"></param>
        public static void BroadcastUnsubscribe(this Subscription subscription, IMqttSender client)
        {
            // Broadcast to peers
            foreach (var peer in Service.Mesh.Members)
            {
                // Send it as a command
                peer.Send(
                    MeshEmitterEvent.Acquire(MeshEventType.Unsubscribe, subscription.ContractKey, subscription.Channel)
                    );
            }
        }
    }
}