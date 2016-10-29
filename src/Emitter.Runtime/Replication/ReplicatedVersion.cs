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
using System.Threading;
using Emitter.Network;
using TNodeKey = System.Int32;

namespace Emitter.Replication
{
    /// <summary>
    /// Represents a version vector.
    /// </summary>
    public sealed class ReplicatedVersion : ISerializable
    {
        /// <summary>
        /// This is the version for the node itself. This way we can avoid storing the node identifier
        /// within the vector and only hit the dictionary when updating someone else's state.
        /// </summary>
        private long Version;

        /// <summary>
        /// Gets our own id.
        /// </summary>
        internal readonly TNodeKey NodeId;

        /// <summary>
        /// Gets the underlying dictionary for storing the version vector. This thing only grows, as
        /// we do not have any limits on how long a machine can be offline. One could think to compact
        /// this vector after some time.
        /// </summary>
        private readonly Dictionary<TNodeKey, long> Vector
            = new Dictionary<TNodeKey, long>();

        /// <summary>
        /// Constructs a new instance of the object.
        /// </summary>
        public ReplicatedVersion()
        {
            //this.NodeId = Service.Mesh.Identifier;
            this.NodeId = 0;
            this.Version = 0;
        }

        /// <summary>
        /// Constructs a new instance of the object.
        /// </summary>
        public ReplicatedVersion(TNodeKey nodeId)
        {
            this.NodeId = nodeId;
            this.Version = 0;
        }

        /// <summary>
        /// Gets our own version.
        /// </summary>
        public long Own
        {
            get { return this.Version; }
        }

        /// <summary>
        /// Increment our own version and return the value.
        /// </summary>
        /// <returns>The latest version.</returns>
        public long Increment()
        {
            return Interlocked.Increment(ref this.Version);
        }

        /// <summary>
        /// Gets the version of a node (except ourself).
        /// </summary>
        /// <param name="nodeId">The node whose version we need to retrieve.</param>
        /// <returns>The observed version of that node.</returns>
        public long Of(TNodeKey nodeId)
        {
            lock (this.Vector)
            {
                long version;
                if (this.Vector.TryGetValue(nodeId, out version))
                    return version;
                return 0;
            }
        }

        /// <summary>
        /// Gets the digest for this version.
        /// </summary>
        /// <returns></returns>
        public int Digest()
        {
            // Construct the string list
            var entities = new SortedSet<string>();
            lock (this.Vector)
            {
                entities.Add(this.NodeId + ":" + this.Version);
                foreach (var kvp in this.Vector)
                    entities.Add(kvp.Key + ":" + kvp.Value);
            }

            // Compute the hash
            int hash = 0;
            foreach (var kvp in entities)
                hash ^= Murmur32.GetHash(kvp);
            return hash;
        }

        /// <summary>
        /// Attempts to update the version vector.
        /// </summary>
        /// <param name="other">The other vector.</param>
        /// <returns></returns>
        public bool TryUpdate(ReplicatedVersion other)
        {
            // Go through all the other vectors
            var newer = false;
            foreach (var kvp in other.Vector)
            {
                if (kvp.Key == this.NodeId)
                    continue;

                var v1 = this.Of(kvp.Key);
                var v2 = kvp.Value;
                if (v2 > v1)
                {
                    this.Vector[kvp.Key] = kvp.Value;
                    newer = true;
                }
            }

            // We updated something
            return newer;
        }

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public void Read(PacketReader reader)
        {
            var count = reader.ReadInt32();
            for (int i = 0; i < count; ++i)
            {
                // Read stuff
                var id = reader.ReadInt32();
                var version = reader.ReadInt64();

                // Add the vector of neighbors
                if (id != this.NodeId)
                    this.Vector.Add(id, version);
            }
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        public void Write(PacketWriter writer)
        {
            lock (this.Vector)
            {
                writer.Write(this.Vector.Count + 1);
                writer.Write(this.NodeId);
                writer.Write(this.Version);

                foreach (var kvp in this.Vector)
                {
                    writer.Write(kvp.Key);
                    writer.Write(kvp.Value);
                }
            }
        }

        /// <summary>
        /// Converts the version to a string representation.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            var entities = new SortedSet<string>();
            lock (this.Vector)
            {
                entities.Add(this.NodeId + ":" + this.Version);
                foreach (var kvp in this.Vector)
                    entities.Add(kvp.Key + ":" + kvp.Value);
            }

            return "[" + String.Join(", ", entities.ToArray()) + "]";
        }
    }
}