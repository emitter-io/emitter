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

using System.Linq;
using Emitter.Security;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a handler for key generation api request.
    /// </summary>
    internal class HandlePresence
    {
        /// <summary>
        /// Attempts to generate the key and returns the result.
        /// </summary>
        /// <param name="client">The remote client.</param>
        /// <param name="channel">The full channel string.</param>
        /// <param name="message">The message to publish.</param>
        public static EmitterResponse Process(IClient client, EmitterChannel channel, PresenceRequest request)
        {
            // Parse the channel
            EmitterChannel info;
            if (!EmitterChannel.TryParse(request.Channel, false, out info))
                return EmitterError.BadRequest;

            // Should be a static (non-wildcard) channel string.
            if (info.Type != ChannelType.Static)
                return EmitterError.BadRequest;

            // Attempt to parse the key, this should be a master key
            SecurityKey channelKey;
            if (!SecurityKey.TryParse(request.Key, out channelKey) || !channelKey.HasPermission(SecurityAccess.Presence) || channelKey.IsExpired)
                return EmitterError.Unauthorized;

            // Make sure the channel name ends with
            if (!request.Channel.EndsWith("/"))
                request.Channel += "/";

            // Subscription
            Subscription sub;
            if (request.Changes)
            {
                // If we requested changes, register a subscription with this new interest
                Subscription.Register(client, channelKey.Contract, request.Channel, SubscriptionInterest.Presence);
            }
            else
            {
                // If the changes flag is set to false, unregister the client from presence
                Subscription.Unregister(client, channelKey.Contract, request.Channel, SubscriptionInterest.Presence, out sub);
            }

            // If we didn't request a presence, send an OK response only
            if (!request.Status)
                return new EmitterResponse(200);

            // Find the subscription
            if (!Subscription.TryGet(channelKey.Contract, request.Channel, out sub))
            {
                // No subscription found, return an empty one
                return new PresenceResponse(request.Channel, 0, null);
            }

            // Current list
            var presenceList = sub.Presence
                .Where(p => !p.Deleted)
                .Take(1000)
                .Select(p => p.Value.AsInfo())
                .ToList();

            // If we requested a presence, prepare a response
            return new PresenceResponse(request.Channel, sub.Occupancy, presenceList);
        }
    }
}