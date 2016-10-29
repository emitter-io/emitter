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
using System.Text;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a monitor usage.
    /// </summary>
    public sealed class MonitorUsage
    {
        /// <summary>
        /// Creates a new usage container.
        /// </summary>
        /// <param name="contract"></param>
        /// <returns></returns>
        public MonitorUsage()
        {
            this.IncomingTraffic = new MonitorGauge();
            this.IncomingMessages = new MonitorGauge();
            this.OutgoingTraffic = new MonitorGauge();
            this.OutgoingMessages = new MonitorGauge();
        }

        /// <summary>
        /// The number of bytes for a contract usage.
        /// </summary>
        public MonitorGauge IncomingTraffic;

        /// <summary>
        /// The number of messages for a contract usage.
        /// </summary>
        public MonitorGauge IncomingMessages;

        /// <summary>
        /// The number of bytes for a contract usage.
        /// </summary>
        public MonitorGauge OutgoingTraffic;

        /// <summary>
        /// The number of messages for a contract usage.
        /// </summary>
        public MonitorGauge OutgoingMessages;

        /// <summary>
        /// The number of messages send+received by this contract.
        /// </summary>
        public PerfCrossSection MessageFrequency = new PerfCrossSection(5);

        /// <summary>
        /// Checks whether the usage is empty.
        /// </summary>
        public bool IsZero
        {
            get
            {
                return this.IncomingTraffic.Value == 0
                    && this.OutgoingTraffic.Value == 0
                    && this.IncomingMessages.Value == 0
                    && this.OutgoingMessages.Value == 0;
            }
        }

        /// <summary>
        /// Reset the usage.
        /// </summary>
        public void Reset()
        {
            // Reset everything now
            IncomingMessages.Reset();
            IncomingTraffic.Reset();
            OutgoingMessages.Reset();
            OutgoingTraffic.Reset();
        }

        /// <summary>
        /// Writes the usage to the stream.
        /// </summary>
        public void Write(StringBuilder writer, int contract)
        {
            // Write to a string builder as JSON
            writer.Append("{\"contract\":");
            writer.Append(contract);
            writer.Append(",\"t\":");
            writer.Append(Timer.UtcNow.ToFileTimeUtc());
            writer.Append(",\"m-in\":");
            writer.Append(this.IncomingMessages.Value);
            writer.Append(",\"m-out\":");
            writer.Append(this.OutgoingMessages.Value);
            writer.Append(",\"b-in\":");
            writer.Append(this.IncomingTraffic.Value);
            writer.Append(",\"b-out\":");
            writer.Append(this.OutgoingTraffic.Value);
            writer.Append('}');
        }
    }
}