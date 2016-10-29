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
using Emitter.Threading;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents an instrumentation entry point.
    /// </summary>
    public abstract class Instrument : ThreadBase
    {
        protected EmitterChannel Info;

        /// <summary>
        /// Constructs a new instance of a profiler, bound to a particular thread.
        /// </summary>
        /// <param name="channel">The channel for emitter.io.</param>
        /// <param name="interval">The interval to publish measurements info.</param>
        public Instrument(string channel, TimeSpan interval) : base(interval)
        {
            this.Channel = channel;

            // Get the channel
            if (!EmitterChannel.TryParse(this.Channel, false, out Info))
                throw new ArgumentException("Invalid channel");
        }

        /// <summary>
        /// The publishing channel.
        /// </summary>
        public string Channel
        {
            get;
            set;
        }
    }
}