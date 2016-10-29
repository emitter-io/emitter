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
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// WebSocket factory that constructs a protocol handler for a particular websocket protocol
    /// </summary>
    internal static class WebSocketFactory
    {
        internal static readonly WebSocketDraft76Upgrade Draft76 = new WebSocketDraft76Upgrade();
        internal static readonly WebSocketHybi13Upgrade Hybi13 = new WebSocketHybi13Upgrade();

        /// <summary>
        /// Builds a websocket handler for the incoming http request.
        /// </summary>
        /// <param name="request">There upgrade request that upgrades the connection to a websocket one.</param>
        /// <param name="onMessage">A delegate to be invoked on each incoming message.</param>
        /// <param name="onClose">A delecate to be invoked when a connection is closed.</param>
        /// <param name="channel">The underlying channel that requests the upgrade</param>
        /// <returns>The websocket protocol handler constructed.</returns>
        public static IWebSocketProtocol GetHandler(HttpRequest request, Emitter.Connection channel, Action<string> onMessage, Action onClose)
        {
            // Get the version number that is requested
            string version = GetVersion(request);

            // Trace a websocket event
            NetTrace.WriteLine("Requested upgrade to version " + version, channel, NetTraceCategory.WebSocket);

            // Get the handler
            switch (version)
            {
                case "76": return WebSocketFactory.Draft76;
                case "7": return WebSocketFactory.Hybi13;
                case "8": return WebSocketFactory.Hybi13;
                case "13": return WebSocketFactory.Hybi13;
                case "75": return WebSocketFactory.Hybi13; // not sure!
            }

            throw new WebSocketException("Unsupported websocket request. Protocol version requested: " + version);
        }

        /// <summary>
        /// Attempts to detect the version of the websocket protocol
        /// </summary>
        /// <param name="request">Incoming HTTP Request</param>
        /// <returns>The version of the protocol to use</returns>
        private static string GetVersion(HttpRequest request)
        {
            string version;
            if (request.TryGetHeader("Sec-WebSocket-Version", out version))
                return version;

            if (request.TryGetHeader("Sec-WebSocket-Draft", out version))
                return version;

            if (request.Headers["Sec-WebSocket-Key1"] != null)
                return "76";

            return "75";
        }
    }
}