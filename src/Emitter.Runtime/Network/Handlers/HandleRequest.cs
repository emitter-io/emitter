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
using Emitter.Text.Json;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a handler for emitter API requests.
    /// </summary>
    internal static unsafe class HandleRequest
    {
        /// <summary>
        /// Attempts to generate the key and returns the result.
        /// </summary>
        /// <param name="client">The remote client.</param>
        /// <param name="info">The full channel string.</param>
        /// <param name="message">The message to publish.</param>
        public static bool TryProcess(IClient client, EmitterChannel info, ArraySegment<byte> message)
        {
            // Is this a special api request?
            if (info.Key != EmitterConst.ApiPrefix)
                return false;

            // The response to send out, default to bad request
            EmitterResponse response = EmitterError.BadRequest;
            try
            {
                switch (info.Target)
                {
                    case 548658350:
                        // Handle the 'keygen' request.
                        response = HandleKeyGen.Process(client, info, message.As<KeyGenRequest>());
                        return true;

                    case 3869262148:
                        // Handle the 'presence' request.
                        response = HandlePresence.Process(client, info, message.As<PresenceRequest>());
                        return true;

                    default:
                        // We've got a bad request
                        response = EmitterError.BadRequest;
                        return true;
                }
            }
            catch (NotImplementedException)
            {
                // We've got a not implemented exception
                response = EmitterError.NotImplemented;
                return true;
            }
            catch (Exception ex)
            {
                // We've got a an internal error
                Service.Logger.Log(ex);
                response = EmitterError.ServerError;
                return true;
            }
            finally
            {
                // Send the response, always
                if (response != null)
                    SendResponse(client, "emitter/" + info.Channel, response);
            }
        }

        /// <summary>
        /// Sends a response to the client.
        /// </summary>
        /// <param name="client">The client to reply to.</param>
        /// <param name="response">The emitte response to send.</param>
        private static void SendResponse(IClient client, string channel, EmitterResponse response)
        {
            // Serialize the response
            var serialized = JsonConvert.SerializeObject(response, Formatting.Indented);
            if (serialized == null)
                return;

            // Send the message out
            var msg = MqttPublishPacket.Acquire();
            msg.Channel = channel;
            msg.Message = serialized.AsUTF8().AsSegment();
            client.Send(msg);
        }
    }
}