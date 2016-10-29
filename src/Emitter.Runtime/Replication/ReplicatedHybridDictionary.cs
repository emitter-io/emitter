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
using Emitter.Network;
using TNodeKey = System.Int32;

namespace Emitter.Replication
{
    /// <summary>
    /// Represents a distributed dictionary CRDT data structure that is eventually consistent.
    /// </summary>
    public class ReplicatedHybridDictionary : ReplicatedDictionary<IReplicated>
    {
        /// <summary>
        /// Creates an isntance of a <see cref="ReplicatedHybridDictionary"/>.
        /// </summary>
        public ReplicatedHybridDictionary() : base() { }

        /// <summary>
        /// Creates an isntance of a <see cref="ReplicatedHybridDictionary"/>.
        /// </summary>
        /// <param name="nodeId">The node identifier.</param>
        public ReplicatedHybridDictionary(TNodeKey nodeId) : base(nodeId) { }

        /// <summary>
        /// Gets the value by the key, adds if if no value was found.
        /// </summary>
        /// <param name="key"></param>
        /// <returns></returns>
        public T GetOrCreate<T>(string key)
            where T : IReplicated
        {
            // Attempt to fetch first
            IReplicated value;
            if (this.TryGetValue(key, out value))
                return (T)value;

            // Add and return
            value = ReplicatedActivator.CreateInstance<T>();
            this.Add(key, value);
            return (T)value;
        }

        /// <summary>
        /// Creates a value that should be used for deserialization.
        /// </summary>
        /// <returns>The created value.</returns>
        protected sealed override IReplicated ReadValuePrefix(PacketReader reader)
        {
            // Read the identity of the type and create an instance for it
            return ReplicatedActivator.CreateInstance(reader.ReadInt32());
        }

        /// <summary>
        /// Writes a prefix for a value in this dictionary.
        /// </summary>
        /// <param name="value">The value we are writing.</param>
        /// <param name="writer">The writer to use.</param>
        protected sealed override void WriteValuePrefix(IReplicated value, PacketWriter writer)
        {
            // Write the identifier for the actual type
            writer.Write(value.GetType().ToIdentifier());
        }
    }
}