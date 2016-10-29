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

namespace Emitter.Security
{
    /// <summary>
    /// Represents a replacement for .NET's Random which uses RNGCryptoProvider instead.
    /// </summary>
    public class SecurityRandom
    {
        private const int BufferSize = 1024;  // must be a multiple of 4
        private byte[] Buffer;
        private int BufferOffset;
        private RandomNumberGenerator Rand;

        /// <summary>
        /// Constructs a new random.
        /// </summary>
        public SecurityRandom()
        {
            Buffer = new byte[BufferSize];
            Rand = RandomNumberGenerator.Create();
            BufferOffset = Buffer.Length;
        }

        [ThreadStatic]
        private static SecurityRandom DefaultGenerator;

        /// <summary>
        /// Gets the default random generator.
        /// </summary>
        public static SecurityRandom Default
        {
            get
            {
                if (DefaultGenerator == null)
                    DefaultGenerator = new SecurityRandom();
                return DefaultGenerator;
            }
        }

        private void FillBuffer()
        {
            Rand.GetBytes(Buffer);
            BufferOffset = 0;
        }

        /// <summary>
        /// Generates a nonnegative random integer.
        /// </summary>
        /// <returns>The generated number.</returns>
        public int Next()
        {
            if (BufferOffset >= Buffer.Length)
            {
                FillBuffer();
            }
            int val = BitConverter.ToInt32(Buffer, BufferOffset) & 0x7fffffff;
            BufferOffset += sizeof(int);
            return val;
        }

        /// <summary>
        /// Generates a nonnegative random integer.
        /// </summary>
        /// <param name="maxValue">The maximum value.</param>
        /// <returns>The generated number.</returns>
        public int Next(int maxValue)
        {
            return Next() % maxValue;
        }

        /// <summary>
        /// Generates a nonnegative random integer.
        /// </summary>
        /// <param name="minValue">The minimum value.</param>
        /// <param name="maxValue">The maximum value.</param>
        /// <returns>The generated number.</returns>
        public int Next(int minValue, int maxValue)
        {
            if (maxValue < minValue)
            {
                throw new ArgumentOutOfRangeException("maxValue must be greater than or equal to minValue");
            }
            int range = maxValue - minValue;
            return minValue + Next(range);
        }

        /// <summary>
        /// Generates a random double.
        /// </summary>
        /// <returns>The generated number.</returns>
        public double NextDouble()
        {
            int val = Next();
            return (double)val / int.MaxValue;
        }

        /// <summary>
        /// Generates random bytes.
        /// </summary>
        /// <param name="buffer">The buffer to fill.</param>
        public void GetBytes(byte[] buffer)
        {
            Rand.GetBytes(buffer);
        }
    }
}