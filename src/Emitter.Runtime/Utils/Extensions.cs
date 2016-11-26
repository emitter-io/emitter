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
using System.IO;
using System.Linq;
using System.Net;
using System.Reflection;
using System.Security.Cryptography;
using System.Text;
using System.Threading.Tasks;
using Emitter.Network;

namespace Emitter
{
    /// <summary>
    /// Cryptography extensions for various classes.
    /// </summary>
    public static class Extensions
    {
        #region Private Members

        /// <summary>
        /// Unix epoch offset.
        /// </summary>
        private static readonly DateTime UnixOffset = new DateTime(1970, 1, 1, 0, 0, 0, 0, DateTimeKind.Utc);

        #endregion Private Members

        #region Threading

        /// <summary>
        /// Iterates through each key/value pair in the dictionary and invokes an action on each one.
        /// </summary>
        /// <typeparam name="TKey"></typeparam>
        /// <typeparam name="TValue"></typeparam>
        /// <param name="dictionary"></param>
        /// <param name="action"></param>
        internal static void ForEach<TKey, TValue>(this IDictionary<TKey, TValue> dictionary, Action<TKey, TValue> action)
        {
            if (dictionary == null)
                throw new ArgumentNullException("dictionary");
            if (action == null)
                throw new ArgumentNullException("action");

            foreach (TKey key in dictionary.Keys)
            {
                action(key, dictionary[key]);
            }
        }

        /// <summary>
        /// This function specifies that the task should be forgotten as we do
        /// not care for its result.
        /// </summary>
        /// <param name="task">The task to forget.</param>
        public static void Forget(this Task task)
        {
        }

        #endregion Threading

        #region Encoding

        /// <summary>
        /// Converts a date time to a unix timestamp.
        /// </summary>
        /// <param name="dateTime">Date time (should be UTC).</param>
        /// <returns>The UNIX timestamp in seconds</returns>
        public static long ToUnixTimestamp(this DateTime dateTime)
        {
            return (long)(dateTime - UnixOffset).TotalSeconds;
        }

        /// <summary>
        /// Generates a hash for the given plain text value and returns hash bytes.
        /// </summary>
        /// <param name="source">Source value to be hashed. The function does not check whether this parameter is null.</param>
        /// <returns>Hash value in bytes.</returns>
        public static byte[] GetSHA1Bytes(this string source)
        {
            byte[] plainTextBytes = Encoding.UTF8.GetBytes(source);
            using (var hash = SHA1.Create())
            {
                return hash.ComputeHash(plainTextBytes);
            }
        }

        /// <summary>
        /// Converts the integer into the big endian byte array.
        /// </summary>
        /// <typeparam name="T"></typeparam>
        /// <param name="source"></param>
        /// <returns></returns>
        public static byte[] ToBigEndianBytes<T>(this int source)
        {
            byte[] bytes;

            var type = typeof(T);
            if (type == typeof(ushort))
                bytes = BitConverter.GetBytes((ushort)source);
            else if (type == typeof(ulong))
                bytes = BitConverter.GetBytes((ulong)source);
            else if (type == typeof(int))
                bytes = BitConverter.GetBytes(source);
            else
                throw new InvalidCastException("Cannot be cast to T");

            if (BitConverter.IsLittleEndian)
                Array.Reverse(bytes);
            return bytes;
        }

        /// <summary>
        /// Converts bytes to a hexadecimal string representation.
        /// </summary>
        /// <param name="bytes"></param>
        /// <returns></returns>
        private static unsafe string ToHex(this byte[] bytes)
        {
            var c = stackalloc char[bytes.Length * 2 + 1];
            int b;
            for (int i = 0; i < bytes.Length; ++i)
            {
                b = bytes[i] >> 4;
                c[i * 2] = (char)(55 + b + (((b - 10) >> 31) & -7));
                b = bytes[i] & 0xF;
                c[i * 2 + 1] = (char)(55 + b + (((b - 10) >> 31) & -7));
            }
            c[bytes.Length * 2] = '\0';
            return new string(c);
        }

        /// <summary>
        /// Converts an number to a hexadecimal representation.
        /// </summary>
        /// <param name="source">The number to convert.</param>
        /// <returns>The hexadecimal representation of the number.</returns>
        public static string ToHex(this int source)
        {
            return ToHex(BitConverter.GetBytes(source));
        }

        /// <summary>
        /// Converts an number to a hexadecimal representation.
        /// </summary>
        /// <param name="source">The number to convert.</param>
        /// <returns>The hexadecimal representation of the number.</returns>
        public static string ToHex(this uint source)
        {
            return ToHex(BitConverter.GetBytes(source));
        }

        /// <summary>
        /// Converts an number to a hexadecimal representation.
        /// </summary>
        /// <param name="source">The number to convert.</param>
        /// <returns>The hexadecimal representation of the number.</returns>
        public static string ToHex(this long source)
        {
            return ToHex(BitConverter.GetBytes(source));
        }

        /// <summary>
        /// Converts an number to a hexadecimal representation.
        /// </summary>
        /// <param name="source">The number to convert.</param>
        /// <returns>The hexadecimal representation of the number.</returns>
        public static string ToHex(this ulong source)
        {
            return ToHex(BitConverter.GetBytes(source));
        }

        #endregion Encoding

        #region IP

        /// <summary>
        /// Gets the node identifier from the endpoint.
        /// </summary>
        /// <param name="ep">The endpoint to convert.</param>
        /// <returns></returns>
        internal static int ToIdentifier(this IPEndPoint ep)
        {
            return Murmur32.GetHash(ep.ToString());
        }

        #endregion IP

        #region Type

        /// <summary>
        /// Gets invokable methods by type
        /// </summary>
        public static MethodInfo[] GetInvokeAt(this Type source, InvokeAtType type)
        {
            var result = new List<MethodInfo>();
            var methods = source.GetMethods(BindingFlags.Static | BindingFlags.Public);

            for (int i = 0; i < methods.Length; ++i)
            {
                var info = methods[i];
                var attr = info.GetCustomAttributes(typeof(InvokeAtAttribute), false)
                    .FirstOrDefault();
                if (attr == null)
                    continue;

                if ((attr as InvokeAtAttribute).Type != type)
                    continue;

                if (info.ReturnType != typeof(void))
                    throw new TargetInvocationException("A function is marked with InvokeAt attribute, but the return type is not System.Void. Source: " + source.ToString(), null);

                if (info.GetParameters().Length != 0)
                    throw new TargetInvocationException("A function is marked with InvokeAt attribute, but contains parameters. Source: " + source.ToString(), null);

                result.Add(info);
            }

            return result.ToArray();
        }

        /// <summary>
        /// Checks whether a type is a subclass of another type.
        /// </summary>
        /// <param name="source"></param>
        /// <param name="other"></param>
        /// <returns></returns>
        public static bool IsSubclassOf(this Type source, Type other)
        {
            return source.GetTypeInfo().IsSubclassOf(other);
        }

        /// <summary>
        /// Gets the murmur hash code for the type.
        /// </summary>
        /// <param name="source"></param>
        /// <returns></returns>
        public static int ToIdentifier(this Type source)
        {
            return Murmur32.GetHash(source.FullName);
        }

        #endregion Type

        #region Stream
#if DOTNET

        public static byte[] GetBuffer(this MemoryStream stream)
        {
            ArraySegment<byte> buffer;
            if (stream.TryGetBuffer(out buffer))
            {
                return buffer.ToArray();
            }
            return null;
        }

        public static byte[] GetBuffer(this ByteStream stream)
        {
            ArraySegment<byte> buffer;
            if (stream.TryGetBuffer(out buffer))
            {
                return buffer.ToArray();
            }
            return null;
        }

#endif
        #endregion Stream
    }
}