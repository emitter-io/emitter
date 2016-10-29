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
using Emitter.Diagnostics;

namespace Emitter.Collections
{
    /// <summary>
    /// This class represents a strongly-typed generic, concurrrent object pool.
    /// </summary>
    /// <typeparam name="T">The type of the items to manage in the pool.</typeparam>
    public class ConcurrentPool<T> : IRecycler<T>, IDisposable
        where T : class, IRecyclable
    {
        /// <summary>
        /// Creates an instance of T for the given recycler.
        /// </summary>
        /// <param name="recycler">The recycler that creates the instance.</param>
        /// <returns>The instance of T.</returns>
        public delegate T CreateInstanceDelegate(IRecycler recycler);

        private int InstancesInUseCount;
        private string Name = String.Empty;
        private ConcurrentQueue<T> Instances = new ConcurrentQueue<T>();
        private CreateInstanceDelegate Constructor = null;
        private ReleaseInstanceDelegate Releaser;

        #region Constructors

        /// <summary>
        /// Constructs a ConcurrentPool object.
        /// </summary>
        /// <param name="name">The name for the <see cref="ConcurrentPool&lt;T&gt;"/> instance.</param>
        public ConcurrentPool(string name) : this(name, null, 0) { }

        /// <summary>
        /// Constructs a ConcurrentPool object.
        /// </summary>
        /// <param name="name">The name for the ConcurrentPool instance.</param>
        /// <param name="constructor">The <see cref="CreateInstanceDelegate"/> delegate that is used to construct the <see cref="IRecyclable"/> instance.</param>
        public ConcurrentPool(string name, CreateInstanceDelegate constructor) : this(name, constructor, 0) { }

        /// <summary>
        /// Constructs a ConcurrentPool object.
        /// </summary>
        /// <param name="name">The name for the ConcurrentPool instance.</param>
        /// <param name="initialCapacity">Initial pool capacity.</param>
        public ConcurrentPool(string name, int initialCapacity) : this(name, null, initialCapacity) { }

        /// <summary>
        /// Constructs a ConcurrentPool object.
        /// </summary>
        /// <param name="name">The name for the ConcurrentPool instance.</param>
        /// <param name="constructor">The <see cref="CreateInstanceDelegate"/> delegate that is used to construct the <see cref="IRecyclable"/> instance.</param>
        /// <param name="initialCapacity">Initial pool capacity.</param>
        public ConcurrentPool(string name, CreateInstanceDelegate constructor, int initialCapacity)
        {
            this.Name = name;
            this.Constructor = constructor;
            this.Releaser = new ReleaseInstanceDelegate((this as IRecycler).Release);
            if (initialCapacity > 0)
            {
                // Create instances
                for (int i = 0; i < initialCapacity; ++i)
                    Instances.Enqueue(CreateInstance());
            }

            // Diagnostics
            NetStat.MemoryPools.Add(this);
        }

        #endregion Constructors

        #region Properties

        /// <summary>
        /// Gets the overall number of elements managed by this pool.
        /// </summary>
        public int Count
        {
            get { return InstancesInUseCount + this.Instances.Count; }
        }

        /// <summary>
        /// Gets the number of available elements currently contained in the pool.
        /// </summary>
        public int AvailableCount
        {
            get { return this.Instances.Count; }
        }

        /// <summary>
        /// Gets the number of elements currently in use and not available in this pool.
        /// </summary>
        public int InUseCount
        {
            get { return InstancesInUseCount; }
        }

        /// <summary>
        /// Gets a value that indicates whether the available pool is empty.
        /// </summary>
        public bool IsEmpty
        {
            get { return this.Instances.IsEmpty; }
        }

#if DEBUG

        /// <summary>
        /// Get the underlying buffer queue (only in debug)
        /// </summary>
        public ConcurrentQueue<T> Queue
        {
            get { return this.Instances; }
        }

#endif
        #endregion Properties

        #region Virtual Members

        /// <summary>
        /// Allocates a new instance of T.
        /// </summary>
        /// <returns>Allocated instance of T.</returns>
        protected virtual T CreateInstance()
        {
            // If there is no constructor defined, create a new one.
            if (this.Constructor == null)
                this.Constructor = _ => Activator.CreateInstance<T>();

            // Create a new instance
            T instance = Constructor(this);
            instance.Bind(this.Releaser);
            return instance;
        }

        #endregion Virtual Members

        #region IRecycler<T> Members

        /// <summary>
        /// Acquires an instance of a recyclable object.
        /// </summary>
        /// <returns>The acquired instance.</returns>
        public virtual T Acquire()
        {
            // Increment the counter
            Interlocked.Increment(ref InstancesInUseCount);

            // Try to get the instance from the queue
            T instance;
            if (Instances.TryDequeue(out instance))
            {
                instance.OnAcquire();
                return instance;
            }

            // Dequeue failed, create a new instance.
            instance = CreateInstance();
            instance.OnAcquire();
            return instance;
        }

        /// <summary>
        /// Releases an instance of a recyclable object back to the pool.
        /// </summary>
        /// <param name="instance">The instance of IRecyclable to release.</param>
        public virtual void Release(T instance)
        {
            // Recycle the instance first.
            instance.Recycle();

            // Put it back to the pool.
            Instances.Enqueue(instance);

            // Decrement the counter
            Interlocked.Decrement(ref InstancesInUseCount);
        }

        #endregion IRecycler<T> Members

        #region IRecycler Members

        /// <summary>
        /// Acquires an instance of a recyclable object.
        /// </summary>
        /// <returns>The acquired instance.</returns>
        IRecyclable IRecycler.Acquire()
        {
            return this.Acquire();
        }

        /// <summary>
        /// Releases an instance of a recyclable object back to the pool.
        /// </summary>
        /// <param name="instance">The instance of IRecyclable to release.</param>
        void IRecycler.Release(IRecyclable instance)
        {
            this.Release(instance as T);
        }

        #endregion IRecycler Members

        #region IDisposable Members

        /// <summary>
        /// Performs application-defined tasks associated with freeing, releasing, or
        /// resetting unmanaged resources.
        /// </summary>
        public void Dispose()
        {
            Dispose(true);
            GC.SuppressFinalize(this);
        }

        /// <summary>
        /// Releases the unmanaged resources used by the ByteSTream class and optionally releases the managed resources.
        /// </summary>
        /// <param name="disposing">
        /// If set to true, release both managed and unmanaged resources, othewise release only unmanaged resources.
        /// </param>
        protected virtual void Dispose(bool disposing)
        {
            while (this.Instances.Count > 0)
            {
                T instance;
                if (this.Instances.TryDequeue(out instance))
                    instance.Bind(null); // Bind it to null and release to the GC
            }
        }

        /// <summary>
        /// Finalizes this concurrent pool
        /// </summary>
        ~ConcurrentPool()
        {
            Dispose(false);
        }

        #endregion IDisposable Members

        #region Object Members

        /// <summary>
        /// Constructs a user-friendly diagnostics string for this pool.
        /// </summary>
        /// <returns>The diagnostincs string.</returns>
        public override string ToString()
        {
            return this.Name;
        }

        #endregion Object Members
    }
}