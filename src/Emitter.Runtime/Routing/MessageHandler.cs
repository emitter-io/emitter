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
using Emitter.Network.Mesh;

namespace Emitter
{
    /// <summary>
    /// Represents the messaging layer.
    /// </summary>
    internal sealed partial class MessageHandler : IMeshHandler
    {
        /// <summary>
        /// Handles frames of the mesh layer.
        /// </summary>
        /// <param name="server">The server sending the frame.</param>
        /// <param name="frame">The data frame to process.</param>
        /// <returns>The processing state of the event.</returns>
        public ProcessingState ProcessFrame(IServer server, ArraySegment<byte> frame)
        {
            MessageFrame.TryParse(frame);
            return ProcessingState.Success;
        }

        /// <summary>
        /// Handles the custom commands, JSON-encoded.
        /// </summary>
        /// <param name="server">The server sending the command.</param>
        /// <param name="packet">The event payload.</param>
        /// <returns>The processing state of the event.</returns>
        public ProcessingState ProcessEvent(IServer server, MeshEvent packet)
        {
            // Must be emitter event
            var ev = packet as MeshEmitterEvent;
            if (ev == null)
                return ProcessingState.Failure;

            Service.Logger.Log("Event: " + ev + " from " + server.EndPoint);

            switch (ev.Type)
            {
                // Handles subscribe event
                case MeshEventType.Subscribe:
                    Subscription.Register(server, ev.Contract, ev.Channel);
                    return ProcessingState.Success;

                // Handles unsubscribe event
                case MeshEventType.Unsubscribe:
                    Subscription.Unregister(server, ev.Contract, ev.Channel);
                    return ProcessingState.Success;

                default:
                    Service.Logger.Log(LogLevel.Info, "Unknown event: " + packet);
                    return ProcessingState.Failure;
            }
        }
    }
}