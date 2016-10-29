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
using Emitter.Network;
using TKey = System.String;

namespace Emitter.Replication
{
    /// <summary>
    /// Represents a value that can be replicated.
    /// </summary>
    public abstract class ReplicatedObject : IReplicated
    {
        /// <summary>
        /// Gets or sets the key of the entry.
        /// </summary>
        public TKey Key
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the parent of the replicated object.
        /// </summary>
        public IReplicated Parent
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the version of the replicated object.
        /// </summary>
        public ReplicatedVersion Version
        {
            get;
            set;
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        /// <param name="since">The minimum version of updates to pack.</param>
        public abstract void Write(PacketWriter writer, ReplicatedVersion since);

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public abstract void Read(PacketReader reader);

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        public virtual void Write(PacketWriter writer)
        {
            throw new NotSupportedException("The replicated object should be serialized with a version parameter.");
        }

        /// <summary>
        /// Compares the value.
        /// </summary>
        /// <param name="other"></param>
        /// <returns></returns>
        public int CompareTo(IReplicated other)
        {
            return (int)(this.Version.Own - other.Version.Own);
        }
    }
}