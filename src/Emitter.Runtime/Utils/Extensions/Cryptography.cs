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
using System.Text;

namespace Emitter
{
    /// <summary>
    /// Cryptography extensions for various classes.
    /// </summary>
    internal static class Cryptography
    {
        /// <summary>
        /// Helper array to speedup conversion
        /// </summary>
        private static string[] Baths = { "00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "0A", "0B", "0C", "0D", "0E", "0F",
                                  "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "1A", "1B", "1C", "1D", "1E", "1F",
                                  "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "2A", "2B", "2C", "2D", "2E", "2F",
                                  "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "3A", "3B", "3C", "3D", "3E", "3F",
                                  "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "4A", "4B", "4C", "4D", "4E", "4F",
                                  "50", "51", "52", "53", "54", "55", "56", "57", "58", "59", "5A", "5B", "5C", "5D", "5E", "5F",
                                  "60", "61", "62", "63", "64", "65", "66", "67", "68", "69", "6A", "6B", "6C", "6D", "6E", "6F",
                                  "70", "71", "72", "73", "74", "75", "76", "77", "78", "79", "7A", "7B", "7C", "7D", "7E", "7F",
                                  "80", "81", "82", "83", "84", "85", "86", "87", "88", "89", "8A", "8B", "8C", "8D", "8E", "8F",
                                  "90", "91", "92", "93", "94", "95", "96", "97", "98", "99", "9A", "9B", "9C", "9D", "9E", "9F",
                                  "A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7", "A8", "A9", "AA", "AB", "AC", "AD", "AE", "AF",
                                  "B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7", "B8", "B9", "BA", "BB", "BC", "BD", "BE", "BF",
                                  "C0", "C1", "C2", "C3", "C4", "C5", "C6", "C7", "C8", "C9", "CA", "CB", "CC", "CD", "CE", "CF",
                                  "D0", "D1", "D2", "D3", "D4", "D5", "D6", "D7", "D8", "D9", "DA", "DB", "DC", "DD", "DE", "DF",
                                  "E0", "E1", "E2", "E3", "E4", "E5", "E6", "E7", "E8", "E9", "EA", "EB", "EC", "ED", "EE", "EF",
                                  "F0", "F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "FA", "FB", "FC", "FD", "FE", "FF" };

        /// <summary>
        /// Generates a hash for the given plain text value and returns a base64-encoded result.
        /// </summary>
        /// <param name="source">Source value to be hashed. The function does not check whether this parameter is null.</param>
        /// <returns>Hash value formatted as a base64-encoded string.</returns>
        public static string GetSHA1Encoded(this string source)
        {
            byte[] plainTextBytes = Encoding.UTF8.GetBytes(source);
            using (var hash = SHA1.Create())
            {
                byte[] hashBytes = hash.ComputeHash(plainTextBytes);
                string hashValue = hashBytes.ToHexString();
                return hashValue;
            }
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
        /// Function converts byte array to it's hexadecimal implementation
        /// </summary>
        /// <param name="ArrayToConvert">Array to be converted</param>
        /// <returns>String to represent given array</returns>
        public static string ToHexString(this byte[] ArrayToConvert)
        {
            int LengthRequired = (ArrayToConvert.Length) * 2;
            StringBuilder tempstr = new StringBuilder(LengthRequired, LengthRequired);
            foreach (byte CurrentElem in ArrayToConvert)
            {
                tempstr.Append(Baths[CurrentElem]);
            }

            return tempstr.ToString();
        }
    }
}