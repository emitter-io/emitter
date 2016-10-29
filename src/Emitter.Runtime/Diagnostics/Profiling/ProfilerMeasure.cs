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
using System.Diagnostics;
using System.Threading;

#if UTILS
using Newtonsoft.Json;
#else

using Emitter.Text.Json;

#endif

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a profiling sample.
    /// </summary>
    public sealed class ProfilerMeasure : IDisposable
    {
        /// <summary>
        /// Creates the sampler instance.
        /// </summary>
        /// <param name="name">The name</param>
        public ProfilerMeasure(string name, string host)
        {
            this.Name = name;
            this.Host = host;
            this.Count = 0;
            this.Average = 0;
            this.Time = 0;
        }

        /// <summary>
        /// The stopwatch used for measurements
        /// </summary>
        [ThreadStatic]
        private static Stopwatch Watch;

        /// <summary>
        /// The hit count of the measure.
        /// </summary>
        [JsonProperty("hit")]
        public volatile int Count;

        /// <summary>
        /// The name of the method.
        /// </summary>
        [JsonProperty("name")]
        public string Name;

        /// <summary>
        /// The average of time (in microseconds)
        /// </summary>
        [JsonProperty("avg")]
        public double Average;

        /// <summary>
        /// The total time, in milliseconds.
        /// </summary>
        [JsonProperty("time")]
        public double Time;

        /// <summary>
        /// The host address.
        /// </summary>
        [JsonProperty("host")]
        public string Host;

        #region Measurement Members

        /// <summary>
        /// Restarts the stopwatch.
        /// </summary>
        internal void Restart()
        {
            // Have one watch per thread
            if (Watch == null)
                Watch = new Stopwatch();

            // Only measure every 10th call to avoid high performance impact
            if (this.Count < 100)
                Watch.Restart();
            else if ((this.Count % (int)(this.Count / 100)) == 0)
                Watch.Restart();
        }

        /// <summary>
        /// Stop measuring.
        /// </summary>
        public void Stop()
        {
            try
            {
                // Increment hit count and check if we are measuring
                var n = Interlocked.Increment(ref Count);

                // We're done mesuring
                if (!Watch.IsRunning)
                {
                    // Estimate total time
                    this.Time += (this.Average / 1000);
                    return;
                }
                Watch.Stop();

                // If we don't have a 100, compute over n
                const int window = 100;
                if (n > window) n = window;

                // Compute a moving average over the last 100 elements
                var value = Watch.Elapsed.TotalMilliseconds;
                var oldavg = this.Average;
                this.Average = ((n - 1) * oldavg + (value * 1000)) / n;
                this.Time += value;
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex.Message);
            }
        }

        /// <summary>
        /// On dispose we should measure the time.
        /// </summary>
        public void Dispose()
        {
            this.Stop();
        }

        #endregion Measurement Members
    }
}