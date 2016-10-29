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
using System.Collections.Generic;
using System.Linq;

namespace Emitter.Collections
{
    /// <summary>
    /// A collection of weak references to objects. Weak references are purged by iteration/count operations, not by add/remove operations.
    /// </summary>
    /// <typeparam name="T">The type of object to hold weak references to.</typeparam>
    /// <remarks>
    /// <para>Since the collection holds weak references to the actual objects, the collection is comprised of both living and dead references. Living references refer to objects that have not been garbage collected, and may be used as normal references. Dead references refer to objects that have been garbage collected.</para>
    /// <para>Dead references do consume resources; each dead reference is a garbage collection handle.</para>
    /// <para>Dead references may be cleaned up by a <see cref="Purge"/> operation. Some properties and methods cause a purge as a side effect; the member documentation specifies whether a purge takes place.</para>
    /// </remarks>
    public sealed class WeakReferenceList<T> : ICollection<T>, IDisposable
        where T : class
    {
        /// <summary>
        /// The actual collection of strongly-typed weak references.
        /// </summary>
        private List<WeakReference<T>> InternalList;

        /// <summary>
        /// The lock for the list
        /// </summary>
        private object Lock;

        /// <summary>
        /// Initializes a new instance of the <see cref="WeakReferenceList&lt;T&gt;"/> class that is empty.
        /// </summary>
        public WeakReferenceList()
        {
            this.InternalList = new List<WeakReference<T>>();
            this.Lock = new object();
        }

        #region ICollection<T> Properties

        /// <summary>
        /// Gets the number of live entries in the collection, causing a purge. O(n).
        /// </summary>
        int ICollection<T>.Count
        {
            get
            {
                lock (Lock)
                {
                    return this.UnsafeLiveList.Count();
                }
            }
        }

        /// <summary>
        /// Gets a value indicating whether the collection is read only. Always returns false.
        /// </summary>
        bool ICollection<T>.IsReadOnly
        {
            get { return false; }
        }

        #endregion ICollection<T> Properties

        /// <summary>
        /// Gets a sequence of live objects from the collection, causing a purge. The entire sequence MUST always be enumerated!
        /// </summary>
        private IEnumerable<T> UnsafeLiveList
        {
            get
            {
                // This implementation uses logic similar to List<T>.RemoveAll, which always has O(n) time.
                //  Some other implementations seen in the wild have O(n*m) time, where m is the number of dead entries.
                //  As m approaches n (e.g., mass object extinctions), their running time approaches O(n^2).
                int writeIndex = 0;
                for (int readIndex = 0; readIndex != this.InternalList.Count; ++readIndex)
                {
                    WeakReference<T> weakReference = this.InternalList[readIndex];
                    T target;
                    if (weakReference.TryGetTarget(out target))
                    {
                        yield return target;

                        if (readIndex != writeIndex)
                        {
                            this.InternalList[writeIndex] = this.InternalList[readIndex];
                        }

                        ++writeIndex;
                    }
                }

                this.InternalList.RemoveRange(writeIndex, this.InternalList.Count - writeIndex);
            }
        }

        /// <summary>
        /// Adds a weak reference to an object to the collection. Does not cause a purge.
        /// </summary>
        /// <param name="item">The object to add a weak reference to.</param>
        public void Add(T item)
        {
            lock (Lock)
            {
                this.InternalList.Add(new WeakReference<T>(item));
            }
        }

        /// <summary>
        /// Removes a weak reference to an object from the collection. Does not cause a purge.
        /// </summary>
        /// <param name="item">The object to remove a weak reference to.</param>
        /// <returns>True if the object was found and removed; false if the object was not found.</returns>
        public bool Remove(T item)
        {
            lock (Lock)
            {
                for (int i = 0; i != this.InternalList.Count; ++i)
                {
                    var weakReference = this.InternalList[i];
                    T target;
                    if (weakReference.TryGetTarget(out target) && target == item)
                    {
                        this.InternalList.RemoveAt(i);
                        return true;
                    }
                }

                return false;
            }
        }

        /// <summary>
        /// Removes all dead objects from the collection.
        /// </summary>
        public void Purge()
        {
            lock (Lock)
            {
                // We cannot simply use List<T>.RemoveAll, because the predicate "x => !x.IsAlive" is not stable.
                IEnumerator<T> enumerator = this.UnsafeLiveList.GetEnumerator();
                while (enumerator.MoveNext())
                {
                }
            }
        }

        /// <summary>
        /// Frees all resources held by the collection.
        /// </summary>
        public void Dispose()
        {
            this.Clear();
        }

        /// <summary>
        /// Empties the collection.
        /// </summary>
        public void Clear()
        {
            lock (Lock)
            {
                this.InternalList.Clear();
            }
        }

        #region ICollection<T> Methods

        /// <summary>
        /// Determines whether the collection contains a specific value.
        /// </summary>
        /// <param name="item">The object to locate.</param>
        /// <returns>True if the collection contains a specific value; false if it does not.</returns>
        bool ICollection<T>.Contains(T item)
        {
            lock (Lock)
            {
                foreach (var reference in this.InternalList)
                {
                    T target;
                    if (reference.TryGetTarget(out target) && target == item)
                        return true;
                }

                return false;
            }
        }

        /// <summary>
        /// Copies all live objects to an array.
        /// </summary>
        /// <param name="array">The destination array.</param>
        /// <param name="arrayIndex">The index to begin writing into the array.</param>
        void ICollection<T>.CopyTo(T[] array, int arrayIndex)
        {
            lock (Lock)
            {
                List<T> ret = new List<T>(this.InternalList.Count);
                ret.AddRange(this.UnsafeLiveList);
                ret.CopyTo(array, arrayIndex);
            }
        }

        #endregion ICollection<T> Methods

        #region IEnumerable<T> Members

        /// <summary>
        /// Gets a sequence of live objects from the collection, causing a purge.
        /// </summary>
        /// <returns>The sequence of live objects.</returns>
        IEnumerator<T> IEnumerable<T>.GetEnumerator()
        {
            lock (Lock)
            {
                var purgedList = new List<T>(this.InternalList.Count);
                purgedList.AddRange(this.UnsafeLiveList);
                return purgedList.GetEnumerator();
            }
        }

        #endregion IEnumerable<T> Members

        #region IEnumerable Members

        /// <summary>
        /// Gets a sequence of live objects from the collection, causing a purge.
        /// </summary>
        /// <returns>The sequence of live objects.</returns>
        System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator()
        {
            return ((IEnumerable<T>)this).GetEnumerator();
        }

        #endregion IEnumerable Members
    }
}