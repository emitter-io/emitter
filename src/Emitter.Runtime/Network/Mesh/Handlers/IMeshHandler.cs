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

namespace Emitter.Network.Mesh
{
    /// <summary>
    /// Represents a contract that a Mesh message handler should implement to handle a received message.
    /// </summary>
    public interface IMeshHandler
    {
        /// <summary>
        /// Handles the messages of the Mesh layer.
        /// </summary>
        /// <param name="server">The server which is sending the frame.</param>
        /// <param name="frame">The data frame to process.</param>
        /// <returns>The processing state of the event.</returns>
        ProcessingState ProcessFrame(IServer server, ArraySegment<byte> frame);

        /// <summary>
        /// Handles the custom events.
        /// </summary>
        /// <param name="server">The server which is sending the command.</param>
        /// <param name="event">The event.</param>
        /// <returns>The processing state of the event.</returns>
        ProcessingState ProcessEvent(IServer server, MeshEvent @event);
    }
}