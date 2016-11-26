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
using System.Threading;
using Emitter.Collections;
using Emitter.Network.Mesh;

namespace Emitter
{
    /// <summary>
    /// Represents a message queue registry.
    /// </summary>
    public sealed class MessageQueue
    {
        #region Static Members

        /// <summary>
        /// The thread used to flush (send) all the frames.
        /// </summary>
        private static Thread FlusherThread;

        /// <summary>
        /// The next identifier for the message queue.
        /// </summary>
        private static int NextId = 1;

        /// <summary>
        /// Initialize the global flusher.
        /// </summary>
        [InvokeAt(InvokeAtType.Initialize)]
        public static void Initialize()
        {
            // Start the publisher loop
            FlusherThread = new Thread(Flush);
            FlusherThread.Start();
        }

        /// <summary>
        /// This sends all the frames though the message bus.
        /// </summary>
        private static void Flush()
        {
            // This task should run while the service is running
            Console.WriteLine("Messaging: Starting publisher on thread #" + Thread.CurrentThread.ManagedThreadId);
            while (Service.IsRunning)
            {
                try
                {
                    // Go through all the connections
                    foreach (var server in Service.Mesh.Members)
                    {
                        // Get the message queue
                        var mq = server.Session as MessageQueue;
                        if (mq == null || server.State != ServerState.Online)
                            continue;

                        // Acquire a lock on the current frame so we can push it to the pending queue
                        lock (mq.PublishLock)
                        {
                            // If frame is empty, ignore
                            if (mq.CurrentFrame.Length > 0)
                            {
                                // The frame is not full, but push it anyway
                                mq.FrameQueue.Enqueue(mq.CurrentFrame);
                                mq.CurrentFrame = MessageFrame.Acquire();
                            }
                        }

                        MessageFrame frame;
                        while (mq.FrameQueue.TryDequeue(out frame))
                        {
                            try
                            {
                                // Send the buffer
                                server.Send(
                                    MeshFrame.Acquire(frame.AsSegment())
                                    );
                            }
                            catch (Exception ex)
                            {
                                // Catch all exceptions here and print it out
                                Service.Logger.Log(ex);
                            }
                            finally
                            {
                                // Once the frame is sent, release it
                                frame.TryRelease();
                            }
                        }
                    }

                    // Wait for the messages to queue up
                    Thread.Sleep(2);
                }
                catch (Exception ex)
                {
                    // Log the exception
                    Service.Logger.Log(ex);
                }
            }
        }

        #endregion Static Members

        private object PublishLock = new object();
        private volatile MessageFrame CurrentFrame;

        private readonly ConcurrentQueue<MessageFrame> FrameQueue
            = new ConcurrentQueue<MessageFrame>();

        /// <summary>
        /// Gets the secondary hash code for the message queue.
        /// </summary>
        internal readonly int SecondaryHash;

        /// <summary>
        /// Gets the number of pending frames.
        /// </summary>
        public int PendingFrames
        {
            get { return this.FrameQueue.Count; }
        }

        /// <summary>
        /// Constructs a new message queue.
        /// </summary>
        public MessageQueue()
        {
            this.CurrentFrame = MessageFrame.Acquire();
            this.SecondaryHash = Filter<uint>.HashInt32(
                (uint)(Interlocked.Increment(ref NextId))
                );
        }

        /// <summary>
        /// Publishes a message to a remote server.
        /// </summary>
        /// <param name="contract">The contract</param>
        /// <param name="channel">The channel to publish to.</param>
        /// <param name="reply">The reply handler.</param>
        public void Enqueue(int contract, string channel, ArraySegment<byte> message)
        {
            lock (this.PublishLock)
            {
                // Try to append the message directly, and exit once we're done
                //this.MsgSend.Increment();
                while (!this.CurrentFrame.TryAppendMessage(contract, channel, message))
                {
                    // The frame is full, enqueue it to the pending queue and acquire a new one
                    this.FrameQueue.Enqueue(this.CurrentFrame);
                    this.CurrentFrame = MessageFrame.Acquire();
                }
            }
        }
    }
}