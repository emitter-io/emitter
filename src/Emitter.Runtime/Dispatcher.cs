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
using System.Collections.Generic;
using Emitter.Collections;
using Emitter.Network;

namespace Emitter
{
    /// <summary>
    /// Represents core operations such as publish, subscribe.
    /// </summary>
    internal static class Dispatcher
    {
        /// <summary>
        /// Publishes a message.
        /// </summary>
        /// <param name="contract">The contract</param>
        /// <param name="message">The message to publish.</param>
        /// <param name="channelHash">The hashed channel id.</param>
        /// <param name="channelName">The full name of the channel.</param>
        /// <param name="ttl">The time-to-live value for the message.</param>
        public static void Publish(int contractId, uint channelHash, string channelName, ArraySegment<byte> message, int ttl = EmitterConst.Transient)
        {
            // We forward to the call with the contract, as this current code path is not the hot path.
            Publish((EmitterContract)Services.Contract.GetByKey(contractId), channelHash, channelName, message, ttl);
        }

        /// <summary>
        /// Publishes a message.
        /// </summary>
        /// <param name="contract">The contract</param>
        /// <param name="message">The message to publish.</param>
        /// <param name="channelHash">The hashed channel id.</param>
        /// <param name="channelName">The full name of the channel.</param>
        /// <param name="ttl">The time-to-live value for the message.</param>
        public static void Publish(EmitterContract contract, uint channelHash, string channelName, ArraySegment<byte> message, int ttl = EmitterConst.Transient)
        {
            try
            {
                // Make sure the TTL is within limits and max is 30 days.
                if (ttl < 0)
                    ttl = EmitterConst.Transient;
                if (ttl > 2592000)
                    ttl = 2592000;

                var contractId = contract.Oid;
                var cid = ((long)contractId << 32) + channelHash;

                // Increment the counter
                Service.MessageReceived?.Invoke(contract, channelName, message.Count);

                // The SSID
                var ssid = EmitterChannel.Ssid(contractId, channelName);

                // Get the subscriptions
                var subs = Subscription.Match(ssid);
                if (subs != null)
                {
                    // Fast-Path: check local subscribers and send it to them
                    ForwardToClients(contractId, channelName, message, subs);

                    // Publish into the messaging cluster once
                    ForwardToServers(contractId, channelName, message, subs);
                }

                // Only call into the storage service if necessary
                if (ttl > 0 && Services.Storage != null)
                {
                    // If we have a storage service, store the message
                    Services.Storage.AppendAsync(contractId, ssid, ttl, message);
                }
            }
            catch (Exception ex)
            {
                Service.Logger.Log(LogLevel.Warning, "Publish failed. Reason: " + ex.Message);
                //Service.Logger.Log(ex);
                return;
            }
        }

        /// <summary>
        /// Forwards a message to a set of subscriptions
        /// </summary>
        internal static void ForwardToClients(int contractId, string channel, ArraySegment<byte> message)
        {
            // Separator must be added
            channel += EmitterConst.Separator;

            // Create an SSID
            var ssid = EmitterChannel.Ssid(contractId, channel);

            // Get the matching subscriptions
            var subs = Subscription.Match(ssid);
            if (subs == null)
                return;

            // Forward to the clients
            ForwardToClients(contractId, channel, message, subs);
        }

        /// <summary>
        /// Forwards a message to a set of subscriptions
        /// </summary>
        private static void ForwardToClients(int contractId, string channel, ArraySegment<byte> message, IEnumerable<Subscription> subscriptions)
        {
            // Iterate through all of subscriptions matched
            foreach (var subscription in subscriptions)
            {
                // We need to hold on to a copy of the array we are going to iterate through
                // as to to this would not break the loop, since the array can/will be changed
                // while we iterate!
                var clients = subscription.MessageSubscribers;
                if (clients == null)
                    continue;

                // We also cache the contract reference
                var contract = subscription.Contract;

                // Send a message to every subscribed client
                for (int i = 0; i < clients.Length; ++i)
                {
                    // Increment the counter
                    Service.MessageSent?.Invoke(contract, channel, message.Count);

                    // Send the message out
                    clients[i].Send(contractId, subscription.Channel, message);
                }
            }
        }

        /// <summary>
        /// Forwards a message to a set of subscriptions
        /// </summary>
        private static void ForwardToServers(int contract, string channel, ArraySegment<byte> message, IEnumerable<Subscription> subscriptions)
        {
            // Send the message out
            using (var filter = Filter<MessageQueue>.Acquire())
            {
                // Iterate through all of subscriptions matched
                foreach (var subscription in subscriptions)
                {
                    // We need to hold on to a copy of the array we are going to iterate through
                    // as to to this would not break the loop, since the array can/will be changed
                    // while we iterate!
                    var servers = subscription.Servers;
                    if (servers == null || servers.Length == 0)
                        continue;

                    // Add the server to the set
                    for (int i = 0; i < servers.Length; ++i)
                    {
                        // Set the bloom filter value for the server we found.
                        var mq = servers[i].Session as MessageQueue;
                        if (mq == null)
                            servers[i].Session = mq = new MessageQueue();
                        filter.Add(mq);
                    }
                }

                // Send a message to every subscribed server
                foreach (var server in Service.Mesh.Members)
                {
                    // Get the message queue and check if we have it in our filter
                    var mq = server.Session as MessageQueue;
                    if (mq == null || !filter.Contains(mq))
                        continue;

                    // Enqueue to the message queue
                    mq.Enqueue(contract, channel, message);
                }
            }
        }

        /// <summary>
        /// Subscribe will express interest in the given subject. The subject
        /// can have wildcards (partial:*, full:>).
        /// </summary>
        /// <param name="client">The client to subscribe.</param>
        /// <param name="contract">The contract for this operation.</param>
        /// <param name="channel">The channel to subscribe to.</param>
        /// <param name="interest">The interest of the subscription.</param>
        /// <returns>The subscription id.</returns>
        public static Subscription Subscribe(IClient client, int contract, string channel, SubscriptionInterest interest)
        {
            // Hook the subscription unregister when the client disconnects, this way
            // we should not leak any memory since the disconnect is reliable.
            var subs = Subscription.Register(client, contract, channel, interest);

            // If we have provided a client, make sure to unsubscribe on disconnect
            if (client != null)
                client.Disconnect += (c) => Unsubscribe(c, contract, channel, SubscriptionInterest.Everything);

            // Return the subscription
            return subs;
        }

        /// <summary>
        /// Subscribe will express interest in the given subject. The subject
        /// can have wildcards (partial:*, full:>).
        /// </summary>
        /// <param name="sender">The sender to subscribe.</param>
        /// <param name="contract">The contract for this operation.</param>
        /// <param name="channel">The channel to subscribe to.</param>
        /// <param name="interest">The interest of the subscription.</param>
        /// <returns>The subscription id.</returns>
        public static Subscription Subscribe(IMqttSender sender, int contract, string channel, SubscriptionInterest interest)
        {
            // Hook the subscription unregister when the client disconnects, this way
            // we should not leak any memory since the disconnect is reliable.
            return Subscription.Register(sender, contract, channel, interest);
        }

        /// <summary>
        /// Unregisters a subscription from the current registry.
        /// </summary>
        /// <param name="client">The client to unsubscribe.</param>
        /// <param name="ssid">The subscription to unregister.</param>
        /// <param name="contract">The contract to unsubscribe.</param>
        /// <param name="interest">The interest of the subscription.</param>
        /// <returns>Whether the subscription was removed or not.</returns>
        public static bool Unsubscribe(IMqttSender client, int contract, string channel, SubscriptionInterest interest)
        {
            // First we attempt to unregister the subscription
            Subscription subscription;
            if (Subscription.Unregister(client, contract, channel, interest, out subscription))
            {
                // Successfully unsubscribed
                return true;
            }
            else
            {
                // We didn't manage to unsubscibe
                return false;
            }
        }
    }
}