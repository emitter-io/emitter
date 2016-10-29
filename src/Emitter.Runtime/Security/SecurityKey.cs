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
using Emitter.Text;
using Emitter.Text.Json;

namespace Emitter.Security
{
    /// <summary>
    /// Represents an api key information.
    /// </summary>
    public struct SecurityKey
    {
        #region Constructors

        /// <summary>
        /// Random number generator (non-cryptographic) for key salting. This is used solely for illusion
        /// of randomness, as the data is encrypted via XTEA algorithm afterwards anyway.
        /// </summary>
        private static readonly Random Random = new Random();

        /// <summary>
        /// Gets the beginning of time for the timestamp.
        /// </summary>
        private static readonly DateTime ExpireBegin = new DateTime(2010, 1, 1, 0, 0, 0);

        /// <summary>
        /// Gets an empty key for comparisons.
        /// </summary>
        public static readonly SecurityKey Empty = default(SecurityKey);

        /// <summary>
        /// Constructor that allows loading a key from its raw bytes.
        /// </summary>
        /// <param name="buffer">The buffer to load.</param>
        private SecurityKey(byte[] buffer)
        {
            if (buffer == null || buffer.Length != 24)
                throw new ArgumentException("Buffer length should be exactly 24 bytes.");
            this.Buffer = buffer;
            this.Offset = 0;
        }

        /// <summary>
        /// Constructor that allows loading a key from its raw bytes.
        /// </summary>
        /// <param name="buffer">The buffer to load.</param>
        /// <param name="offset">The offset in this buffer.</param>
        private SecurityKey(byte[] buffer, int offset)
        {
            if (buffer == null || offset + 24 > buffer.Length)
                throw new ArgumentException("Buffer length should be at least 24 bytes.");
            this.Buffer = buffer;
            this.Offset = offset;
        }

        /// <summary>
        /// Gets the underlying buffer of the key.
        /// </summary>
        private readonly byte[] Buffer;

        /// <summary>
        /// Gets the offset where the api key starts in this buffer.
        /// </summary>
        private readonly int Offset;

        #endregion Constructors

        #region Public Properties

        /// <summary>
        /// Gets whether the key is empty or not.
        /// </summary>
        public bool IsEmpty
        {
            get { return this.Buffer == null; }
        }

        /// <summary>
        /// Gets or sets a random salt of the key.
        /// </summary>
        private ushort Salt
        {
            get
            {
                // Read the value
                return (ushort)((this.Buffer[this.Offset] << 8) | this.Buffer[this.Offset + 1]);
            }
            set
            {
                // Write the value
                this.Buffer[this.Offset] = (byte)(value >> 8);
                this.Buffer[this.Offset + 1] = (byte)value;
            }
        }

        /// <summary>
        /// Gets or sets a master key id.
        /// </summary>
        public ushort Master
        {
            get
            {
                // Read the value
                return (ushort)
                       (this.Buffer[this.Offset + 2] << 8
                     | (this.Buffer[this.Offset + 3]));
            }
            set
            {
                // Write the value
                this.Buffer[this.Offset + 2] = (byte)(value >> 8);
                this.Buffer[this.Offset + 3] = (byte)value;
            }
        }

        /// <summary>
        /// Gets or sets the contract id.
        /// </summary>
        public int Contract
        {
            get
            {
                // Read the value
                return (int)
                       (this.Buffer[this.Offset + 4] << 24
                     | (this.Buffer[this.Offset + 5] << 16)
                     | (this.Buffer[this.Offset + 6] << 8)
                     | (this.Buffer[this.Offset + 7]));
            }
            set
            {
                // Write the value
                this.Buffer[this.Offset + 4] = (byte)(value >> 24);
                this.Buffer[this.Offset + 5] = (byte)(value >> 16);
                this.Buffer[this.Offset + 6] = (byte)(value >> 8);
                this.Buffer[this.Offset + 7] = (byte)value;
            }
        }

        /// <summary>
        /// Gets or sets the signature of the contract, servers as
        /// a failsafe validation. To validate an api key both contract id
        /// and the signature should be found in the database.
        /// </summary>
        public int Signature
        {
            get
            {
                // Read the value
                return (int)
                       (this.Buffer[this.Offset + 8] << 24
                     | (this.Buffer[this.Offset + 9] << 16)
                     | (this.Buffer[this.Offset + 10] << 8)
                     | (this.Buffer[this.Offset + 11]));
            }
            set
            {
                // Write the value
                this.Buffer[this.Offset + 8] = (byte)(value >> 24);
                this.Buffer[this.Offset + 9] = (byte)(value >> 16);
                this.Buffer[this.Offset + 10] = (byte)(value >> 8);
                this.Buffer[this.Offset + 11] = (byte)value;
            }
        }

        /// <summary>
        /// Gets or sets the permission flags.
        /// </summary>
        public SecurityAccess Permissions
        {
            get
            {
                // Read the value
                return (SecurityAccess)(uint)
                       (this.Buffer[this.Offset + 12] << 24
                     | (this.Buffer[this.Offset + 13] << 16)
                     | (this.Buffer[this.Offset + 14] << 8)
                     | (this.Buffer[this.Offset + 15]));
            }
            set
            {
                // Write the value
                var val = (uint)value;
                this.Buffer[this.Offset + 12] = (byte)(val >> 24);
                this.Buffer[this.Offset + 13] = (byte)(val >> 16);
                this.Buffer[this.Offset + 14] = (byte)(val >> 8);
                this.Buffer[this.Offset + 15] = (byte)val;
            }
        }

        /// <summary>
        /// Gets or sets the target for the key.
        /// </summary>
        public uint Target
        {
            get
            {
                // Read the value
                return (uint)
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
        /// Gets or sets the expiration date for the key.
        /// </summary>
        public DateTime Expires
        {
            get
            {
                // Read the expiration value
                var expire = (uint)
                             (this.Buffer[this.Offset + 20] << 24
                           | (this.Buffer[this.Offset + 21] << 16)
                           | (this.Buffer[this.Offset + 22] << 8)
                           | (this.Buffer[this.Offset + 23]));

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
                this.Buffer[this.Offset + 20] = (byte)(expire >> 24);
                this.Buffer[this.Offset + 21] = (byte)(expire >> 16);
                this.Buffer[this.Offset + 22] = (byte)(expire >> 8);
                this.Buffer[this.Offset + 23] = (byte)expire;
            }
        }

        #endregion Public Properties

        #region Public Members

        /// <summary>
        /// Gets the encrypted value of the key.
        /// </summary>
        public string Value
        {
            get
            {
                // Then encrypt the key using the master key
                return Base64.UrlEncoding.ToBase(
                    this.Encrypt()
                    );
            }
        }

        /// <summary>
        /// Gets whether the key has expired or not.
        /// </summary>
        public bool IsExpired
        {
            get
            {
                if (this.Expires == DateTime.MinValue)
                    return false;
                return this.Expires < DateTime.UtcNow;
            }
        }

        /// <summary>
        /// Gets whether the key is a master key.
        /// </summary>
        public bool IsMaster
        {
            get { return this.Permissions == SecurityAccess.Master; }
        }

        /// <summary>
        /// Convert the key to a readable string representation for debugging purposes.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            return JsonConvert.SerializeObject(this, Formatting.Indented);
        }

        #endregion Public Members

        #region Encrypt/Decrypt Members

        /// <summary>
        /// Encrypts the key and returns the encrypted buffer.
        /// </summary>
        /// <returns>Encrypted buffer.</returns>
        private byte[] Encrypt()
        {
            // The buffer for encryption
            var buffer = new byte[24];
            buffer[0] = this.Buffer[this.Offset];
            buffer[1] = this.Buffer[this.Offset + 1];

            // First XOR the entire array with the salt
            for (int i = 2; i < 24; i += 2)
            {
                buffer[i] = (byte)(this.Buffer[this.Offset + i] ^ buffer[0]);
                buffer[i + 1] = (byte)(this.Buffer[this.Offset + i + 1] ^ buffer[1]);
            }

            // Then encrypt the key using the master key
            return SecurityEncryption.Encrypt(buffer);
        }

        /// <summary>
        /// Decrypts the key and returns the decrypted buffer.
        /// </summary>
        /// <param name="buffer">The key to decrypt.</param>
        /// <param name="offset">The offset to start the decryption at.</param>
        /// <returns></returns>
        private static void Decrypt(byte[] buffer, int offset)
        {
            // First decrypt the key
            SecurityEncryption.Decrypt(buffer, offset);

            // Then XOR the entire array with the salt
            for (int i = (offset + 2); i < (offset + 24); i += 2)
            {
                buffer[i] = (byte)(buffer[i] ^ buffer[offset]);
                buffer[i + 1] = (byte)(buffer[i + 1] ^ buffer[offset + 1]);
            }
        }

        #endregion Encrypt/Decrypt Members

        #region TryParse and Create

        /// <summary>
        /// Attempts to parse and decrypt the API key.
        /// </summary>
        /// <param name="value">The buffer containing the key.</param>
        /// <param name="key">The key to parse.</param>
        /// <returns>Whether we successfully parsed the key or not.</returns>
        public static bool TryParse(string value, out SecurityKey key)
        {
            try
            {
                // Decrypt the buffer first
                var buffer = Base64.UrlEncoding.FromBase(value);
                SecurityKey.Decrypt(buffer, 0);

                // Attempt to decrypt
                key = new SecurityKey(buffer);
                return true;
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                // Key parsing failure
                key = default(SecurityKey);
                return false;
            }
        }

        /// <summary>
        /// Attempts to parse and decrypt the API key.
        /// </summary>
        /// <param name="value">The buffer containing the key.</param>
        /// <param name="key">The key to parse.</param>
        /// <returns>Whether we successfully parsed the key or not.</returns>
        public static bool TryParse(byte[] value, out SecurityKey key)
        {
            try
            {
                // Decrypt the buffer first
                SecurityKey.Decrypt(value, 0);

                // Attempt to decrypt
                key = new SecurityKey(value);
                return true;
            }
            catch
            {
                // Key parsing failure
                key = default(SecurityKey);
                return false;
            }
        }

        /// <summary>
        /// Attempts to parse and decrypt the API key.
        /// </summary>
        /// <param name="value">The buffer containing the key.</param>
        /// <param name="key">The key to parse.</param>
        /// <returns>Whether we successfully parsed the key or not.</returns>
        public static bool TryParse(ArraySegment<byte> value, out SecurityKey key)
        {
            try
            {
                // Decrypt the buffer first
                SecurityKey.Decrypt(value.Array, value.Offset);

                // Attempt to decrypt
                key = new SecurityKey(value.Array, value.Offset);
                return true;
            }
            catch
            {
                // Key parsing failure
                key = default(SecurityKey);
                return false;
            }
        }

        /// <summary>
        /// Creates a new <see cref="SecurityKey"/> without any parameters.
        /// </summary>
        public static SecurityKey Create()
        {
            // Create the internal buffer and salt
            var key = new SecurityKey(
                 new byte[24]
                 );

            // Generate a random salt
            key.Salt = (ushort)Random.Next(ushort.MaxValue);
            return key;
        }

        #endregion TryParse and Create
    }

    /// <summary>
    /// Represents the access of the key.
    /// </summary>
    [Flags]
    public enum SecurityAccess : uint
    {
        /// <summary>
        /// Key has no privileges.
        /// </summary>
        None = 0,

        /// <summary>
        /// Key should be allowed to generate other keys.
        /// </summary>
        Master = 1 << 0,

        /// <summary>
        /// Key should be allowed to subscribe to the target channel.
        /// </summary>
        Read = 1 << 1,

        /// <summary>
        /// Key should be allowed to publish to the target channel.
        /// </summary>
        Write = 1 << 2,

        /// <summary>
        /// Key should be allowed to write to the message history of the target channel.
        /// </summary>
        Store = 1 << 3,

        /// <summary>
        /// Key should be allowed to write to read the message history of the target channel.
        /// </summary>
        Load = 1 << 4,

        /// <summary>
        /// Key should be allowed to query the presence on the target channel.
        /// </summary>
        Presence = 1 << 5,

        /// <summary>
        /// Key should be allowed to read and write to the target channel.
        /// </summary>
        ReadWrite = Read | Write,

        /// <summary>
        /// Key should be allowed to read and write the message history.
        /// </summary>
        StoreLoad = Store | Load
    }
}