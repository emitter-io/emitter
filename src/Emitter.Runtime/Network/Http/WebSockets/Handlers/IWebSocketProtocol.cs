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
    /// The contract that is defined for handling web socket protocol.
    /// </summary>
    public interface IWebSocketProtocol
    {
        /// <summary>
        /// Upgrades the connection to the particular protocol. Handles the handshake.
        /// </summary>
        /// <param name="context"><see cref="ProcessingContext"/> for the current connection.</param>
        /// <param name="httpContext"><see cref="HttpContext"/> for the current connection.</param>
        /// <returns>The handlers that have been inserted in the pipeline.</returns>
        WebSocketPipeline Upgrade(ProcessingContext context, HttpContext httpContext);
    }

    /// <summary>
    /// Represents a pair of handlers for a websocket protocol.
    /// </summary>
    public struct WebSocketPipeline
    {
        /// <summary>
        /// Constructs a pipeline response.
        /// </summary>
        /// <param name="encoder">The processor used for encoding.</param>
        /// <param name="decoder">The processor used for decoding.</param>
        public WebSocketPipeline(Processor encoder, Processor decoder)
        {
            this.Encoder = encoder;
            this.Decoder = decoder;
        }

        /// <summary>
        /// Gets the decoder used for decoding websocket requests.
        /// </summary>
        public readonly Processor Decoder;

        /// <summary>
        /// Gets the encoder used for encoding websocket requests.
        /// </summary>
        public readonly Processor Encoder;
    }
}