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
using Emitter.Security;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a handler of various Emitter requests.
    /// </summary>
    internal static unsafe class HandleUnsubscribe
    {
        /// <summary>
        /// Occurs when the remote client attempts to unsubscribe from a hub.
        /// </summary>
        /// <param name="client">The remote client.</param>
        /// <param name="channel">The full channel string.</param>
        public static EmitterEventCode Process(IClient client, string channel)
        {
            try
            {
                // Parse the channel
                EmitterChannel info;
                if (!EmitterChannel.TryParse(channel, true, out info))
                    return EmitterEventCode.BadRequest;

                // Attempt to parse the key
                SecurityKey key;
                if (!SecurityKey.TryParse(info.Key, out key))
                    return EmitterEventCode.BadRequest;

                // Has the key expired?
                if (key.IsExpired)
                    return EmitterEventCode.Unauthorized;

                // Have we already unsubscribed?
                //if (client[channel] == null)
                //    return EmitterEventCode.Success;

                // Attempt to fetch the contract using the key. Underneath, it's cached.
                var contract = Services.Contract.GetByKey(key.Contract) as EmitterContract;
                if (contract == null)
                    return EmitterEventCode.NotFound;

                // Check if the payment state is valid
                if (contract.Status == EmitterContractStatus.Refused)
                    return EmitterEventCode.PaymentRequired;

                // Validate the contract
                if (!contract.Validate(ref key))
                    return EmitterEventCode.Unauthorized;

                // Check if the key has the permission to read here
                if (!key.HasPermission(SecurityAccess.Read))
                    return EmitterEventCode.Unauthorized;

                // Check if the key has the permission for the required channel
                if (info.Target != key.Target)
                    return EmitterEventCode.Unauthorized;

                // Get the subscription id from the storage
                if (!Dispatcher.Unsubscribe(client, key.Contract, info.Channel, SubscriptionInterest.Messages))
                    return EmitterEventCode.Unauthorized;

                // Successfully unsubscribed
                return EmitterEventCode.Success;
            }
            catch (NotImplementedException)
            {
                // We've got a not implemented exception
                return EmitterEventCode.NotImplemented;
            }
            catch (Exception ex)
            {
                // We need to log it
                Service.Logger.Log(ex);

                // We've got a an internal error
                return EmitterEventCode.ServerError;
            }
        }
    }
}