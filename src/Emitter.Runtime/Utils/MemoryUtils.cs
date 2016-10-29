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
using System.Text;

namespace Emitter
{
    /// <summary>
    /// Represents a memory utility that can be used for faster memcpy.
    /// </summary>
    public static class Memory
    {
        /// <summary>
        /// Copies a chunk of memory.
        /// </summary>
        /// <param name="source">The source buffer to read.</param>
        /// <param name="sourceOffset">The source offset to read.</param>
        /// <param name="destination">The destination buffer to write.</param>
        /// <param name="destinationOffset">The destination offset to write.</param>
        /// <param name="length">The lenght of bytes to copy.</param>
        public unsafe static void Copy(byte[] source, int sourceOffset, byte[] destination, int destinationOffset, int length)
        {
            // Write to the block
            if (length < 512)
            {
                // Copy memory with a fast method
                fixed (byte* ptr1 = destination)
                fixed (byte* ptr2 = source)
                {
                    // Go to the appropriate offsets
                    var dest = ptr1 + destinationOffset;
                    var src = ptr2 + sourceOffset;

                    if (length >= 16)
                    {
                        do
                        {
                            *(long*)dest = *(long*)src;
                            *(long*)(dest + 8) = *(long*)(src + 8);
                            dest += 16;
                            src += 16;
                        }
                        while ((length -= 16) >= 16);
                    }
                    if (length > 0)
                    {
                        if ((length & 8) != 0)
                        {
                            *(long*)dest = *(long*)src;
                            dest += 8;
                            src += 8;
                        }
                        if ((length & 4) != 0)
                        {
                            *(int*)dest = *(int*)src;
                            dest += 4;
                            src += 4;
                        }
                        if ((length & 2) != 0)
                        {
                            *(short*)dest = *(short*)src;
                            dest += 2;
                            src += 2;
                        }
                        if ((length & 1) != 0)
                        {
                            byte* finalByte = dest;
                            dest = finalByte + 1;
                            byte* finalByteSrc = src;
                            src = finalByteSrc + 1;
                            *finalByte = *finalByteSrc;
                        }
                    }
                }
            }
            else
            {
                // Copy memory with standard method
                Buffer.BlockCopy(source, sourceOffset, destination, destinationOffset, length);
            }
        }

        /// <summary>
        /// Copies the string.
        /// </summary>
        /// <param name="src">The bytes containing an UTF-8 string.</param>
        /// <param name="length">The length of the string.</param>
        /// <returns></returns>
        public static unsafe string CopyString(byte* src, int length)
        {
            var str = new char[length];
            fixed (char* pStr = str)
            {
                Encoding.UTF8.GetChars(src, length, pStr, length);
                return new string(str);
            }
        }
    }
}