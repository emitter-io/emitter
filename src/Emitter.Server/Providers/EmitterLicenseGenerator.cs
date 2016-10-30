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