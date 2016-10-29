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
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;

namespace Emitter
{
    /// <summary>
    /// Represents a trie for subscriptions with a reverse-pattern search.
    /// </summary>
    internal sealed class SubscriptionTrie : ReverseTrie<Subscription>
    {
        /// <summary>
        /// Constructs the Subscription matching trie.
        /// </summary>
        public SubscriptionTrie() : base(-1) { }

        /// <summary>
        /// Attempts to add a subscription to the trie.
        /// </summary>
        /// <param name="sub">The subscription to add.</param>
        /// <returns>Whether the subscription was added or not.</returns>
        public bool TryAdd(Subscription sub)
        {
            return this.TryAdd(EmitterChannel.Ssid(sub.ContractKey, sub.Channel), sub);
        }

        /// <summary>
        /// Attempts to remove a subscription from the trie.
        /// </summary>
        /// <param name="sub">The subscription to remove.</param>
        /// <returns>Whether the subscription was removed or not.</returns>
        public bool TryRemove(Subscription sub)
        {
            Subscription removed;
            return this.TryRemove(EmitterChannel.Ssid(sub.ContractKey, sub.Channel), 0, out removed);
        }

        /// <summary>
        /// Attempts to remove a subscription from the trie.
        /// </summary>
        /// <param name="contract">The contract for the subscription to remove.</param>
        /// <param name="channel">The channel for the subscription to remove.</param>
        /// <param name="sub">The removed item.</param>
        /// <returns>Whether the subscription was removed or not.</returns>
        public bool TryRemove(int contract, string channel, out Subscription sub)
        {
            return this.TryRemove(EmitterChannel.Ssid(contract, channel), 0, out sub);
        }

        /// <summary>
        /// Attempts to fetch a subscription from the trie.
        /// </summary>
        /// <param name="sub">The subscription retrieved.</param>
        /// <param name="channel">The channel for the subscription to find.</param>
        /// <param name="contract">The channel for the subscription to find.</param>
        /// <returns>Whether the subscription was found or not.</returns>
        public bool TryGetValue(int contract, string channel, out Subscription sub)
        {
            return this.TryGetValue(EmitterChannel.Ssid(contract, channel), 0, out sub);
        }

        /// <summary>
        /// Adds or updates the subscription.
        /// </summary>
        /// <param name="channel">The channel for the subscription to find.</param>
        /// <param name="contract">The channel for the subscription to find.</param>
        /// <param name="addFunc">The add function to execute.</param>
        /// <param name="updateFunc">The update function to execute.</param>
        /// <returns></returns>
        public Subscription AddOrUpdate(int contract, string channel, Func<Subscription> addFunc, Func<Subscription, Subscription> updateFunc)
        {
            return this.AddOrUpdate(EmitterChannel.Ssid(contract, channel), 0, addFunc, updateFunc);
        }
    }

    /// <summary>
    /// Represents a trie with a reverse-pattern search.
    /// </summary>
    /// <typeparam name="T"></typeparam>
    internal class ReverseTrie<T>
        where T : class
    {
        // The cutoff point
        private const int CutoffPoint = 64;

        // An array or a map
        private volatile ConcurrentDictionary<uint, ReverseTrie<T>> Children;

        private volatile Entry[] Flat;

        // The level and the value of the node
        private readonly short Level = 0;

        private T Value = default(T);

        /// <summary>
        /// Constructs a node of the trie.
        /// </summary>
        public ReverseTrie() : this(-1)
        {
        }

        /// <summary>
        /// Constructs a node of the trie.
        /// </summary>
        /// <param name="level">The level of this node within the trie.</param>
        public ReverseTrie(short level)
        {
            this.Level = level;
            this.Children = null;
            this.Flat = new Entry[0];
        }

        /// <summary>
        /// Adds a value to the Trie.
        /// </summary>
        /// param name="key">The key for the value.</param>
        /// param name="value">The value to add</param>
        public bool TryAdd(uint[] key, T value)
        {
            return TryAdd(key, 0, value);
        }

        /// <summary>
        /// Adds a value to the Trie.
        /// </summary>
        /// <param name="key">The key for the value.</param>
        /// <param name="position">The position of the value.</param>
        /// <param name="value">The value to add</param>
        public bool TryAdd(uint[] key, int position, T value)
        {
            if (position >= key.Length)
            {
                lock (this)
                {
                    // There's already a value
                    if (this.Value != default(T))
                        return false;

                    this.Value = value;
                    return true;
                }
            }

            // Create a child
            return ChildGetOrAdd(key, position)
                .TryAdd(key, position + 1, value);
        }

        /// <summary>
        /// Attempts to remove a specific key from the Trie.
        /// </summary>
        public bool TryRemove(uint[] key, int position, out T value)
        {
            if (position >= key.Length)
            {
                lock (this)
                {
                    // There's no value
                    value = this.Value;
                    if (this.Value == default(T))
                        return false;

                    this.Value = default(T);
                    return true;
                }
            }

            // Remove from the child
            ReverseTrie<T> child;
            if (ChildTryGet(key[position], out child))
                return child.TryRemove(key, position + 1, out value);

            value = default(T);
            return false;
        }

        /// <summary>
        /// Attempts to fetch a value.
        /// </summary>
        public bool TryGetValue(uint[] key, int position, out T value)
        {
            if (position >= key.Length)
            {
                lock (this)
                {
                    // Set the value to return
                    value = this.Value;

                    // If there's no value
                    if (this.Value == default(T))
                        return false;
                    return true;
                }
            }

            // Create a child
            ReverseTrie<T> child;
            if (ChildTryGet(key[position], out child))
                return child.TryGetValue(key, position + 1, out value);

            value = default(T);
            return false;
        }

        /// <summary>
        /// Gets or adds a specific value.
        /// </summary>
        public T GetOrAdd(uint[] key, int position, Func<T> addFunc)
        {
            if (position >= key.Length)
            {
                lock (this)
                {
                    // There's already a value
                    if (this.Value != default(T))
                        return this.Value;

                    // No value, add it
                    this.Value = addFunc();
                    return this.Value;
                }
            }

            // Add a child
            return ChildGetOrAdd(key, position)
                .GetOrAdd(key, position + 1, addFunc);
        }

        /// <summary>
        /// Adds or updates a specific value.
        /// </summary>
        /// <param name="key"></param>
        /// <param name="position"></param>
        /// <returns></returns>
        public T AddOrUpdate(uint[] key, int position, Func<T> addFunc, Func<T, T> updateFunc)
        {
            if (position >= key.Length)
            {
                lock (this)
                {
                    // There's already a value
                    if (this.Value != default(T))
                        return updateFunc(this.Value);

                    // No value, add it
                    this.Value = addFunc();
                    return this.Value;
                }
            }

            return ChildGetOrAdd(key, position)
                .AddOrUpdate(key, position + 1, addFunc, updateFunc);
        }

        /// <summary>
        /// Retrieves a set of values.
        /// </summary>
        /// <param name="query">The query to retrieve.</param>
        /// <param name="position">The position.</param>
        /// <returns></returns>
        public IEnumerable<T> Match(uint[] query)
        {
            // Get the matching stack
            var matches = new ReverseTrie<T>[8];
            var matchesCount = 0;
            var queryLength = query.Length;
            ReverseTrie<T> childNode, current;

            // Push a new match to the stack
            if (matchesCount + 1 >= matches.Length)
                Array.Resize(ref matches, (matches.Length + 1) * 2);
            matches[matchesCount++] = this;

            // While we have matches
            while (matchesCount != 0)
            {
                // Pop the current value from the stack
                current = matches[--matchesCount];
                matches[matchesCount] = default(ReverseTrie<T>);

                if (current.Value != default(T))
                    yield return current.Value;

                var level = current.Level + 1;
                if (level >= queryLength)
                    continue;

                if (current.ChildTryGet(1815237614, out childNode))
                {
                    if (matchesCount + 1 >= matches.Length)
                        Array.Resize(ref matches, (matches.Length + 1) * 2);
                    matches[matchesCount++] = childNode;
                }

                if (current.ChildTryGet(query[level], out childNode))
                {
                    if (matchesCount + 1 >= matches.Length)
                        Array.Resize(ref matches, (matches.Length + 1) * 2);
                    matches[matchesCount++] = childNode;
                }
            }
        }

        /// <summary>
        /// Gets all values in the trie.
        /// </summary>
        /// <returns></returns>
        public IEnumerable<T> GetAllValues()
        {
            // Yield all the values in the current node
            if (this.Value != default(T))
                yield return this.Value;

            // Get children
            var next = this.Children != null
                ? this.Children.Values
                : FlatGetChildren();

            foreach (var leaf in next)
            {
                var values = leaf.GetAllValues();
                foreach (var value in values)
                    yield return value;
            }
        }

        /// <summary>
        /// Gets the children nodes for search.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private IEnumerable<ReverseTrie<T>> GetChildren(uint[] query, int position)
        {
            // Check for the wildcard
            ReverseTrie<T> childNode;
            if (ChildTryGet(1815237614, out childNode))
                yield return childNode;

            // Check for the leaf
            if (ChildTryGet(query[position], out childNode))
                yield return childNode;
        }

        /// <summary>
        /// Gets a specific child for removal.
        /// </summary>
        private ReverseTrie<T> GetChild(uint[] query, int position)
        {
            // Check for the leaf
            ReverseTrie<T> childNode;
            return ChildTryGet(query[position], out childNode)
                    ? childNode : null;
        }

        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public bool ChildTryGet(uint key, out ReverseTrie<T> value)
        {
            if (this.Children != null)
                return this.Children.TryGetValue(key, out value);

            var idx = FlatSearch(key);
            if (idx < 0)
            {
                value = null;
                return false;
            }

            value = this.Flat[idx].Value;
            return true;
        }

        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private ReverseTrie<T> ChildGetOrAdd(uint[] key, int position)
        {
            return this.Children == null
                ? FlatGetOrAdd(key[position], new ReverseTrie<T>((short)position))
                : this.Children.GetOrAdd(key[position], new ReverseTrie<T>((short)position));
        }

        #region Flat Types

        /// <summary>
        /// Gets all the children nodes.
        /// </summary>
        /// <returns></returns>
        private IEnumerable<ReverseTrie<T>> FlatGetChildren()
        {
            for (int i = 0; i < this.Flat.Length; ++i)
                yield return this.Flat[i].Value;
        }

        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private int FlatSearch(uint key)
        {
            int i = 0;
            for (; i < this.Flat.Length; ++i)
            {
                if (this.Flat[i].Key == key)
                    return i;
            }
            return -1;
        }

        public ReverseTrie<T> FlatGetOrAdd(uint key, ReverseTrie<T> value)
        {
            var idx = FlatSearch(key);
            if (idx < 0)
            {
                lock (this.Flat)
                {
                    // First, we check if we need to promote this node to a dictionary.
                    var length = this.Flat.Length;
                    if (length >= CutoffPoint)
                    {
                        var map = new ConcurrentDictionary<uint, ReverseTrie<T>>(Environment.ProcessorCount, 16);
                        for (int i = 0; i < length; ++i)
                        {
                            var item = this.Flat[i];
                            map.TryAdd(item.Key, item.Value);
                        }

                        // Scrap the old array how that we have a dictionary
                        this.Children = map;
                        this.Flat = null;
                    }
                    else
                    {
                        // If we still have space, copy and add
                        var target = this.Flat;
                        var result = new Entry[target.Length + 1];
                        target.CopyTo(result, 0);
                        result[target.Length] = new Entry(key, value);
                        this.Flat = result;
                    }
                }

                return value;
            }
            else
            {
                return this.Flat[idx].Value;
            }
        }

        #endregion Flat Types

        #region Nested Types

        private struct Entry
        {
            public Entry(uint key, ReverseTrie<T> value)
            {
                this.Key = key;
                this.Value = value;
            }

            public uint Key;
            public ReverseTrie<T> Value;
        }

        #endregion Nested Types
    }
}