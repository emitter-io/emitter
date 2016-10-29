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
    /// Represents a decoder for simple TCP packets.
    /// </summary>
    internal sealed unsafe class TcpTransport : TransportBase
    {
        public TcpTransport()
        {
            this.State = TransportState.Open;
        }

        /// <summary>
        /// Gets the encoder for this transport.
        /// </summary>
        public sealed override Processor Encoder
        {
            get { return null; }
        }

        /// <summary>
        /// Gets the decoder for this transport.
        /// </summary>
        public sealed override Processor Decoder
        {
            get { return null; }
        }

        /// <summary>
        /// Writes a packet to the transport.
        /// </summary>
        /// <param name="packet">The packet to send.</param>
        /// <returns>Whether the packet was written successfully or not.</returns>
        public sealed override bool Write(Packet packet)
        {
            // Websocket simply sends the packet
            this.Channel.Send(packet);
            return true;
        }
    }
}