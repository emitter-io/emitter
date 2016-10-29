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
using System.Collections;
using System.Collections.Generic;
using System.Diagnostics;
using System.Threading;

namespace Emitter.Collections
{
    /// <summary>
    /// Represents a strongly typed list of objects that can be accessed by index.
    /// Provides methods to search, sort, and manipulate lists.
    ///
    /// Thread-safe and can be accessed by multiple threads on read.
    /// </summary>
    /// <typeparam name="T">The type of elements in the list.</typeparam>
    [DebuggerDisplay("Count = {Count}")]
    public class ConcurrentList<T> : DisposableObject, IList<T>, ICollection<T>, IEnumerable<T>, IList, ICollection, IEnumerable
    {
        #region Constructors
        private readonly List<T> List;

        private readonly ReaderWriterLockSlim Lock = new ReaderWriterLockSlim();

        /// <summary>
        /// Initializes a new instance of the ConcurrentList&lt;T&gt; class that is empty and has the default initial capacity.
        /// </summary>
        public ConcurrentList()
        {
            List = new List<T>();
        }

        /// <summary>
        /// Initializes a new instance of the ConcurrentList&lt;T&gt; class that contains elements copied from the specified
        /// collection and has sufficient capacity to accommodate the number of elements copied.
        /// <param name="collection">The collection whose elements are copied to the new list.</param>
        /// </summary>
        public ConcurrentList(IEnumerable<T> collection)
        {
            List = new List<T>(collection);
        }

        /// <summary>
        /// I Initializes a new instance of the ConcurrentList&lt;T&gt; class that is empty and has the specified
        /// initial capacity.
        /// <param name="capacity">The number of elements that the new list can initially store.</param>
        /// </summary>
        public ConcurrentList(int capacity)
        {
            List = new List<T>(capacity);
        }

        #endregion Constructors

        #region IList<T> Members

        /// <summary>
        /// Gets the number of elements actually contained in the list.
        /// </summary>
        public int Count
        {
            get
            {
                this.Lock.EnterReadLock();
                try
                {
                    return List.Count;
                }
                finally
                {
                    this.Lock.ExitReadLock();
                }
            }
        }

        /// <summary>
        /// Gets or sets the element at the specified index.
        /// </summary>
        /// <param name="index">The zero-based index of the element to get or set.</param>
        /// <returns>The element at the specified index.</returns>
        public T this[int index]
        {
            get
            {
                this.Lock.EnterReadLock();
                try
                {
                    return List[index];
                }
                finally
                {
                    this.Lock.ExitReadLock();
                }
            }
            set
            {
                this.Lock.EnterWriteLock();
                try
                {
                    List[index] = value;
                }
                finally
                {
                    this.Lock.ExitWriteLock();
                }
            }
        }

        /// <summary>
        /// Adds an object to the end of the ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <param name="item">The object to be added to the end of the ConcurrentList&lt;T&gt;.
        /// The value can be null for reference types.</param>
        public void Add(T item)
        {
            this.Lock.EnterWriteLock();
            try
            {
                List.Add(item);
            }
            finally
            {
                this.Lock.ExitWriteLock();
            }
        }

        /// <summary>
        /// Removes all elements from the ConcurrentList&lt;T&gt;.
        /// </summary>
        public void Clear()
        {
            this.Lock.EnterWriteLock();
            try
            {
                List.Clear();
            }
            finally
            {
                this.Lock.ExitWriteLock();
            }
        }

        /// <summary>
        /// Determines whether an element is in the ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <param name="item">The object to locate in the ConcurrentList&lt;T&gt;. The value
        /// can be null for reference types.</param>
        /// <returns>true if item is found in the ConcurrentList&lt;T&gt;; otherwise, false.</returns>
        public bool Contains(T item)
        {
            this.Lock.EnterReadLock();
            try
            {
                return List.Contains(item);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        /// <summary>
        /// Copies the entire ConcurrentList&lt;T&gt;> to a compatible one-dimensional
        /// array, starting at the beginning of the target array.
        /// </summary>
        /// <param name="array">The one-dimensional System.Array that is the destination of the elements
        /// copied from ConcurrentList&lt;T&gt;. The System.Array must have
        /// zero-based indexing.</param>
        public void CopyTo(T[] array)
        {
            this.Lock.EnterReadLock();
            try
            {
                List.CopyTo(array);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        /// <summary>
        /// Copies the entire ConcurrentList&lt;T&gt; to a compatible one-dimensional
        /// array, starting at the specified index of the target array.
        /// </summary>
        /// <param name="array">The one-dimensional System.Array that is the destination of the elements
        /// copied from ConcurrentList&lt;T&gt;. The System.Array must have
        /// zero-based indexing.</param>
        /// <param name="arrayIndex">The zero-based index in array at which copying begins.</param>
        public void CopyTo(T[] array, int arrayIndex)
        {
            this.Lock.EnterReadLock();
            try
            {
                List.CopyTo(array, arrayIndex);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        /// <summary>
        /// Removes the first occurrence of a specific object from the ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <param name="item">The object to remove from the ConcurrentList&lt;T&gt;. The value
        /// can be null for reference types.</param>
        /// <returns>true if item is successfully removed; otherwise, false. This method also
        /// returns false if item was not found in the ConcurrentList&lt;T&gt;.</returns>
        public bool Remove(T item)
        {
            this.Lock.EnterWriteLock();
            try
            {
                return List.Remove(item);
            }
            finally
            {
                this.Lock.ExitWriteLock();
            }
        }

        /// <summary>
        /// Returns an enumerator that iterates through the ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <returns>A thread-safe enumerator.</returns>
        public IEnumerator<T> GetEnumerator()
        {
            return new ConcurrentEnumerator(List, Lock);
        }

        /// <summary>
        /// Searches for the specified object and returns the zero-based index of the
        /// first occurrence within the entire ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <param name="item">The object to locate in the ConcurrentList&lt;T&gt;. The value
        /// can be null for reference types.</param>
        /// <returns>The zero-based index of the first occurrence of item within the entire ConcurrentList&lt;T&gt;,
        /// if found; otherwise, –1.</returns>
        public int IndexOf(T item)
        {
            this.Lock.EnterReadLock();
            try
            {
                return List.IndexOf(item);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        /// <summary>
        /// Inserts an element into the ConcurrentList&lt;T&gt; at the specified index.
        /// </summary>
        /// <param name="index">The zero-based index at which item should be inserted.</param>
        /// <param name="item">The object to insert. The value can be null for reference types.</param>
        public void Insert(int index, T item)
        {
            this.Lock.EnterWriteLock();
            try
            {
                List.Insert(index, item);
            }
            finally
            {
                this.Lock.ExitWriteLock();
            }
        }

        /// <summary>
        /// Removes the element at the specified index of the ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <param name="index">The zero-based index of the element to remove.</param>
        public void RemoveAt(int index)
        {
            this.Lock.EnterWriteLock();
            try
            {
                List.RemoveAt(index);
            }
            finally
            {
                this.Lock.ExitWriteLock();
            }
        }

        /// <summary>
        /// Returns a read-only System.Collections.Generic.IList&lt;T&gt; wrapper for the current collection.
        /// </summary>
        /// <returns>A System.Collections.ObjectModel.ReadOnlyCollection&lt;T&gt; that acts as a read-only
        ///  wrapper around the current ConcurrentList&lt;T&gt;.</returns>
        public ReadOnlyCollection<T> AsReadOnly()
        {
            this.Lock.EnterReadLock();
            try
            {
                return new ReadOnlyCollection<T>(this);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        /// <summary>
        /// Searches for an element that matches the conditions defined by the specified
        /// predicate, and returns the first occurrence within the entire list.
        /// </summary>
        /// <param name="match">
        /// The System.Predicate&lt;T&gt; delegate that defines the conditions of the element
        /// to search for.
        /// </param>
        /// <returns>
        /// The first element that matches the conditions defined by the specified predicate,
        /// if found; otherwise, the default value for type T.
        /// </returns>
        public T Find(Predicate<T> match)
        {
            this.Lock.EnterReadLock();
            try
            {
                return List.Find(match);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        #endregion IList<T> Members

        #region IList Members

        /// <summary>
        /// Returns an enumerator that iterates through a collection.
        /// </summary>
        /// <returns>An <see cref="System.Collections.IEnumerator"/> object that can be used to iterate through the collection.</returns>
        IEnumerator IEnumerable.GetEnumerator()
        {
            return new ConcurrentEnumerator(List, Lock);
        }

        /// <summary>
        /// Gets or sets the element at the specified index.
        /// </summary>
        /// <param name="index">The zero-based index of the element to get or set.</param>
        /// <returns>The element at the specified index.</returns>
        object IList.this[int index]
        {
            get
            {
                this.Lock.EnterReadLock();
                try
                {
                    return List[index];
                }
                finally
                {
                    this.Lock.ExitReadLock();
                }
            }
            set
            {
                this.Lock.EnterWriteLock();
                try
                {
                    List[index] = (T)value;
                }
                finally
                {
                    this.Lock.ExitWriteLock();
                }
            }
        }

        /// <summary>
        /// Gets a value indicating whether the System.Collections.IList is read-only.
        /// </summary>
        bool IList.IsReadOnly
        {
            get { return false; }
        }

        /// <summary>
        /// Adds an item to the System.Collections.IList.
        /// </summary>
        /// <param name="value">The object to add to the System.Collections.IList.</param>
        /// <returns>The position into which the new element was inserted, or -1 to indicate that
        /// the item was not inserted into the collection.</returns>
        int IList.Add(object value)
        {
            this.Lock.EnterWriteLock();
            try
            {
                List.Add((T)value);
                return List.IndexOf((T)value);
            }
            finally
            {
                this.Lock.ExitWriteLock();
            }
        }

        /// <summary>
        /// Determines whether an element is in the ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <param name="value">The object to locate in the ConcurrentList&lt;T&gt;. The value
        /// can be null for reference types.</param>
        /// <returns>true if item is found in the ConcurrentList&lt;T&gt;; otherwise, false.</returns>
        bool IList.Contains(object value)
        {
            this.Lock.EnterReadLock();
            try
            {
                return List.Contains((T)value);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        /// <summary>
        /// Searches for the specified object and returns the zero-based index of the
        /// first occurrence within the entire ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <param name="value">The object to locate in the ConcurrentList&lt;T&gt;. The value
        /// can be null for reference types.</param>
        /// <returns>The zero-based index of the first occurrence of item within the entire ConcurrentList&lt;T&gt;,
        /// if found; otherwise, –1.</returns>
        int IList.IndexOf(object value)
        {
            this.Lock.EnterReadLock();
            try
            {
                return List.IndexOf((T)value);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        /// <summary>
        /// Inserts an element into the ConcurrentList&lt;T&gt; at the specified index.
        /// </summary>
        /// <param name="index">The zero-based index at which item should be inserted.</param>
        /// <param name="value">The object to insert. The value can be null for reference types.</param>
        void IList.Insert(int index, object value)
        {
            this.Lock.EnterWriteLock();
            try
            {
                List.Insert(index, (T)value);
            }
            finally
            {
                this.Lock.ExitWriteLock();
            }
        }

        /// <summary>
        /// Gets a value indicating whether the System.Collections.IList has a fixed size.
        /// </summary>
        bool IList.IsFixedSize
        {
            get { return false; }
        }

        /// <summary>
        /// Removes the first occurrence of a specific object from the ConcurrentList&lt;T&gt;.
        /// </summary>
        /// <param name="value">The object to remove from the ConcurrentList&lt;T&gt;. The value
        /// can be null for reference types.</param>
        /// <returns>true if item is successfully removed; otherwise, false. This method also
        /// returns false if item was not found in the ConcurrentList&lt;T&gt;.</returns>
        void IList.Remove(object value)
        {
            this.Lock.EnterWriteLock();
            try
            {
                List.Remove((T)value);
            }
            finally
            {
                this.Lock.ExitWriteLock();
            }
        }

        #endregion IList Members

        #region ICollection Members

        bool ICollection<T>.IsReadOnly
        {
            get { return false; }
        }

        void ICollection.CopyTo(Array array, int index)
        {
            this.Lock.EnterReadLock();
            try
            {
                List.CopyTo((T[])array, index);
            }
            finally
            {
                this.Lock.ExitReadLock();
            }
        }

        bool ICollection.IsSynchronized
        {
            get { return false; }
        }

        object ICollection.SyncRoot
        {
            get { return null; }
        }

        #endregion ICollection Members

        #region Concurrent Enumerator Implementation

        /// <summary>
        /// A thread-safe IEnumerator implementation.
        /// </summary>
        private struct ConcurrentEnumerator : IEnumerator<T>
        {
            private readonly IEnumerator<T> Enumerator;
            private readonly ReaderWriterLockSlim Lock;

            public ConcurrentEnumerator(List<T> list, ReaderWriterLockSlim @lock)
            {
                // First, acquire the read lock
                @lock.EnterReadLock();

                // Get the enumerator implementation
                Enumerator = list.GetEnumerator();
                Lock = @lock;
            }

            /// <summary>
            /// Gets the element in the collection at the current position of the enumerator.
            /// </summary>
            public T Current
            {
                get { return Enumerator.Current; }
            }

            /// <summary>
            /// Gets the element in the collection at the current position of the enumerator.
            /// </summary>
            object IEnumerator.Current
            {
                get { return Enumerator.Current; }
            }

            /// <summary>
            /// Disposes the enumerator
            /// </summary>
            public void Dispose()
            {
                // This will be called when foreach loop finishes
                this.Lock.ExitReadLock();

                // Dispose the inner as we do not need it anymore
                Enumerator.Dispose();
            }

            /// <summary>
            /// Advances the enumerator to the next element of the collection.
            /// </summary>
            /// <returns>true if the enumerator was successfully advanced to the next element; false
            /// if the enumerator has passed the end of the collection.</returns>
            public bool MoveNext()
            {
                return Enumerator.MoveNext();
            }

            /// <summary>
            /// Sets the enumerator to its initial position, which is before the first element
            /// in the collection.
            /// </summary>
            public void Reset()
            {
                Enumerator.Reset();
            }
        }

        #endregion Concurrent Enumerator Implementation
    }
}