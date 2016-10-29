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
using Emitter.Diagnostics;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Represents a decoder of WebSocket upgrade requests.
    /// </summary>
    public static unsafe class WebSocketUpgradeHandler
    {
        private const string UrlWebSocket = "/socket.io/1/websocket";
        private const string UrlFlashSocket = "/socket.io/1/flashsocket";
        private const string UrlPatternA = "/socket.io/?eio=3&transport=websocket";
        private const string UrlPatternB = "/engine.io/?eio=3&transport=websocket";

        /// <summary>
        /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
        /// </summary>
        /// <param name="channel">The through which the packet is coming/going out.</param>
        /// <param name="context">The packet context for this operation.</param>
        /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
        public static ProcessingState Process(Emitter.Connection channel, ProcessingContext context)
        {
            // Only handle stuff with HTTP Context
            var httpContext = context.GetSession<HttpContext>();
            if (httpContext == null)
                return ProcessingState.Failure;

            // Must have Upgrade: websocket
            string upgradeType;
            if (!httpContext.Request.TryGetHeader("Upgrade", out upgradeType))
                return ProcessingState.Failure;
            if (String.IsNullOrWhiteSpace(upgradeType))
                return ProcessingState.Failure;
            if (upgradeType.ToLowerInvariant() != "websocket")
                return ProcessingState.Failure;

            // This should not be tested with Fiddler2 or another proxy
            NetTrace.WriteLine("Websocket transport requested", channel, NetTraceCategory.WebSocket);

            // Get the websocket handler
            var handler = WebSocketFactory.GetHandler(httpContext.Request, channel, null, null);
            if (handler == null)
                throw new WebSocketException("Requested websocket version is not supported");

            // Upgrade to websocket (handshake)
            var upgrade = handler.Upgrade(context, httpContext);

            // Trace the websocket upgrade
            NetTrace.WriteLine("WebSocket upgrade done", channel, NetTraceCategory.WebSocket);

            // If it's socket.io 1.x, setup the transport
            /*if (httpContext.Request.IsFabric)
            {
                // Set the pipeline
                channel.Encoding.PipelineAddBeforeOrLast(
                    upgrade.Encoder,
                    Encode.WebSocket);
                channel.Decoding.PipelineAddAfterOrFirst(
                    upgrade.Decoder,
                    Decode.WebSocket);

                // Make sure the session is set
                var session = FabricRegistry.Get(httpContext);
                if (session != null)
                {
                    // Unbind the old one
                    channel.Client.UnbindChannel(channel);

                    // Set the session
                    channel.Client = session;
                    channel.FabricSession = session.Handle;
                    channel.FabricEncoding = httpContext.GetFabricEncoding();
                }
            }*/

            // Trace the channel upgrade
            NetTrace.WriteLine(channel + " was upgraded to websocket transport", channel, NetTraceCategory.Channel);

            // There still might be SSL encoding in the pipeline
            return ProcessingState.Stop;
        }
    }
}