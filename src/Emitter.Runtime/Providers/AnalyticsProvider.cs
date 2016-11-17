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
using Emitter.Network.Http;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents an analytics provider for tracking various events.
    /// </summary>
    public abstract class AnalyticsProvider : Provider
    {
        /// <summary>
        /// Tracks an event.
        /// </summary>
        /// <param name="eventCategory">The event category.</param>
        /// <param name="eventName">The name of the event.</param>
        public abstract void Track(string eventCategory, string eventName);
    }

    /// <summary>
    /// Represents an implementation of an analytics provider, where the data is sent
    /// to meta.emitter.io/v1/collect endpoint. This is fed into Google Analytics
    /// for simple statistics gathering.
    /// </summary>
    public class EmitterAnalyticsProvider : AnalyticsProvider
    {
        /// <summary>
        /// The collection endpoint.
        /// </summary>
        private const string Endpoint = "http://meta.emitter.io/v1/collect";

        /// <summary>
        /// Tracks an event.
        /// </summary>
        /// <param name="eventCategory">The event category.</param>
        /// <param name="eventName">The name of the event.</param>
        public override void Track(string eventCategory, string eventName)
        {
            try
            {
                // Get the fingerprint for this machine
                var uid = Service.Providers.Resolve<AddressProvider>().GetFingerprint();
                var url = string.Format("{0}?uid={1}&ec={2}&en={3}", Endpoint, uid, eventCategory, eventName);

                // Fire and forget
                HttpUtility.GetAsync(url, 5000).Forget();
            }
            catch
            {
                // Ignore if the tracking failed.
            }
        }
    }
}