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

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents mesh command packet.
    /// </summary>
    public abstract class MeshPacket : Packet
    {
        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="Reader">PacketReader used to serialize the packet.</param>
        public abstract void Read(PacketReader Reader);

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="Writer">PacketWriter used to deserialize the packet.</param>
        public abstract void Write(PacketWriter Writer);
    }

    /// <summary>
    /// Represents a command type.
    /// </summary>
    public enum MeshEventType : byte
    {
        /// <summary>
        /// Represents an error event.
        /// </summary>
        Error = 0,

        /// <summary>
        /// Represents a ping event, used for heartbeats between mesh servers.
        /// </summary>
        Ping = 1,

        /// <summary>
        /// Represents a ping ack, used for heartbeats between servers.
        /// </summary>
        PingAck = 2,

        /// <summary>
        /// Represents a mesh handshake request.
        /// </summary>
        Handshake = 3,

        /// <summary>
        /// Represents a mesh handshake response.
        /// </summary>
        HandshakeAck = 4,

        /// <summary>
        /// Represents a gossip message with a digest (scuttlebutt).
        /// </summary>
        GossipDigest = 5,

        /// <summary>
        /// Represents a gossip message asking for an update.
        /// </summary>
        GossipSince = 6,

        /// <summary>
        /// Represents a gossip message which contains an update about a state.
        /// </summary>
        GossipUpdate = 7,

        /// <summary>
        /// Represents a subscribe command, notifying the server to subscribe to a particular channel.
        /// </summary>
        Subscribe = 10,

        /// <summary>
        /// Represents an unsubscribe command, telling the server to unsubscribe from a particular channel.
        /// </summary>
        Unsubscribe = 11,

        /// <summary>
        /// Represents a custom event
        /// </summary>
        Custom = 255
    }
}