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

namespace Emitter.Network.Http
{
    /// <summary>
    /// Provides a type-safe way to manipulate HTTP cookies.
    /// </summary>
    public sealed class HttpCookieCollection : NameObjectCollectionBase
    {
        #region Internal Constructors
        private bool AutoFill = false;

        internal HttpCookieCollection(bool auto_fill, bool read_only)
        {
            this.AutoFill = auto_fill;
            this.IsReadOnly = read_only;
        }

        internal HttpCookieCollection(string cookies)
        {
            if (cookies == null || cookies == "")
                return;

            string[] cookie_components = cookies.Split(';');
            foreach (string kv in cookie_components)
            {
                int pos = kv.IndexOf('=');
                if (pos == -1)
                {
                    /* XXX ugh */
                    continue;
                }
                else
                {
                    string key = kv.Substring(0, pos);
                    string val = kv.Substring(pos + 1);

                    Add(new HttpCookie(key.Trim(), val.Trim()));
                }
            }
        }

        #endregion Internal Constructors

        /// <summary>
        /// Initializes a new instance of the <see cref="HttpCookieCollection"/> class.
        /// </summary>
        public HttpCookieCollection()
        {
        }

        /// <summary>
        /// Adds the specified cookie to the cookie collection.
        /// </summary>
        /// <param name="cookie">The <see cref="HttpCookie"/> to add.</param>
        public void Add(HttpCookie cookie)
        {
            if (BaseGet(cookie.Name) != null)
                return;

            BaseAdd(cookie.Name, cookie);
        }

        /// <summary>
        /// Clears all cookies from the cookie collection.
        /// </summary>
        public void Clear()
        {
            BaseClear();
        }

        /// <summary>
        /// Copies this cookie colklection to an array.
        /// </summary>
        /// <param name="array">The destination array to copy to.</param>
        /// <param name="index">The starting index.</param>
        public void CopyTo(Array array, int index)
        {
            /* XXX this is kind of gross and inefficient
                * since it makes a copy of the superclass's
                * list */
            object[] values = BaseGetAllValues();
            values.CopyTo(array, index);
        }

        /// <summary>
        /// Returns the key (name) of the cookie at the specified numerical index
        /// </summary>
        /// <param name="index">The index of the key to retrieve from the collection. </param>
        /// <returns>The name of the cookie specified by index.</returns>
        public string GetKey(int index)
        {
            HttpCookie cookie = (HttpCookie)BaseGet(index);
            if (cookie == null)
                return null;
            else
                return cookie.Name;
        }

        /// <summary>
        /// Removes the cookie with the specified name from the collection
        /// </summary>
        /// <param name="name">The name of the cookie to remove from the collection. </param>
        public void Remove(string name)
        {
            BaseRemove(name);
        }

        /// <summary>
        /// Updates the value of an existing cookie in a cookie collection.
        /// </summary>
        /// <param name="cookie">The <see cref="HttpCookie"/> object to update. </param>
        public void Set(HttpCookie cookie)
        {
            BaseSet(cookie.Name, cookie);
        }

        /// <summary>
        /// Updates the value of an existing cookie in a cookie collection.
        /// </summary>
        /// <param name="name">The name of <see cref="HttpCookie"/> object to update. </param>
        /// <param name="value">The value to set. </param>
        public void Set(string name, string value)
        {
            var existing = BaseGet(name) as HttpCookie;
            if (existing == null)
            {
                // A new cookie
                HttpCookie cookie = new HttpCookie(name, value);
                cookie.Expires = Timer.Now + TimeSpan.FromMinutes(20);
                BaseSet(cookie.Name, cookie);
            }
            else
            {
                // Update the existing
                BaseSet(existing.Name, existing);
            }
        }

        /// <summary>
        /// Updates the value of an existing cookie in a cookie collection.
        /// </summary>
        /// <param name="name">The name of <see cref="HttpCookie"/> object to update. </param>
        /// <param name="value">The value to set. </param>
        /// <param name="expiresIn">The amount of time in which the cookie expires.</param>
        public void Set(string name, string value, TimeSpan expiresIn)
        {
            var existing = BaseGet(name) as HttpCookie;
            if (existing == null)
            {
                // A new cookie
                HttpCookie cookie = new HttpCookie(name, value);
                cookie.Expires = Timer.Now + expiresIn;
                BaseSet(cookie.Name, cookie);
            }
            else
            {
                // Update the existing
                BaseSet(existing.Name, existing);
            }
        }

        /// <summary>
        /// Updates the value of an existing cookie in a cookie collection.
        /// </summary>
        /// <param name="name">The name of <see cref="HttpCookie"/> object to update. </param>
        /// <param name="value">The value to set. </param>
        /// <param name="expires">The date when the cookie expires.</param>
        public void Set(string name, string value, DateTime expires)
        {
            var existing = BaseGet(name) as HttpCookie;
            if (existing == null)
            {
                // A new cookie
                HttpCookie cookie = new HttpCookie(name, value);
                cookie.Expires = expires;
                BaseSet(cookie.Name, cookie);
            }
            else
            {
                // Update the existing
                BaseSet(existing.Name, existing);
            }
        }

        /// <summary>
        /// Returns the <see cref="HttpCookie"/> item with the specified index from the cookie collection.
        /// </summary>
        /// <param name="index">The index of the cookie to retrieve from the collection.</param>
        /// <returns>The <see cref="HttpCookie"/> specified by index.</returns>
        public HttpCookie Get(int index)
        {
            return (HttpCookie)BaseGet(index);
        }

        /// <summary>
        /// Returns the <see cref="HttpCookie"/> item with the specified name from the cookie collection.
        /// </summary>
        /// <param name="name">The name of the cookie to retrieve from the collection.</param>
        /// <returns>The <see cref="HttpCookie"/> specified by name.</returns>
        public HttpCookie Get(string name)
        {
            return (HttpCookie)BaseGet(name);
        }

        /// <summary>
        /// Returns the string representation of the cookie's value with the specified name from the cookie collection.
        /// </summary>
        /// <param name="name">The name of the cookie to retrieve from the collection.</param>
        /// <returns>The string representation of the value.</returns>
        public string GetString(string name)
        {
            var cookie = BaseGet(name) as HttpCookie;
            if (cookie == null)
                return String.Empty;
            if (cookie.Value == null)
                return String.Empty;
            return cookie.Value.ToString();
        }

        /// <summary>
        /// Gets the cookie with the specified numerical index from the cookie collection.
        /// </summary>
        /// <param name="index">The index of the cookie to retrieve from the collection. </param>
        /// <returns>The <see cref="HttpCookie"/> specified by index.</returns>
        public HttpCookie this[int index]
        {
            get
            {
                return (HttpCookie)BaseGet(index);
            }
        }

        /// <summary>
        /// Gets the cookie with the specified name from the cookie collection.
        /// </summary>
        /// <param name="name">The hame of the cookie to retrieve from the collection</param>
        /// <returns>The <see cref="HttpCookie"/> specified by name.</returns>
        public HttpCookie this[string name]
        {
            get
            {
                HttpCookie cookie = (HttpCookie)BaseGet(name);
                if (!IsReadOnly && AutoFill && cookie == null)
                {
                    cookie = new HttpCookie(name);
                    BaseAdd(name, cookie);
                }
                return cookie;
            }
        }

        /// <summary>
        /// Gets a string array containing all the keys (cookie names) in the cookie collection.
        /// </summary>
        public string[] AllKeys
        {
            get
            {
                /* XXX another inefficient copy due to
                    * lack of exposure from the base
                    * class */
                string[] keys = new string[Keys.Count];
                for (int i = 0; i < Keys.Count; i++)
                    keys[i] = Keys[i];

                return keys;
            }
        }
    }
}