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

#if UTILS
using Newtonsoft.Json;
#else

using Emitter.Text.Json;

#endif

namespace Emitter
{
    /// <summary>
    /// Represents the statistics for the emitter.io node, published inside emitter.io itself.
    /// </summary>
    internal partial class EmitterStatus
    {
        [JsonIgnore]
        public DateTime LastUpdate = DateTime.MinValue;

        [JsonProperty("node")]
        public string Node;

        [JsonProperty("host")]
        public string Host;

        [JsonProperty("time")]
        public DateTime Time;

        [JsonProperty("machine")]
        public string Machine;

        [JsonProperty("cpu")]
        public double CPU;

        [JsonProperty("memory")]
        public double Memory;

        [JsonProperty("uptime")]
        public TimeSpan Uptime;

        [JsonProperty("network")]
        public EmitterStatusNetwork Network;
    }

    /// <summary>
    /// Represents the statistics for the emitter.io node, published inside emitter.io itself.
    /// </summary>
    internal class EmitterStatusNetwork
    {
        [JsonProperty("connections")]
        public int Connections;

        [JsonProperty("avg-pps-in")]
        public double AveragePPSIncoming;

        [JsonProperty("avg-pps-out")]
        public double AveragePPSOutgoing;

        [JsonProperty("avg-mps-in")]
        public double AverageMPSIncoming;

        [JsonProperty("avg-mps-out")]
        public double AverageMPSOutgoing;

        [JsonProperty("avg-bps-in")]
        public double AverageBPSIncoming;

        [JsonProperty("avg-bps-out")]
        public double AverageBPSOutgoing;

        [JsonProperty("compression")]
        public double Compression;
    }
}