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

namespace Emitter.Security
{
    /// <summary>
    /// Represents a security encryption algorithm for the keys.
    /// </summary>
    public class SecurityEncryption
    {
        #region Static Members

        /// <summary>
        /// The cryptographic key used by the encryption algorithm.
        /// </summary>
        private static uint[] Key;

        /// <summary>
        /// Sets the key to use for the encryption algorithm.
        /// </summary>
        /// <param name="key">The key to use.</param>
        public static void SupplyKey(string key)
        {
            if (key.Length != 22)
                throw new ArgumentException("Key provided is invalid.");

            var data = Base64.UrlEncoding.FromBase(key);
            if (data.Length != 16)
                throw new ArgumentException("Key provided is invalid.");

            // Load the key to the appropriate format
            var keyArray = new uint[4];
            for (int i = 0; i < 4; ++i)
            {
                keyArray[i] = (uint)(
                    (data[(4 * i) + 0] << 24) |
                    (data[(4 * i) + 1] << 16) |
                    (data[(4 * i) + 2] << 8) |
                     data[(4 * i) + 3]);
            }

            // Set the key
            Key = keyArray;
        }

        #endregion Static Members

        /// <summary>
		/// Encrypts the data.
		/// </summary>
		/// <param name="data">The key data to encrypt.</param>
		/// <returns>Encrypted data.</returns>
		internal static byte[] Encrypt(byte[] data)
        {
            if (data.Length != 24)
                throw new ArgumentException("The security key should be 24-bytes long.");

            uint r, sum;
            unchecked
            {
                for (int i = 0; i < data.Length; i += 8)
                {
                    uint y = data[i];
                    y <<= 8; y |= data[i + 1];
                    y <<= 8; y |= data[i + 2];
                    y <<= 8; y |= data[i + 3];

                    uint z = data[i + 4];
                    z <<= 8; z |= data[i + 5];
                    z <<= 8; z |= data[i + 6];
                    z <<= 8; z |= data[i + 7];

                    // Encipher the block
                    sum = 0;
                    for (r = 0; r < 32; ++r)
                    {
                        y += (((z << 4) ^ (z >> 5)) + z) ^ (sum + Key[sum & 3]);
                        sum += 0x9E3779B9;
                        z += (((y << 4) ^ (y >> 5)) + y) ^ (sum + Key[(sum >> 11) & 3]);
                    }

                    // Set to the current block
                    data[i] = (byte)(y >> 24);
                    data[i + 1] = (byte)(y >> 16);
                    data[i + 2] = (byte)(y >> 8);
                    data[i + 3] = (byte)(y);
                    data[i + 4] = (byte)(z >> 24);
                    data[i + 5] = (byte)(z >> 16);
                    data[i + 6] = (byte)(z >> 8);
                    data[i + 7] = (byte)(z);
                }

                return data;
            }
        }

        /// <summary>
        /// Decrypts the data.
        /// </summary>
        /// <param name="data">The data to decrypt.</param>
        /// <param name="offset">The starting offset in the data.</param>
        /// <returns>Decrypted buffer.</returns>
        internal static void Decrypt(byte[] data, int offset)
        {
            uint r, sum;
            unchecked
            {
                for (var i = offset; i < (offset + 24); i += 8)
                {
                    uint y = data[i];
                    y <<= 8; y |= data[i + 1];
                    y <<= 8; y |= data[i + 2];
                    y <<= 8; y |= data[i + 3];

                    uint z = data[i + 4];
                    z <<= 8; z |= data[i + 5];
                    z <<= 8; z |= data[i + 6];
                    z <<= 8; z |= data[i + 7];

                    // Decipher the block
                    sum = 0xC6EF3720;
                    for (r = 0; r < 32; ++r)
                    {
                        z -= (((y << 4) ^ (y >> 5)) + y) ^ (sum + Key[(sum >> 11) & 3]);
                        sum -= 0x9E3779B9;
                        y -= (((z << 4) ^ (z >> 5)) + z) ^ (sum + Key[sum & 3]);
                    }

                    // Set to the current block
                    data[i] = (byte)(y >> 24);
                    data[i + 1] = (byte)(y >> 16);
                    data[i + 2] = (byte)(y >> 8);
                    data[i + 3] = (byte)(y);
                    data[i + 4] = (byte)(z >> 24);
                    data[i + 5] = (byte)(z >> 16);
                    data[i + 6] = (byte)(z >> 8);
                    data[i + 7] = (byte)(z);
                }
            }
        }
    }
}