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
using Emitter.Text;
using Emitter.Text.Json;

namespace Emitter.Security
{
    /// <summary>
    /// Represents a security license for the service.
    /// </summary>
    public sealed class SecurityLicense
    {
        #region Static Members

        /// <summary>
        /// Gets or sets currently used security license.
        /// </summary>
        public static SecurityLicense Current;

        /// <summary>
        /// Decrypts the license with the public key.
        /// </summary>
        /// <param name="data"></param>
        /// <param name="keySize"></param>
        /// <returns></returns>
        public static SecurityLicense LoadAndVerify(string data)
        {
            try
            {
                // We should have the data
                if (String.IsNullOrWhiteSpace(data))
                    throw new SecurityLicenseException("No license was found, please provide a valid license key through the configuration file, an EMITTER_LICENSE environment variable or a valid vault key 'secrets/emitter/license'.");

                // Load the license
                Current = new SecurityLicense(
                    Base64.UrlEncoding.FromBase(data)
                );

                // Expired?
                if (DateTime.UtcNow > Current.Expires && Current.Expires > new DateTime(2000, 1, 1))
                    throw new SecurityLicenseException("The license file provided has expired.");

                // Set the encryption key for the service
                SecurityEncryption.SupplyKey(Current.EncryptionKey);

                // Return the license now, but scramble the encryption key
                Current.EncryptionKey = null;
                return Current;
            }
            catch (SecurityLicenseException ex)
            {
                // We do not return an exception, as the license is invalid anyway
                Service.Logger.Log(LogLevel.Error, ex.Message);
                return null;
            }
            catch (Exception)
            {
                // We do not return an exception, as the license is invalid anyway
                Service.Logger.Log(LogLevel.Error, "Invalid license was provided, please check the configuration file or use EMITTER_LICENSE environment variable to provide a valid license key.");
                return null;
            }
        }

        #endregion Static Members

        #region Constructors

        /// <summary>
        /// Gets the beginning of time for the timestamp.
        /// </summary>
        private static readonly DateTime ExpireBegin = new DateTime(2010, 1, 1, 0, 0, 0);

        /// <summary>
        /// Gets the underlying buffer of the key.
        /// </summary>
        private readonly byte[] Buffer;

        /// <summary>
        /// Gets the offset where the api key starts in this buffer.
        /// </summary>
        private readonly int Offset;

        /// <summary>
        /// Constructs an empty license.
        /// </summary>
        public SecurityLicense()
        {
            var size = 32;
            this.Buffer = new byte[size];
            this.Offset = 0;
        }

        /// <summary>
        /// Constructor that allows loading a license from its raw bytes.
        /// </summary>
        /// <param name="buffer">The buffer to load.</param>
        private SecurityLicense(byte[] buffer)
        {
            var size = 32;
            if (buffer == null || buffer.Length != size)
                throw new ArgumentException("Buffer length should be exactly " + size + " bytes.");
            this.Buffer = buffer;
            this.Offset = 0;
        }

        #endregion Constructors

        #region Properties

        /// <summary>
        /// Gets or sets the encryption key.
        /// </summary>
        [JsonProperty("key")]
        public string EncryptionKey
        {
            get
            {
                return Base64.UrlEncoding.ToBase(
                    new ArraySegment<byte>(this.Buffer, 0, 16).ToArray()
                    );
            }
            set
            {
                if (value == null)
                {
                    for (int i = 0; i < 16; ++i)
                        this.Buffer[this.Offset + i] = 42;
                    return;
                }

                if (value.Length != 22)
                    throw new ArgumentException("Key provided is invalid.");

                var data = Base64.UrlEncoding.FromBase(value);
                if (data.Length != 16)
                    throw new ArgumentException("Key provided is invalid.");

                // Write the key to the buffer
                Array.Copy(data, this.Buffer, 16);
            }
        }

        /// <summary>
        /// Gets or sets the contract id.
        /// </summary>
        [JsonProperty("contract")]
        public int Contract
        {
            get
            {
                // Read the value
                return (int)
                       (this.Buffer[this.Offset + 16] << 24
                     | (this.Buffer[this.Offset + 17] << 16)
                     | (this.Buffer[this.Offset + 18] << 8)
                     | (this.Buffer[this.Offset + 19]));
            }
            set
            {
                // Write the value
                this.Buffer[this.Offset + 16] = (byte)(value >> 24);
                this.Buffer[this.Offset + 17] = (byte)(value >> 16);
                this.Buffer[this.Offset + 18] = (byte)(value >> 8);
                this.Buffer[this.Offset + 19] = (byte)value;
            }
        }

        /// <summary>
        /// Gets or sets the signature of the contract, servers as
        /// a failsafe validation. To validate an api key both contract id
        /// and the signature should be found in the database.
        /// </summary>
        [JsonProperty("signature")]
        public int Signature
        {
            get
            {
                // Read the value
                return (int)
                       (this.Buffer[this.Offset + 20] << 24
                     | (this.Buffer[this.Offset + 21] << 16)
                     | (this.Buffer[this.Offset + 22] << 8)
                     | (this.Buffer[this.Offset + 23]));
            }
            set
            {
                // Write the value
                this.Buffer[this.Offset + 20] = (byte)(value >> 24);
                this.Buffer[this.Offset + 21] = (byte)(value >> 16);
                this.Buffer[this.Offset + 22] = (byte)(value >> 8);
                this.Buffer[this.Offset + 23] = (byte)value;
            }
        }

        /// <summary>
        /// Gets or sets the expiration date for the license.
        /// </summary>
        [JsonProperty("expires")]
        public DateTime Expires
        {
            get
            {
                // Read the expiration value
                var expire = (uint)
                             (this.Buffer[this.Offset + 24] << 24
                           | (this.Buffer[this.Offset + 25] << 16)
                           | (this.Buffer[this.Offset + 26] << 8)
                           | (this.Buffer[this.Offset + 27]));

                // Return the expiration date time
                return expire == 0
                    ? DateTime.MinValue
                    : ExpireBegin.AddSeconds(expire);
            }
            set
            {
                if (value < ExpireBegin && value != DateTime.MinValue)
                    throw new ArgumentException("The expiration date should be after the year 2010.");

                // Calculate the expiration date, in seconds
                var expire = value == DateTime.MinValue
                    ? (uint)0
                    : (uint)((value - ExpireBegin).TotalSeconds);

                // Write the value
                this.Buffer[this.Offset + 24] = (byte)(expire >> 24);
                this.Buffer[this.Offset + 25] = (byte)(expire >> 16);
                this.Buffer[this.Offset + 26] = (byte)(expire >> 8);
                this.Buffer[this.Offset + 27] = (byte)expire;
            }
        }

        /// <summary>
        /// Gets or sets the license type.
        /// </summary>
        [JsonProperty("type")]
        public SecurityLicenseType Type
        {
            get
            {
                // Read the value
                return (SecurityLicenseType)(uint)
                       (this.Buffer[this.Offset + 28] << 24
                     | (this.Buffer[this.Offset + 29] << 16)
                     | (this.Buffer[this.Offset + 30] << 8)
                     | (this.Buffer[this.Offset + 31]));
            }
            set
            {
                // Write the value
                var val = (uint)value;
                this.Buffer[this.Offset + 28] = (byte)(val >> 24);
                this.Buffer[this.Offset + 29] = (byte)(val >> 16);
                this.Buffer[this.Offset + 30] = (byte)(val >> 8);
                this.Buffer[this.Offset + 31] = (byte)val;
            }
        }

        #endregion Properties

        #region Public Members

        /// <summary>
        /// Converts the license to string representation.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            return JsonConvert.SerializeObject(this, Formatting.Indented);
        }

        /// <summary>
        /// Encrypts the license information with the private key.
        /// </summary>
        /// <param name="keySize"></param>
        public string Sign()
        {
            return Base64.UrlEncoding.ToBase(this.Buffer);
        }

        #endregion Public Members
    }

    /// <summary>
    /// Represents the license type.
    /// </summary>
    public enum SecurityLicenseType : uint
    {
        Unknown = 0,
        Cloud = 1,
        OnPremise = 2
    }

    /// <summary>
    /// Represents a security license exception.
    /// </summary>
    public class SecurityLicenseException : Exception
    {
        public SecurityLicenseException(string message) : base(message)
        {
        }
    }
}