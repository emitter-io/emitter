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
using System.Collections;
using System.Runtime.CompilerServices;

namespace Emitter.Collections
{
    /// <summary>
    /// Represents a bloom filter.
    /// </summary>
    /// <typeparam name="T">Item type </typeparam>
    internal sealed class Filter<T> : RecyclableObject
    {
        private readonly int HashFunctionCount;
        private readonly BitArray HashBits;
        private readonly HashFunction GetHashSecondary;

        private readonly static ConcurrentPool<Filter<T>> Pool
            = new ConcurrentPool<Filter<T>>("Bloom Filters", (c) => new Filter<T>(2048), Environment.ProcessorCount);

        /// <summary>
        /// Acquires a new bloom filter instance.
        /// </summary>
        /// <returns></returns>
        public static Filter<T> Acquire()
        {
            return Pool.Acquire();
        }

        /// <summary>
        /// Creates a new Bloom filter, specifying an error rate of 1/capacity, using the optimal size for the underlying data structure based on the desired capacity and error rate, as well as the optimal number of hash functions.
        /// A secondary hash function will be provided for you if your type T is either string or int. Otherwise an exception will be thrown. If you are not using these types please use the overload that supports custom hash functions.
        /// </summary>
        /// <param name="capacity">The anticipated number of items to be added to the filter. More than this number of items can be added, but the error rate will exceed what is expected.</param>
        public Filter(int capacity)
            : this(capacity, null)
        {
        }

        /// <summary>
        /// Creates a new Bloom filter, using the optimal size for the underlying data structure based on the desired capacity and error rate, as well as the optimal number of hash functions.
        /// A secondary hash function will be provided for you if your type T is either string or int. Otherwise an exception will be thrown. If you are not using these types please use the overload that supports custom hash functions.
        /// </summary>
        /// <param name="capacity">The anticipated number of items to be added to the filter. More than this number of items can be added, but the error rate will exceed what is expected.</param>
        /// <param name="errorRate">The accepable false-positive rate (e.g., 0.01F = 1%)</param>
        public Filter(int capacity, float errorRate)
            : this(capacity, errorRate, null)
        {
        }

        /// <summary>
        /// Creates a new Bloom filter, specifying an error rate of 1/capacity, using the optimal size for the underlying data structure based on the desired capacity and error rate, as well as the optimal number of hash functions.
        /// </summary>
        /// <param name="capacity">The anticipated number of items to be added to the filter. More than this number of items can be added, but the error rate will exceed what is expected.</param>
        /// <param name="hashFunction">The function to hash the input values. Do not use GetHashCode(). If it is null, and T is string or int a hash function will be provided for you.</param>
        public Filter(int capacity, HashFunction hashFunction)
            : this(capacity, BestErrorRate(capacity), hashFunction)
        {
        }

        /// <summary>
        /// Creates a new Bloom filter, using the optimal size for the underlying data structure based on the desired capacity and error rate, as well as the optimal number of hash functions.
        /// </summary>
        /// <param name="capacity">The anticipated number of items to be added to the filter. More than this number of items can be added, but the error rate will exceed what is expected.</param>
        /// <param name="errorRate">The accepable false-positive rate (e.g., 0.01F = 1%)</param>
        /// <param name="hashFunction">The function to hash the input values. Do not use GetHashCode(). If it is null, and T is string or int a hash function will be provided for you.</param>
        public Filter(int capacity, float errorRate, HashFunction hashFunction)
            : this(capacity, errorRate, hashFunction, BestM(capacity, errorRate), BestK(capacity, errorRate))
        {
        }

        /// <summary>
        /// Creates a new Bloom filter.
        /// </summary>
        /// <param name="capacity">The anticipated number of items to be added to the filter. More than this number of items can be added, but the error rate will exceed what is expected.</param>
        /// <param name="errorRate">The accepable false-positive rate (e.g., 0.01F = 1%)</param>
        /// <param name="hashFunction">The function to hash the input values. Do not use GetHashCode(). If it is null, and T is string or int a hash function will be provided for you.</param>
        /// <param name="m">The number of elements in the BitArray.</param>
        /// <param name="k">The number of hash functions to use.</param>
        public Filter(int capacity, float errorRate, HashFunction hashFunction, int m, int k)
        {
            // validate the params are in range
            if (capacity < 1)
            {
                throw new ArgumentOutOfRangeException("capacity", capacity, "capacity must be > 0");
            }

            if (errorRate >= 1 || errorRate <= 0)
            {
                throw new ArgumentOutOfRangeException("errorRate", errorRate, string.Format("errorRate must be between 0 and 1, exclusive. Was {0}", errorRate));
            }

            // from overflow in bestM calculation
            if (m < 1)
            {
                throw new ArgumentOutOfRangeException(string.Format("The provided capacity and errorRate values would result in an array of length > int.MaxValue. Please reduce either of these values. Capacity: {0}, Error rate: {1}", capacity, errorRate));
            }

            // set the secondary hash function
            if (hashFunction == null)
            {
                if (typeof(T) == typeof(string))
                {
                    this.GetHashSecondary = HashString;
                }
                else if (typeof(T) == typeof(int))
                {
                    this.GetHashSecondary = HashInt32;
                }
                else if (typeof(T) == typeof(MessageQueue))
                {
                    this.GetHashSecondary = HashMessageQueue;
                }
                else
                {
                    throw new ArgumentNullException("hashFunction", "Please provide a hash function for your type T, when T is not a string or int.");
                }
            }
            else
            {
                this.GetHashSecondary = hashFunction;
            }

            this.HashFunctionCount = k;
            this.HashBits = new BitArray(m);
        }

        /// <summary>
        /// Recycles the filter.
        /// </summary>
        public override void Recycle()
        {
            this.Clear();
        }

        /// <summary>
        /// A function that can be used to hash input.
        /// </summary>
        /// <param name="input">The values to be hashed.</param>
        /// <returns>The resulting hash code.</returns>
        public delegate int HashFunction(T input);

        /// <summary>
        /// The ratio of false to true bits in the filter. E.g., 1 true bit in a 10 bit filter means a truthiness of 0.1.
        /// </summary>
        public double Truthiness
        {
            get
            {
                return (double)this.TrueBits() / this.HashBits.Length;
            }
        }

        /// <summary>
        /// Adds a new item to the filter. It cannot be removed.
        /// </summary>
        /// <param name="item">The item.</param>
        public void Add(T item)
        {
            // start flipping bits for each hash of item
            int primaryHash = item.GetHashCode();
            int secondaryHash = this.GetHashSecondary(item);
            for (int i = 0; i < this.HashFunctionCount; i++)
            {
                int hash = this.ComputeHash(primaryHash, secondaryHash, i);
                this.HashBits[hash] = true;
            }
        }

        /// <summary>
        /// Checks for the existance of the item in the filter for a given probability.
        /// </summary>
        /// <param name="item"> The item. </param>
        /// <returns> The <see cref="bool"/>. </returns>
        public bool Contains(T item)
        {
            int primaryHash = item.GetHashCode();
            int secondaryHash = this.GetHashSecondary(item);
            for (int i = 0; i < this.HashFunctionCount; i++)
            {
                int hash = this.ComputeHash(primaryHash, secondaryHash, i);
                if (this.HashBits[hash] == false)
                {
                    return false;
                }
            }

            return true;
        }

        /// <summary>
        /// Clears the bloom filter by resetting the bits to zero.
        /// </summary>
        public void Clear()
        {
            this.HashBits.SetAll(false);
        }

        /// <summary>
        /// The best k.
        /// </summary>
        /// <param name="capacity"> The capacity. </param>
        /// <param name="errorRate"> The error rate. </param>
        /// <returns> The <see cref="int"/>. </returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static int BestK(int capacity, float errorRate)
        {
            return (int)Math.Round(Math.Log(2.0) * BestM(capacity, errorRate) / capacity);
        }

        /// <summary>
        /// The best m.
        /// </summary>
        /// <param name="capacity"> The capacity. </param>
        /// <param name="errorRate"> The error rate. </param>
        /// <returns> The <see cref="int"/>. </returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static int BestM(int capacity, float errorRate)
        {
            return (int)Math.Ceiling(capacity * Math.Log(errorRate, (1.0 / Math.Pow(2, Math.Log(2.0)))));
        }

        /// <summary>
        /// The best error rate.
        /// </summary>
        /// <param name="capacity"> The capacity. </param>
        /// <returns> The <see cref="float"/>. </returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static float BestErrorRate(int capacity)
        {
            float c = (float)(1.0 / capacity);
            if (c != 0)
            {
                return c;
            }

            // default
            // http://www.cs.princeton.edu/courses/archive/spring02/cs493/lec7.pdf
            return (float)Math.Pow(0.6185, int.MaxValue / capacity);
        }

        /// <summary>
        /// Gets the message queue hash code.
        /// </summary>
        /// <param name="input"></param>
        /// <returns></returns>
        internal static int HashMessageQueue(T input)
        {
            return (input as MessageQueue).SecondaryHash;
        }

        /// <summary>
        /// Hashes a 32-bit signed int using Thomas Wang's method v3.1 (http://www.concentric.net/~Ttwang/tech/inthash.htm).
        /// Runtime is suggested to be 11 cycles.
        /// </summary>
        /// <param name="input">The integer to hash.</param>
        /// <returns>The hashed result.</returns>
        internal static int HashInt32(T input)
        {
            uint? x = input as uint?;
            unchecked
            {
                x = ~x + (x << 15); // x = (x << 15) - x- 1, as (~x) + y is equivalent to y - x - 1 in two's complement representation
                x = x ^ (x >> 12);
                x = x + (x << 2);
                x = x ^ (x >> 4);
                x = x * 2057; // x = (x + (x << 3)) + (x<< 11);
                x = x ^ (x >> 16);
                return (int)x;
            }
        }

        /// <summary>
        /// Hashes a string using Bob Jenkin's "One At A Time" method from Dr. Dobbs (http://burtleburtle.net/bob/hash/doobs.html).
        /// Runtime is suggested to be 9x+9, where x = input.Length.
        /// </summary>
        /// <param name="input">The string to hash.</param>
        /// <returns>The hashed result.</returns>
        internal static int HashString(T input)
        {
            string s = input as string;
            int hash = 0;

            for (int i = 0; i < s.Length; i++)
            {
                hash += s[i];
                hash += (hash << 10);
                hash ^= (hash >> 6);
            }

            hash += (hash << 3);
            hash ^= (hash >> 11);
            hash += (hash << 15);
            return hash;
        }

        /// <summary>
        /// Count the number of true bits.
        /// </summary>
        /// <returns> The <see cref="int"/>. </returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private int TrueBits()
        {
            int output = 0;
            for (int i = 0; i < this.HashBits.Length; ++i)
            {
                if (this.HashBits[i])
                    ++output;
            }
            return output;
        }

        /// <summary>
        /// Performs Dillinger and Manolios double hashing.
        /// </summary>
        /// <param name="primaryHash"> The primary hash. </param>
        /// <param name="secondaryHash"> The secondary hash. </param>
        /// <param name="i"> The i. </param>
        /// <returns> The <see cref="int"/>. </returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private int ComputeHash(int primaryHash, int secondaryHash, int i)
        {
            return Math.Abs(
                (primaryHash + (i * secondaryHash)) % this.HashBits.Length
                );
        }
    }
}