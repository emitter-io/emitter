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
    /// Represents an error response.
    /// </summary>
    internal sealed class EmitterError : EmitterResponse
    {
        /// <summary>
        /// Constructs a new error.
        /// </summary>
        /// <param name="status">The status code for the error.</param>
        /// <param name="message">The message for the error.</param>
        public EmitterError(EmitterEventCode status, string message)
        {
            this.Status = (int)status;
            this.Message = message;
        }

        /// <summary>
        /// The error message for this error.
        /// </summary>
        [JsonProperty("message", NullValueHandling = NullValueHandling.Ignore)]
        public string Message;

        public static readonly EmitterError NotImplemented = new EmitterError(EmitterEventCode.NotImplemented, "The server either does not recognize the request method, or it lacks the ability to fulfill the request.");
        public static readonly EmitterError BadRequest = new EmitterError(EmitterEventCode.BadRequest, "The request was invalid or cannot be otherwise served.");
        public static readonly EmitterError Unauthorized = new EmitterError(EmitterEventCode.Unauthorized, "The security key provided is not authorized to perform this operation.");
        public static readonly EmitterError PaymentRequired = new EmitterError(EmitterEventCode.PaymentRequired, "The request can not be served, as the payment is required to proceed.");
        public static readonly EmitterError Forbidden = new EmitterError(EmitterEventCode.Forbidden, "The request is understood, but it has been refused or access is not allowed.");
        public static readonly EmitterError NotFound = new EmitterError(EmitterEventCode.NotFound, "The resource requested does not exist.");
        public static readonly EmitterError ServerError = new EmitterError(EmitterEventCode.ServerError, "An unexpected condition was encountered and no more specific message is suitable.");
    }
}