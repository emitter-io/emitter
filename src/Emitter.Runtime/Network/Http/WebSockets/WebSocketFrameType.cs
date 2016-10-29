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

namespace Emitter.Network.Http
{
    /// <summary>
    /// Defines a type of a websocket frame
    /// </summary>
    public enum WebSocketFrameType : byte
    {
        /// <summary>
        /// Represents a websocket frame which is a continuation of a previous frame.
        /// </summary>
        Continuation = 0,

        /// <summary>
        /// Represents a websocket textual frame.
        /// </summary>
        Text = 1,

        /// <summary>
        /// Represents a websocket binary frame.
        /// </summary>
        Binary = 2,

        /// <summary>
        /// Represents a websocket frame which is used to close the connection.
        /// </summary>
        Close = 8,

        /// <summary>
        /// Represents a websocket frame which is used to ping the remote end point.
        /// </summary>
        Ping = 9,

        /// <summary>
        /// Represents a websocket frame which is used to reply to the ping of the remote end point.
        /// </summary>
        Pong = 10,
    }
}