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
using System.Security.Cryptography;
using Emitter.Security;
using Emitter.Text;

namespace Emitter
{
    /// <summary>
    /// Represents a utility class that can be used to generate licenses.
    /// </summary>
    public class EmitterLicenseGenerator
    {
        /// <summary>
        /// Gets the default license generator.
        /// </summary>
        public static readonly EmitterLicenseGenerator Default
            = new EmitterLicenseGenerator();

        /// <summary>
        /// Generate a new license.
        /// </summary>
        /// <returns>The security license generated.</returns>
        public SecurityLicense GenerateLicense()
        {
            var license = new SecurityLicense();
            using (var random = RandomNumberGenerator.Create())
            {
                var encryptionKey = new byte[16];
                random.GetBytes(encryptionKey);

                var contract = new byte[4];
                random.GetBytes(contract);

                var signature = new byte[4];
                random.GetBytes(contract);

                // Generate contents
                license.Type = SecurityLicenseType.OnPremise;
                license.EncryptionKey = Base64.UrlEncoding.ToBase(encryptionKey);
                license.Contract = BitConverter.ToInt32(contract, 0);
                license.Signature = BitConverter.ToInt32(signature, 0);
                return license;
            }
        }

        /// <summary>
        /// Generate a new secret key for the license.
        /// </summary>
        /// <param name="license">The license to use.</param>
        /// <returns>The secret key that can be used for channel key generation.</returns>
        public SecurityKey GenerateSecretKey(SecurityLicense license)
        {
            SecurityLicense.LoadAndVerify(license.Sign());
            var secretKey = SecurityKey.Create();
            secretKey.Master = (ushort)1;                   // Also store the reference to itself
            secretKey.Contract = license.Contract;          // Store the contract id
            secretKey.Signature = license.Signature;        // The signature of the contract
            secretKey.Permissions = SecurityAccess.Master;  // Permission of 1 means it's a master key
            secretKey.Target = 0;                           // Master key does not have a target
            return secretKey;
        }
    }
}