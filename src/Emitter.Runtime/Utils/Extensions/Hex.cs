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

namespace System.Linq
{
    public static class HexExtensions
    {
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
    }
}