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
using Emitter.Network.Http;
using Emitter.Security;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a handler for key generation api request.
    /// </summary>
    internal class HandleKeyGen : IHttpHandler
    {
        /// <summary>
        /// Attempts to generate the key and returns the result.
        /// </summary>
        /// <param name="client">The remote client.</param>
        /// <param name="channel">The full channel string.</param>
        /// <param name="message">The message to publish.</param>
        public static EmitterResponse Process(IClient client, EmitterChannel channel, KeyGenRequest request)
        {
            // Parse the channel
            EmitterChannel info;
            if (!EmitterChannel.TryParse(request.Channel, false, out info))
                return EmitterError.BadRequest;

            // Should be a static (non-wildcard) channel string.
            if (info.Type != ChannelType.Static)
                return EmitterError.BadRequest;

            // Attempt to parse the key, this should be a master key
            SecurityKey masterKey;
            if (!SecurityKey.TryParse(request.Key, out masterKey) || !masterKey.IsMaster || masterKey.IsExpired)
                return EmitterError.Unauthorized;

            // Attempt to fetch the contract using the key. Underneath, it's cached.
            var contract = Services.Contract.GetByKey(masterKey.Contract) as EmitterContract;
            if (contract == null)
                return EmitterError.NotFound;

            // Validate the contract
            if (!contract.Validate(ref masterKey))
                return EmitterError.Unauthorized;

            // Generate the key
            var key = SecurityKey.Create();
            key.Master = masterKey.Master;
            key.Contract = contract.Oid;
            key.Signature = contract.Signature;
            key.Permissions = request.Access;
            key.Target = SecurityHash.GetHash(request.Channel);
            key.Expires = request.Expires;

            return new KeyGenResponse()
            {
                Key = key.Value,
                Channel = request.Channel
            };
        }

        /// <summary>
        /// Enables processing of HTTP Web requests by a custom HttpHandler that implements
        /// the IHttpHandler interface.
        /// </summary>
        /// <param name="context">An HttpContext object that provides references to the intrinsic
        /// server objects (for example, Request, Response, Session, and Server) used
        /// to service HTTP requests.</param>
        public bool CanHandle(HttpContext context, HttpVerb verb, string url)
        {
            return verb == HttpVerb.Get && url.EndsWith("/keygen");
        }

        /// <summary>
        /// Checks whether the handler can handle an incoming request or not
        /// </summary>
        /// <param name="context">An HttpContext object that provides references to the intrinsic
        /// server objects (for example, Request, Response, Session, and Server) used
        /// to service HTTP requests.</param>
        /// <param name="verb">Verb of the request</param>
        /// <param name="url">Url passed in parameter</param>
        /// <returns></returns>
        public void ProcessRequest(HttpContext context)
        {
            context.Response.Write(this.Content.Value);
        }

        /// <summary>
        /// Load the content lazily.
        /// </summary>
        private Lazy<byte[]> Content = new Lazy<byte[]>(() =>
        {
            var response = HttpUtility.Get("http://s3-eu-west-1.amazonaws.com/cdn.emitter.io/web/keygen.html", 30000);
            if (response.Success && response.HasValue)
                return response.Value.AsUTF8();
            return ArrayUtils<byte>.Empty;
        });
    }
}