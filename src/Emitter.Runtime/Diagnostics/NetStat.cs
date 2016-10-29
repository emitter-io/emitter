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
using System.Linq;
using Emitter.Collections;
using Emitter.Threading;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Provides global statistics and monitoring
    /// </summary>
    public static class NetStat
    {
        #region Monitoring hooks
        private static PerSecondTimer InternalTimer = new PerSecondTimer();
        private static AtomicInt64 LastCpuUsageTime = 0;
        private static AtomicInt64 LastCpuLogTime = System.DateTime.Now.Ticks;
        private static AtomicInt64 Tick = 0;

        /// <summary>
        /// Automatically called by Emitter to initialize the monitoring.
        /// </summary>
        [InvokeAt(InvokeAtType.Initialize)]
        public static void Initialize()
        {
            NetStat.InternalTimer.Start();
        }

        /// <summary>
        /// Automatically called by Emitter to stop the monitoring.
        /// </summary>
        [InvokeAt(InvokeAtType.Terminate)]
        public static void Terminate()
        {
            NetStat.InternalTimer.Stop();
        }

        #endregion Monitoring hooks

        /// <summary>
        /// Gets the name of the process
        /// </summary>
        public static readonly string ProcessName = Process.GetCurrentProcess().ProcessName;

        /// <summary>
        /// Gets current process information
        /// </summary>
        public static readonly Process ProcessInfo = Process.GetCurrentProcess();

        /// <summary>
        /// Gets the date-time of the service startup
        /// </summary>
        public static readonly DateTime StartupDate = DateTime.Now;

        /// <summary>
        /// Gets the total amount of incoming packets since the service startup
        /// </summary>
        public static AtomicInt64 PacketsIncoming = 0;

        /// <summary>
        /// Gets the total amount of outgoing packets since the service startup
        /// </summary>
        public static AtomicInt64 PacketsOutgoing = 0;

        /// <summary>
        /// Gets the total amount of incoming number of bytes since the service startup
        /// </summary>
        public static AtomicInt64 BytesIncoming = 0;

        /// <summary>
        /// Gets the total amount of outgoing number of bytes since the service startup
        /// </summary>
        public static AtomicInt64 BytesOutgoing = 0;

        /// <summary>
        /// Gets the average amount of incoming packets per second
        /// </summary>
        public static AtomicDouble AveragePPSIncoming = 0;

        /// <summary>
        /// Gets the average amount of outgoing packets per second
        /// </summary>
        public static AtomicDouble AveragePPSOutgoing = 0;

        /// <summary>
        /// Gets the average amount of incoming bytes per second
        /// </summary>
        public static AtomicDouble AverageBPSIncoming = 0;

        /// <summary>
        /// Gets the average amount of outgoing bytes per second
        /// </summary>
        public static AtomicDouble AverageBPSOutgoing = 0;

        /// <summary>
        /// Gets the average working set.
        /// </summary>
        public static AtomicDouble AverageWorkingSet = 0;

        /// <summary>
        /// Gets the average global cpu usage.
        /// </summary>
        public static AtomicDouble AverageCpuUsage = 0;

        /// <summary>
        /// Gets the total amount of currently used byte arrays
        /// </summary>
        public static AtomicInt64 UsedByteArrays = 0;

        /// <summary>
        /// Gets the average compression ratio.
        /// </summary>
        public static PerfAverage Compression = new PerfAverage();

        /// <summary>
        /// Gets the memory pools registry used by this server.
        /// </summary>
        public static readonly WeakReferenceList<IRecycler> MemoryPools = new WeakReferenceList<IRecycler>();

        #region Calculated Properties

        /// <summary>
        /// Gets the current uptime of the service.
        /// </summary>
        public static TimeSpan Uptime
        {
            get { return DateTime.Now - NetStat.StartupDate; }
        }

        /// <summary>
        /// Gets the file name of the current process.
        /// </summary>
        public static string ProcessFileName
        {
            get
            {
                if (ProcessInfo == null || ProcessInfo.Modules.Count == 0)
                    return null;
                return ProcessInfo.Modules[0].FileName;
            }
        }

        #endregion Calculated Properties

        private sealed class PerSecondTimer : Timer
        {
            private PerfCrossSection PPSIncomingSampler;
            private PerfCrossSection PPSOutgoingSampler;
            private PerfCrossSection BPSIncomingSampler;
            private PerfCrossSection BPSOutgoingSampler;
            private PerfCrossSection WorkingSetSampler;
            private PerfCrossSection CpuUsageSampler;

            public PerSecondTimer() : base(TimeSpan.FromSeconds(1), TimeSpan.FromSeconds(1))
            {
                PPSIncomingSampler = new PerfCrossSection(5);
                PPSOutgoingSampler = new PerfCrossSection(5);
                BPSIncomingSampler = new PerfCrossSection(5);
                BPSOutgoingSampler = new PerfCrossSection(5);
                WorkingSetSampler = new PerfCrossSection(5);
                CpuUsageSampler = new PerfCrossSection(5);
            }

            protected override void OnTick()
            {
                try
                {
                    PPSIncomingSampler.CumulativeDelta((double)PacketsIncoming);
                    PPSOutgoingSampler.CumulativeDelta((double)PacketsOutgoing);
                    BPSIncomingSampler.CumulativeDelta((double)BytesIncoming);
                    BPSOutgoingSampler.CumulativeDelta((double)BytesOutgoing);
                    AveragePPSIncoming = PPSIncomingSampler.Average();
                    AveragePPSOutgoing = PPSOutgoingSampler.Average();
                    AverageBPSIncoming = BPSIncomingSampler.Average();
                    AverageBPSOutgoing = BPSOutgoingSampler.Average();

                    // Sample working set (using GC for Mono)
                    WorkingSetSampler.Sample(GC.GetTotalMemory(false));
                    AverageWorkingSet = WorkingSetSampler.Average();

                    // Sample CPU
                    long currentLogTime = System.DateTime.Now.Ticks;
                    long currentCpuUsageTime = ProcessInfo.TotalProcessorTime.Ticks;
                    long timeDiff = currentLogTime - LastCpuLogTime.Value;
                    if (timeDiff != 0)
                    {
                        CpuUsageSampler.Sample(
                            ((currentCpuUsageTime - LastCpuUsageTime.Value) * 100 / timeDiff) / Service.ProcessorCount
                            );
                        AverageCpuUsage = CpuUsageSampler.Average();
                    }
                    LastCpuLogTime = currentLogTime;
                    LastCpuUsageTime = currentCpuUsageTime;

                    // Collect
                    Compression.Collect();

                    // Increment the tick
                    NetStat.Tick.Increment();
                }
                catch (Exception ex)
                {
                    // Log the exception
                    Service.Logger.Log(ex);
                }
            }
        }
    }
}