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
using System.Collections.Specialized;
using System.Linq;
using System.Text;

namespace Emitter.Network.Http
{
    [Flags]
    internal enum CookieFlags : byte
    {
        Secure = 1,
        HttpOnly = 2
    }

    /// <summary>
    /// Provides a type-safe way to create and manipulate individual HTTP cookies.
    /// </summary>
    public sealed class HttpCookie
    {
        private string fName;
        private string fDomain;
        private string fPath = "/";
        private CookieFlags fFlags = 0;
        private DateTime fExpires = DateTime.MinValue;
        private NameValueCollection fValues;

        /// <summary>
        /// Creates and names a new cookie.
        /// </summary>
        /// <param name="name">The name of the cookie.</param>
        public HttpCookie(string name)
        {
            this.fName = name;
            fValues = new CookieNVC();
            Value = "";
        }

        /// <summary>
        /// Creates, names, and assigns a value to a new cookie.
        /// </summary>
        /// <param name="name">The name of the cookie.</param>
        /// <param name="value">The value of the cookie.</param>
        public HttpCookie(string name, string value)
            : this(name)
        {
            Value = value;
        }

        /// <summary>
        /// Gets a string representing the header of the cookie.
        /// </summary>
        internal string GetCookieHeader()
        {
            StringBuilder builder = new StringBuilder("");

            builder.Append(fName);
            builder.Append("=");
            builder.Append(Value);

            if (fDomain != null)
            {
                builder.Append("; domain=");
                builder.Append(fDomain);
            }

            if (fPath != null)
            {
                builder.Append("; path=");
                builder.Append(fPath);
            }

            if (fExpires != DateTime.MinValue)
            {
                builder.Append("; expires=");
                builder.Append(fExpires.ToUniversalTime().ToString("r"));
            }

            if ((fFlags & CookieFlags.Secure) != 0)
            {
                builder.Append("; secure");
            }

            if ((fFlags & CookieFlags.HttpOnly) != 0)
            {
                builder.Append("; HttpOnly");
            }

            return builder.ToString();
        }

        /// <summary>
        /// Gets or sets the domain to associate the cookie with.
        /// </summary>
        public string Domain
        {
            get
            {
                return fDomain;
            }
            set
            {
                fDomain = value;
            }
        }

        /// <summary>
        /// Gets or sets the expiration date and time for the cookie.
        /// </summary>
        public DateTime Expires
        {
            get
            {
                return fExpires;
            }
            set
            {
                fExpires = value;
            }
        }

        /// <summary>
        /// Gets a value indicating whether a cookie has subkeys.
        /// </summary>
        public bool HasKeys
        {
            get
            {
                return fValues.HasKeys();
            }
        }

        /// <summary>
        /// Gets a shortcut to the <see cref="HttpCookie"/>.Values property.
        /// </summary>
        /// <param name="key">The key of the cookie property.</param>
        /// <returns>The value of the property.</returns>
        public string this[string key]
        {
            get
            {
                return fValues[key];
            }
            set
            {
                fValues[key] = value;
            }
        }

        /// <summary>
        /// Gets or sets the name of a cookie.
        /// </summary>
        public string Name
        {
            get
            {
                return fName;
            }
            set
            {
                fName = value;
            }
        }

        /// <summary>
        /// Gets or sets the virtual path to transmit with the current cookie.
        /// </summary>
        public string Path
        {
            get
            {
                return fPath;
            }
            set
            {
                fPath = value;
            }
        }

        /// <summary>
        /// Gets or sets a value indicating whether to transmit the cookie using Secure Sockets Layer (SSL)--that is, over HTTPS only.
        /// </summary>
        public bool Secure
        {
            get
            {
                return (fFlags & CookieFlags.Secure) == CookieFlags.Secure;
            }
            set
            {
                if (value)
                    fFlags |= CookieFlags.Secure;
                else
                    fFlags &= ~CookieFlags.Secure;
            }
        }

        /// <summary>
        /// Gets or sets an individual cookie value.
        /// </summary>
        public string Value
        {
            get
            {
                return fValues.ToString();
            }
            set
            {
                fValues.Clear();

                if (value != null && value != "")
                {
                    string[] components = value.Split('&');
                    foreach (string kv in components)
                    {
                        int pos = kv.IndexOf('=');
                        if (pos < 1)
                        {
                            fValues.Add(null, kv);
                        }
                        else
                        {
                            string key = kv.Substring(0, pos);
                            string val = kv.Substring(pos + 1);

                            fValues.Add(key, val);
                        }
                    }
                }
            }
        }

        /// <summary>
        /// Gets a collection of key/value pairs that are contained within a single cookie object.
        /// </summary>
        public NameValueCollection Values
        {
            get
            {
                return fValues;
            }
        }

        /// <summary>
        /// Gets or sets a value that specifies whether a cookie is accessible by client-side script.
        /// </summary>
		public bool HttpOnly
        {
            get
            {
                return (fFlags & CookieFlags.HttpOnly) == CookieFlags.HttpOnly;
            }

            set
            {
                fFlags |= CookieFlags.HttpOnly;
            }
        }

        /*
		 * simple utility class that just overrides ToString
		 * to get the desired behavior for
		 * HttpCookie.Values
		 */

        private class CookieNVC : NameValueCollection
        {
            public override string ToString()
            {
                StringBuilder builder = new StringBuilder("");

                bool first_key = true;
                foreach (string key in Keys)
                {
                    if (!first_key)
                        builder.Append("&");

                    bool first_val = true;
                    foreach (string v in GetValues(key))
                    {
                        if (!first_val)
                            builder.Append("&");

                        if (key != null)
                        {
                            builder.Append(key);
                            builder.Append("=");
                        }
                        builder.Append(v);
                        first_val = false;
                    }
                    first_key = false;
                }

                return builder.ToString();
            }

            /* MS's implementation has the interesting quirk that if you do:
             * cookie.Values[null] = "foo"
             * it clears out the rest of the values.
             */

            public override void Set(string name, string value)
            {
                if (this.IsReadOnly)
                    throw new NotSupportedException("Collection is read-only");

                if (name == null)
                    Clear();

                base.Set(name, value);
            }
        }
    }
}