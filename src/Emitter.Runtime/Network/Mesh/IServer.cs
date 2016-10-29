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
using System.Net;

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a contract for a Mesh Server.
    /// </summary>
    public interface IServer : IPacketSender
    {
        /// <summary>
        /// Gets the server identifier.
        /// </summary>
        IPEndPoint EndPoint { get; }

        /// <summary>
        /// Gets the state of the node.
        /// </summary>
        ServerState State { get; }

        /// <summary>
        /// Gets or sets the session object for the server.
        /// </summary>
        object Session { get; set; }
    }
}