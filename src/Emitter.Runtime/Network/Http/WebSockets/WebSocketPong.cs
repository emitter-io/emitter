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

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a pong packet send on a websocket ping event.
    /// </summary>
    internal sealed class WebSocketPong : BytePacket
    {
        #region Static Members

        /// <summary>
        /// The pool of transport packets.
        /// </summary>
        private readonly static PacketPool<WebSocketPong> Pool =
            new PacketPool<WebSocketPong>("WebSocket Pongs", (p) => new WebSocketPong());

        /// <summary>
        /// Acquires a <see cref="WebSocketPong"/> instance.
        /// </summary>
        /// <param name="payload">The payload to copy.</param>
        /// <returns> A <see cref="WebSocketPong"/> that can be sent to the remote client.</returns>
        public static WebSocketPong Acquire(BufferSegment payload)
        {
            var packet = Pool.Acquire();
            packet.Buffer = payload.AsArray();
            return packet;
        }

        #endregion Static Members
    }
}