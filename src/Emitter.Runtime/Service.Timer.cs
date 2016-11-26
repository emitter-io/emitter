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
using System.Diagnostics;
using System.IO;
using System.Threading;
using Emitter.Collections;

namespace Emitter
{
    /// <summary>
    /// Represents the priority used to schedule different instances of <see cref="Timer"/> class.
    /// </summary>
    public enum TimerPriority
    {
        /// <summary>
        /// Specifies that the <see cref="Timer"/> should be checked on every tick.
        /// </summary>
        EveryTick,

        /// <summary>
        /// Specifies that the <see cref="Timer"/> should be checked every 10 milliseconds.
        /// </summary>
        TenMilliseconds,

        /// <summary>
        /// Specifies that the <see cref="Timer"/> should be checked every 25 milliseconds.
        /// </summary>
        TwentyFiveMilliseconds,

        /// <summary>
        /// Specifies that the <see cref="Timer"/> should be checked every 50 milliseconds.
        /// </summary>
        FiftyMilliseconds,

        /// <summary>
        /// Specifies that the <see cref="Timer"/> should be checked every 250 milliseconds.
        /// </summary>
        TwoHundredFiftyMilliseconds,

        /// <summary>
        /// Specifies that the <see cref="Timer"/> should be checked every second.
        /// </summary>
        OneSecond,

        /// <summary>
        /// Specifies that the <see cref="Timer"/> should be checked every 5 seconds.
        /// </summary>
        FiveSeconds,

        /// <summary>
        /// Specifies that the <see cref="Timer"/> should be checked every minute.
        /// </summary>
        OneMinute
    }

    /// <summary>
    /// Represents the method that will handle OnTick method of a <see cref="Timer"/> instance.
    /// </summary>
    public delegate void TimerCallback();

    /// <summary>
    /// Represents the method that will handle OnTick method of a <see cref="Timer"/> instance.
    /// </summary>
    /// <param name="state">The state that should be passed to the method.</param>
    public delegate void TimerStateCallback(object state);

    /// <summary>
    /// Represents the method that will handle OnTick method of a <see cref="Timer"/> instance.
    /// </summary>
    /// <typeparam name="T">The type of the state that should be passed to the method.</typeparam>
    /// <param name="state">The state that should be passed to the method.</param>
    public delegate void TimerStateCallback<T>(T state);

    /// <summary>
    /// Provides a mechanism for executing a method at specified intervals. All timers are executed in the same Thread.
    /// </summary>
    public class Timer
    {
        #region Fields

        // Constants
        private const int TimerPriorities = 8;

        // Fields
        private DateTime fNext;

        private TimeSpan fDelay;
        private TimeSpan fInterval;
        private bool fRunning;
        private int fIndex;
        private int fCount;
        private TimerPriority fPriority;
        private ArrayList<Timer> fList;
        private int fListHandle;
        private bool fQueued;

        // Static Fields
        private static ConcurrentQueue<Timer> fQueue = new ConcurrentQueue<Timer>();

        private static int fBreakCount = 20000;
        private static TimeSpan fIdleCycleInterval = TimeSpan.FromMilliseconds(16);
        private static DateTime fMainClock = DateTime.Now;
        private static DateTime fMainClockUtc = DateTime.UtcNow;

        #endregion Fields

        #region Constructors

        /// <summary>
        /// Constructs a new <see cref="Timer"/> instance.
        /// </summary>
        /// <param name="delay">The delay before the start of the timer.</param>
        public Timer(TimeSpan delay) : this(delay, TimeSpan.Zero, 1)
        {
        }

        /// <summary>
        /// Constructs a new <see cref="Timer"/> instance.
        /// </summary>
        /// <param name="delay">The delay before the start of the timer.</param>
        /// <param name="interval">The interval between two ticks.</param>
        public Timer(TimeSpan delay, TimeSpan interval) : this(delay, interval, 0)
        {
        }

        /// <summary>
        /// Constructs a new <see cref="Timer"/> instance.
        /// </summary>
        /// <param name="delay">The delay before the start of the timer.</param>
        /// <param name="interval">The interval between two ticks.</param>
        /// <param name="count">The amount of ticks after which the timer should stop.</param>
        public Timer(TimeSpan delay, TimeSpan interval, int count)
        {
            fDelay = delay;
            fInterval = interval;
            fCount = count;
        }

        /// <summary>
        /// Returns a string representation of the <see cref="Timer"/> instance.
        /// </summary>
        /// <returns>Returns a string representation of the <see cref="Timer"/> instance.</returns>
        public override string ToString()
        {
#if !DOTNET
            if (this is CallTimer)
                return (this as CallTimer).Callback.Method.ToString();
            if (this is StateCallTimer)
                return (this as StateCallTimer).Callback.Method.ToString();
#endif
            return GetType().FullName;
        }

        private static string FormatDelegate(Delegate callback)
        {
            if (callback == null)
                return "null";

#if !DOTNET
            return String.Format("{0}.{1}", callback.Method.DeclaringType.FullName, callback.Method.Name);
#else
            return callback.ToString();
#endif
        }

        #endregion Constructors

        #region Properties

        /// <summary>
        /// Invoked if the timer OnTick() method has thrown an exception.
        /// </summary>
        public event TimerExceptionHandler OnError;

        /// <summary>
        /// Gets or sets the priority of this timer.
        /// </summary>
        public TimerPriority Priority
        {
            get
            {
                return fPriority;
            }
            set
            {
                if (fPriority != value)
                {
                    fPriority = value;

                    if (fRunning)
                        Scheduler.PriorityChange(this, (int)fPriority);
                }
            }
        }

        /// <summary>
        /// Gets the time of next callback execution.
        /// </summary>
        public DateTime Next
        {
            get { return fNext; }
        }

        /// <summary>
        /// Gets or sets the delay of the timer.
        /// </summary>
        public TimeSpan Delay
        {
            get { return fDelay; }
            set { fDelay = value; }
        }

        /// <summary>
        /// Gets or sets the interval of the timer.
        /// </summary>
        public TimeSpan Interval
        {
            get { return fInterval; }
            set { fInterval = value; }
        }

        /// <summary>
        /// Gets or sets whether this timer is running or not.
        /// </summary>
        public bool Running
        {
            get
            {
                return fRunning;
            }
            set
            {
                if (value)
                {
                    Start();
                }
                else
                {
                    Stop();
                }
            }
        }

        /// <summary>
        /// Gets or sets the break count (the maximum amount of timers to be sliced at once)
        /// </summary>
        public static int BreakCount
        {
            get { return fBreakCount; }
            set { fBreakCount = value; }
        }

        /// <summary>
        /// Gets or sets the idle interval on which a new cycle should be forced
        /// </summary>
        public static TimeSpan IdleCycleInterval
        {
            get { return fIdleCycleInterval; }
            set { fIdleCycleInterval = value; }
        }

        /// <summary>
        /// Gets the current time (an approximation)
        /// </summary>
        public static DateTime Now
        {
            get { return fMainClock; }
        }

        /// <summary>
        /// Gets the current universal time (an approximation)
        /// </summary>
        public static DateTime UtcNow
        {
            get { return fMainClockUtc; }
        }

        /// <summary>
        /// Gets the amount of active timers
        /// </summary>
        internal static int Count
        {
            get { return Scheduler.Count; }
        }

        #endregion Properties

        #region Start, Stop and OnTick

        /// <summary>
        /// Starts the timer.
        /// </summary>
        public void Start()
        {
            if (!fRunning)
            {
                fRunning = true;
                Scheduler.AddTimer(this);
            }
        }

        /// <summary>
        /// Stops the timer.
        /// </summary>
        public void Stop()
        {
            if (fRunning)
            {
                fRunning = false;
                Scheduler.RemoveTimer(this);
            }
        }

        /// <summary>
        /// Restarts the timer.
        /// </summary>
        public void Restart()
        {
            this.Stop();
            this.Start();
        }

        /// <summary>
        /// Main timer method that is called on every tick.
        /// </summary>
        protected virtual void OnTick()
        {
        }

        #endregion Start, Stop and OnTick

        #region GetPriority

        /// <summary>
        /// Computes timer priority based on the given interval.
        /// </summary>
        /// <param name="interval">The interval between two ticks.</param>
        /// <returns>The computed priority of the timer.</returns>
        public static TimerPriority GetPriority(TimeSpan interval)
        {
            if (interval >= TimeSpan.FromMinutes(1.0))
                return TimerPriority.FiveSeconds;

            if (interval >= TimeSpan.FromSeconds(10.0))
                return TimerPriority.OneSecond;

            if (interval >= TimeSpan.FromSeconds(5.0))
                return TimerPriority.TwoHundredFiftyMilliseconds;

            if (interval >= TimeSpan.FromSeconds(2.5))
                return TimerPriority.FiftyMilliseconds;

            if (interval >= TimeSpan.FromSeconds(1.0))
                return TimerPriority.TwentyFiveMilliseconds;

            if (interval >= TimeSpan.FromSeconds(0.5))
                return TimerPriority.TenMilliseconds;

            return TimerPriority.EveryTick;
        }

        #endregion GetPriority

        #region DelayCall(..)

        /// <summary>
        /// Invokes the specified callback function after the specified delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the callback.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer DelayCall(TimeSpan delay, TimerCallback callback)
        {
            Timer t = new CallTimer(delay, TimeSpan.Zero, 1, callback);
            t.Priority = GetPriority(delay);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes the specified callback function after the specified delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the callback.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer DelayCall(TimeSpan delay, TimerStateCallback callback, object state)
        {
            Timer t = new StateCallTimer(delay, TimeSpan.Zero, 1, callback, state);
            t.Priority = GetPriority(delay);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes the specified callback function after the specified delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the callback.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer DelayCall<T>(TimeSpan delay, TimerStateCallback<T> callback, T state)
        {
            Timer t = new StateCallTimer<T>(delay, TimeSpan.Zero, 1, callback, state);
            t.Priority = GetPriority(delay);
            t.Start();
            return t;
        }

        #endregion DelayCall(..)

        #region PeriodicCall(..)

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall(TimeSpan interval, TimerCallback callback)
        {
            Timer t = new CallTimer(TimeSpan.Zero, interval, 0, callback);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the first callback.</param>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall(TimeSpan delay, TimeSpan interval, TimerCallback callback)
        {
            Timer t = new CallTimer(delay, interval, 0, callback);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the first callback.</param>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="count">Specifies the maximum amount of calls. (0 = unlimited)</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall(TimeSpan delay, TimeSpan interval, int count, TimerCallback callback)
        {
            Timer t = new CallTimer(delay, interval, 0, callback);
            t.Priority = (count == 1) ? GetPriority(delay) : GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="count">Specifies the maximum amount of calls. (0 = unlimited)</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall(TimeSpan interval, int count, TimerCallback callback)
        {
            Timer t = new CallTimer(TimeSpan.Zero, interval, 0, callback);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall(TimeSpan interval, TimerStateCallback callback, object state)
        {
            Timer t = new StateCallTimer(TimeSpan.Zero, interval, 0, callback, state);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the first callback.</param>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall(TimeSpan delay, TimeSpan interval, TimerStateCallback callback, object state)
        {
            Timer t = new StateCallTimer(delay, interval, 0, callback, state);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the first callback.</param>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="count">Specifies the maximum amount of calls. (0 = unlimited)</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall(TimeSpan delay, TimeSpan interval, int count, TimerStateCallback callback, object state)
        {
            Timer t = new StateCallTimer(delay, interval, 0, callback, state);
            t.Priority = (count == 1) ? GetPriority(delay) : GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="count">Specifies the maximum amount of calls. (0 = unlimited)</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall(TimeSpan interval, int count, TimerStateCallback callback, object state)
        {
            Timer t = new StateCallTimer(TimeSpan.Zero, interval, 0, callback, state);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall<T>(TimeSpan interval, TimerStateCallback<T> callback, T state)
        {
            Timer t = new StateCallTimer<T>(TimeSpan.Zero, interval, 0, callback, state);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the first callback.</param>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall<T>(TimeSpan delay, TimeSpan interval, TimerStateCallback<T> callback, T state)
        {
            Timer t = new StateCallTimer<T>(delay, interval, 0, callback, state);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="delay">Specifies the amount of time to wait before invoking the first callback.</param>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="count">Specifies the maximum amount of calls. (0 = unlimited)</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall<T>(TimeSpan delay, TimeSpan interval, int count, TimerStateCallback<T> callback, T state)
        {
            Timer t = new StateCallTimer<T>(delay, interval, 0, callback, state);
            t.Priority = (count == 1) ? GetPriority(delay) : GetPriority(interval);
            t.Start();
            return t;
        }

        /// <summary>
        /// Invokes repeatedly the specified callback function on the specified interval and/or delay.
        /// </summary>
        /// <param name="interval">Specifies the interval of the repetition.</param>
        /// <param name="count">Specifies the maximum amount of calls. (0 = unlimited)</param>
        /// <param name="callback">Specifies the callback to invoke.</param>
        /// <param name="state">Specifies a state object which will be passed to the callback function.</param>
        /// <returns>Returns the timer object used for this operation.</returns>
        public static Timer PeriodicCall<T>(TimeSpan interval, int count, TimerStateCallback<T> callback, T state)
        {
            Timer t = new StateCallTimer<T>(TimeSpan.Zero, interval, 0, callback, state);
            t.Priority = GetPriority(interval);
            t.Start();
            return t;
        }

        #endregion PeriodicCall(..)

        #region Private - Call Timers

        private class CallTimer : Timer
        {
            private TimerCallback fCallback;

            public TimerCallback Callback { get { return fCallback; } }

            public CallTimer(TimeSpan delay, TimeSpan interval, int count, TimerCallback callback)
                : base(delay, interval, count)
            {
                fCallback = callback;
            }

            protected override void OnTick()
            {
                if (fCallback != null)
                    fCallback();
            }

            public override string ToString()
            {
                return String.Format("Call[{0}]", FormatDelegate(fCallback));
            }
        }

        private class StateCallTimer : Timer
        {
            private TimerStateCallback fCallback;
            private object fState;

            public TimerStateCallback Callback { get { return fCallback; } }

            public StateCallTimer(TimeSpan delay, TimeSpan interval, int count, TimerStateCallback callback, object state)
                : base(delay, interval, count)
            {
                fCallback = callback;
                fState = state;
            }

            protected override void OnTick()
            {
                if (fCallback != null)
                    fCallback(fState);
            }

            public override string ToString()
            {
                return String.Format("StateCall[{0}]", FormatDelegate(fCallback));
            }
        }

        private class StateCallTimer<T> : Timer
        {
            private TimerStateCallback<T> fCallback;
            private T fState;

            public TimerStateCallback<T> Callback { get { return fCallback; } }

            public StateCallTimer(TimeSpan delay, TimeSpan interval, int count, TimerStateCallback<T> callback, T state)
                : base(delay, interval, count)
            {
                fCallback = callback;
                fState = state;
            }

            protected override void OnTick()
            {
                if (fCallback != null)
                    fCallback(fState);
            }

            public override string ToString()
            {
                return String.Format("StateCall[{0}]", FormatDelegate(fCallback));
            }
        }

        #endregion Private - Call Timers

        #region Private - Scheduler

        /// <summary>
        /// Represents a static scheduler thread.
        /// </summary>
        internal static class Scheduler
        {
            private static ConcurrentQueue<TimerChangeEntry> fChangeQueue = new ConcurrentQueue<TimerChangeEntry>();
            private static AutoResetEvent fSignal = new AutoResetEvent(false);

            private static DateTime[] fNextPriorities = new DateTime[TimerPriorities];

            private static TimeSpan[] fPriorityDelays = new TimeSpan[TimerPriorities]
            {
                TimeSpan.Zero,
                TimeSpan.FromMilliseconds( 10.0 ),
                TimeSpan.FromMilliseconds( 25.0 ),
                TimeSpan.FromMilliseconds( 50.0 ),
                TimeSpan.FromMilliseconds( 250.0 ),
                TimeSpan.FromSeconds( 1.0 ),
                TimeSpan.FromSeconds( 5.0 ),
                TimeSpan.FromMinutes( 1.0 )
            };

            private static ArrayList<Timer>[] fTimers = new ArrayList<Timer>[TimerPriorities]
            {
                new ArrayList<Timer>(),
                new ArrayList<Timer>(),
                new ArrayList<Timer>(),
                new ArrayList<Timer>(),
                new ArrayList<Timer>(),
                new ArrayList<Timer>(),
                new ArrayList<Timer>(),
                new ArrayList<Timer>(),
            };

            #region API Methods

            internal static int Count
            {
                get
                {
                    int count = 0;
                    lock (fTimers)
                    {
                        for (int i = 0; i < TimerPriorities; ++i)
                            count += fTimers[i].Count;
                    }
                    return count;
                }
            }

            internal static void DumpInfo(TextWriter tw)
            {
                for (int i = 0; i < TimerPriorities; ++i)
                {
                    tw.WriteLine("Priority: {0}", (TimerPriority)i);
                    tw.WriteLine();

                    Dictionary<string, List<Timer>> hash = new Dictionary<string, List<Timer>>();

                    fTimers[i].ForEach(t =>
                    {
                        string key = t.ToString();

                        List<Timer> list;
                        hash.TryGetValue(key, out list);

                        if (list == null)
                            hash[key] = list = new List<Timer>();

                        list.Add(t);
                    });

                    foreach (KeyValuePair<string, List<Timer>> kv in hash)
                    {
                        string key = kv.Key;
                        List<Timer> list = kv.Value;

                        tw.WriteLine("Type: {0}; Count: {1}; Percent: {2}%", key, list.Count, (int)(100 * (list.Count / (double)fTimers[i].RawCount)));
                    }

                    tw.WriteLine();
                    tw.WriteLine();
                }
            }

            public static void Change(Timer t, int newIndex, bool isAdd)
            {
                fChangeQueue.Enqueue(TimerChangeEntry.GetInstance(t, newIndex, isAdd));
                fSignal.Set();
            }

            public static void AddTimer(Timer t)
            {
                if (t == null)
                    return;

                Change(t, (int)t.Priority, true);
            }

            public static void PriorityChange(Timer t, int newPrio)
            {
                if (t == null)
                    return;

                Change(t, newPrio, false);
            }

            public static void RemoveTimer(Timer t)
            {
                if (t == null)
                    return;

                Change(t, -1, false);
            }

            public static void Set()
            {
                fSignal.Set();
            }

            #endregion API Methods

            #region Scheduler

            /// <summary>
            /// This is the scheduler thread that fills the queues and schedules the tasks.
            /// </summary>
            internal static void Slice()
            {
                fMainClock = DateTime.Now;
                fMainClockUtc = fMainClock.ToUniversalTime();

                ProcessChangeQueue();
                for (int i = 0; i < fTimers.Length; ++i)
                {
                    if (fMainClock < fNextPriorities[i])
                        break;

                    fNextPriorities[i] = fMainClock + fPriorityDelays[i];
                    fTimers[i].ForEach(t =>
                    {
                        if (!t.fQueued && fMainClock > t.fNext)
                        {
                            t.fQueued = true;
                            lock (fQueue)
                                fQueue.Enqueue(t);

                            if (t.fCount != 0 && (++t.fIndex >= t.fCount))
                            {
                                t.Stop();
                            }
                            else
                            {
                                t.fNext = fMainClock + t.fInterval;
                            }
                        }
                    });
                }
            }

            private static void ProcessChangeQueue()
            {
                while (fChangeQueue.Count > 0)
                {
                    try
                    {
                        TimerChangeEntry tce;
                        if (!fChangeQueue.TryDequeue(out tce))
                            continue;
                        if (tce == null)
                            continue;

                        Timer timer = tce.Timer;
                        int newIndex = tce.NewIndex;

                        if (timer.fList != null)
                            timer.fList.Remove(timer.fListHandle);

                        if (tce.IsAdd)
                        {
                            timer.fNext = fMainClock + timer.fDelay;
                            timer.fIndex = 0;
                        }

                        if (newIndex >= 0)
                        {
                            timer.fList = fTimers[newIndex];
                            timer.fListHandle = timer.fList.Add(timer);
                        }
                        else // Remove
                        {
                            timer.fList = null;
                            //timer.OnError = null;
                        }

                        tce.Free();
                    }
                    catch (Exception ex)
                    {
                        // Log the weird exception
                        Service.Logger.Log(ex);
                    }
                }
            }

            #endregion Scheduler

            #region Private Class - TimerChangeEntry

            private class TimerChangeEntry
            {
                public Timer Timer;
                public int NewIndex;
                public bool IsAdd;

                private TimerChangeEntry(Timer t, int newIndex, bool isAdd)
                {
                    Timer = t;
                    NewIndex = newIndex;
                    IsAdd = isAdd;
                }

                public void Free()
                {
                    Timer = null;
                    InstancePool.Enqueue(this);
                }

                private static Queue<TimerChangeEntry> InstancePool = new Queue<TimerChangeEntry>();

                public static TimerChangeEntry GetInstance(Timer t, int newIndex, bool isAdd)
                {
                    TimerChangeEntry e;

                    if (InstancePool.Count > 0)
                    {
                        e = InstancePool.Dequeue();

                        if (e == null)
                            e = new TimerChangeEntry(t, newIndex, isAdd);
                        else
                        {
                            e.Timer = t;
                            e.NewIndex = newIndex;
                            e.IsAdd = isAdd;
                        }
                    }
                    else
                    {
                        e = new TimerChangeEntry(t, newIndex, isAdd);
                    }

                    return e;
                }
            }

            #endregion Private Class - TimerChangeEntry
        }

        #endregion Private - Scheduler

        #region Private - Executor

        /// <summary>
        /// Represents a static executor thread
        /// </summary>
        internal static class Executor
        {
            /// <summary>
            /// Slices the timer thread and calls OnTick() method.
            /// </summary>
            internal static void Run()
            {
                var measureSlice = new Stopwatch();
                var measureTimer = new Stopwatch();
                while (!Service.IsStopping)
                {
                    try
                    {
                        //Removed lock (fQueue) wrapping the entire block, seemed unnecessary since it's a concurrent queue.
                        measureSlice.Start();
                        int index = 0;
                        while (index < fBreakCount && fQueue.Count != 0)
                        {
                            Timer t;
                            if (!fQueue.TryDequeue(out t))
                                return;

                            try
                            {
                                measureTimer.Start();
                                t.OnTick();
                                measureTimer.Stop();
                                if (measureTimer.ElapsedMilliseconds > 50)
                                {
                                    Service.Logger.Log(
                                        LogLevel.Warning,
                                        String.Format("Slow Timer: {0}, {1}ms.", t.ToString(), measureTimer.ElapsedMilliseconds)
                                        );
                                }
                                measureTimer.Reset();
                            }
                            catch (Exception ex)
                            {
                                TimerExceptionEventArgs args = new TimerExceptionEventArgs(t, ex);
                                if (t.OnError != null) // There is a specific handler attached, execute that one
                                    t.OnError(args);

                                if (!args.Handled) // Still not handled, try the global handler
                                    Service.InvokeUnhandledTimerException(args);

                                if (args.StopTimer)
                                    t.Stop();

                                if (!args.Handled)
                                    throw;
                            }

                            t.fQueued = false;
                            ++index;
                        }

                        // Calculate the time we have to sleep
                        measureSlice.Stop();
                        var interval = 1000000 * 33; // 33 millisecond interval intended
                        var sleepDuration = (int)Math.Max((interval - measureSlice.ElapsedTicks) / 1000000, 1);

                        // Sleep the specified time, aiming for predictable flush interval
                        Thread.Sleep(sleepDuration);
                        measureSlice.Reset();
                    }
                    catch (Exception ex)
                    {
                        // Log the exception, but do not stop the thread.
                        Service.Logger.Log(ex);
                    }
                }
            }
        }

        #endregion Private - Executor
    }
}