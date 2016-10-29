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

namespace Emitter.Network
{
    /// <summary>
    /// Represents a data packet that can be send remotely.
    /// </summary>
    public abstract class Packet : RecyclableObject
    {
        /// <summary>
        /// Construcs a new instance of a <see cref="Packet"/>.
        /// </summary>
        public Packet()
        {
        }

        #region Public Properties

        /// <summary>
        /// Gets or sets the connection the packet originated from.
        /// </summary>
        public Emitter.Connection Origin;

        /// <summary>
        /// Gets or sets the lifetime policy of the packet.
        /// </summary>
        public PacketLifetime Lifetime;

        /// <summary>
        /// Gets the direction of this packet.
        /// </summary>
        public virtual PacketDirection Direction
        {
            get { return PacketDirection.Incoming; }
        }

        #endregion Public Properties

        #region IRecyclable Members

        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        public override void Recycle()
        {
            // Reset all other members
            Origin = null;
            Lifetime = PacketLifetime.Automatic;
        }

        #endregion IRecyclable Members
    }

    /// <summary>
    /// Represents a direction of the packet.
    /// </summary>
    public enum PacketDirection
    {
        /// <summary>
        /// Incoming direction, the client is sending the packet to the server.
        /// </summary>
        Incoming,

        /// <summary>
        /// Outgoing direction, the server is sending the packet to the client.
        /// </summary>
        Outgoing
    }

    /// <summary>
    /// Represents a complex type for packet serialization.
    /// </summary>
    public interface ISerializable
    {
        /// <summary>
        /// Reads the complex type from the stream.
        /// </summary>
        /// <param name="reader">The deserialization reader.</param>
        void Read(PacketReader reader);

        /// <summary>
        /// Writes the complex type to the stream.
        /// </summary>
        /// <param name="writer">The serialization writer.</param>
        void Write(PacketWriter writer);
    }

    /// <summary>
    /// Represents the lifetime management of the packet.
    /// </summary>
    public enum PacketLifetime : byte
    {
        /// <summary>
        /// The packet's lifetime should be managed by the runtime. This means
        /// that the packet can only be used once, and once it is sent the
        /// packet will be automatically disposed/released back to the memory
        /// pool.
        /// </summary>
        Automatic = 0,

        /// <summary>
        /// The packet lifetime should be managed manually, the user will
        /// be responsible releasing or disposing the packet.
        /// </summary>
        Manual = 1
    }
}