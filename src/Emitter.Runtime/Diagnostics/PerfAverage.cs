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
    /// Represents a sliding window cross section
    /// </summary>
    public class PerfAverage
    {
        private AtomicDouble Aggregate;
        private AtomicInt32 Count;

        /// <summary>
        /// The current value.
        /// </summary>
        public double Value;

        /// <summary>
        /// Sample a new value.
        /// </summary>
        /// <param name="value">The value to sample.</param>
        public void Sample(double value)
        {
            this.Count.Increment();
            this.Aggregate.Add(value);
        }

        /// <summary>
        /// Collects the average value.
        /// </summary>
        /// <returns></returns>
        public void Collect()
        {
            lock (this)
            {
                var count = this.Count.Value;
                var aggregate = this.Aggregate.Value;
                if (count == 0)
                    return;

                var old = this.Value;
                var current = aggregate / count;

                this.Count = 0;
                this.Aggregate = 0;

                this.Value = (current + old) / 2;
            }
        }
    }
}