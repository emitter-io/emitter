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
using System.Collections.Specialized;
using System.IO;
using System.Linq;
using Emitter.Providers;

namespace Emitter.Network.Http
{
    /// <summary>
    /// Provides access to session-state values as well as session-level settings and lifetime management methods.
    /// </summary>
    public sealed class HttpSession : DisposableObject, ICollection, IEnumerable
    {
        /// <summary>
        /// Values of the session.
        /// </summary>
        private HttpSessionDictionary Values = new HttpSessionDictionary();

        /// <summary>
        /// Creates a new <see cref="HttpSession"/> object.
        /// </summary>
        public HttpSession()
        {
            this.Key = Service.Providers
                .Resolve<SecurityProvider>()
                .CreateSessionToken();
            this.Timeout = TimeSpan.FromMinutes(20);
            this.Expires = Timer.Now + this.Timeout;
        }

        /// <summary>
        /// Gets the key/identifier of this <see cref="HttpSession"/>.
        /// </summary>
        public string Key
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets or sets when this session will expire.
        /// </summary>
        public DateTime Expires
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets the timeout of this session object. Default is 20 minutes.
        /// </summary>
        public TimeSpan Timeout
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets a session value by name.
        /// </summary>
        /// <param name="name">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public object this[string name]
        {
            get { return this.Values[name]; }
            set { this.Values[name] = value; }
        }

        /// <summary>
        /// Gets or sets a session value by index.
        /// </summary>
        /// <param name="index">The index of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public object this[int index]
        {
            get { return this.Values[index]; }
            set
            {
                string key = this.Values.Keys[index];
                this.Values[key] = value;
            }
        }

        /// <summary>
        /// Gets all the keys in this <see cref="HttpSession"/> object.
        /// </summary>
        public NameObjectCollectionBase.KeysCollection Keys
        {
            get { return this.Values.Keys; }
        }

        /// <summary>
        /// Gets the number of items in the session-state collection.
        /// </summary>
        public int Count
        {
            get { return this.Values.Count; }
        }

        /// <summary>
        /// Gets whether this object is synchronized or not.
        /// </summary>
        public bool IsSynchronized
        {
            get { return false; }
        }

        /// <summary>
        /// Gets the synchronization root for this object.
        /// </summary>
        public object SyncRoot
        {
            get { return this; }
        }

        /// <summary>
        /// Adds a new item to the session-state collection.
        /// </summary>
        /// <param name="name">The name of the item to add to the session-state collection.</param>
        /// <param name="value">The value of the item to add to the session-state collection.</param>
        public void Add(string name, object value)
        {
            this.Values[name] = value;
        }

        /// <summary>
        /// Removes all keys and values from the session-state collection.
        /// </summary>
        public void Clear()
        {
            if (this.Values != null)
                this.Values.Clear();
        }

        /// <summary>
        /// Deletes an item from the session-state collection.
        /// </summary>
        /// <param name="name">The name of the item to delete from the session-state collection.</param>
        public void Remove(string name)
        {
            try
            {
                if (this[name] != null)
                    this.Remove(name);
            }
            catch { }
        }

        /// <summary>
        /// Removes all keys and values from the session-state collection.
        /// </summary>
        public void RemoveAll()
        {
            if (this.Values != null)
                this.Values.Clear();
        }

        /// <summary>
        /// Removes the item from the collection at the specified index.
        /// </summary>
        /// <param name="index">The index to remove the item at.</param>
        public void RemoveAt(int index)
        {
            if (this.Values != null)
                this.Values.RemoveAt(index);
        }

        /// <summary>
        /// Gets the enumerator for this <see cref="HttpSession"/> object.
        /// </summary>
        /// <returns>Returns the enumerator for this <see cref="HttpSession"/> object.</returns>
        public IEnumerator GetEnumerator()
        {
            return this.Values.GetEnumerator();
        }

        /// <summary>
        /// Copies the values of this <see cref="HttpSession"/> to the destination array.
        /// </summary>
        /// <param name="array">The destination array.</param>
        /// <param name="index">The destination index to start to.</param>
        public void CopyTo(Array array, int index)
        {
            NameObjectCollectionBase.KeysCollection all = Keys;
            for (int i = 0; i < all.Count; i++)
                array.SetValue(this.Values[all[i]], i + index);
        }

        #region Helper Members

        /// <summary>
        /// Gets a session value as a string or empty string if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public string GetAsStringOrEmpty(string key)
        {
            return GetAsString(key, String.Empty);
        }

        /// <summary>
        /// Gets a session value as a string or a null string if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public string GetAsString(string key)
        {
            return GetAsString(key, null);
        }

        /// <summary>
        /// Gets a session value as a single value or false if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public float GetAsSingle(string key)
        {
            // Check if there's a key
            if (this[key] == null)
                return -1;

            // Check if it's a boolean value that we have
            object value = this[key];
            if (value is Single)
                return (Single)value;

            // Attempt to parse
            Single result;
            string text = value.ToString();
            if (Single.TryParse(text, out result))
                return result;

            return -1;
        }

        /// <summary>
        /// Gets a session value as a double value or false if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public double GetAsDouble(string key)
        {
            // Check if there's a key
            if (this[key] == null)
                return -1;

            // Check if it's a boolean value that we have
            object value = this[key];
            if (value is Double)
                return (Double)value;

            // Attempt to parse
            Double result;
            string text = value.ToString();
            if (Double.TryParse(text, out result))
                return result;

            return -1;
        }

        /// <summary>
        /// Gets a session value as a int16 value or false if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public short GetAsInt16(string key)
        {
            // Check if there's a key
            if (this[key] == null)
                return -1;

            // Check if it's a boolean value that we have
            object value = this[key];
            if (value is Int16)
                return (Int16)value;

            // Attempt to parse
            Int16 result;
            string text = value.ToString();
            if (Int16.TryParse(text, out result))
                return result;

            return -1;
        }

        /// <summary>
        /// Gets a session value as a int32 value or false if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public int GetAsInt32(string key)
        {
            // Check if there's a key
            if (this[key] == null)
                return -1;

            // Check if it's a boolean value that we have
            object value = this[key];
            if (value is Int32)
                return (Int32)value;

            // Attempt to parse
            Int32 result;
            string text = value.ToString();
            if (Int32.TryParse(text, out result))
                return result;

            return -1;
        }

        /// <summary>
        /// Gets a session value as a int64 value or false if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public long GetAsInt64(string key)
        {
            // Check if there's a key
            if (this[key] == null)
                return -1;

            // Check if it's a boolean value that we have
            object value = this[key];
            if (value is Int64)
                return (Int64)value;

            // Attempt to parse
            Int64 result;
            string text = value.ToString();
            if (Int64.TryParse(text, out result))
                return result;

            return -1;
        }

        /// <summary>
        /// Gets a session value as a string or default string if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <param name="defaultValue">The default value of the string.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public string GetAsString(string key, string defaultValue)
        {
            return this[key] != null
                ? this[key].ToString()
                : defaultValue;
        }

        /// <summary>
        /// Gets a session value as a boolean value or false if no key was found.
        /// </summary>
        /// <param name="key">The key name of the session value.</param>
        /// <returns>The session-state value with the specified name.</returns>
        public bool GetAsBoolean(string key)
        {
            // Check if there's a key
            if (this[key] == null)
                return false;

            // Check if it's a boolean value that we have
            object value = this[key];
            if (value is Boolean)
                return (Boolean)value;

            // Attempt to parse
            bool result;
            string text = value.ToString();
            if (Boolean.TryParse(text, out result))
                return result;

            return false;
        }

        #endregion Helper Members

        #region Static Members

        /// <summary>
        /// The name of the cookie that contains the session id.
        /// </summary>
        public static readonly string CookieKey = "spike-session";

        /// <summary>
        /// Check if session id is present in the cookie. If there is no session id
        /// we create a new session.
        /// </summary>
        /// <param name="context">The context to operate on.</param>
        internal static void OnRequest(HttpContext context)
        {
            // Session variable
            HttpSession session;

            // Get the cookie
            HttpCookie cookie = context.Request.Cookies.Get(HttpSession.CookieKey);
            if (cookie != null)
            {
                // Get the session object we have in the cache
                session = Service.Http.GetSession(cookie.Value);
                if (session != null)
                {
                    // Set the session we got from the cache
                    context.Session = session;
                    session.Expires = Timer.Now + session.Timeout;

                    // Disable caching a cookie
                    context.Response.Headers["Cache-control"] = "private";

                    // We must let it live longer
                    context.Response.Cookies.Set(HttpSession.CookieKey, session.Key,
                        session.Expires);
                    return;
                }

                // Else: a cookie specifies a session which was not found in the cache
                // we must skip and create a new session. Unfortunately we won't have any data
                // here.
            }

            // Create a new session key
            session = new HttpSession();

            // Set the session to the new session
            context.Session = session;

            // Put it to the cache
            Service.Http.PutSession(session);

            // Set to both request and response
            context.Response.Headers["Cache-control"] = "private";
            context.Request.Cookies.Set(HttpSession.CookieKey, session.Key, session.Expires);
            context.Response.Cookies.Set(HttpSession.CookieKey, session.Key, session.Expires);
        }

        #endregion Static Members
    }

    /// <summary>
    /// Represents an event that is sent by the <see cref="HttpProvider"/> when a session is created.
    /// </summary>
    /// <param name="session">The session that is created.</param>
    /// <param name="key">The key of the session.</param>
    public delegate void HttpSessionCreate(string key, HttpSession session);

    /// <summary>
    /// Represents an event that is sent by the <see cref="HttpProvider"/> when a session expires.
    /// </summary>
    /// <param name="session">The session that is about to expire.</param>
    /// <param name="key">The key of the session.</param>
    public delegate void HttpSessionExpire(string key, HttpSession session);

    /// <summary>
    /// Represents internal session dictionary.
    /// </summary>
    internal class HttpSessionDictionary : NameObjectCollectionBase
    {
        private static ArrayList types;
        private bool _dirty;

        static HttpSessionDictionary()
        {
            types = new ArrayList();
            types.Add("");
            types.Add(typeof(string));
            types.Add(typeof(int));
            types.Add(typeof(bool));
            types.Add(typeof(DateTime));
            types.Add(typeof(Decimal));
            types.Add(typeof(Byte));
            types.Add(typeof(Char));
            types.Add(typeof(Single));
            types.Add(typeof(Double));
            types.Add(typeof(short));
            types.Add(typeof(long));
            types.Add(typeof(ushort));
            types.Add(typeof(uint));
            types.Add(typeof(ulong));
        }

        public HttpSessionDictionary()
        {
        }

        internal void Clear()
        {
            _dirty = true;
            BaseClear();
        }

        internal string GetKey(int index)
        {
            return BaseGetKey(index);
        }

        internal void Remove(string s)
        {
            BaseRemove(s);
            _dirty = true;
        }

        internal void RemoveAt(int index)
        {
            BaseRemoveAt(index);
            _dirty = true;
        }

        internal void Serialize(BinaryWriter w)
        {
            w.Write(Count);
            foreach (string key in Keys)
            {
                w.Write(key);
                object value = BaseGet(key);
                if (value == null)
                {
                    w.Write(16); // types.Count + 1
                    continue;
                }

                SerializeByType(w, value);
            }
        }

        private static void SerializeByType(BinaryWriter w, object value)
        {
            var type = value.GetType();
            int i = types.IndexOf(type);
            if (i == -1)
                throw new InvalidOperationException("Unable to serialize the type " + type);

            w.Write(i);
            switch (i)
            {
                case 1: w.Write((string)value); break;
                case 2: w.Write((int)value); break;
                case 3: w.Write((bool)value); break;
                case 4:
                    w.Write(((DateTime)value).Ticks);
                    break;

                case 5:
                    w.Write((decimal)value);
                    break;

                case 6:
                    w.Write((byte)value);
                    break;

                case 7:
                    w.Write((char)value);
                    break;

                case 8:
                    w.Write((float)value);
                    break;

                case 9:
                    w.Write((double)value);
                    break;

                case 10:
                    w.Write((short)value);
                    break;

                case 11:
                    w.Write((long)value);
                    break;

                case 12:
                    w.Write((ushort)value);
                    break;

                case 13:
                    w.Write((uint)value);
                    break;

                case 14:
                    w.Write((ulong)value);
                    break;
            }
        }

        internal static HttpSessionDictionary Deserialize(BinaryReader r)
        {
            var result = new HttpSessionDictionary();
            for (int i = r.ReadInt32(); i > 0; i--)
            {
                string key = r.ReadString();
                int index = r.ReadInt32();
                if (index == 16)
                    result[key] = null;
                else
                    result[key] = DeserializeFromIndex(index, r);
            }

            return result;
        }

        private static object DeserializeFromIndex(int index, BinaryReader r)
        {
            object value = null;
            switch (index)
            {
                case 1:
                    value = r.ReadString();
                    break;

                case 2:
                    value = r.ReadInt32();
                    break;

                case 3:
                    value = r.ReadBoolean();
                    break;

                case 4:
                    value = new DateTime(r.ReadInt64());
                    break;

                case 5:
                    value = r.ReadDecimal();
                    break;

                case 6:
                    value = r.ReadByte();
                    break;

                case 7:
                    value = r.ReadChar();
                    break;

                case 8:
                    value = r.ReadSingle();
                    break;

                case 9:
                    value = r.ReadDouble();
                    break;

                case 10:
                    value = r.ReadInt16();
                    break;

                case 11:
                    value = r.ReadInt64();
                    break;

                case 12:
                    value = r.ReadUInt16();
                    break;

                case 13:
                    value = r.ReadUInt32();
                    break;

                case 14:
                    value = r.ReadUInt64();
                    break;
            }

            return value;
        }

        internal bool Dirty
        {
            get { return _dirty; }
            set { _dirty = value; }
        }

        internal object this[string s]
        {
            get { return BaseGet(s); }
            set
            {
                BaseSet(s, value);
                _dirty = true;
            }
        }

        public object this[int index]
        {
            get { return BaseGet(index); }
            set
            {
                BaseSet(index, value);
                _dirty = true;
            }
        }
    }
}