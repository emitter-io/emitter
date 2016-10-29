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
using Emitter.Network;
using TKey = System.String;
using TNodeKey = System.Int32;

namespace Emitter.Replication
{
    /// <summary>
    /// Represends a contract for an entity that can be replicated.
    /// </summary>
    public interface IReplicated : ISerializable, IComparable<IReplicated>
    {
        /// <summary>
        /// Gets or sets the key of the entry.
        /// </summary>
        TKey Key
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the parent of the replicated object.
        /// </summary>
        IReplicated Parent
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the version of the replicated object.
        /// </summary>
        ReplicatedVersion Version
        {
            get;
            set;
        }

        /// <summary>
        /// Writes the complex type to the stream.
        /// </summary>
        /// <param name="writer">The serialization writer.</param>
        /// <param name="since">The minimum version to replicate.</param>
        void Write(PacketWriter writer, ReplicatedVersion since);
    }

    /// <summary>
    /// Notifies that a child was updated
    /// </summary>
    public interface IReplicatedCollection : IEnumerable
    {
        /// <summary>
        /// Occurs when a child is updated.
        /// </summary>
        /// <param name="key">The key to update.</param>
        void OnChildUpdate(TKey key);

        /// <summary>
        /// Merges in another replicated value.
        /// </summary>
        /// <param name="other">The other value to merge.</param>
        void MergeIn(TNodeKey source, IReplicatedCollection other);

        /// <summary>
        /// Gets the enumerable which iterates throug all entries, including the tombstones.
        /// </summary>
        /// <returns>The enumerable.</returns>
        IEnumerable GetEntries();
    }
}