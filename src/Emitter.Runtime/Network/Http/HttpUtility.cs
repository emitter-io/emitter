//
// System.Web.HttpUtility
//
// Authors:
//   Patrik Torstensson (Patrik.Torstensson@labs2.com)
//   Wictor Wilén (decode/encode functions) (wictor@ibizkit.se)
//   Tim Coleman (tim@timcoleman.com)
//   Gonzalo Paniagua Javier (gonzalo@ximian.com)
//
// Copyright (C) 2005 Novell, Inc (http://www.novell.com)
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//

using System;
using System.Collections;
using System.Collections.Generic;
using System.Collections.Specialized;
using System.Globalization;
using System.IO;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Provides methods for encoding and decoding URLs when processing Web requests.
    /// </summary>
    public sealed class HttpUtility
    {
        #region Fields
        private static Hashtable HttpEntities;
        private static object Lock = new object();

        private static Hashtable Entities
        {
            get
            {
                lock (Lock)
                {
                    if (HttpEntities == null)
                        InitEntities();

                    return HttpEntities;
                }
            }
        }

        #endregion Fields

        #region Constructors

        private static void InitEntities()
        {
            // Build the hash table of HTML entity references.  This list comes
            // from the HTML 4.01 W3C recommendation.
            HttpEntities = new Hashtable();
            HttpEntities.Add("nbsp", '\u00A0');
            HttpEntities.Add("iexcl", '\u00A1');
            HttpEntities.Add("cent", '\u00A2');
            HttpEntities.Add("pound", '\u00A3');
            HttpEntities.Add("curren", '\u00A4');
            HttpEntities.Add("yen", '\u00A5');
            HttpEntities.Add("brvbar", '\u00A6');
            HttpEntities.Add("sect", '\u00A7');
            HttpEntities.Add("uml", '\u00A8');
            HttpEntities.Add("copy", '\u00A9');
            HttpEntities.Add("ordf", '\u00AA');
            HttpEntities.Add("laquo", '\u00AB');
            HttpEntities.Add("not", '\u00AC');
            HttpEntities.Add("shy", '\u00AD');
            HttpEntities.Add("reg", '\u00AE');
            HttpEntities.Add("macr", '\u00AF');
            HttpEntities.Add("deg", '\u00B0');
            HttpEntities.Add("plusmn", '\u00B1');
            HttpEntities.Add("sup2", '\u00B2');
            HttpEntities.Add("sup3", '\u00B3');
            HttpEntities.Add("acute", '\u00B4');
            HttpEntities.Add("micro", '\u00B5');
            HttpEntities.Add("para", '\u00B6');
            HttpEntities.Add("middot", '\u00B7');
            HttpEntities.Add("cedil", '\u00B8');
            HttpEntities.Add("sup1", '\u00B9');
            HttpEntities.Add("ordm", '\u00BA');
            HttpEntities.Add("raquo", '\u00BB');
            HttpEntities.Add("frac14", '\u00BC');
            HttpEntities.Add("frac12", '\u00BD');
            HttpEntities.Add("frac34", '\u00BE');
            HttpEntities.Add("iquest", '\u00BF');
            HttpEntities.Add("Agrave", '\u00C0');
            HttpEntities.Add("Aacute", '\u00C1');
            HttpEntities.Add("Acirc", '\u00C2');
            HttpEntities.Add("Atilde", '\u00C3');
            HttpEntities.Add("Auml", '\u00C4');
            HttpEntities.Add("Aring", '\u00C5');
            HttpEntities.Add("AElig", '\u00C6');
            HttpEntities.Add("Ccedil", '\u00C7');
            HttpEntities.Add("Egrave", '\u00C8');
            HttpEntities.Add("Eacute", '\u00C9');
            HttpEntities.Add("Ecirc", '\u00CA');
            HttpEntities.Add("Euml", '\u00CB');
            HttpEntities.Add("Igrave", '\u00CC');
            HttpEntities.Add("Iacute", '\u00CD');
            HttpEntities.Add("Icirc", '\u00CE');
            HttpEntities.Add("Iuml", '\u00CF');
            HttpEntities.Add("ETH", '\u00D0');
            HttpEntities.Add("Ntilde", '\u00D1');
            HttpEntities.Add("Ograve", '\u00D2');
            HttpEntities.Add("Oacute", '\u00D3');
            HttpEntities.Add("Ocirc", '\u00D4');
            HttpEntities.Add("Otilde", '\u00D5');
            HttpEntities.Add("Ouml", '\u00D6');
            HttpEntities.Add("times", '\u00D7');
            HttpEntities.Add("Oslash", '\u00D8');
            HttpEntities.Add("Ugrave", '\u00D9');
            HttpEntities.Add("Uacute", '\u00DA');
            HttpEntities.Add("Ucirc", '\u00DB');
            HttpEntities.Add("Uuml", '\u00DC');
            HttpEntities.Add("Yacute", '\u00DD');
            HttpEntities.Add("THORN", '\u00DE');
            HttpEntities.Add("szlig", '\u00DF');
            HttpEntities.Add("agrave", '\u00E0');
            HttpEntities.Add("aacute", '\u00E1');
            HttpEntities.Add("acirc", '\u00E2');
            HttpEntities.Add("atilde", '\u00E3');
            HttpEntities.Add("auml", '\u00E4');
            HttpEntities.Add("aring", '\u00E5');
            HttpEntities.Add("aelig", '\u00E6');
            HttpEntities.Add("ccedil", '\u00E7');
            HttpEntities.Add("egrave", '\u00E8');
            HttpEntities.Add("eacute", '\u00E9');
            HttpEntities.Add("ecirc", '\u00EA');
            HttpEntities.Add("euml", '\u00EB');
            HttpEntities.Add("igrave", '\u00EC');
            HttpEntities.Add("iacute", '\u00ED');
            HttpEntities.Add("icirc", '\u00EE');
            HttpEntities.Add("iuml", '\u00EF');
            HttpEntities.Add("eth", '\u00F0');
            HttpEntities.Add("ntilde", '\u00F1');
            HttpEntities.Add("ograve", '\u00F2');
            HttpEntities.Add("oacute", '\u00F3');
            HttpEntities.Add("ocirc", '\u00F4');
            HttpEntities.Add("otilde", '\u00F5');
            HttpEntities.Add("ouml", '\u00F6');
            HttpEntities.Add("divide", '\u00F7');
            HttpEntities.Add("oslash", '\u00F8');
            HttpEntities.Add("ugrave", '\u00F9');
            HttpEntities.Add("uacute", '\u00FA');
            HttpEntities.Add("ucirc", '\u00FB');
            HttpEntities.Add("uuml", '\u00FC');
            HttpEntities.Add("yacute", '\u00FD');
            HttpEntities.Add("thorn", '\u00FE');
            HttpEntities.Add("yuml", '\u00FF');
            HttpEntities.Add("fnof", '\u0192');
            HttpEntities.Add("Alpha", '\u0391');
            HttpEntities.Add("Beta", '\u0392');
            HttpEntities.Add("Gamma", '\u0393');
            HttpEntities.Add("Delta", '\u0394');
            HttpEntities.Add("Epsilon", '\u0395');
            HttpEntities.Add("Zeta", '\u0396');
            HttpEntities.Add("Eta", '\u0397');
            HttpEntities.Add("Theta", '\u0398');
            HttpEntities.Add("Iota", '\u0399');
            HttpEntities.Add("Kappa", '\u039A');
            HttpEntities.Add("Lambda", '\u039B');
            HttpEntities.Add("Mu", '\u039C');
            HttpEntities.Add("Nu", '\u039D');
            HttpEntities.Add("Xi", '\u039E');
            HttpEntities.Add("Omicron", '\u039F');
            HttpEntities.Add("Pi", '\u03A0');
            HttpEntities.Add("Rho", '\u03A1');
            HttpEntities.Add("Sigma", '\u03A3');
            HttpEntities.Add("Tau", '\u03A4');
            HttpEntities.Add("Upsilon", '\u03A5');
            HttpEntities.Add("Phi", '\u03A6');
            HttpEntities.Add("Chi", '\u03A7');
            HttpEntities.Add("Psi", '\u03A8');
            HttpEntities.Add("Omega", '\u03A9');
            HttpEntities.Add("alpha", '\u03B1');
            HttpEntities.Add("beta", '\u03B2');
            HttpEntities.Add("gamma", '\u03B3');
            HttpEntities.Add("delta", '\u03B4');
            HttpEntities.Add("epsilon", '\u03B5');
            HttpEntities.Add("zeta", '\u03B6');
            HttpEntities.Add("eta", '\u03B7');
            HttpEntities.Add("theta", '\u03B8');
            HttpEntities.Add("iota", '\u03B9');
            HttpEntities.Add("kappa", '\u03BA');
            HttpEntities.Add("lambda", '\u03BB');
            HttpEntities.Add("mu", '\u03BC');
            HttpEntities.Add("nu", '\u03BD');
            HttpEntities.Add("xi", '\u03BE');
            HttpEntities.Add("omicron", '\u03BF');
            HttpEntities.Add("pi", '\u03C0');
            HttpEntities.Add("rho", '\u03C1');
            HttpEntities.Add("sigmaf", '\u03C2');
            HttpEntities.Add("sigma", '\u03C3');
            HttpEntities.Add("tau", '\u03C4');
            HttpEntities.Add("upsilon", '\u03C5');
            HttpEntities.Add("phi", '\u03C6');
            HttpEntities.Add("chi", '\u03C7');
            HttpEntities.Add("psi", '\u03C8');
            HttpEntities.Add("omega", '\u03C9');
            HttpEntities.Add("thetasym", '\u03D1');
            HttpEntities.Add("upsih", '\u03D2');
            HttpEntities.Add("piv", '\u03D6');
            HttpEntities.Add("bull", '\u2022');
            HttpEntities.Add("hellip", '\u2026');
            HttpEntities.Add("prime", '\u2032');
            HttpEntities.Add("Prime", '\u2033');
            HttpEntities.Add("oline", '\u203E');
            HttpEntities.Add("frasl", '\u2044');
            HttpEntities.Add("weierp", '\u2118');
            HttpEntities.Add("image", '\u2111');
            HttpEntities.Add("real", '\u211C');
            HttpEntities.Add("trade", '\u2122');
            HttpEntities.Add("alefsym", '\u2135');
            HttpEntities.Add("larr", '\u2190');
            HttpEntities.Add("uarr", '\u2191');
            HttpEntities.Add("rarr", '\u2192');
            HttpEntities.Add("darr", '\u2193');
            HttpEntities.Add("harr", '\u2194');
            HttpEntities.Add("crarr", '\u21B5');
            HttpEntities.Add("lArr", '\u21D0');
            HttpEntities.Add("uArr", '\u21D1');
            HttpEntities.Add("rArr", '\u21D2');
            HttpEntities.Add("dArr", '\u21D3');
            HttpEntities.Add("hArr", '\u21D4');
            HttpEntities.Add("forall", '\u2200');
            HttpEntities.Add("part", '\u2202');
            HttpEntities.Add("exist", '\u2203');
            HttpEntities.Add("empty", '\u2205');
            HttpEntities.Add("nabla", '\u2207');
            HttpEntities.Add("isin", '\u2208');
            HttpEntities.Add("notin", '\u2209');
            HttpEntities.Add("ni", '\u220B');
            HttpEntities.Add("prod", '\u220F');
            HttpEntities.Add("sum", '\u2211');
            HttpEntities.Add("minus", '\u2212');
            HttpEntities.Add("lowast", '\u2217');
            HttpEntities.Add("radic", '\u221A');
            HttpEntities.Add("prop", '\u221D');
            HttpEntities.Add("infin", '\u221E');
            HttpEntities.Add("ang", '\u2220');
            HttpEntities.Add("and", '\u2227');
            HttpEntities.Add("or", '\u2228');
            HttpEntities.Add("cap", '\u2229');
            HttpEntities.Add("cup", '\u222A');
            HttpEntities.Add("int", '\u222B');
            HttpEntities.Add("there4", '\u2234');
            HttpEntities.Add("sim", '\u223C');
            HttpEntities.Add("cong", '\u2245');
            HttpEntities.Add("asymp", '\u2248');
            HttpEntities.Add("ne", '\u2260');
            HttpEntities.Add("equiv", '\u2261');
            HttpEntities.Add("le", '\u2264');
            HttpEntities.Add("ge", '\u2265');
            HttpEntities.Add("sub", '\u2282');
            HttpEntities.Add("sup", '\u2283');
            HttpEntities.Add("nsub", '\u2284');
            HttpEntities.Add("sube", '\u2286');
            HttpEntities.Add("supe", '\u2287');
            HttpEntities.Add("oplus", '\u2295');
            HttpEntities.Add("otimes", '\u2297');
            HttpEntities.Add("perp", '\u22A5');
            HttpEntities.Add("sdot", '\u22C5');
            HttpEntities.Add("lceil", '\u2308');
            HttpEntities.Add("rceil", '\u2309');
            HttpEntities.Add("lfloor", '\u230A');
            HttpEntities.Add("rfloor", '\u230B');
            HttpEntities.Add("lang", '\u2329');
            HttpEntities.Add("rang", '\u232A');
            HttpEntities.Add("loz", '\u25CA');
            HttpEntities.Add("spades", '\u2660');
            HttpEntities.Add("clubs", '\u2663');
            HttpEntities.Add("hearts", '\u2665');
            HttpEntities.Add("diams", '\u2666');
            HttpEntities.Add("quot", '\u0022');
            HttpEntities.Add("amp", '\u0026');
            HttpEntities.Add("lt", '\u003C');
            HttpEntities.Add("gt", '\u003E');
            HttpEntities.Add("OElig", '\u0152');
            HttpEntities.Add("oelig", '\u0153');
            HttpEntities.Add("Scaron", '\u0160');
            HttpEntities.Add("scaron", '\u0161');
            HttpEntities.Add("Yuml", '\u0178');
            HttpEntities.Add("circ", '\u02C6');
            HttpEntities.Add("tilde", '\u02DC');
            HttpEntities.Add("ensp", '\u2002');
            HttpEntities.Add("emsp", '\u2003');
            HttpEntities.Add("thinsp", '\u2009');
            HttpEntities.Add("zwnj", '\u200C');
            HttpEntities.Add("zwj", '\u200D');
            HttpEntities.Add("lrm", '\u200E');
            HttpEntities.Add("rlm", '\u200F');
            HttpEntities.Add("ndash", '\u2013');
            HttpEntities.Add("mdash", '\u2014');
            HttpEntities.Add("lsquo", '\u2018');
            HttpEntities.Add("rsquo", '\u2019');
            HttpEntities.Add("sbquo", '\u201A');
            HttpEntities.Add("ldquo", '\u201C');
            HttpEntities.Add("rdquo", '\u201D');
            HttpEntities.Add("bdquo", '\u201E');
            HttpEntities.Add("dagger", '\u2020');
            HttpEntities.Add("Dagger", '\u2021');
            HttpEntities.Add("permil", '\u2030');
            HttpEntities.Add("lsaquo", '\u2039');
            HttpEntities.Add("rsaquo", '\u203A');
            HttpEntities.Add("euro", '\u20AC');
        }

        /// <summary>
        /// Initializes a new instance of the <see cref="HttpUtility"/> class.
        /// </summary>
        public HttpUtility()
        {
        }

        #endregion Constructors

        #region Methods

        /// <summary>
        /// Retrieves the subdomain from the specified URL.
        /// </summary>
        /// <param name="url">The URL from which to retrieve the subdomain.</param>
        /// <returns>The subdomain if it exist, otherwise null.</returns>
        public static string GetSubDomain(Uri url)
        {
            if (url.HostNameType == UriHostNameType.Dns)
                return GetSubDomain(url.Host);

            return null;
        }

        /// <summary>
        /// Retrieves the subdomain from the specified URL.
        /// </summary>
        /// <param name="host">The URL from which to retrieve the subdomain.</param>
        /// <returns>The subdomain if it exist, otherwise null.</returns>
        public static string GetSubDomain(string host)
        {
            if (String.IsNullOrWhiteSpace(host))
                return null;

            if (host.Split('.').Length > 2)
            {
                int lastIndex = host.LastIndexOf(".");
                int index = host.LastIndexOf(".", lastIndex - 1);
                return host.Substring(0, index);
            }

            return null;
        }

        /// <summary>
        /// Minimally converts a string into an HTML-encoded string and sends the encoded string to a TextWriter output stream.
        /// </summary>
        /// <param name="s">The string to encode</param>
        /// <param name="output">A TextWriter output stream.</param>
        public static void HtmlAttributeEncode(string s, TextWriter output)
        {
            output.Write(HtmlAttributeEncode(s));
        }

        /// <summary>
        /// Minimally converts a string to an HTML-encoded string
        /// </summary>
        /// <param name="s">The string to encode.</param>
        /// <returns>An encoded string.</returns>
        public static string HtmlAttributeEncode(string s)
        {
            if (null == s)
                return null;

            if (s.IndexOf('&') == -1 && s.IndexOf('"') == -1)
                return s;

            StringBuilder output = new StringBuilder();
            foreach (char c in s)
                switch (c)
                {
                    case '&':
                        output.Append("&amp;");
                        break;

                    case '"':
                        output.Append("&quot;");
                        break;

                    default:
                        output.Append(c);
                        break;
                }

            return output.ToString();
        }

        /// <summary>
        /// Converts a string that has been encoded for transmission in a URL into a decoded string.
        /// </summary>
        /// <param name="str">The string to decode.</param>
        /// <returns>A decoded string.</returns>
        public static string UrlDecode(string str)
        {
            return UrlDecode(str, Encoding.UTF8);
        }

        /// <summary>
        /// Converts a URL-encoded string into a decoded string, using the specified encoding object.
        /// </summary>
        /// <param name="s">The string to decode.</param>
        /// <param name="e">The Encoding that specifies the decoding scheme.</param>
        /// <returns>A decoded string.</returns>
        public static string UrlDecode(string s, Encoding e)
        {
            if (null == s)
                return null;

            if (s.IndexOf('%') == -1 && s.IndexOf('+') == -1)
                return s;

            if (e == null)
                e = Encoding.UTF8;

            StringBuilder output = new StringBuilder();
            long len = s.Length;
            MemoryStream bytes = new MemoryStream();
            int xchar;

            for (int i = 0; i < len; i++)
            {
                if (s[i] == '%' && i + 2 < len && s[i + 1] != '%')
                {
                    if (s[i + 1] == 'u' && i + 5 < len)
                    {
                        if (bytes.Length > 0)
                        {
                            output.Append(GetChars(bytes, e));
                            bytes.SetLength(0);
                        }

                        xchar = GetChar(s, i + 2, 4);
                        if (xchar != -1)
                        {
                            output.Append((char)xchar);
                            i += 5;
                        }
                        else
                        {
                            output.Append('%');
                        }
                    }
                    else if ((xchar = GetChar(s, i + 1, 2)) != -1)
                    {
                        bytes.WriteByte((byte)xchar);
                        i += 2;
                    }
                    else
                    {
                        output.Append('%');
                    }
                    continue;
                }

                if (bytes.Length > 0)
                {
                    output.Append(GetChars(bytes, e));
                    bytes.SetLength(0);
                }

                if (s[i] == '+')
                {
                    output.Append(' ');
                }
                else
                {
                    output.Append(s[i]);
                }
            }

            if (bytes.Length > 0)
            {
                output.Append(GetChars(bytes, e));
            }

            bytes = null;
            return output.ToString();
        }

        /// <summary>
        /// Converts a URL-encoded byte array into a decoded string using the specified decoding object
        /// </summary>
        /// <param name="bytes">The array of bytes to decode.</param>
        /// <param name="e">The Encoding that specifies the decoding scheme.</param>
        /// <returns>A decoded string.</returns>
        public static string UrlDecode(byte[] bytes, Encoding e)
        {
            if (bytes == null)
                return null;

            return UrlDecode(bytes, 0, bytes.Length, e);
        }

        private static char[] GetChars(MemoryStream b, Encoding e)
        {
            return e.GetChars(b.GetBuffer(), 0, (int)b.Length);
        }

        private static int GetInt(byte b)
        {
            char c = (char)b;
            if (c >= '0' && c <= '9')
                return c - '0';

            if (c >= 'a' && c <= 'f')
                return c - 'a' + 10;

            if (c >= 'A' && c <= 'F')
                return c - 'A' + 10;

            return -1;
        }

        private static int GetChar(byte[] bytes, int offset, int length)
        {
            int value = 0;
            int end = length + offset;
            for (int i = offset; i < end; i++)
            {
                int current = GetInt(bytes[i]);
                if (current == -1)
                    return -1;
                value = (value << 4) + current;
            }

            return value;
        }

        private static int GetChar(string str, int offset, int length)
        {
            int val = 0;
            int end = length + offset;
            for (int i = offset; i < end; i++)
            {
                char c = str[i];
                if (c > 127)
                    return -1;

                int current = GetInt((byte)c);
                if (current == -1)
                    return -1;
                val = (val << 4) + current;
            }

            return val;
        }

        private static char[] hexChars = "0123456789abcdef".ToCharArray();
        private const string notEncoded = "!'()*-._";

        private static void UrlEncodeChar(char c, Stream result, bool isUnicode)
        {
            if (c > 255)
            {
                //FIXME: what happens when there is an internal error?
                //if (!isUnicode)
                //	throw new ArgumentOutOfRangeException ("c", c, "c must be less than 256");
                int idx;
                int i = (int)c;

                result.WriteByte((byte)'%');
                result.WriteByte((byte)'u');
                idx = i >> 12;
                result.WriteByte((byte)hexChars[idx]);
                idx = (i >> 8) & 0x0F;
                result.WriteByte((byte)hexChars[idx]);
                idx = (i >> 4) & 0x0F;
                result.WriteByte((byte)hexChars[idx]);
                idx = i & 0x0F;
                result.WriteByte((byte)hexChars[idx]);
                return;
            }

            if (c > ' ' && notEncoded.IndexOf(c) != -1)
            {
                result.WriteByte((byte)c);
                return;
            }
            if (c == ' ')
            {
                result.WriteByte((byte)'+');
                return;
            }
            if ((c < '0') ||
                (c < 'A' && c > '9') ||
                (c > 'Z' && c < 'a') ||
                (c > 'z'))
            {
                if (isUnicode && c > 127)
                {
                    result.WriteByte((byte)'%');
                    result.WriteByte((byte)'u');
                    result.WriteByte((byte)'0');
                    result.WriteByte((byte)'0');
                }
                else
                    result.WriteByte((byte)'%');

                int idx = ((int)c) >> 4;
                result.WriteByte((byte)hexChars[idx]);
                idx = ((int)c) & 0x0F;
                result.WriteByte((byte)hexChars[idx]);
            }
            else
                result.WriteByte((byte)c);
        }

        /// <summary>
        /// Converts a URL-encoded byte array into a decoded string using the specified encoding object, starting at the specified position in the array, and continuing for the specified number of bytes.
        /// </summary>
        /// <param name="bytes">The array of bytes to decode.</param>
        /// <param name="offset">The position in the byte to begin decoding.</param>
        /// <param name="count">The number of bytes to decode.</param>
        /// <param name="e"> Encoding object that specifies the decoding scheme.</param>
        /// <returns>A decoded string.</returns>
        public static string UrlDecode(byte[] bytes, int offset, int count, Encoding e)
        {
            if (bytes == null)
                return null;
            if (count == 0)
                return String.Empty;

            if (bytes == null)
                throw new ArgumentNullException("bytes");

            if (offset < 0 || offset > bytes.Length)
                throw new ArgumentOutOfRangeException("offset");

            if (count < 0 || offset + count > bytes.Length)
                throw new ArgumentOutOfRangeException("count");

            StringBuilder output = new StringBuilder();
            MemoryStream acc = new MemoryStream();

            int end = count + offset;
            int xchar;
            for (int i = offset; i < end; i++)
            {
                if (bytes[i] == '%' && i + 2 < count && bytes[i + 1] != '%')
                {
                    if (bytes[i + 1] == (byte)'u' && i + 5 < end)
                    {
                        if (acc.Length > 0)
                        {
                            output.Append(GetChars(acc, e));
                            acc.SetLength(0);
                        }
                        xchar = GetChar(bytes, i + 2, 4);
                        if (xchar != -1)
                        {
                            output.Append((char)xchar);
                            i += 5;
                            continue;
                        }
                    }
                    else if ((xchar = GetChar(bytes, i + 1, 2)) != -1)
                    {
                        acc.WriteByte((byte)xchar);
                        i += 2;
                        continue;
                    }
                }

                if (acc.Length > 0)
                {
                    output.Append(GetChars(acc, e));
                    acc.SetLength(0);
                }

                if (bytes[i] == '+')
                {
                    output.Append(' ');
                }
                else
                {
                    output.Append((char)bytes[i]);
                }
            }

            if (acc.Length > 0)
            {
                output.Append(GetChars(acc, e));
            }

            acc = null;
            return output.ToString();
        }

        /// <summary>
        /// Converts a URL-encoded array of bytes into a decoded array of bytes.
        /// </summary>
        /// <param name="bytes">The array of bytes to decode.</param>
        /// <returns>A decoded array of bytes.</returns>
        public static byte[] UrlDecodeToBytes(byte[] bytes)
        {
            if (bytes == null)
                return null;

            return UrlDecodeToBytes(bytes, 0, bytes.Length);
        }

        /// <summary>
        /// Converts a URL-encoded string into a decoded array of bytes.
        /// </summary>
        /// <param name="str">The string to decode.</param>
        /// <returns>A decoded array of bytes.</returns>
        public static byte[] UrlDecodeToBytes(string str)
        {
            return UrlDecodeToBytes(str, Encoding.UTF8);
        }

        /// <summary>
        /// Converts a URL-encoded string into a decoded array of bytes using the specified decoding object.
        /// </summary>
        /// <param name="str">The string to decode.</param>
        /// <param name="e">The Encoding object that specifies the decoding scheme.</param>
        /// <returns>A decoded array of bytes.</returns>
        public static byte[] UrlDecodeToBytes(string str, Encoding e)
        {
            if (str == null)
                return null;

            if (e == null)
                throw new ArgumentNullException("e");

            return UrlDecodeToBytes(e.GetBytes(str));
        }

        /// <summary>
        /// Converts a URL-encoded array of bytes into a decoded array of bytes, starting at the specified position in the array and continuing for the specified number of bytes.
        /// </summary>
        /// <param name="bytes">The array of bytes to decode.</param>
        /// <param name="offset">The position in the byte array at which to begin decoding.</param>
        /// <param name="count">The number of bytes to decode.</param>
        /// <returns>A decoded array of bytes.</returns>
        public static byte[] UrlDecodeToBytes(byte[] bytes, int offset, int count)
        {
            if (bytes == null)
                return null;
            if (count == 0)
                return new byte[0];

            int len = bytes.Length;
            if (offset < 0 || offset >= len)
                throw new ArgumentOutOfRangeException("offset");

            if (count < 0 || offset > len - count)
                throw new ArgumentOutOfRangeException("count");

            MemoryStream result = new MemoryStream();
            int end = offset + count;
            for (int i = offset; i < end; i++)
            {
                char c = (char)bytes[i];
                if (c == '+')
                {
                    c = ' ';
                }
                else if (c == '%' && i < end - 2)
                {
                    int xchar = GetChar(bytes, i + 1, 2);
                    if (xchar != -1)
                    {
                        c = (char)xchar;
                        i += 2;
                    }
                }
                result.WriteByte((byte)c);
            }

            return result.ToArray();
        }

        /// <summary>
        /// Encodes a URL string.
        /// </summary>
        /// <param name="str">The text to encode.</param>
        /// <returns>An encoded string.</returns>
        public static string UrlEncode(string str)
        {
            return UrlEncode(str, Encoding.UTF8);
        }

        /// <summary>
        /// Encodes a URL string using the specified encoding object.
        /// </summary>
        /// <param name="s">The text to encode.</param>
        /// <param name="e">The Encoding object that specifies the encoding scheme.</param>
        /// <returns>An encoded string.</returns>
        public static string UrlEncode(string s, Encoding e)
        {
            if (s == null)
                return null;

            if (s == "")
                return "";

            byte[] bytes = e.GetBytes(s);
            return Encoding.ASCII.GetString(UrlEncodeToBytes(bytes, 0, bytes.Length));
        }

        /// <summary>
        /// Converts a byte array into an encoded URL string.
        /// </summary>
        /// <param name="bytes">The array of bytes to encode.</param>
        /// <returns>An encoded string.</returns>
        public static string UrlEncode(byte[] bytes)
        {
            if (bytes == null)
                return null;

            if (bytes.Length == 0)
                return "";

            return Encoding.ASCII.GetString(UrlEncodeToBytes(bytes, 0, bytes.Length));
        }

        /// <summary>
        /// Converts a byte array into a URL-encoded string, starting at the specified position in the array and continuing for the specified number of bytes.
        /// </summary>
        /// <param name="bytes">The array of bytes to encode.</param>
        /// <param name="offset">The position in the byte array at which to begin encoding.</param>
        /// <param name="count">The number of bytes to encode.</param>
        /// <returns>An encoded string.</returns>
        public static string UrlEncode(byte[] bytes, int offset, int count)
        {
            if (bytes == null)
                return null;

            if (bytes.Length == 0)
                return "";

            return Encoding.ASCII.GetString(UrlEncodeToBytes(bytes, offset, count));
        }

        /// <summary>
        /// Converts a string into a URL-encoded array of bytes.
        /// </summary>
        /// <param name="str"></param>
        /// <returns>An encoded array of bytes.</returns>
        public static byte[] UrlEncodeToBytes(string str)
        {
            return UrlEncodeToBytes(str, Encoding.UTF8);
        }

        /// <summary>
        /// Converts a string into a URL-encoded array of bytes using the specified encoding object.
        /// </summary>
        /// <param name="str">The string to encode</param>
        /// <param name="e">The Encoding that specifies the encoding scheme.</param>
        /// <returns>An encoded array of bytes.</returns>
        public static byte[] UrlEncodeToBytes(string str, Encoding e)
        {
            if (str == null)
                return null;

            if (str == "")
                return new byte[0];

            byte[] bytes = e.GetBytes(str);
            return UrlEncodeToBytes(bytes, 0, bytes.Length);
        }

        /// <summary>
        /// Converts an array of bytes into a URL-encoded array of bytes.
        /// </summary>
        /// <param name="bytes">The array of bytes to encode.</param>
        /// <returns>An encoded array of bytes.</returns>
        public static byte[] UrlEncodeToBytes(byte[] bytes)
        {
            if (bytes == null)
                return null;

            if (bytes.Length == 0)
                return new byte[0];

            return UrlEncodeToBytes(bytes, 0, bytes.Length);
        }

        /// <summary>
        /// Converts an array of bytes into a URL-encoded array of bytes, starting at the specified position in the array and continuing for the specified number of bytes.
        /// </summary>
        /// <param name="bytes">The array of bytes to encode.</param>
        /// <param name="offset">The position in the byte array at which to begin encoding.</param>
        /// <param name="count">The number of bytes to encode.</param>
        /// <returns>An encoded array of bytes.</returns>
        public static byte[] UrlEncodeToBytes(byte[] bytes, int offset, int count)
        {
            if (bytes == null)
                return null;

            int len = bytes.Length;
            if (len == 0)
                return new byte[0];

            if (offset < 0 || offset >= len)
                throw new ArgumentOutOfRangeException("offset");

            if (count < 0 || count > len - offset)
                throw new ArgumentOutOfRangeException("count");

            MemoryStream result = new MemoryStream(count);
            int end = offset + count;
            for (int i = offset; i < end; i++)
                UrlEncodeChar((char)bytes[i], result, false);

            return result.ToArray();
        }

        /// <summary>
        /// Converts a string into a Unicode string.
        /// </summary>
        /// <param name="str">The string to convert.</param>
        /// <returns>A Unicode string in %UnicodeValue notation.</returns>
        public static string UrlEncodeUnicode(string str)
        {
            if (str == null)
                return null;

            return Encoding.ASCII.GetString(UrlEncodeUnicodeToBytes(str));
        }

        /// <summary>
        /// Converts a Unicode string into an array of bytes.
        /// </summary>
        /// <param name="str">The string to convert.</param>
        /// <returns>A byte array.</returns>
        public static byte[] UrlEncodeUnicodeToBytes(string str)
        {
            if (str == null)
                return null;

            if (str == "")
                return new byte[0];

            MemoryStream result = new MemoryStream(str.Length);
            foreach (char c in str)
            {
                UrlEncodeChar(c, result, true);
            }
            return result.ToArray();
        }

        /// <summary>
        /// Encodes a URL Path.
        /// </summary>
        /// <param name="value">The path to encode.</param>
        /// <returns>Encoded url path.</returns>
        public static string UrlPathEncode(string value)
        {
            if (String.IsNullOrEmpty(value))
                return value;

            MemoryStream result = new MemoryStream();
            int length = value.Length;
            for (int i = 0; i < length; i++)
                UrlPathEncodeChar(value[i], result);

            return Encoding.ASCII.GetString(result.ToArray());
        }

        internal static void UrlPathEncodeChar(char c, Stream result)
        {
            if (c < 33 || c > 126)
            {
                byte[] bIn = Encoding.UTF8.GetBytes(c.ToString());
                for (int i = 0; i < bIn.Length; i++)
                {
                    result.WriteByte((byte)'%');
                    int idx = ((int)bIn[i]) >> 4;
                    result.WriteByte((byte)hexChars[idx]);
                    idx = ((int)bIn[i]) & 0x0F;
                    result.WriteByte((byte)hexChars[idx]);
                }
            }
            else if (c == ' ')
            {
                result.WriteByte((byte)'%');
                result.WriteByte((byte)'2');
                result.WriteByte((byte)'0');
            }
            else
                result.WriteByte((byte)c);
        }

        /// <summary>
        /// Decodes an HTML-encoded string and returns the decoded string.
        /// </summary>
        /// <param name="s">The HTML string to decode. </param>
        /// <returns>The decoded text.</returns>
        public static string HtmlDecode(string s)
        {
            if (s == null)
                throw new ArgumentNullException("s");

            if (s.IndexOf('&') == -1)
                return s;

            StringBuilder entity = new StringBuilder();
            StringBuilder output = new StringBuilder();
            int len = s.Length;
            // 0 -> nothing,
            // 1 -> right after '&'
            // 2 -> between '&' and ';' but no '#'
            // 3 -> '#' found after '&' and getting numbers
            int state = 0;
            int number = 0;
            bool have_trailing_digits = false;

            for (int i = 0; i < len; i++)
            {
                char c = s[i];
                if (state == 0)
                {
                    if (c == '&')
                    {
                        entity.Append(c);
                        state = 1;
                    }
                    else
                    {
                        output.Append(c);
                    }
                    continue;
                }

                if (c == '&')
                {
                    state = 1;
                    if (have_trailing_digits)
                    {
                        entity.Append(number.ToString(CultureInfo.InvariantCulture));
                        have_trailing_digits = false;
                    }

                    output.Append(entity.ToString());
                    entity.Length = 0;
                    entity.Append('&');
                    continue;
                }

                if (state == 1)
                {
                    if (c == ';')
                    {
                        state = 0;
                        output.Append(entity.ToString());
                        output.Append(c);
                        entity.Length = 0;
                    }
                    else
                    {
                        number = 0;
                        if (c != '#')
                        {
                            state = 2;
                        }
                        else
                        {
                            state = 3;
                        }
                        entity.Append(c);
                    }
                }
                else if (state == 2)
                {
                    entity.Append(c);
                    if (c == ';')
                    {
                        string key = entity.ToString();
                        if (key.Length > 1 && Entities.ContainsKey(key.Substring(1, key.Length - 2)))
                            key = Entities[key.Substring(1, key.Length - 2)].ToString();

                        output.Append(key);
                        state = 0;
                        entity.Length = 0;
                    }
                }
                else if (state == 3)
                {
                    if (c == ';')
                    {
                        if (number > 65535)
                        {
                            output.Append("&#");
                            output.Append(number.ToString(CultureInfo.InvariantCulture));
                            output.Append(";");
                        }
                        else
                        {
                            output.Append((char)number);
                        }
                        state = 0;
                        entity.Length = 0;
                        have_trailing_digits = false;
                    }
                    else if (Char.IsDigit(c))
                    {
                        number = number * 10 + ((int)c - '0');
                        have_trailing_digits = true;
                    }
                    else
                    {
                        state = 2;
                        if (have_trailing_digits)
                        {
                            entity.Append(number.ToString(CultureInfo.InvariantCulture));
                            have_trailing_digits = false;
                        }
                        entity.Append(c);
                    }
                }
            }

            if (entity.Length > 0)
            {
                output.Append(entity.ToString());
            }
            else if (have_trailing_digits)
            {
                output.Append(number.ToString(CultureInfo.InvariantCulture));
            }
            return output.ToString();
        }

        /// <summary>
        /// Decodes an HTML-encoded string and sends the resulting output to a TextWriter output stream.
        /// </summary>
        /// <param name="s">The HTML string to decode</param>
        /// <param name="output">The TextWriter output stream containing the decoded string. </param>
        public static void HtmlDecode(string s, TextWriter output)
        {
            if (s != null)
                output.Write(HtmlDecode(s));
        }

        /// <summary>
        /// HTML-encodes a string and returns the encoded string.
        /// </summary>
        /// <param name="s">The text string to encode. </param>
        /// <returns>The HTML-encoded text.</returns>
        public static string HtmlEncode(string s)
        {
            if (s == null)
                return null;

            StringBuilder output = new StringBuilder();

            foreach (char c in s)
                switch (c)
                {
                    case '&':
                        output.Append("&amp;");
                        break;

                    case '>':
                        output.Append("&gt;");
                        break;

                    case '<':
                        output.Append("&lt;");
                        break;

                    case '"':
                        output.Append("&quot;");
                        break;

                    default:
                        // MS starts encoding with &# from 160 and stops at 255.
                        // We don't do that. One reason is the 65308/65310 unicode
                        // characters that look like '<' and '>'.
                        if (c > 159)
                        {
                            output.Append("&#");
                            output.Append(((int)c).ToString(CultureInfo.InvariantCulture));
                            output.Append(";");
                        }
                        else
                        {
                            output.Append(c);
                        }
                        break;
                }
            return output.ToString();
        }

        /// <summary>
        /// HTML-encodes a string and sends the resulting output to a TextWriter output stream.
        /// </summary>
        /// <param name="s">The string to encode. </param>
        /// <param name="output">The TextWriter output stream containing the encoded string. </param>
        public static void HtmlEncode(string s, TextWriter output)
        {
            if (s != null)
                output.Write(HtmlEncode(s));
        }

        /// <summary>
        /// Parses a query string into a NameValueCollection using UTF8 encoding.
        /// </summary>
        /// <param name="content">The query string to parse.</param>
        /// <returns>A NameValueCollection of query parameters and values.</returns>
        public static NameValueCollection ParseQueryString(byte[] content)
        {
            return ParseQueryString(Encoding.UTF8.GetString(content), Encoding.UTF8);
        }

        /// <summary>
        /// Parses a query string into a NameValueCollection using UTF8 encoding.
        /// </summary>
        /// <param name="query">The query string to parse.</param>
        /// <returns>A NameValueCollection of query parameters and values.</returns>
        public static NameValueCollection ParseQueryString(string query)
        {
            return ParseQueryString(query, Encoding.UTF8);
        }

        /// <summary>
        /// Parses a query string into a NameValueCollection using the specified Encoding.
        /// </summary>
        /// <param name="query">The query string to parse.</param>
        /// <param name="encoding">The Encoding to use.</param>
        /// <returns>A NameValueCollection of query parameters and values.</returns>
        public static NameValueCollection ParseQueryString(string query, Encoding encoding)
        {
            if (query == null)
                throw new ArgumentNullException("query");
            if (encoding == null)
                throw new ArgumentNullException("encoding");
            if (query.Length == 0 || (query.Length == 1 && query[0] == '?'))
                return new NameValueCollection();
            if (query[0] == '?')
                query = query.Substring(1);

            NameValueCollection result = new NameValueCollection();
            ParseQueryString(query, encoding, result);
            return result;
        }

        internal static void ParseQueryString(string query, Encoding encoding, NameValueCollection result)
        {
            if (query.Length == 0)
                return;

            int namePos = 0;
            while (namePos <= query.Length)
            {
                int valuePos = -1, valueEnd = -1;
                for (int q = namePos; q < query.Length; q++)
                {
                    if (valuePos == -1 && query[q] == '=')
                    {
                        valuePos = q + 1;
                    }
                    else if (query[q] == '&')
                    {
                        valueEnd = q;
                        break;
                    }
                }

                string name, value;
                if (valuePos == -1)
                {
                    name = null;
                    valuePos = namePos;
                }
                else
                {
                    name = UrlDecode(query.Substring(namePos, valuePos - namePos - 1), encoding);
                }
                if (valueEnd < 0)
                {
                    namePos = -1;
                    valueEnd = query.Length;
                }
                else
                {
                    namePos = valueEnd + 1;
                }
                value = UrlDecode(query.Substring(valuePos, valueEnd - valuePos), encoding);

                result.Add(name, value);
                if (namePos == -1) break;
            }
        }

        /// <summary>
        /// Parses the URL and returns the base and the query parts.
        /// </summary>
        /// <param name="url">The url string to parse.</param>
        /// <param name="page">The page (base) part as output argument.</param>
        /// <param name="query">The query part as output argument.</param>
        /// <returns>Whether the URL was parsed successfuly or not.</returns>
        public static bool ParseUrl(string url, out string page, out string query)
        {
            page = null;
            query = null;
            int sIdx = url.IndexOf('?');

            if (url == null)
                throw new ArgumentNullException("url");
            if (url.Length == 0)
                return false;
            if (sIdx >= 0)
            {
                page = url.Substring(0, sIdx);
                if (url.Length - page.Length > 0)
                    query = url.Substring(sIdx, url.Length - page.Length);
                return true;
            }
            else
            {
                page = url;
                return true;
            }
        }

        /// <summary>
        /// Parses the URL and returns the base/page part.
        /// </summary>
        /// <param name="url">The url string to parse.</param>
        /// <returns>The base/page part.</returns>
        public static string GetPageBase(string url)
        {
            string page;
            string query;
            ParseUrl(url, out page, out query);
            return page;
        }

        /// <summary>
        /// Parses the URL and returns the query string part.
        /// </summary>
        /// <param name="url">The url string to parse.</param>
        /// <returns>The query string part</returns>
        public static string GetQuery(string url)
        {
            string page;
            string query;
            ParseUrl(url, out page, out query);
            return query;
        }

        /// <summary>
        /// Checks if the buffer starts as a http request
        /// </summary>
        internal static bool IsHttp(byte[] buffer)
        {
            const int offset = 0;

            // First, is it possible that this is a HTTP request
            if (buffer[offset] == 'P' ||
                buffer[offset] == 'G' ||
                buffer[offset] == 'H' ||
                buffer[offset] == 'D' ||
                buffer[offset] == 'O' ||
                buffer[offset] == 'T' ||
                buffer[offset] == 'C')
            {
                // Next, check if it contains a HTTP Verb
                if (buffer[offset] == 'P' &&
                    buffer[offset + 1] == 'O' &&
                    buffer[offset + 2] == 'S' &&
                    buffer[offset + 3] == 'T')
                    return true;
                if (buffer[offset] == 'G' &&
                    buffer[offset + 1] == 'E' &&
                    buffer[offset + 2] == 'T')
                    return true;
                if (buffer[offset] == 'P' &&
                    buffer[offset + 1] == 'U' &&
                    buffer[offset + 2] == 'T')
                    return true;
                if (buffer[offset] == 'H' &&
                    buffer[offset + 1] == 'E' &&
                    buffer[offset + 2] == 'A' &&
                    buffer[offset + 3] == 'D')
                    return true;
                if (buffer[offset] == 'D' &&
                    buffer[offset + 1] == 'E' &&
                    buffer[offset + 2] == 'L' &&
                    buffer[offset + 3] == 'E' &&
                    buffer[offset + 4] == 'T' &&
                    buffer[offset + 5] == 'E')
                    return true;
                if (buffer[offset] == 'O' &&
                    buffer[offset + 1] == 'P' &&
                    buffer[offset + 2] == 'T' &&
                    buffer[offset + 3] == 'I' &&
                    buffer[offset + 4] == 'O' &&
                    buffer[offset + 5] == 'N' &&
                    buffer[offset + 6] == 'S')
                    return true;
                if (buffer[offset] == 'T' &&
                    buffer[offset + 1] == 'R' &&
                    buffer[offset + 2] == 'A' &&
                    buffer[offset + 3] == 'C' &&
                    buffer[offset + 4] == 'E')
                    return true;
                if (buffer[offset] == 'C' &&
                    buffer[offset + 1] == 'O' &&
                    buffer[offset + 2] == 'N' &&
                    buffer[offset + 3] == 'N' &&
                    buffer[offset + 4] == 'E' &&
                    buffer[offset + 5] == 'C' &&
                    buffer[offset + 6] == 'T')
                    return true;
            }

            return false;
        }

        #endregion Methods

        #region Http Querying

        /// <summary>
        /// Issues a GET http request and returns the answer string.
        /// </summary>
        /// <param name="url">The url of the http request.</param>
        /// <param name="timeout">Timeout for operation.</param>
        /// <param name="answer">The parsed answer string</param>
        /// <returns>Whether it returned 200 or not.</returns>
        public static Expected<string> Get(string url, int timeout)
        {
            try
            {
                var task = GetAsync(url, timeout);
                task.Wait();
                return task.Result;
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                return ex;
            }
        }

        /// <summary>
        /// Issues a GET http request and returns the answer string.
        /// </summary>
        /// <param name="url">The url of the http request.</param>
        /// <param name="timeout">Timeout for operation.</param>
        /// <param name="answer">The parsed answer string</param>
        /// <param name="headers">Optional headers</param>
        /// <returns>Whether it returned 200 or not.</returns>
        public static async Task<Expected<string>> GetAsync(string url, int timeout, params KeyValuePair<string, string>[] headers)
        {
            try
            {
                if (!url.StartsWith("http://") && !url.StartsWith("https://"))
                    url = "http://" + url;

                using (var client = new HttpClient())
                {
                    foreach (var header in headers)
                    {
                        client.DefaultRequestHeaders.TryAddWithoutValidation(header.Key, header.Value);
                    }

                    var task = client.GetStringAsync(url);
                    if (await Task.WhenAny(task, Task.Delay(timeout)) == task)
                    {
                        if (task.Status == TaskStatus.Faulted)
                            return task.Exception;

                        return task.Result;
                    }
                    else
                    {
                        // Timed out
                        return new TimeoutException();
                    }
                }
            }
            catch (Exception ex)
            {
                if (ex.InnerException != null)
                    ex = ex.InnerException;
                return ex;
            }
        }

        /// <summary>
        /// Issues a POST http request and returns the answer string.
        /// </summary>
        /// <param name="url">The url of the http request.</param>
        /// <param name="timeout">Timeout for operation.</param>
        /// <param name="content">The content to post.</param>
        /// <returns>Whether it returned 200 or not.</returns>
        public static async Task<Expected<string>> PostAsync(string url, HttpContent content, int timeout)
        {
            try
            {
                if (!url.StartsWith("http://") && !url.StartsWith("https://"))
                    url = "http://" + url;

                using (var client = new HttpClient())
                {
                    var task = client.PostAsync(url, content);
                    if (await Task.WhenAny(task, Task.Delay(timeout)) == task)
                    {
                        if (task.Status == TaskStatus.Faulted)
                            return task.Exception;

                        return await task.Result.Content.ReadAsStringAsync();
                    }
                    else
                    {
                        // Timed out
                        return new TimeoutException();
                    }
                }
            }
            catch (Exception ex)
            {
                //if (ex.InnerException != null)
                //    ex = ex.InnerException;
                Service.Logger.Log(ex);
                return ex;
            }
        }

        #endregion Http Querying
    }
}