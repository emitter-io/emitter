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
using System.Globalization;
using System.Threading;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a value.
    /// </summary>
    public struct MonitorGauge
    {
        /// <summary>
        /// The invariant culture.
        /// </summary>
        private static readonly CultureInfo Invariant = CultureInfo.InvariantCulture;

        /// <summary>
        /// The value of the Gauge.
        /// </summary>
        public volatile int Value;

        /// <summary>
        /// Increments the value by a specified amount.
        /// </summary>
        /// <param name="value">The amount to increment by.</param>
        public void IncrementBy(int value)
        {
            Interlocked.Add(ref this.Value, value);
        }

        /// <summary>
        /// Increments the value by one.
        /// </summary>
        public void Increment()
        {
            Interlocked.Increment(ref this.Value);
        }

        /// <summary>
        /// Decrements the value by one.
        /// </summary>
        public void Decrement()
        {
            Interlocked.Decrement(ref this.Value);
        }

        /// <summary>
        /// Sets the value to a specified number.
        /// </summary>
        /// <param name="value">The target value.</param>
        public void Set(int value)
        {
            int oldValue;
            int newValue;

            do
            {
                oldValue = this.Value;
                newValue = value;
            } while (Interlocked.CompareExchange(ref this.Value, newValue, oldValue) != oldValue);
        }

        /// <summary>
        /// Resets the value to zero.
        /// </summary>
        public void Reset()
        {
            int oldValue;
            int newValue;

            do
            {
                oldValue = this.Value;
                newValue = 0;
            } while (Interlocked.CompareExchange(ref this.Value, newValue, oldValue) != oldValue);
        }

        /// <summary>
        /// Gets the value in a string format.
        /// </summary>
        /// <returns>The value of the gauge.</returns>
        public override string ToString()
        {
            return Value.ToString(Invariant);
        }
    }
}