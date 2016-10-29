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
using Emitter.Text.Json;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a key generation request.
    /// </summary>
    internal class KeyGenRequest
    {
        /// <summary>
        /// The master key for this request.
        /// </summary>
        [JsonProperty("key")]
        public string Key = null;

        /// <summary>
        /// The target channel for this request.
        /// </summary>
        [JsonProperty("channel")]
        public string Channel = null;

        /// <summary>
        /// The target access type (r, w, rw).
        /// </summary>
        [JsonProperty("type", NullValueHandling = NullValueHandling.Ignore)]
        public string Type = null;

        /// <summary>
        /// The time to live for the key.
        /// </summary>
        [JsonProperty("ttl", NullValueHandling = NullValueHandling.Ignore)]
        public int Ttl = 0;

        /// <summary>
        /// Gets the security access type.
        /// </summary>
        [JsonIgnore]
        public SecurityAccess Access
        {
            get
            {
                var required = SecurityAccess.None;
                if (this.Type.Contains("r"))
                    required |= SecurityAccess.Read;
                if (this.Type.Contains("w"))
                    required |= SecurityAccess.Write;
                if (this.Type.Contains("s"))
                    required |= SecurityAccess.Store;
                if (this.Type.Contains("l"))
                    required |= SecurityAccess.Load;
                if (this.Type.Contains("p"))
                    required |= SecurityAccess.Presence;
                return required;
            }
        }

        /// <summary>
        /// Gets the expiration date.
        /// </summary>
        public DateTime Expires
        {
            get
            {
                // Generate expiration date
                return this.Ttl > 0
                    ? DateTime.UtcNow.AddSeconds(this.Ttl)
                    : DateTime.MinValue;
            }
        }
    }
}