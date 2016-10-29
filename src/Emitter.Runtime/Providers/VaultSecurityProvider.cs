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
using System.Threading.Tasks;
using Emitter.Security;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a security provider where the secrets are stored in Hashicorp Vault.
    /// </summary>
    public sealed class VaultSecurityProvider : SecurityProvider
    {
        private readonly VaultClient Vault;
        private readonly string AppId;
        private readonly string UserId;

        /// <summary>
        /// Gets the provider friendly name.
        /// </summary>
        public override string Name
        {
            get { return "Vault"; }
        }

        /// <summary>
        /// Constructs a new instance of a provider.
        /// </summary>
        /// <param name="appId">The application ID to use.</param>
        /// <param name="userId">The user id</param>
        /// <param name="vaultAddress">The address of the vault</param>
        public VaultSecurityProvider(string vaultAddress, string appId, string userId) : base()
        {
            this.AppId = appId;
            this.UserId = userId;
            this.Vault = new VaultClient(vaultAddress);
        }

        /// <summary>
        /// Gets the secret from the provider.
        /// </summary>
        /// <param name="secretName">The name of the secret to retrieve.</param>
        /// <returns></returns>
        public override async Task<string> GetSecretAsync(string secretName)
        {
            if (!this.Vault.IsAuthenticated)
                await this.Vault.AuthenticateAsync(this.AppId, this.UserId);

            return (await this.Vault.ReadAsync<VaultData>(secretName))?.Value;
        }

        /// <summary>
        /// Gets the credentials asynchronously.
        /// </summary>
        /// <param name="credentialsName">The name </param>
        /// <returns></returns>
        public override async Task<ICredentials> GetCredentialsAsync(string credentialsName)
        {
            if (!this.Vault.IsAuthenticated)
                await this.Vault.AuthenticateAsync(this.AppId, this.UserId);

            var response = await this.Vault.GetAsync<AwsCredentials>("/aws/sts/" + credentialsName);
            if (response == null)
                return null;

            var credentials = response.Data;
            credentials.Duration = TimeSpan.FromSeconds(response.Duration);
            credentials.Expires = DateTime.UtcNow + credentials.Duration;
            return credentials;
        }
    }
}