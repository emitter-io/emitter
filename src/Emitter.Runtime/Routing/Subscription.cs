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
using System.Linq;
using Emitter.Network;
using Emitter.Network.Mesh;
using Emitter.Replication;
using Emitter.Text.Json;

namespace Emitter
{
    /// <summary>
    /// a class to hold subscription-related data
    /// </summary>
    internal sealed class Subscription : IDisposable
    {
        #region Static Members

        /// <summary>
        /// The registry of current subscriptions.
        /// </summary>
        private static readonly SubscriptionTrie Index = new SubscriptionTrie();

        /// <summary>
        /// Static constructor.
        /// </summary>
        static Subscription()
        {
            // When a new node joins
            Service.NodeConnect += OnNodeConnect;
        }

        /// <summary>
        /// Occurs when a new node joins.
        /// </summary>
        /// <param name="e"></param>
        private static void OnNodeConnect(ClusterEventArgs e)
        {
            // The node to notify
            var node = e.Node;

            // Send current subscriptions
            foreach (var sub in Index.GetAllValues())
            {
                // We're only interested in subscriptions with clients connected
                if (sub.MessageSubscribers.Length == 0)
                    continue;

                sub.NotifySubscribe(node);
            }
        }

        /// <summary>
        /// Creates a new subscription.
        /// </summary>
        /// <param name="client">The client to subscribe.</param>
        /// <param name="contract">The contract that subscribes.</param>
        /// <param name="channel">The channel to subscribe to.</param>
        /// <returns>The subscription data structure.</returns>
        public static unsafe Subscription Register(IMqttSender client, int contract, string channel, SubscriptionInterest interest)
        {
            // Register the subscription
            var subs = Index.AddOrUpdate(contract, channel,
                () => new Subscription(contract, channel, client, interest),
                (v) =>
                {
                    // If we're simply updating an existing one, increment the number of clients we have
                    v.Subscribe(client, interest);
                    return v;
                });

            // Every time a client subscribes, broadcast it through the network
            subs.BroadcastSubscribe(client);

            // Return the subscription
            return subs;
        }

        /// <summary>
        /// Unregisters a subscription from the current registry.
        /// </summary>
        /// <param name="client">The client to unsubscribe.</param>
        /// <param name="contract">The contract.</param>
        /// <param name="channel"></param>
        /// <returns>Whether the subscription was removed or not.</returns>
        public static bool Unregister(IMqttSender client, int contract, string channel, SubscriptionInterest interest, out Subscription subs)
        {
            // Make the subscription id
            var ssid = EmitterChannel.Ssid(contract, channel);

            // Get the subscription first
            if (!Index.TryGetValue(ssid, 0, out subs))
                return false;

            // Validate, only the correct contract can unsubscribe
            if (subs.ContractKey != contract)
                return false;

            // Unregister a client
            subs.Unsubscribe(client, interest);
            return true;
        }

        /// <summary>
        /// Attempts to retrieve a subscription.
        /// </summary>
        /// <param name="contract">The contract use for retrieval.</param>
        /// <param name="channel">The channel use for retrieval</param>
        /// <param name="sub">The subscription, if found.</param>
        /// <returns></returns>
        public static bool TryGet(int contract, string channel, out Subscription sub)
        {
            return Index.TryGetValue(contract, channel, out sub);
        }

        /// <summary>
        /// Creates a new subscription.
        /// </summary>
        /// <param name="server">The server to subscribe.</param>
        /// <param name="contract">The contract that subscribes.</param>
        /// <param name="channel">The channel to subscribe to.</param>
        /// <returns>The subscription data structure.</returns>
        public static unsafe Subscription Register(IServer server, int contract, string channel)
        {
            // Register the subscription
            return Index.AddOrUpdate(contract, channel,
                () => new Subscription(contract, channel, server),
                (v) =>
                {
                    // If we're simply updating an existing one, register the new server
                    v.Subscribe(server);
                    return v;
                });
        }

        /// <summary>
        /// Unregisters a subscription from the current registry.
        /// </summary>
        /// <param name="server">The server to unsubscribe.</param>
        /// <param name="contract">The contract.</param>
        /// <param name="channel"></param>
        /// <returns>Whether the subscription was removed or not.</returns>
        public static bool Unregister(IServer server, int contract, string channel)
        {
            // Make the subscription id
            var ssid = EmitterChannel.Ssid(contract, channel);

            // Get the subscription first
            Subscription subs;
            if (!Index.TryGetValue(ssid, 0, out subs))
                return false;

            // Validate, only the correct contract can unsubscribe
            if (subs.ContractKey != contract)
                return false;

            // Unregister a server
            subs.Unsubscribe(server);
            return false;
        }

        /// <summary>
        /// Gets the matching subscriptions.
        /// </summary>
        /// <param name="ssid">The ssid to match.</param>
        /// <returns>The subscription or null if none was found.</returns>
        public static IEnumerable<Subscription> Match(uint[] ssid)
        {
            return Index.Match(ssid);
        }

        #endregion Static Members

        #region Constructors

        /// <summary>
        /// Constructs a new subscription.
        /// </summary>
        /// <param name="contract">The contract that subscribes.</param>
        /// <param name="channel">The channel to subscribe to.</param>
        /// <param name="client">The first member.</param>
        /// <param name="interest">The interest of the subsriber</param>
        public Subscription(int contract, string channel, IMqttSender client, SubscriptionInterest interest)
        {
            this.ContractKey = contract;
            this.Channel = channel;

            // Get the contract and set the presence reference
            this.Presence = this.Contract.Info.GetOrCreate<SubscriptionPresence>(channel);
            this.Presence.Change += OnPresenceChange;
            this.Subscribe(client, interest);
        }

        /// <summary>
        /// Constructs a new subscription.
        /// </summary>
        /// <param name="contract">The contract that subscribes.</param>
        /// <param name="channel">The channel to subscribe to.</param>
        /// <param name="server">The first member.</param>
        public Subscription(int contract, string channel, IServer server)
        {
            this.ContractKey = contract;
            this.Channel = channel;

            // Get the contract and set the presence reference
            this.Presence = this.Contract.Info.GetOrCreate<SubscriptionPresence>(channel);
            this.Presence.Change += OnPresenceChange;
            this.Subscribe(server);
        }

        #endregion Constructors

        #region Public Properties

        /// <summary>
        /// Gets or sets the cached contract value.
        /// </summary>
        private volatile EmitterContract ContractValue;

        /// <summary>
        /// The list of registered clients who are interested in presence events.
        /// </summary>
        public IMqttSender[] PresenceSubscribers = new IMqttSender[0];

        /// <summary>
        /// The list of registered clients who are interested in messages.
        /// </summary>
        public IMqttSender[] MessageSubscribers = new IMqttSender[0];

        /// <summary>
        /// The list of registered servers.
        /// </summary>
        public IServer[] Servers = new IServer[0];

        /// <summary>
        /// The presence list.
        /// </summary>
        public readonly SubscriptionPresence Presence;

        /// <summary>
        /// Gets or sets the contract id for this subscription.
        /// </summary>
        public int ContractKey;

        /// <summary>
        /// Gets or sets the channel for this subscription.
        /// </summary>
        public string Channel;

        /// <summary>
        /// Gets the number of clients and servers currently subscribed to this.
        /// </summary>
        public int Listeners
        {
            get { return this.MessageSubscribers.Length + this.Servers.Length; }
        }

        /// <summary>
        /// Gets the presence occupancy. The amount of clients subscribed cluster-wide.
        /// </summary>
        public int Occupancy
        {
            get { return this.Presence.Count; }
        }

        /// <summary>
        /// Gets whether the subscription is empty and contains no listeners or nor presence subscribers
        /// </summary>
        public bool Empty
        {
            get { return this.Listeners == 0 && this.PresenceSubscribers.Length == 0; }
        }

        /// <summary>
        /// Retrieves a contract, lazily
        /// </summary>
        public EmitterContract Contract
        {
            get
            {
                if (this.ContractValue == null)
                {
                    lock (this)
                    {
                        this.ContractValue = (EmitterContract)Services.Contract.GetByKey(this.ContractKey);
                    }
                }

                return this.ContractValue;
            }
        }

        #endregion Public Properties

        #region Messaging Members

        /// <summary>
        /// Subscribes a client by adding it to an appropriate membership list.
        /// </summary>
        /// <param name="client">The client to subscribe.</param>
        /// <param name="interest">The interests to register for.</param>
        /// <returns></returns>
        public void Subscribe(IMqttSender client, SubscriptionInterest interest)
        {
            lock (this)
            {
                if (interest.HasFlag(SubscriptionInterest.Messages))
                {
                    // Add to the membership list
                    ArrayUtils.AddUnique(ref this.MessageSubscribers, client);

                    // Add to the presence. Since the presence is inside the registry, it will be automatically
                    // synchronized and eventually consistent across the cluster.
                    if (client.Id != ConnectionId.Empty)
                        this.Presence.Add(client.Id.ToString(), new SubscriptionPresenceInfo(client));
                }

                if (interest.HasFlag(SubscriptionInterest.Presence))
                {
                    // Subscribe to presence events
                    ArrayUtils.AddUnique(ref this.PresenceSubscribers, client);
                }
            }
        }

        /// <summary>
        /// Removes a client from the memebership list.
        /// </summary>
        /// <param name="client">The client to remove.</param>
        /// <returns>The current membership.</returns>
        public void Unsubscribe(IMqttSender client, SubscriptionInterest interest)
        {
            lock (this)
            {
                // Remove from the membership
                if (interest.HasFlag(SubscriptionInterest.Messages) && ArrayUtils.Remove(ref this.MessageSubscribers, client) >= 0)
                {
                    // Remove the client from the presence
                    if (client.Id != ConnectionId.Empty)
                    {
                        Console.WriteLine("Removing from presence: " + client.Id);
                        this.Presence.Remove(client.Id.ToString());
                    }

                    if (this.Listeners == 0)
                    {
                        // Broadcast unsubscribe if there's no more clients left
                        this.BroadcastUnsubscribe(client);
                    }
                }

                if (interest.HasFlag(SubscriptionInterest.Presence))
                {
                    // Unregister the client presence as well
                    ArrayUtils.Remove(ref this.PresenceSubscribers, client);
                }

                // Dispose the subscription if empty
                if (this.Empty)
                    this.Dispose();
            }
        }

        /// <summary>
        /// Adds a client to the subscription membership list.
        /// </summary>
        /// <param name="server">The client to add.</param>
        public void Subscribe(IServer server)
        {
            lock (this)
            {
                // Add to the array
                ArrayUtils.AddUnique(ref this.Servers, server);
            }
        }

        /// <summary>
        /// Removes a client from the memebership list.
        /// </summary>
        /// <param name="server">The client to remove.</param>
        /// <returns>The current membership.</returns>
        public void Unsubscribe(IServer server)
        {
            lock (this)
            {
                // Remove from the array
                ArrayUtils.Remove(ref this.Servers, server);

                // Dispose the subscription if empty
                if (this.Empty)
                    this.Dispose();
            }
        }

        #endregion Messaging Members

        #region Join/Leave Events

        /// <summary>
        /// Occurs when a new entry is added/removed or merged into the presence list.
        /// </summary>
        /// <param name="value">The entry value.</param>
        /// <param name="source">The source node.</param>
        /// <param name="isMerging">Whether it is merging or not.</param>
        private void OnPresenceChange(ref ReplicatedDictionary<SubscriptionPresenceInfo>.Entry value, int source, bool isMerging)
        {
            if (!value.Deleted)
            {
                var data = value.Value;
                this.OnClientJoin(value.Value.AsInfo());
            }
            else
            {
                this.OnClientLeave(value.Key);
            }
        }

        /// <summary>
        /// Occurs when we've been told that a cliend has joined a subscription.
        /// </summary>
        /// <param name="clientId">The client id.</param>
        /// <param name="username">The username of the connection.</param>
        private void OnClientJoin(PresenceInfo who)
        {
            // If there's no subscribers for presence, ignore
            if (this.PresenceSubscribers.Length == 0)
                return;

            // Prepare a notification
            var serialized = JsonConvert.SerializeObject(new PresenceNotification(PresenceEvent.Subscribe, this.Channel, who, this.Occupancy), Formatting.Indented);
            if (serialized == null)
                return;

            // Since we have a subscription, iterate and notify presence change
            foreach (var client in this.PresenceSubscribers)
                client.Send(this.ContractKey, "emitter/presence/", serialized.AsUTF8().AsSegment());
        }

        /// <summary>
        /// Occurs when we've been told that a cliend has left a subscription.
        /// </summary>
        /// <param name="clientId">The client id.</param>
        /// <param name="username">The username of the connection.</param>
        private void OnClientLeave(string key)
        {
            // If there's no subscribers for presence, ignore
            if (this.PresenceSubscribers.Length == 0)
                return;

            var who = new PresenceInfo();
            who.Id = key;

            // Prepare a notification
            var serialized = JsonConvert.SerializeObject(new PresenceNotification(PresenceEvent.Unsubscribe, this.Channel, who, this.Occupancy), Formatting.Indented);
            if (serialized == null)
                return;

            // Since we have a subscription, iterate and notify presence change
            foreach (var client in this.PresenceSubscribers)
                client.Send(this.ContractKey, "emitter/presence/", serialized.AsUTF8().AsSegment());
        }

        #endregion Join/Leave Events

        #region IDisposable Memers

        /// <summary>
        /// Disposes the subscription.
        /// </summary>
        public void Dispose()
        {
            try
            {
                // Unbind events
                if (this.Presence != null)
                    this.Presence.Change -= this.OnPresenceChange;

                // Get the ssid
                var ssid = EmitterChannel.Ssid(this.ContractKey, this.Channel);

                // Remove the subscription from the registry
                Subscription removed;
                Index.TryRemove(ssid, 0, out removed);
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
        }

        #endregion IDisposable Memers
    }
}