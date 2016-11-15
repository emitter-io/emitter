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
using System.Runtime.InteropServices;

namespace Emitter
{
    internal unsafe struct TextRange
    {
        private const int toUpper = 'a' - 'A';
        private string CachedString;

        public static readonly TextRange Empty = new TextRange(null, 0);
        public readonly byte* pStart;
        public readonly int Length;

        public TextRange(byte* pStart, int length)
        {
            this.pStart = pStart;
            this.Length = length;
            this.CachedString = null;
        }

        public unsafe bool Is(string match)
        {
            if (match == null)
            {
                if (this.Length == 0)
                    return true;

                return false;
            }

            int matchLen = match.Length;
            if (matchLen != this.Length)
                return false;

            byte* p = this.pStart;

            fixed (char* pString = match)
            {
                char* pMatch = pString;

                for (int i = 0; i < matchLen; i++, p++, pMatch++)
                {
                    byte a = *p;
                    byte b = (byte)*pMatch;

                    // convert key to upper
                    if (a >= 'a' &&
                        a <= 'z')
                    {
                        a -= toUpper;
                    }

                    if (b >= 'a' &&
                        b <= 'z')
                    {
                        b -= toUpper;
                    }

                    if (a != b)
                        return false;
                }
            }

            return true;
        }

        public unsafe bool Contains(string match)
        {
            if (string.IsNullOrEmpty(match))
                return false;

            int length = this.Length;
            int matchLen = match.Length;

            if (length < matchLen)
                return false;

            int scanCount = length - matchLen + 1;

            byte* pStart = this.pStart;

            for (int i = 0; i < scanCount; i++)
            {
                byte* p = pStart + i;

                int j = 0;

                for (; j < matchLen; j++, p++)
                {
                    byte a = *p;
                    byte b = (byte)match[j];

                    // convert key to upper
                    if (a >= 'a' &&
                        a <= 'z')
                    {
                        a -= toUpper;
                    }

                    if (b >= 'a' &&
                        b <= 'z')
                    {
                        b -= toUpper;
                    }

                    if (a != b)
                        break;
                }

                if (j == matchLen)
                    return true;
            }

            return false;
        }

        public byte[] GetBytes()
        {
            byte[] raw = new byte[this.Length];
            Marshal.Copy((IntPtr)this.pStart, raw, 0, this.Length);
            return raw;
        }

        public override string ToString()
        {
            if (this.CachedString != null)
                return this.CachedString;

            if (this.pStart == null ||
                this.Length == 0)
            {
                this.CachedString = string.Empty;
                return this.CachedString;
            }

            //this.CachedString = new string((sbyte*) this.pStart, 0, this.Length, Encoding.ASCII);
            this.CachedString = Memory.CopyString(this.pStart, this.Length);
            return this.CachedString;
        }
    }
}