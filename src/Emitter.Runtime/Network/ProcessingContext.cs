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
using System.IO;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using Emitter.Collections;
using Emitter.Diagnostics;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a processing context for an individual incoming or outgoing packet.
    /// </summary>
    public sealed class ProcessingContext : RecyclableObject
    {
        #region Constructor

        // Holds reusable contexts
        private static ConcurrentPool<ProcessingContext> Pool =
            new ConcurrentPool<ProcessingContext>("Processing Contexts", (c) => new ProcessingContext());

        // Fields
        private ConcurrentQueue<BufferSegment> LaterQueue = new ConcurrentQueue<BufferSegment>();

        private ProcessingSettings Settings;

        private readonly Processor[] Processors = new Processor[6];
        private int Current = 0;
        private int Count = 0;

        // Shortcuts for speedup
        private ConcurrentQueue<BufferSegment> Segments;

        private BufferProvider BufferProvider;
        private object ReceiveLock;

        /// <summary>
        /// Constructs a new <see cref="ProcessingContext"/>.
        /// </summary>
        private ProcessingContext() { }

        /// <summary>
        /// Acquires a new processing context for the given channel and settings.
        /// </summary>
        /// <param name="channel">The channel to acquire the context for.</param>
        /// <param name="settings">The type of the context.</param>
        /// <returns>An acquired <see cref="ProcessingContext"/>.</returns>
        internal static ProcessingContext Acquire(Emitter.Connection channel, ProcessingSettings settings)
        {
            // Acquire a new procesing context
            var context = Pool.Acquire();

            // Fills a new pipeline;
            var processors = settings.Processors;
            for (int i = 0; i < processors.Length; ++i)
                context.Processors[i] = processors[i];
            context.Count = processors.Length;

            // Fills the fields accordingly
            context.Channel = channel;
            context.Settings = settings;
            context.Segments = settings.Segments;
            context.BufferProvider = settings.BufferProvider;
            context.Session = null;
            context.Buffer = null;
            context.ReceiveLock = settings.Segments;

            return context;
        }

        /// <summary>
        /// Recycles the context.
        /// </summary>
        public override void Recycle()
        {
            // Reset properties
            this.Settings = null;
            this.Channel = null;
            this.Buffer = null;
            this.Segments = null;
            this.BufferProvider = null;
            this.Current = 0;
            this.Count = 0;
            this.ReceiveLock = null;

            // If we have a session object set
            if (this.Session != null)
            {
                // Attempt to release or dispose the session
                if (this.Session is RecyclableObject)
                    ((RecyclableObject)this.Session).TryRelease();
                else if (this.Session is IDisposable)
                    ((IDisposable)this.Session).Dispose();
            }

            // Reset the session
            this.Session = null;

            // If we have a packet set, attempt to release it
            if (this.Packet != null && this.Packet.Lifetime == PacketLifetime.Automatic && this.Packet.IsPooled)
            {
                this.Packet.TryRelease();
                this.Packet = null;
            }
        }

        /// <summary>
        /// Gets the identifier of the context, for debugging.
        /// </summary>
        /// <returns></returns>
        private string GetIdentity()
        {
            return "Ctx #" + this.GetHashCode();
        }

        #endregion Constructor

        #region Public Properties

        /// <summary>
        /// Gets the sender or recipient channel for the packet or buffer.
        /// </summary>
        public Connection Channel
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets the sender or recipient client for the packet or buffer.
        /// </summary>
        public IClient Client
        {
            get { return this.Channel.Client; }
        }

        /// <summary>
        /// Gets or sets the buffer segment.
        /// </summary>
        public BufferSegment Buffer
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets or sets the incoming or outgoing packet.
        /// </summary>
        public Packet Packet
        {
            get;
            set;
        }

        /// <summary>
        /// Get the maximum memory that can be used for encoding/decoding.
        /// </summary>
        public int MaxMemory
        {
            get { return this.Settings.BufferProvider.MaxMemory; }
        }

        #endregion Public Properties

        #region Session Members

        /// <summary>
        /// Gets or sets a session object for storing various intermediate parsing state.
        /// </summary>
        public object Session
        {
            get;
            set;
        }

        /// <summary>
        /// Attempts to cast the session object to the specified type. Returns null if the session object is not
        /// present or unable to cast.
        /// </summary>
        /// <typeparam name="T">The type to cast to.</typeparam>
        /// <returns>The strongly-typed session object or null.</returns>
        public T GetSession<T>() where T : class
        {
            if (this.Session == null)
                return null;

            return this.Session as T;
        }

        #endregion Session Members

        #region Buffer Members

        /// <summary>
        /// Writes a memory stream to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="stream">The memory stream to write to this buffer.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment BufferWrite(MemoryStream stream)
        {
            lock (this.ReceiveLock)
            {
                return this.BufferProvider.Write(stream);
            }
        }

        /// <summary>
        /// Writes a buffer segment to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="segment">The segment to write to this buffer.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment BufferWrite(ArraySegment<byte> segment)
        {
            lock (this.ReceiveLock)
            {
                return this.BufferProvider.Write(segment.Array, segment.Offset, segment.Count);
            }
        }

        /// <summary>
        /// Writes a buffer segment to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="segment">The segment to write to this buffer.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment BufferWrite(BufferSegment segment)
        {
            lock (this.ReceiveLock)
            {
                return this.BufferProvider.Write(segment.Array, segment.Offset, segment.Length);
            }
        }

        /// <summary>
        /// Writes a memory stream to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="stream">The memory stream to write to this buffer.</param>
        /// <param name="offset">The starting offset in the byte array.</param>
        /// <param name="length">The amount of bytes to write.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment BufferWrite(MemoryStream stream, int offset, int length)
        {
            lock (this.ReceiveLock)
            {
                return this.BufferProvider.Write(stream, offset, length);
            }
        }

        /// <summary>
        /// Writes the array of bytes to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="buffer">The array of bytes to write to this buffer.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment BufferWrite(params byte[] buffer)
        {
            lock (this.ReceiveLock)
            {
                return this.BufferProvider.Write(buffer);
            }
        }

        /// <summary>
        /// Writes the array of bytes to the buffer and returns a segment which can be used for various operations.
        /// </summary>
        /// <param name="buffer">The array of bytes to write to this buffer.</param>
        /// <param name="offset">The starting offset in the byte array.</param>
        /// <param name="length">The amount of bytes to write.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment BufferWrite(byte[] buffer, int offset, int length)
        {
            lock (this.ReceiveLock)
            {
                return this.BufferProvider.Write(buffer, offset, length);
            }
        }

        /// <summary>
        /// Reserves a specific segment which can be used for various operations.
        /// </summary>
        /// <param name="length">The amount of bytes to reserve.</param>
        /// <returns>A delimited buffer segment.</returns>
        public BufferSegment BufferReserve(int length)
        {
            lock (this.ReceiveLock)
            {
                return this.BufferProvider.Reserve(length);
            }
        }

        #endregion Buffer Members

        #region Pipeline Members

        /// <summary>
        /// Clears everything in the current pipeline and enqueues a next packet processor a
        /// fter the current processor have been executed.
        /// </summary>
        /// <param name="processors">The packet processors to insert.</param>
        public void Redirect(params Processor[] processors)
        {
            lock (this.ReceiveLock)
            {
                // Clear first
                this.Current = 0;
                this.Count = 0;

                // Add after the current
                for (int i = 0; i < processors.Length; ++i)
                    this.Processors[i] = processors[i];
                this.Count = processors.Length;
            }
        }

        #endregion Pipeline Members

        #region Throttle Members

        /// <summary>
        /// Swaps the buffer and frees the memory used by the current one.
        /// </summary>
        /// <param name="newBuffer">New buffer to put in the Buffer property.</param>
        public void SwitchBuffer(BufferSegment newBuffer)
        {
            lock (this.ReceiveLock)
            {
                if (this.Buffer != null)
                    this.Buffer.TryRelease();

                this.Buffer = newBuffer;
            }
        }

        /// <summary>
        /// Throttles a buffer segment. It will be processed later.
        /// </summary>
        /// <param name="segment">The segment to process later.</param>
        public unsafe void Throttle(BufferSegment segment)
        {
            lock (this.ReceiveLock)
            {
                // Put in the working queue.
                if (segment != null)
                {
                    this.Segments.Enqueue(segment);
                }
            }
        }

        /// <summary>
        /// Throttles a buffer segment by creating a new segment from the current buffer. It will
        /// be processed later.
        /// </summary>
        /// <param name="offset">The offset specifying where to perform a buffer split.</param>
        public unsafe void Throttle(int offset)
        {
            lock (this.ReceiveLock)
            {
                // First, check if there's something to actually throttle
                if (this.Buffer.Length <= offset)
                    return;

                // Get the subsegment
                var segment = this.Buffer.Split(offset);

                // There might be several packets in the same segment. We need to specify
                // that one is decoded and forward only that one to the next decoder.
                // However, we must not discard the segment completely as we might loose data!

                // Put in the working queue.
                this.Segments.Enqueue(segment);
            }
        }

        /// <summary>
        /// Similar to throttle, simply execute this in the next iteration.
        /// </summary>
        /// <param name="offset"></param>
        public void Later(int offset)
        {
            lock (this.ReceiveLock)
            {
                // First, check if there's something to actually throttle
                if (this.Buffer.Length <= offset)
                    return;

                // Get the subsegment
                var segment = this.Buffer.Split(offset);

                // Put in the waiting queue.
                this.LaterQueue.Enqueue(segment);
            }
        }

        /// <summary>
        /// Similar to throttle, simply execute this in the next iteration.
        /// </summary>
        /// <param name="segment">The segment to process later.</param>
        public void Later(BufferSegment segment)
        {
            lock (this.ReceiveLock)
            {
                // Put in the waiting queue.
                this.LaterQueue.Enqueue(segment);
            }
        }

        #endregion Throttle Members

        #region OnReceive

        /// <summary>
        /// Processes the incoming data in the context.
        /// </summary>
        ///<param name="incoming">The socket input to read from</param>
        internal Task OnReceive(BufferSegment incoming)
        {
            // Process the receive
            Monitor.Enter(this.ReceiveLock);
            try
            {
                // Increment incoming bytes
                NetStat.BytesIncoming.Add(incoming.Length);

                // If we have previously throttled segments, glue the last two of them
                // together. Only two will suffice as it has cumulative effect.
                bool throttledElement = !this.Segments.IsEmpty;

                // If we weren't able to decrypt, continue at the next receive()
                if (incoming == null)
                    return Task.CompletedTask;

                // We have now a decrypted buffer, enqueue it to the processor queue
                this.Segments.Enqueue(incoming);
                byte start = incoming.Array[incoming.Offset];

                // While we can dequeue, process each.
                while (this.Segments.Count > 0)
                {
                    // If there was some data throttled previously and only two elements left,
                    // glue them together.
                    if (throttledElement && this.Segments.Count == 2)
                    {
                        BufferSegment leftSegment;
                        BufferSegment rightSegment;

                        this.Segments.TryDequeue(out leftSegment);
                        this.Segments.TryDequeue(out rightSegment);
                        this.Buffer = leftSegment.Join(rightSegment);

                        leftSegment.TryRelease();
                        rightSegment.TryRelease();
                    }
                    else
                    {
                        // Dequeue just one segment and process it
                        BufferSegment segment;
                        this.Segments.TryDequeue(out segment);

                        // Set the buffer to the dequeued segment
                        this.Buffer = segment;
                    }

                    // Check the contents
                    //Console.WriteLine("[{0}] {1} ({2})", this.Channel.Handle, this.Buffer.ViewAsSsl, this.Channel.IsSecure);

                    NetStat.PacketsIncoming.Increment();
                    this.Current = 0;

                    // While we have a worker and the connection buffer is still alive
                    while (this.Current < this.Count && !this.BufferProvider.IsDisposed)
                    {
                        // Get the processor and schedule the next one. The 'current' might be adjusted during
                        // the process if Redirect() is called.
                        var process = this.Processors[this.Current++];

                        // Execute the process
                        var result = process(this.Channel, this);
                        if (result == ProcessingState.InsufficientData)
                        {
                            // Insufficient data was returned, we need to get the current buffer and keep
                            // it for later. When new data arrives, we need to glue them together in one
                            // buffer segment. Only two will suffice as it has cumulative effect.
                            Throttle(this.Buffer);
                            return Task.CompletedTask;
                        }
                        else if (result == ProcessingState.Stop)
                        {
                            // The processor have told us to stop processing, we must free the buffer.
                            if (this.Buffer != null)
                                this.Buffer.TryRelease();
                            break;
                        }
                        else if (result == ProcessingState.HandleLater)
                        {
                            // This particular buffer needs to be handled later, we need to requeue it back
                            // on the Segments queue.
                            if (this.Buffer != null)
                                this.LaterQueue.Enqueue(this.Buffer);
                            break;
                        }
                    }
                }
            }
#if DEBUG
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
#endif
            finally
            {
                // If we have some segments we have scheduled for later handling, enqueue them now
                BufferSegment laterSegment;
                while (this.LaterQueue.TryDequeue(out laterSegment))
                    this.Segments.Enqueue(laterSegment);

                // Unlock
                Monitor.Exit(this.ReceiveLock);
            }

            return Task.CompletedTask;
        }

        #endregion OnReceive

        #region OnSend

        /// <summary>
        /// Processes outgoing data.
        /// </summary>
        /// <param name="packet">The packet to send to the remote endpoint.</param>
        /// <returns>The buffer segment to send to the remote client.</returns>
        internal unsafe BufferSegment OnSend(Packet packet)
        {
            // Increment packet counter
            NetStat.PacketsOutgoing.Increment();

            // Set the packet to send
            this.Packet = packet;
            this.Current = 0;

            // While we have a worker
            while (this.Current < this.Count)
            {
                // Get the processor and schedule the next one. The 'current' might be adjusted during
                // the process if Redirect() is called.
                var process = this.Processors[this.Current++];

                // Execute the process
                switch (process(this.Channel, this))
                {
                    case ProcessingState.Stop:
                        return this.Buffer;

                    case ProcessingState.InsufficientData:
                        throw new InvalidOperationException("An encoding processor can not return InsufficientData return code.");
                }
            }

            return this.Buffer;
        }

        #endregion OnSend
    }
}