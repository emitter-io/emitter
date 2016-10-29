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
using System.Text;

namespace Emitter.Security
{
    /// <summary>
    /// Provides Murmur3 non-cryptographic hashing.
    /// </summary>
    public static class SecurityHash
    {
        private const UInt32 Seed = 37;

        /// <summary>
        /// Computes MurmurHash3 on this set of bytes and returns the calculated hash value.
        /// </summary>
        /// <param name="data">The data to compute the hash of.</param>
        /// <returns>A 32bit hash value.</returns>
        public static string GetHex(string data)
        {
            // Compute the hash
            return SecurityHash.GetHex(
                Encoding.UTF8.GetBytes(data)
                );
        }

        /// <summary>
        /// Computes MurmurHash3 on this set of bytes and returns the calculated hash value.
        /// </summary>
        /// <param name="data">The data to compute the hash of.</param>
        /// <returns>A 32bit hash value.</returns>
        public static uint GetHash(string data)
        {
            // Compute the hash
            var buffer = SecurityHash.GetBytes(Encoding.UTF8.GetBytes(data));
            return (uint)
                   (buffer[0] << 24
                 | (buffer[1] << 16)
                 | (buffer[2] << 8)
                 | (buffer[3]));
        }

        /// <summary>
        /// Computes MurmurHash3 on this set of bytes and returns the calculated hash value.
        /// </summary>
        /// <param name="data">The data to compute the hash of.</param>
        /// <returns>A 32bit hash value.</returns>
        public static uint GetHash(byte[] data)
        {
            // Compute the hash
            var buffer = SecurityHash.GetBytes(data);
            return (uint)
                   (buffer[0] << 24
                 | (buffer[1] << 16)
                 | (buffer[2] << 8)
                 | (buffer[3]));
        }

        /// <summary>
        /// Computes MurmurHash3 on this set of bytes and returns the calculated hash value.
        /// </summary>
        /// <param name="data">The data to compute the hash of.</param>
        /// <returns>A 32bit hash value.</returns>
        public static string GetHex(byte[] data)
        {
            // Compute the hash
            var bytes = SecurityHash.GetBytes(data);

            // Convert to string
            char[] chars = new char[bytes.Length * 2];
            byte current;
            for (int y = 0, x = 0; y < bytes.Length; ++y, ++x)
            {
                current = ((byte)(bytes[y] >> 4));
                chars[x] = (char)(current > 9 ? current + 0x37 : current + 0x30);
                current = ((byte)(bytes[y] & 0xF));
                chars[++x] = (char)(current > 9 ? current + 0x37 : current + 0x30);
            }

            // Get the hash of the string representation
            return new string(chars);
        }

        /// <summary>
        /// Computes MurmurHash3 on this set of bytes and returns the calculated hash value.
        /// </summary>
        /// <param name="data">The data to compute the hash of.</param>
        /// <returns>A 32bit hash value.</returns>
        public static byte[] GetBytes(byte[] data)
        {
            const UInt32 c1 = 0xcc9e2d51;
            const UInt32 c2 = 0x1b873593;

            int curLength = data.Length; /* Current position in byte array */
            int length = curLength;   /* the const length we need to fix tail */
            UInt32 h1 = Seed;
            UInt32 k1 = 0;

            /* body, eat stream a 32-bit int at a time */
            Int32 currentIndex = 0;
            while (curLength >= 4)
            {
                /* Get four bytes from the input into an UInt32 */
                k1 = (UInt32)(data[currentIndex++]
                  | data[currentIndex++] << 8
                  | data[currentIndex++] << 16
                  | data[currentIndex++] << 24);

                /* bitmagic hash */
                k1 *= c1;
                k1 = Rotl32(k1, 15);
                k1 *= c2;

                h1 ^= k1;
                h1 = Rotl32(h1, 13);
                h1 = h1 * 5 + 0xe6546b64;
                curLength -= 4;
            }

            /* tail, the reminder bytes that did not make it to a full int */
            /* (this switch is slightly more ugly than the C++ implementation
             * because we can't fall through) */
            switch (curLength)
            {
                case 3:
                    k1 = (UInt32)(data[currentIndex++]
                      | data[currentIndex++] << 8
                      | data[currentIndex++] << 16);
                    k1 *= c1;
                    k1 = Rotl32(k1, 15);
                    k1 *= c2;
                    h1 ^= k1;
                    break;

                case 2:
                    k1 = (UInt32)(data[currentIndex++]
                      | data[currentIndex++] << 8);
                    k1 *= c1;
                    k1 = Rotl32(k1, 15);
                    k1 *= c2;
                    h1 ^= k1;
                    break;

                case 1:
                    k1 = (UInt32)(data[currentIndex++]);
                    k1 *= c1;
                    k1 = Rotl32(k1, 15);
                    k1 *= c2;
                    h1 ^= k1;
                    break;
            };

            // finalization, magic chants to wrap it all up
            h1 ^= (UInt32)length;
            h1 = Mix(h1);

            // convert back to 4 bytes
            byte[] key = new byte[4];
            key[0] = (byte)(h1);
            key[1] = (byte)(h1 >> 8);
            key[2] = (byte)(h1 >> 16);
            key[3] = (byte)(h1 >> 24);
            return key;
        }

        private static UInt32 Rotl32(UInt32 x, byte r)
        {
            return (x << r) | (x >> (32 - r));
        }

        private static UInt32 Mix(UInt32 h)
        {
            h ^= h >> 16;
            h *= 0x85ebca6b;
            h ^= h >> 13;
            h *= 0xc2b2ae35;
            h ^= h >> 16;
            return h;
        }
    }
}