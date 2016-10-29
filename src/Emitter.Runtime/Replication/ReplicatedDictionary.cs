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
using System.Collections.Generic;
using System.Linq;
using Emitter.Network;
using TKey = System.String;
using TNodeKey = System.Int32;

namespace Emitter.Replication
{
    /// <summary>
    /// Represents a distributed dictionary CRDT data structure that is eventually consistent.
    /// </summary>
    public abstract class ReplicatedDictionary<TValue> : ReplicatedObject, IEnumerable<ReplicatedDictionary<TValue>.Entry>, IReplicated, IReplicatedCollection
        where TValue : IReplicated
    {
        #region Constructors

        /// <summary>
        /// The counter using for caching the size.
        /// </summary>
        private int Size = 0;

        /// <summary>
        /// The underlying dictionary.
        /// </summary>
        private readonly Dictionary<TKey, Entry> Map;

        /// <summary>
        /// Constructs a new distributed dictionary.
        /// </summary>
        public ReplicatedDictionary()
        {
            this.Map = new Dictionary<TKey, Entry>();
        }

        /// <summary>
        /// Constructs a new distributed dictionary.
        /// </summary>
        /// <param name="nodeId">The node id.</param>
        public ReplicatedDictionary(TNodeKey nodeId)
        {
            //this.NodeId = nodeId;
            this.Version = new ReplicatedVersion(nodeId);
            this.Map = new Dictionary<TKey, Entry>();
        }

        #endregion Constructors

        #region Events

        /// <summary>
        /// Occurs when an entry is being added/removed to the dictionary.
        /// </summary>
        public event ChangeEventHandler Change;

        /// <summary>
        /// Occurs when an entry is being added/removed to the dictionary.
        /// </summary>
        /// <param name="source">The source node.</param>
        /// <param name="newEntry">The entry being added/removed or merged.</param>
        /// <param name="isMerging">Whether it occurs during a merge or not.</param>
        private void OnChange(ref Entry newEntry, TNodeKey source, bool isMerging)
        {
            //Console.WriteLine("Change: {0}{1} at {2} {3}", newEntry.Deleted ? "-" : "+", newEntry.Key, source, isMerging ? "merging" : "updating");

            // Recalculate the counter. This is not very optimized, but at least this is correct.
            this.Size = this.Count();

            // By default, invoke the event
            this.Change?.Invoke(ref newEntry, source, isMerging);
        }

        /// <summary>
        /// Occurs when a child is updated.
        /// </summary>
        /// <param name="key">The key to update.</param>
        public void OnChildUpdate(TKey key)
        {
            Entry entry;
            if (this.Map.TryGetValue(key, out entry))
            {
                // Update the entry
                entry.Version = this.Version.Own;
                this.Map[key] = entry;

                // Propagate upwards
                var notify = this.Parent as IReplicatedCollection;
                if (notify != null)
                    notify.OnChildUpdate(this.Key);
            }
        }

        #endregion Events

        #region Public Members

        /// <summary>
        /// Gets the number of active elements in the dictionary.
        /// </summary>
        public int Count
        {
            get { return this.Size; }
        }

        /// <summary>
        /// Gets the value by the key, adds if if no value was found.
        /// </summary>
        /// <param name="key"></param>
        /// <returns></returns>
        public TValue GetOrCreate(string key)
        {
            // Attempt to fetch first
            TValue value;
            if (this.TryGetValue(key, out value))
                return value;

            // Add and return
            value = ReplicatedActivator.CreateInstance<TValue>();
            this.Add(key, value);
            return value;
        }

        /// <summary>
        /// Determines whether the <see cref="DistributedDictionary{TKey, TValue}"/> contains an element
        //  with the specified key.
        /// </summary>
        /// <param name="key">The key to locate.</param>
        /// <returns></returns>
        public bool ContainsKey(TKey key)
        {
            lock (this.Map)
            {
                Entry entry;
                if (this.Map.TryGetValue(key, out entry))
                    return !entry.Deleted;
                return false;
            }
        }

        /// <summary>
        /// Adds an element with the provided key and value to the <see cref="DistributedDictionary{TKey, TValue}"/>.
        /// </summary>
        /// <param name="key">The object to use as the key of the element to add.</param>
        /// <param name="value">The object to use as the value of the element to add.</param>
        public void Add(TKey key, TValue value)
        {
            this.Update(key, value, false);
        }

        /// <summary>
        /// Removes the element with the specified key from the <see cref="DistributedDictionary{TKey, TValue}"/>.
        /// </summary>
        /// <param name="key">The key of the element to remove.</param>
        /// <returns>true if the element is successfully removed; otherwise, false. </returns>
        public bool Remove(TKey key)
        {
            return this.Update(key, default(TValue), true);
        }

        /// <summary>
        /// Gets the value associated with the specified key.
        /// </summary>
        /// <param name="key"> The key whose value to get.</param>
        /// <param name="value">When this method returns, the value associated with the specified key, if the
        /// key is found; otherwise, the default value for the type of the value parameter.
        /// This parameter is passed uninitialized.</param>
        /// <returns>true if the dictionary contains an element with the specified key; otherwise, false.</returns>
        public bool TryGetValue(TKey key, out TValue value)
        {
            lock (this.Map)
            {
                Entry entry;
                if (this.Map.TryGetValue(key, out entry) && !entry.Deleted)
                {
                    value = entry.Value;
                    return true;
                }

                value = default(TValue);
                return false;
            }
        }

        #endregion Public Members

        #region Private Update/Merge/Reconcile Members

        /// <summary>
        /// Attempts to update a value of a key.
        /// </summary>
        /// <param name="key">The key.</param>
        /// <param name="value">The value for the key.</param>
        /// <param name="deleted">Whether the key is deleted or not.</param>
        private bool Update(TKey key, TValue value, bool deleted)
        {
            // Create the entry and set it
            var newEntry = new Entry() { Version = this.Version.Increment(), Deleted = deleted, From = this.Version.NodeId, Key = key, Value = value };
            this.Map[key] = newEntry;

            // Set the parent
            if (value != null)
            {
                value.Parent = this;
                value.Version = this.Version;
                value.Key = key;
            }

            // Notify that we are merging
            this.OnChange(ref newEntry, this.Version.NodeId, false);

            // Propagate upwards
            var notify = this.Parent as IReplicatedCollection;
            if (notify != null)
                notify.OnChildUpdate(this.Key);
            return true;
        }

        /// <summary>
        /// Merges in another replicated value.
        /// </summary>
        /// <param name="other">The other value to merge.</param>
        public void MergeIn(TNodeKey source, IReplicatedCollection other)
        {
            var entries = other.GetEntries();
            foreach (Entry entry in entries)
            {
                Entry newEntry = entry;
                this.MergeIn(source, ref newEntry);
            }
        }

        /// <summary>
        /// Attempts to merge a value of a key.
        /// </summary>
        /// <param name="source">The source node.</param>
        /// <param name="newEntry">The new entry to merge</param>
        private void MergeIn(TNodeKey source, ref Entry newEntry)
        {
            lock (this.Map)
            {
                // Get or create the entry
                Entry oldEntry;
                var contains = this.Map.TryGetValue(newEntry.Key, out oldEntry);

                // Check if they are both collections, in which case we'll need to merge them
                var newCollection = newEntry.Value as IReplicatedCollection;
                var oldCollection = oldEntry.Value as IReplicatedCollection;
                if (oldCollection != null)
                {
                    // Make sure we match parent and version
                    newEntry.Value.Parent = this;
                    newEntry.Value.Version = this.Version;

                    // If there's already a value, attempt to merge in first
                    oldCollection.MergeIn(source, newCollection);
                }
                else
                {
                    // If there's no key in the set yet, or the current one has the same ts, add it. Also, reconcile
                    // a conflict if one occurs.
                    if (!contains
                        || newEntry.Version > oldEntry.Version
                        || (newEntry.Version == oldEntry.Version && this.Reconcile(ref newEntry, ref oldEntry)))
                    {
                        // Make sure we match parent and version
                        if (newEntry.Value != null)
                        {
                            newEntry.Value.Parent = this;
                            newEntry.Value.Version = this.Version;
                        }

                        // Set the thing to the new entry
                        this.Map[newEntry.Key] = newEntry;

                        // Notify that we are merging
                        this.OnChange(ref newEntry, source, true);
                    }
                }
            }
        }

        /// <summary>
        /// Resolves the conflict and returns whether the new value wins.
        /// </summary>
        /// <param name="newEntry">The new entry.</param>
        /// <param name="oldEntry">The existing entry.</param>
        /// <returns>Returns whether the new value should win.</returns>
        protected virtual bool Reconcile(ref Entry newEntry, ref Entry oldEntry)
        {
            // 1. Pick delete wins strategy
            if (newEntry.Deleted != oldEntry.Deleted)
                return newEntry.Deleted;

            // 2. Delete wins
            if (newEntry.Value == null)
                return true;

            // 3. Pick value strategy
            return newEntry.Value.CompareTo(oldEntry.Value) > 0;
        }

        #endregion Private Update/Merge/Reconcile Members

        #region Read/Write Members

        /// <summary>
        /// Creates a value that should be used for deserialization.
        /// </summary>
        /// <param name="reader">The reader to use.</param>
        /// <returns>The created value.</returns>
        protected virtual TValue ReadValuePrefix(PacketReader reader)
        {
            return ReplicatedActivator.CreateInstance<TValue>();
        }

        /// <summary>
        /// Writes a prefix for a value in this dictionary.
        /// </summary>
        /// <param name="value">The value we are writing.</param>
        /// <param name="writer">The writer to use.</param>
        protected virtual void WriteValuePrefix(TValue value, PacketWriter writer)
        {
        }

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public override void Read(PacketReader reader)
        {
            // Read the header
            if (this.Parent == null)
            {
                var source = reader.ReadInt32();
                this.Version = reader.ReadSerializable<ReplicatedVersion>();
            }

            // Read the amount of updates we have
            var count = reader.ReadInt32();
            for (int i = 0; i < count; ++i)
            {
                var entry = new Entry();
                entry.Version = reader.ReadInt64();
                entry.Deleted = reader.ReadBoolean();
                entry.From = reader.ReadInt32();
                entry.Key = reader.ReadString();
                if (!entry.Deleted)
                {
                    // Read the entry
                    var value = this.ReadValuePrefix(reader);
                    value.Parent = this;
                    value.Version = this.Version;
                    value.Key = entry.Key;
                    value.Read(reader);
                    entry.Value = value;
                }

                // Set the entry
                this.Map[entry.Key] = entry;
            }
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        /// <param name="since">The minimum version of updates to pack.</param>
        public override void Write(PacketWriter writer, ReplicatedVersion since)
        {
            // Write the version vector only for root element
            if (this.Parent == null)
            {
                writer.Write(this.Version.NodeId);
                writer.Write(this.Version);
            }

            // Write a length (bookmark it)
            var bookmark = writer.Position;
            writer.Write(0);

            // Iterate through the values matching the 'since' constraint
            var count = 0;
            foreach (var entry in this.Map.Values)
            {
                // Ignore non-matching versions
                if (entry.Version < since.Of(entry.From))
                    continue;

                // Write the header
                writer.Write(entry.Version);
                writer.Write(entry.Deleted);
                writer.Write(entry.From);
                writer.Write(entry.Key);

                // Write the value only if we have it
                if (!entry.Deleted)
                {
                    // Write the prefix if required
                    this.WriteValuePrefix(entry.Value, writer);

                    // Write the value itself
                    writer.Write(entry.Value, since);
                }

                // Increment our pack count
                count++;
            }

            // Go back and write the final count
            var last = writer.Position;
            writer.Position = bookmark;
            writer.Write(count);
            writer.Position = last;
        }

        #endregion Read/Write Members

        #region IEnumerable Members

        /// <summary>
        /// Returns the enumerator that iterates through the collection.
        /// </summary>
        /// <returns></returns>
        public IEnumerator<Entry> GetEnumerator()
        {
            // TODO: improve
            return this.Map
                .Values
                .Where(v => !v.Deleted)
                .GetEnumerator();
        }

        /// <summary>
        /// Returns the enumerator that iterates through the collection.
        /// </summary>
        /// <returns></returns>
        IEnumerator IEnumerable.GetEnumerator()
        {
            return this.GetEnumerator();
        }

        /// <summary>
        /// Gets the enumerable which iterates throug all entries, including the tombstones.
        /// </summary>
        /// <returns>The enumerable.</returns>
        IEnumerable IReplicatedCollection.GetEntries()
        {
            return this.Map.Values;
        }

        #endregion IEnumerable Members

        #region Nested Entry

        /// <summary>
        /// Represents an entry in the distributed dictionary.
        /// </summary>
        public struct Entry
        {
            /// <summary>
            /// Gets or sets the version of the entry.
            /// </summary>
            public long Version;

            /// <summary>
            /// Gets or sets whether the entry is dead.
            /// </summary>
            public bool Deleted;

            /// <summary>
            /// Gets or sets the node id which owns this entry.
            /// </summary>
            public TNodeKey From;

            /// <summary>
            /// Gets or sets the key of the entry.
            /// </summary>
            public TKey Key;

            /// <summary>
            /// Gets or sets the associated value with the entry.
            /// </summary>
            public TValue Value;
        }

        /// <summary>
        /// Encapsulates a method that handles a merge event.
        /// </summary>
        /// <typeparam name="T"></typeparam>
        /// <param name="source">The source node.</param>
        /// <param name="value">The entry type.</param>
        public delegate void ChangeEventHandler(ref Entry value, TNodeKey source, bool isMerging);

        #endregion Nested Entry
    }
}