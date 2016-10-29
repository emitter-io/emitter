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

namespace Emitter.Network
{
    /// <summary>
    /// Represents an engine.io transport contract.
    /// </summary>
    public interface ITransport : IDisposable
    {
        /// <summary>
        /// Gets the encoder for this transport.
        /// </summary>
        Processor Encoder { get; }

        /// <summary>
        /// Gets the decoder for this transport.
        /// </summary>
        Processor Decoder { get; }

        /// <summary>
        /// Gets or sets the CORS Origin.
        /// </summary>
        string Origin { get; set; }

        /// <summary>
        /// Gets or sets a channel that can be used for sending the data back to the remote client.
        /// </summary>
        Connection Channel { get; set; }

        /// <summary>
        /// Gets or sets the state of the transport.
        /// </summary>
        TransportState State { get; set; }

        /// <summary>
        /// Writes a packet to the transport.
        /// </summary>
        /// <param name="packet">The packet to send.</param>
        /// <returns>Whether the packet was written successfully or not.</returns>
        bool Write(Packet packet);
    }

    /// <summary>
    /// Represents a transport type for the Fabric.
    /// </summary>
    public enum TransportType : byte
    {
        /// <summary>
        /// Represents a TCP transport type.
        /// </summary>
        Tcp = 0,

        /// <summary>
        /// Represents a UDP transport type.
        /// </summary>
        Udp = 1
    }

    /// <summary>
    /// Represents type of a engine.io packet.
    /// </summary>
    public enum TransportEvent : byte
    {
        /// <summary>
        /// Represents an 'open' event.
        /// </summary>
        Open = 0,

        /// <summary>
        /// Represents a 'close' event.
        /// </summary>
        Close = 1,

        /// <summary>
        /// Represents a 'ping' event.
        /// </summary>
        Ping = 2,

        /// <summary>
        /// Represents a 'pong' event.
        /// </summary>
        Pong = 3,

        /// <summary>
        /// Represents a 'message' event.
        /// </summary>
        Message = 4,

        /// <summary>
        /// Represents an 'upgrade' event.
        /// </summary>
        Upgrade = 5,

        /// <summary>
        /// Represents a 'no operation' event.
        /// </summary>
        Noop = 6
    }

    /// <summary>
    /// Represents type of a engine.io packet.
    /// </summary>
    public enum TransportEncoding : byte
    {
        /// <summary>
        /// Represents utf-8 string encoding.
        /// </summary>
        Text = 0,

        /// <summary>
        /// Represents binary encoding.
        /// </summary>
        Binary = 1
    }

    /// <summary>
    /// Represents the state of the transport.
    /// </summary>
    public enum TransportState : byte
    {
        /// <summary>
        /// Represents a closed transport.
        /// </summary>
        Closed = 0,

        /// <summary>
        /// Represents an open transport.
        /// </summary>
        Open = 1,

        /// <summary>
        /// Represents a transport currently being upgraded.
        /// </summary>
        Upgrading = 2
    }
}