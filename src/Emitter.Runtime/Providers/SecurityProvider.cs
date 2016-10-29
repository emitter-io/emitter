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
using System.Security.Cryptography;
using System.Threading.Tasks;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider that provides a security mechanism for Emitter Engine.
    /// </summary>
    public abstract class SecurityProvider : Provider
    {
        private static RandomNumberGenerator fRandomGenerator;

        private static readonly char[] fEncoding = new char[] {
        'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p',
        'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5'
        };

        /// <summary>
        /// Generates a secure session token.
        /// </summary>
        /// <returns>Returns a secure session token.</returns>
        public string CreateSessionToken()
        {
            if (fRandomGenerator == null)
                fRandomGenerator = RandomNumberGenerator.Create();

            byte[] data = new byte[15];
            fRandomGenerator.GetBytes(data);
            return Encode(data);
        }

        private static string Encode(byte[] buffer)
        {
            char[] chArray = new char[0x18];
            int num2 = 0;
            for (int i = 0; i < 15; i += 5)
            {
                int num4 = ((buffer[i] | (buffer[i + 1] << 8)) | (buffer[i + 2] << 0x10)) | (buffer[i + 3] << 0x18);
                int index = num4 & 0x1f;
                chArray[num2++] = fEncoding[index];
                index = (num4 >> 5) & 0x1f;
                chArray[num2++] = fEncoding[index];
                index = (num4 >> 10) & 0x1f;
                chArray[num2++] = fEncoding[index];
                index = (num4 >> 15) & 0x1f;
                chArray[num2++] = fEncoding[index];
                index = (num4 >> 20) & 0x1f;
                chArray[num2++] = fEncoding[index];
                index = (num4 >> 0x19) & 0x1f;
                chArray[num2++] = fEncoding[index];
                num4 = ((num4 >> 30) & 3) | (buffer[i + 4] << 2);
                index = num4 & 0x1f;
                chArray[num2++] = fEncoding[index];
                index = (num4 >> 5) & 0x1f;
                chArray[num2++] = fEncoding[index];
            }
            return new string(chArray);
        }

        /// <summary>
        /// Gets the secret from the provider.
        /// </summary>
        /// <param name="secretName">The name of the secret to retrieve.</param>
        /// <returns>The secret value.</returns>
        public virtual string GetSecret(string secretName)
        {
            var task = this.GetSecretAsync(secretName);
            task.Wait();
            return task.Result;
        }

        /// <summary>
        /// Gets the secret from the provider.
        /// </summary>
        /// <param name="secretName">The name of the secret to retrieve.</param>
        /// <returns>The secret value.</returns>
        public abstract Task<string> GetSecretAsync(string secretName);

        /// <summary>
        /// Gets the credentials asynchronously.
        /// </summary>
        /// <param name="credentialsName"></param>
        /// <returns></returns>
        public abstract Task<ICredentials> GetCredentialsAsync(string credentialsName);
    }

    /// <summary>
    /// Represents a provider that provides the default security mechanism.
    /// </summary>
    public class EnvironmentSecurityProvider : SecurityProvider
    {
        /// <summary>
        /// Gets the provider friendly name.
        /// </summary>
        public override string Name
        {
            get { return "Environment"; }
        }

        /// <summary>
        /// Gets the secret from the provider.
        /// </summary>
        /// <param name="secretName">The name of the secret to retrieve.</param>
        /// <returns>The secret value.</returns>
        public override Task<string> GetSecretAsync(string secretName)
        {
            var name = secretName.ToUpper().Replace("/", "_");
            return Task.FromResult(Environment.GetEnvironmentVariable(name));
        }

        /// <summary>
        /// Gets the credentials asynchronously.
        /// </summary>
        /// <param name="credentialsName"></param>
        /// <returns></returns>
        public override Task<ICredentials> GetCredentialsAsync(string credentialsName)
        {
            throw new NotImplementedException();
        }
    }
}