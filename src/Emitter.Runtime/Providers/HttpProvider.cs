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
using System.Collections.Concurrent;
using System.IO;
using System.Linq;
using System.Text.RegularExpressions;
using Emitter.Collections;
using Emitter.Network.Http;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider used to handle <see cref="IHttpHandler"/> instances.
    /// </summary>
    public abstract class HttpProvider : Provider
    {
        #region Session Members

        /// <summary>
        /// An event that occurs when a session is created.
        /// </summary>
        public event HttpSessionCreate SessionCreated;

        /// <summary>
        /// An event that occurs when a session expires.
        /// </summary>
        public event HttpSessionExpire SessionExpired;

        /// <summary>
        /// Invokes session created event.
        /// </summary>
        /// <param name="session">The parameter session.</param>
        /// <param name="key">The key of the session.</param>
        protected void InvokeSessionCreate(string key, HttpSession session)
        {
            // Invoke the event
            if (this.SessionCreated != null)
                this.SessionCreated(key, session);
        }

        /// <summary>
        /// Invokes session expired event.
        /// </summary>
        /// <param name="session">The parameter session.</param>
        /// <param name="key">The key of the session.</param>
        protected void InvokeSessionExpire(string key, HttpSession session)
        {
            // Invoke the event
            if (this.SessionExpired != null)
                this.SessionExpired(key, session);
        }

        /// <summary>
        /// Gets the session from the store by the session key.
        /// </summary>
        /// <param name="key">The key/identifier of the session to get.</param>
        /// <returns>The <see cref="HttpSession"/> object retrieved or null if no session was found.</returns>
        public abstract HttpSession GetSession(string key);

        /// <summary>
        /// Puts the session to the cache.
        /// </summary>
        /// <param name="session">The <see cref="HttpSession"/> object to put to the cache.</param>
        public abstract void PutSession(HttpSession session);

        #endregion Session Members

        #region Handler Members

        /// <summary>
        /// Gets the handler to be used for a particular verb and url
        /// </summary>
        public abstract IHttpHandler GetHandler(HttpContext context, HttpVerb verb, String url);

        /// <summary>
        /// Finds a handler instances of the handler
        /// </summary>
        public abstract IHttpHandler FindHandler(Predicate<IHttpHandler> match);

        /// <summary>
        /// Registers a handler to the provider
        /// </summary>
        /// <param name="handler">The handler to register</param>
        public abstract void Register(IHttpHandler handler);

        /// <summary>
        /// Unregisters a handler from the provider
        /// </summary>
        /// <param name="handler">The handler to unregister</param>
        public abstract void Unregister(IHttpHandler handler);

        #endregion Handler Members

        #region Mime Members

        /// <summary>
        /// MimeMap
        /// </summary>
        private HttpMimeMap MimeMap = new HttpMimeMap();

        /// <summary>
        /// Gets the <see cref="HttpMimeMap"/> instance that can be used to map file extensions to
        /// the associated MIME type.
        /// </summary>
        public HttpMimeMap Mime
        {
            get { return MimeMap; }
        }

        #endregion Mime Members

        #region Host Members

        /// <summary>
        /// Helper method that allows hosting a static file.
        /// </summary>
        /// <param name="urlRegex">The url path regular expression that is used to access the file.</param>
        /// <param name="filePath">The path to the file on the local machine.</param>
        /// <returns>A <see cref="IHttpHandler"/> reference to the handler that is used to host.</returns>
        public virtual IHttpHandler Host(Regex urlRegex, string filePath)
        {
            return Host((c) => c.Request.HttpVerb == HttpVerb.Get && urlRegex.IsMatch(c.Request.Path), filePath);
        }

        /// <summary>
        /// Helper method that allows hosting a static file.
        /// </summary>
        /// <param name="urlPath">The url path that is used to access the file.</param>
        /// <param name="filePath">The path to the file on the local machine.</param>
        /// <returns>A <see cref="IHttpHandler"/> reference to the handler that is used to host.</returns>
        public virtual IHttpHandler Host(string urlPath, string filePath)
        {
            return Host((c) => c.Request.HttpVerb == HttpVerb.Get && c.Request.Path == urlPath, filePath);
        }

        /// <summary>
        /// Helper method that allows hosting a static file.
        /// </summary>
        /// <param name="verb">Http verb that is used to get the file.</param>
        /// <param name="urlPath">The url path that is used to access the file.</param>
        /// <param name="filePath">The path to the file on the local machine.</param>
        /// <returns>A <see cref="IHttpHandler"/> reference to the handler that is used to host.</returns>
        public virtual IHttpHandler Host(HttpVerb verb, string urlPath, string filePath)
        {
            return Host((c) => c.Request.HttpVerb == verb && c.Request.Path == urlPath, filePath);
        }

        /// <summary>
        /// Helper method that allows hosting a static file.
        /// </summary>
        /// <param name="condition">The dynamic condition that checks whether the file should be sent to the client or not</param>
        /// <param name="filePath">The path to the file on the local machine.</param>
        /// <returns>A <see cref="IHttpHandler"/> reference to the handler that is used to host.</returns>
        public virtual IHttpHandler Host(Func<HttpContext, bool> condition, string filePath)
        {
            return Host(condition, (context) =>
            {
                HttpRequest request = context.Request;
                HttpResponse response = context.Response;

                var file = new FileInfo(filePath);
                var mime = Mime.GetMime(file.Extension);
                if (!file.Exists || mime == null)
                {
                    response.Status = "404";
                    response.Write("Not found");
                    return;
                }

                var ifModifiedSince = request.Headers.Get("If-Modified-Since");
                if (!String.IsNullOrEmpty(ifModifiedSince))
                {
                    DateTime cached;
                    if (DateTime.TryParse(ifModifiedSince, out cached))
                    {
                        if (cached < file.LastWriteTime)
                        {
                            response.Status = "304";
                            return;
                        }
                    }
                }

                // Write headers
                response.Headers.Set("Date", DateTime.UtcNow.ToString("R"));
                response.Headers.Set("Last-Modified", file.LastWriteTimeUtc.ToString("R"));
                response.ContentType = mime;

                // Write the resource body
                byte[] data = File.ReadAllBytes(file.FullName);
                response.Write(data, 0, data.Length);
            });
        }

        /// <summary>
        /// Helper method that allows registering a handler represented by a delegate.
        /// </summary>
        /// <param name="condition">The dynamic condition that checks whether the handler should be executed or not</param>
        /// <param name="handler">The handler to execute. It should fill HttpResponse in order to reply to the HTTP request.</param>
        /// <returns>A <see cref="IHttpHandler"/> reference to the handler that is used to host.</returns>
        public virtual IHttpHandler Host(Func<HttpContext, bool> condition, Action<HttpContext> handler)
        {
            // Create a new handler
            IHttpHandler h = new QuickHttpHandler(condition, handler);

            // Register the handler
            this.Register(h);

            // And return it so we can hold on to the reference for unregister.
            return h;
        }

        #endregion Host Members
    }

    /// <summary>
    /// Represents a provider used to handle <see cref="IHttpHandler"/> instances.
    /// </summary>
    public sealed class DefaultHttpProvider : HttpProvider
    {
        private ConcurrentList<IHttpHandler> Handlers = new ConcurrentList<IHttpHandler>();
        private ConcurrentDictionary<string, HttpSession> SessionStore = new ConcurrentDictionary<string, HttpSession>();

        /// <summary>
        /// Constructs an instance of a default http provider.
        /// </summary>
        public DefaultHttpProvider()
        {
            // Setup the timer that checks periodically for live sessions and kills them if they've expired
            Timer.PeriodicCall(TimeSpan.FromSeconds(30), OnCheckAlive);
        }

        /// <summary>
        /// Registers a handler to the provider.
        /// </summary>
        /// <param name="handler">The handler to register.</param>
        public override void Register(IHttpHandler handler)
        {
            if (!Handlers.Contains(handler))
                Handlers.Add(handler);
        }

        /// <summary>
        /// Unregisters a handler from the provider.
        /// </summary>
        /// <param name="handler">The handler to unregister.</param>
        public override void Unregister(IHttpHandler handler)
        {
            if (Handlers.Contains(handler))
                Handlers.Remove(handler);
        }

        /// <summary>
        /// Gets the handler to be used for a particular verb and url.
        /// </summary>
        public override IHttpHandler GetHandler(HttpContext context, HttpVerb verb, string url)
        {
            url = url.ToLower();
            foreach (IHttpHandler handler in Handlers)
            {
                if (handler.CanHandle(context, verb, url))
                {
                    return handler;
                }
            }
            return null;
        }

        /// <summary>
        /// Finds a handler instances by the type of the handler.
        /// </summary>
        public override IHttpHandler FindHandler(Predicate<IHttpHandler> match)
        {
            return this.Handlers.Find(match);
        }

        /// <summary>
        /// Gets the session from the store by the session key.
        /// </summary>
        /// <param name="key">The key/identifier of the session to get.</param>
        /// <returns>The <see cref="HttpSession"/> object retrieved or null if no session was found.</returns>
        public override HttpSession GetSession(string key)
        {
            HttpSession session;
            if (this.SessionStore.TryGetValue(key, out session))
                return session;
            return null;
        }

        /// <summary>
        /// Puts the session to the cache.
        /// </summary>
        /// <param name="session">The <see cref="HttpSession"/> object to put to the cache.</param>
        public override void PutSession(HttpSession session)
        {
            var key = session.Key;
            this.SessionStore.AddOrUpdate(key, session, (k, s) =>
                session);

            // Invoke the event
            this.InvokeSessionCreate(key, session);
        }

        /// <summary>
        /// Invoked by a timer that checks periodically for live sessions and kills them if they've expired.
        /// </summary>
        private void OnCheckAlive()
        {
            var now = Timer.Now;
            var keys = SessionStore.Keys;
            foreach (var key in keys)
            {
                HttpSession session;
                if (this.SessionStore.TryGetValue(key, out session))
                {
                    if (session.Expires <= now)
                    {
                        // Try to remove the expired session
                        if (this.SessionStore.TryRemove(key, out session))
                        {
                            // Invoke the event
                            this.InvokeSessionExpire(key, session);
                        }
                    }
                }
            }
        }
    }

    /// <summary>
    /// A default handler used for host providers
    /// </summary>
    internal sealed class QuickHttpHandler : IHttpHandler
    {
        private Func<HttpContext, bool> Condition;
        private Action<HttpContext> Handler;

        public QuickHttpHandler(Func<HttpContext, bool> condition, Action<HttpContext> handler)
        {
            this.Condition = condition;
            this.Handler = handler;
        }

        public bool CanHandle(HttpContext context, HttpVerb verb, string url)
        {
            return Condition(context);
        }

        public void ProcessRequest(HttpContext context)
        {
            Handler(context);
        }
    }
}