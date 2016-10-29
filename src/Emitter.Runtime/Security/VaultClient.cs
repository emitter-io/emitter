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
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Threading.Tasks;
using Emitter.Network.Http;
using Emitter.Text.Json;

namespace Emitter.Security
{
    /// <summary>
    /// Represents a vault client.
    /// </summary>
    public class VaultClient
    {
        private readonly string VaultUrl;
        private string Token;

        /// <summary>
        /// Constructs a new instance of a provider.
        /// </summary>
        /// <param name="vaultAddress">The address of the vault</param>
        public VaultClient(string vaultAddress) : base()
        {
            this.VaultUrl = string.Format("http://{0}:8200", vaultAddress);
        }

        /// <summary>
        /// Checks whether we are authenticated or not.
        /// </summary>
        public bool IsAuthenticated
        {
            get { return this.Token != null; }
        }

        /// <summary>
        /// Performs vault authentication.
        /// </summary>
        /// <param name="app"></param>
        /// <param name="user"></param>
        /// <returns></returns>
        public async Task AuthenticateAsync(string app, string user)
        {
            try
            {
                // Authenticate using app-id authentication
                var response = await PostAsync("/auth/app-id/login", new
                {
                    app_id = app,
                    user_id = user
                });

                // Unable to authentify within Vault
                if (response == null || response.Authentication == null)
                {
                    Service.Logger.Log(LogLevel.Error, "Unable to perform vault authentication for user " + user);
                    await Task.Delay(120000);
                    this.Token = null;
                    return;
                }

                // Set the token in the client itself
                this.Token = response.Authentication.Token;
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                this.Token = null;
            }
        }

        /// <summary>
        /// Reads the secret from the vault.
        /// </summary>
        /// <param name="secretName">The name of the secret to read.</param>
        /// <returns></returns>
        public async Task<T> ReadAsync<T>(string secretName)
        {
            try
            {
                // Authenticate using app-id authentication
                var response = await GetAsync<T>("/secret/" + secretName);
                if (response == null)
                    return default(T);

                // Return the secret data
                return response.Data;
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                return default(T);
            }
        }

        /// <summary>
        /// Posts a request to Vault.
        /// </summary>
        public async Task<VaultResponse<T>> GetAsync<T>(string url)
        {
            var response = await HttpUtility.GetAsync(this.VaultUrl + "/v1" + url, 5000,
                new KeyValuePair<string, string>("X-Vault-Token", this.Token)
                );

            if (response.Success)
                return JsonConvert.DeserializeObject<VaultResponse<T>>(response.Value);
            return null;
        }

        /// <summary>
        /// Posts a request to Vault.
        /// </summary>
        public async Task<VaultResponse> PostAsync(string url, object request)
        {
            var content = new StringContent(JsonConvert.SerializeObject(request));
            if (this.Token != null)
                content.Headers.Add("X-Vault-Token", this.Token);

            var response = await HttpUtility.PostAsync(this.VaultUrl + "/v1" + url, content, 5000);
            if (response.Success)
                return JsonConvert.DeserializeObject<VaultResponse>(response.Value);
            return null;
        }
    }
}