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

using Emitter.Network;

namespace Emitter.Replication
{
    /// <summary>
    /// Represents a digest of a replicated version.
    /// </summary>
    public struct ReplicatedVersionDigest : ISerializable
    {
        /// <summary>
        /// Gets or sets the digest value.
        /// </summary>
        public int Value;

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public void Read(PacketReader reader)
        {
            this.Value = reader.ReadInt32();
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        public void Write(PacketWriter writer)
        {
            writer.Write(this.Value);
        }
    }
}