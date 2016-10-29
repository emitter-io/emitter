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

namespace Emitter.Text
{
    /// <remarks>
    /// Based on http://www.csharp411.com/convert-binary-to-base64-string/
    /// </remarks>
    public sealed class Base64
    {
        private const char PaddingChar = '=';
        private readonly byte[] Map;

        public readonly char[] CharacterSet;
        public readonly bool PaddingEnabled;

        private const string CharacterSetBase = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

        public readonly char PlusChar;
        public readonly char SlashChar;

        public static readonly Base64 Default = new Base64('+', '/', true);
        public static readonly Base64 DefaultNoPadding = new Base64('+', '/', false);
        public static readonly Base64 UrlEncoding = new Base64('-', '_', false);
        public static readonly Base64 XmlEncoding = new Base64('_', ':', false);
        public static readonly Base64 RegExEncoding = new Base64('!', '-', false);
        public static readonly Base64 FileEncoding = new Base64('+', '-', false);

        public Base64(char plusChar, char slashChar, bool paddingEnabled)
            : this((CharacterSetBase + plusChar + slashChar).ToCharArray(), paddingEnabled)
        {
            PlusChar = plusChar;
            SlashChar = slashChar;
        }

        private Base64(char[] characterSet, bool paddingEnabled)
        {
            PaddingEnabled = paddingEnabled;
            CharacterSet = characterSet;
            Map = Create(CharacterSet);
        }

        /// <summary>
        /// Encodes a byte array to a base64 string.
        /// </summary>
        /// <param name="data"></param>
        /// <returns></returns>
		public unsafe string ToBase(byte[] data)
        {
            int length;
            if (data == null || 0 == (length = data.Length))
                return string.Empty;

            fixed (byte* _d = data)
                return ToBase(_d, length);
        }

        /// <summary>
        /// Encodes a byte array to base64 string.
        /// </summary>
        /// <param name="data"></param>
        /// <param name="length"></param>
        /// <returns></returns>
        public unsafe string ToBase(byte* data, int length)
        {
            if (length == 0)
                return string.Empty;

            fixed (char* _cs = CharacterSet)
            {
                byte* d = data;
                int padding = length % 3;
                if (padding > 0)
                    padding = 3 - padding;
                int blocks = (length - 1) / 3 + 1;

                int l = blocks * 4;

                char[] _s = new char[l];

                fixed (char* _sp = _s)
                {
                    char* sp = _sp;
                    byte b1, b2, b3;

                    for (int i = 1; i < blocks; i++)
                    {
                        b1 = *d++;
                        b2 = *d++;
                        b3 = *d++;

                        *sp++ = _cs[(b1 & 0xFC) >> 2];
                        *sp++ = _cs[(b2 & 0xF0) >> 4 | (b1 & 0x03) << 4];
                        *sp++ = _cs[(b3 & 0xC0) >> 6 | (b2 & 0x0F) << 2];
                        *sp++ = _cs[b3 & 0x3F];
                    }

                    bool pad2 = padding == 2;
                    bool pad1 = padding > 0;

                    b1 = *d++;
                    b2 = pad2 ? (byte)0 : *d++;
                    b3 = pad1 ? (byte)0 : *d++;

                    *sp++ = _cs[(b1 & 0xFC) >> 2];
                    *sp++ = _cs[(b2 & 0xF0) >> 4 | (b1 & 0x03) << 4];
                    *sp++ = pad2 ? '=' : _cs[(b3 & 0xC0) >> 6 | (b2 & 0x0F) << 2];
                    *sp++ = pad1 ? '=' : _cs[b3 & 0x3F];

                    if (!PaddingEnabled)
                    {
                        if (pad2) l--;
                        if (pad1) l--;
                    }
                }

                return new string(_s, 0, l);
            }
        }

        public byte[] FromBase(string data)
        {
            int length = data == null ? 0 : data.Length;
            if (length == 0)
                return new byte[0];

            unsafe
            {
                fixed (char* _p = data.ToCharArray())
                {
                    char* p2 = _p;

                    int blocks = (length - 1) / 4 + 1;
                    int bytes = blocks * 3;

                    int padding = blocks * 4 - length;

                    if (length > 2 && p2[length - 2] == PaddingChar)
                        padding = 2;
                    else if (length > 1 && p2[length - 1] == PaddingChar)
                        padding = 1;

                    byte[] _data = new byte[bytes - padding];

                    byte temp1, temp2;
                    byte* dp;

                    fixed (byte* _d = _data)
                    {
                        dp = _d;

                        for (int i = 1; i < blocks; i++)
                        {
                            temp1 = Map[*p2++];
                            temp2 = Map[*p2++];

                            *dp++ = (byte)((temp1 << 2) | ((temp2 & 0x30) >> 4));
                            temp1 = Map[*p2++];
                            *dp++ = (byte)(((temp1 & 0x3C) >> 2) | ((temp2 & 0x0F) << 4));
                            temp2 = Map[*p2++];
                            *dp++ = (byte)(((temp1 & 0x03) << 6) | temp2);
                        }

                        temp1 = Map[*p2++];
                        temp2 = Map[*p2++];

                        *dp++ = (byte)((temp1 << 2) | ((temp2 & 0x30) >> 4));

                        temp1 = Map[*p2++];

                        if (padding != 2)
                            *dp++ = (byte)(((temp1 & 0x3C) >> 2) | ((temp2 & 0x0F) << 4));

                        temp2 = Map[*p2++];
                        if (padding == 0)
                            *dp++ = (byte)(((temp1 & 0x03) << 6) | temp2);
                    }
                    return _data;
                }
            }
        }

        private static byte[] Create(char[] characterSet)
        {
            var x = new byte[123];
            for (byte i = 0; i < characterSet.Length; i++)
            {
                x[characterSet[i]] = i;
            }

            return x;
        }
    }
}