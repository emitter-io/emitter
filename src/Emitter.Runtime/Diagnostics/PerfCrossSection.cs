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

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a sliding window cross section
    /// </summary>
    public class PerfCrossSection
    {
        private double[] Samples;
        private double PreviousSample = 0;
        private int Index = 0;
        private int Size = 0;

        /// <summary>
        /// Constructs a new instance of <see cref="PerfCrossSection"/>
        /// </summary>
        /// <param name="windowSize">The size of the window for the cross-section. </param>
        public PerfCrossSection(int windowSize)
        {
            Size = windowSize;
            Samples = new double[Size];
        }

        /// <summary>
        /// Gets current (latest) value of the sampler
        /// </summary>
        public double Value
        {
            get { return PreviousSample; }
        }

        /// <summary>
        /// Collects a sample
        /// </summary>
        public void Sample(double sample)
        {
            PreviousSample = Samples[Index] = sample;
            Index++;
            if (Index >= Size)
                Index = 0;
        }

        /// <summary>
        /// Collects a delta sample
        /// </summary>
        public void Delta(double sample)
        {
            double previous = Index == 0 ? Samples[Size - 1] : Samples[Index - 1];
            PreviousSample = Samples[Index] = sample - previous;
            Index++;
            if (Index >= Size)
                Index = 0;
        }

        /// <summary>
        /// Collects a cumulative delta sample, where the base should be substracted
        /// </summary>
        public void CumulativeDelta(double sample)
        {
            Samples[Index] = sample - PreviousSample;
            Index++;
            if (Index >= Size)
                Index = 0;
            PreviousSample = sample;
        }

        /// <summary>
        /// Gets the average of the window
        /// </summary>
        public double Average()
        {
            return Samples.Average();
        }

        /// <summary>
        /// Gets the minimum value of the window
        /// </summary>
        public double Min()
        {
            return Samples.Min();
        }

        /// <summary>
        /// Gets the maximum value of the window
        /// </summary>
        public double Max()
        {
            return Samples.Max();
        }

        /// <summary>
        /// Gets the sum of the values in the window
        /// </summary>
        public double Sum()
        {
            return Samples.Sum();
        }
    }
}