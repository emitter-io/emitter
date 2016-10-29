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
using Emitter.Providers;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a processing settings to be applied for an individual incoming or outgoing packet.
    /// </summary>
    public sealed class ProcessingSettings : IDisposable
    {
        #region Constructor

        // A throttle queue
        internal ConcurrentQueue<BufferSegment> Segments =
            new ConcurrentQueue<BufferSegment>();

        // A default pipeline to apply on every processing.
        private LinkedList<Processor> Pipeline =
            new LinkedList<Processor>();

        private Processor[] PipelineCache;
        private object Lock = new object();
        private ProcessingType Type;

        /// <summary>
        /// Constructs a new <see cref="ProcessingContext"/>.
        /// </summary>
        /// <param name="type">The type of the pipeline to set.</param>
        /// <param name="processors">The processors to set.</param>
        internal ProcessingSettings(ProcessingType type, Processor[] processors)
        {
            if (processors == null)
                throw new ArgumentNullException("processors");

            // A buffer provieder per settings
            this.BufferProvider = new BufferProvider();
            this.Type = type;

            // Push the pipeline
            for (int i = 0; i < processors.Length; ++i)
                this.Pipeline.AddLast(processors[i]);

            // Cache the pipeline
            this.PipelineCache = this.Pipeline.ToArray();
        }

        /// <summary>
        /// Constructs a new <see cref="ProcessingContext"/>.
        /// </summary>
        internal ProcessingSettings(params Processor[] processors)
        {
            // A buffer provieder per settings
            this.BufferProvider = new BufferProvider();

            // Push the pipeline
            for (int i = 0; i < processors.Length; ++i)
                this.Pipeline.AddLast(processors[i]);

            // Cache the pipeline
            this.PipelineCache = this.Pipeline.ToArray();
        }

        #endregion Constructor

        #region Public Properties

        /// <summary>
        /// Gets the buffer provider who provided the corresponding buffer.
        /// </summary>
        internal BufferProvider BufferProvider
        {
            get;
            private set;
        }

        /// <summary>
        /// Gets the list of processors to apply for each <see cref="ProcessingContext"/>.
        /// </summary>
        public Processor[] Processors
        {
            get { return this.PipelineCache; }
        }

        /// <summary>
        /// Gets the hash code for the settings.
        /// </summary>
        /// <returns></returns>
        public override int GetHashCode()
        {
            lock (this.Lock)
            {
                int hash = 23;
                for (int i = 0; i < this.PipelineCache.Length; ++i)
                    hash = hash * 37 + this.PipelineCache[i].GetHashCode();
                return hash;
            }
        }

        #endregion Public Properties

        #region Pipeline Members

        /// <summary>
        /// Ensures only one instance of the processor on at the head of the pipeline.
        /// </summary>
        /// <param name="processor">The processor to queue to the default pipeline.</param>
        public void PipelineEnsureFirst(Processor processor)
        {
            if (processor == null)
                return;

            lock (this.Lock)
            {
                // Try to remove
                this.PipelineRemove(processor);

                // To the linked list
                this.Pipeline.AddFirst(processor);

                // Cache the pipeline
                this.PipelineCache = this.Pipeline.ToArray();
            }
        }

        /// <summary>
        /// Adds a processor to the first place in the pipeline.
        /// </summary>
        /// <param name="processor">The processor to queue to the default pipeline.</param>
        public void PipelineAddFirst(Processor processor)
        {
            if (processor == null)
                return;

            lock (this.Lock)
            {
                // To the linked list
                this.Pipeline.AddFirst(processor);

                // Cache the pipeline
                this.PipelineCache = this.Pipeline.ToArray();
            }
        }

        /// <summary>
        /// Adds a processor to the last place in the pipeline.
        /// </summary>
        /// <param name="processor">The processor to queue to the default pipeline.</param>
        public void PipelineAddLast(Processor processor)
        {
            if (processor == null)
                return;

            lock (this.Lock)
            {
                // To the linked list
                this.Pipeline.AddLast(processor);

                // Cache the pipeline
                this.PipelineCache = this.Pipeline.ToArray();
            }
        }

        /// <summary>
        /// Adds a processor to the last place in the pipeline or before the specified one.
        /// </summary>
        /// <param name="processor">The processor to queue to the default pipeline.</param>
        /// <param name="target">The target processor to find.</param>
        public void PipelineAddBeforeOrLast(Processor target, Processor processor)
        {
            if (processor == null)
                return;

            lock (this.Lock)
            {
                // Get the index
                var index = this.Pipeline.FindLast(target);
                if (index != null)
                {
                    // To the linked list
                    this.Pipeline.AddBefore(index, processor);

                    // Cache the pipeline
                    this.PipelineCache = this.Pipeline.ToArray();
                }
                else
                {
                    // Add last
                    this.PipelineAddLast(processor);
                }
            }
        }

        /// <summary>
        /// Adds a processor to the last place in the pipeline or after the specified one.
        /// </summary>
        /// <param name="processor">The processor to queue to the default pipeline.</param>
        /// <param name="target">The target processor to find.</param>
        public void PipelineAddAfterOrLast(Processor target, Processor processor)
        {
            if (processor == null)
                return;

            lock (this.Lock)
            {
                // Get the index
                var index = this.Pipeline.FindLast(target);
                if (index != null)
                {
                    // To the linked list
                    this.Pipeline.AddAfter(index, processor);

                    // Cache the pipeline
                    this.PipelineCache = this.Pipeline.ToArray();
                }
                else
                {
                    // Add last
                    this.PipelineAddLast(processor);
                }
            }
        }

        /// <summary>
        /// Adds a processor to the last place in the pipeline or before the specified one.
        /// </summary>
        /// <param name="processor">The processor to queue to the default pipeline.</param>
        /// <param name="target">The target processor to find.</param>
        public void PipelineAddBeforeOrFirst(Processor target, Processor processor)
        {
            if (processor == null)
                return;

            lock (this.Lock)
            {
                // Get the index
                var index = this.Pipeline.FindLast(target);
                if (index != null)
                {
                    // To the linked list
                    this.Pipeline.AddBefore(index, processor);

                    // Cache the pipeline
                    this.PipelineCache = this.Pipeline.ToArray();
                }
                else
                {
                    // Add last
                    this.PipelineAddFirst(processor);
                }
            }
        }

        /// <summary>
        /// Adds a processor to the last place in the pipeline or after the specified one.
        /// </summary>
        /// <param name="processor">The processor to queue to the default pipeline.</param>
        /// <param name="target">The target processor to find.</param>
        public void PipelineAddAfterOrFirst(Processor target, Processor processor)
        {
            if (processor == null)
                return;

            lock (this.Lock)
            {
                // Get the index
                var index = this.Pipeline.FindLast(target);
                if (index != null)
                {
                    // To the linked list
                    this.Pipeline.AddAfter(index, processor);

                    // Cache the pipeline
                    this.PipelineCache = this.Pipeline.ToArray();
                }
                else
                {
                    // Add last
                    this.PipelineAddFirst(processor);
                }
            }
        }

        /// <summary>
        /// Removes a specified processor from the pipeline.
        /// </summary>
        /// <param name="processor">The processor to remove.</param>
        /// <returns>Whether the processor was successfully removed or not.</returns>
        public bool PipelineRemove(Processor processor)
        {
            if (processor == null)
                return false;

            lock (this.Lock)
            {
                while (this.Pipeline.Remove(processor)) ;
                return true;
            }
        }

        /// <summary>
        /// Gets the first element in the pipeline.
        /// </summary>
        /// <returns>The <see cref="Processor"/> retrieved.</returns>
        public Processor PipelineGetFirst()
        {
            lock (this.Lock)
            {
                return this.Pipeline.First.Value;
            }
        }

        /// <summary>
        /// Gets the last element in the pipeline.
        /// </summary>
        /// <returns>The <see cref="Processor"/> retrieved.</returns>
        public Processor PipelineGetLast()
        {
            lock (this.Lock)
            {
                return this.Pipeline.Last.Value;
            }
        }

        #endregion Pipeline Members

        #region Configuration Members

        /// <summary>
        /// Applies a configuration to the settings.
        /// </summary>
        /// <param name="channel">The channel owns the settings.</param>
        internal void ConfigureFor(Emitter.Connection channel)
        {
            // Configure the maximum memory
            var provider = Service.Providers.Resolve<ClientProvider>();
            if (provider != null)
                this.BufferProvider.MaxMemory = provider.GetMaxMemoryFor(channel, this.Type);
        }

        #endregion Configuration Members

        #region IDisposable Members

        /// <summary>
        /// Called by the GC when the object is finalized.
        /// </summary>
        ~ProcessingSettings()
        {
            // Forward to OnDispose
            this.OnDispose(false);
        }

        /// <summary>
        /// Invoked when the object pool is disposing.
        /// </summary>
        /// <param name="isDisposing">Whether the OnDispose was called by a finalizer or a dispose method</param>
        private void OnDispose(bool isDisposing)
        {
            try
            {
                // Free the buffer provider
                if (this.BufferProvider != null)
                    this.BufferProvider.Dispose();
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
        }

        /// <summary>
        /// Disposes the pinned objects.
        /// </summary>
        public void Dispose()
        {
            this.OnDispose(true);
            GC.SuppressFinalize(this);
        }

        #endregion IDisposable Members
    }
}