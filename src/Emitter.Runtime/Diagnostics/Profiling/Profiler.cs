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
using System.Collections.Generic;
using System.Linq;
using System.Text;
using Emitter.Security;
using Emitter.Text.Json;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a profiler.
    /// </summary>
    public sealed class Profiler : Instrument
    {
        #region Static Members

        /// <summary>
        /// Gets the default profiler.
        /// </summary>
        public readonly static Profiler Default = new Profiler();

        /// <summary>
        /// Starts the default profiler.
        /// </summary>
        [InvokeAt(InvokeAtType.Initialize)]
        public static void Initialize()
        {
            Default.Start();
        }

        #endregion Static Members

        #region Constructors

        /// <summary>
        /// Creates a new profiler.
        /// </summary>
        private Profiler() : base(
            "profile" + EmitterConst.Separator + EmitterStatus.Address + EmitterConst.Separator,
            TimeSpan.FromMilliseconds(100))
        {
        }

        #endregion Constructors

        #region Measurement Members

        /// <summary>
        /// The samplers registry
        /// </summary>
        private readonly ConcurrentDictionary<string, ProfilerMeasure> Registry
            = new ConcurrentDictionary<string, ProfilerMeasure>();

        /// <summary>
        /// M
        /// </summary>
        /// <param name="method"></param>
        /// <returns></returns>
        public ProfilerMeasure Measure(string name)
        {
            // Get the sampler for the measurement
            var key = name;
            var sampler = Registry.GetOrAdd(key, (k) => new ProfilerMeasure(name, EmitterStatus.Address));
            sampler.Restart();
            return sampler;
        }

        /// <summary>
        /// Gets the mesurement message.
        /// </summary>
        protected override void OnExecute()
        {
            try
            {
                // Get the measurements
                var measures = new List<ProfilerMeasure>();
                foreach (var value in Registry.Values)
                {
                    var clone = new ProfilerMeasure(value.Name, value.Host);
                    clone.Count = value.Count;
                    clone.Average = value.Average;
                    clone.Time = value.Time;
                    measures.Add(clone);
                }

                // Serialize
                var message = Encoding.UTF8
                    .GetBytes(JsonConvert.SerializeObject(measures))
                    .AsSegment();
                if (message != default(ArraySegment<byte>))
                {
                    // Publish the message in the cluster, don't need to store it as it might slow everything down
                    Dispatcher.Publish(SecurityLicense.Current.Contract, Info.Target, this.Channel, message, 60);
                }
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
        }

        #endregion Measurement Members
    }
}