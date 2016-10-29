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
using Emitter.Text.Json;

namespace Emitter.Security
{
    /// <summary>
    /// Represents app-id authentication response.
    /// </summary>
    public class VaultResponse
    {
        /// <summary>
        /// Gets the lease identifier.
        /// </summary>
        [JsonProperty("lease_id")]
        public string Id;

        /// <summary>
        /// Gets whether the lease is renewable or not.
        /// </summary>
        [JsonProperty("renewable")]
        public bool Renewable;

        /// <summary>
        /// Gets the lease duration.
        /// </summary>
        [JsonProperty("lease_duration")]
        public int Duration;

        /// <summary>
        /// Gets the auth response.
        /// </summary>
        [JsonProperty("auth")]
        public VaultAuth Authentication;
    }

    /// <summary>
    /// Represents app-id authentication response.
    /// </summary>
    public class VaultResponse<T> : VaultResponse
    {
        /// <summary>
        /// Gets the data for this response.
        /// </summary>
        [JsonProperty("data")]
        public T Data;
    }

    /// <summary>
    /// Represents app-id authentication response.
    /// </summary>
    public class VaultData
    {
        /// <summary>
        /// Gets the secret value.
        /// </summary>
        [JsonProperty("value")]
        public string Value;
    }

    /// <summary>
    /// Represents app-id authentication response.
    /// </summary>
    public class VaultAuth
    {
        /// <summary>
        /// Gets the client token.
        /// </summary>
        [JsonProperty("client_token")]
        public string Token;

        /// <summary>
        /// Gets the list of policies
        /// </summary>
        [JsonProperty("policies")]
        public string[] Policies;

        /// <summary>
        /// Gets whether the lease is renewable or not.
        /// </summary>
        [JsonProperty("renewable")]
        public bool Renewable;

        /// <summary>
        /// Gets the lease duration.
        /// </summary>
        [JsonProperty("lease_duration")]
        public int Duration;
    }

    public class AwsCredentials : ICredentials
    {
        /// <summary>
        /// The access key.
        /// </summary>
        [JsonProperty("access_key")]
        public string AccessKey { get; set; }

        /// <summary>
        /// The secret key.
        /// </summary>
        [JsonProperty("secret_key")]
        public string SecretKey { get; set; }

        /// <summary>
        /// The token.
        /// </summary>
        [JsonProperty("security_token")]
        public string Token { get; set; }

        /// <summary>
        /// The duration of the credentials.
        /// </summary>
        [JsonIgnore]
        public TimeSpan Duration { get; set; }

        /// <summary>
        /// The expiration date of the credentials.
        /// </summary>
        [JsonIgnore]
        public DateTime Expires { get; set; }
    }
}