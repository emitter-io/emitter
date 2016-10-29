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
    internal sealed class ArrayList<T> : IDisposable, IArrayList<T>
        where T : class
    {
        /// <summary>
        /// Number of elements in the array (including freed ones)
        /// </summary>
        public int RawCount;

        /// <summary>
        /// Number of used elements in the array
        /// </summary>
        public int Count;

        /// <summary>
        /// Underlying items array
        /// </summary>
        public T[] Items;

        private int FreeListCount;
        private int[] FreeList;
        private byte[] UsedList;

        public ArrayList()
        {
            this.Items = new T[4];
            this.FreeList = new int[4];
            this.UsedList = new byte[4];
        }

        public ArrayList(int initialCapacity)
        {
            this.Items = new T[initialCapacity];
            this.FreeList = new int[initialCapacity];
            this.UsedList = new byte[initialCapacity];
        }

        /// <summary>
        /// Adds the item into a free space in the Items array and returns the handle to the item
        /// </summary>
        /// <param name="item">Item to add to Items array</param>
        /// <returns>Handle to the added item</returns>
        public int Add(T item)
        {
            if (this.FreeListCount > 0)
            {
                this.FreeListCount--;
                int handle = this.FreeList[this.FreeListCount];
                this.UsedList[handle] = 1;
                this.Count++;

                this.Items[handle] = item;
                return handle;
            }
            else
            {
                int handle = this.RawCount;
                if (this.RawCount < this.Items.Length)
                {
                    this.UsedList[handle] = 1;
                    this.Count++;

                    this.Items[this.RawCount++] = item;
                    return handle;
                }
                else
                {
                    // have to reallocate
                    int newSize = this.RawCount;
                    newSize |= newSize >> 1;
                    newSize |= newSize >> 2;
                    newSize |= newSize >> 4;
                    newSize |= newSize >> 8;
                    newSize |= newSize >> 16;
                    newSize++;

                    Array.Resize(ref this.Items, newSize);
                    Array.Resize(ref this.UsedList, newSize);

                    this.Items[this.RawCount] = item;
                    this.UsedList[handle] = 1;
                    this.Count++;
                    this.RawCount++;
                    return handle;
                }
            }
        }

        /// <summary>
        /// Adds the item into a free space in the Items array and returns the handle to the item
        /// </summary>
        /// <param name="item">Item to add to Items array</param>
        /// <returns>Handle to the added item</returns>
        public int Add(ref T item)
        {
            if (this.FreeListCount > 0)
            {
                this.FreeListCount--;
                int handle = this.FreeList[this.FreeListCount];
                this.UsedList[handle] = 1;
                this.Count++;

                this.Items[handle] = item;
                return handle;
            }
            else
            {
                int handle = this.RawCount;
                if (this.RawCount < this.Items.Length)
                {
                    this.UsedList[handle] = 1;
                    this.Count++;

                    this.Items[this.RawCount++] = item;
                    return handle;
                }
                else
                {
                    // have to reallocate
                    int newSize = this.RawCount;
                    newSize |= newSize >> 1;
                    newSize |= newSize >> 2;
                    newSize |= newSize >> 4;
                    newSize |= newSize >> 8;
                    newSize |= newSize >> 16;
                    newSize++;

                    Array.Resize(ref this.Items, newSize);
                    Array.Resize(ref this.UsedList, newSize);

                    this.Items[this.RawCount] = item;
                    this.UsedList[handle] = 1;
                    this.Count++;
                    this.RawCount++;
                    return handle;
                }
            }
        }

        /// <summary>
        ///  Adds a range of items into free spaces in the Items array
        /// </summary>
        /// <param name="items">Range of items to add</param>
        public void AddRange(T[] items)
        {
            int amountOfItems = items.Length;
            if (amountOfItems < this.FreeListCount)
            {
                // Fill all dirty spaces
                int freePosition;
                for (int i = 1; i <= amountOfItems; ++i)
                {
                    freePosition = this.FreeList[this.FreeListCount - i];
                    this.Items[freePosition] = items[amountOfItems - i];
                    this.UsedList[freePosition] = 1;
                }
                this.FreeListCount -= amountOfItems;
                this.Count += amountOfItems;
            }
            else
            {
                // Can be improved
                for (int i = 0; i < amountOfItems; ++i)
                {
                    Add(items[i]);
                }
            }
        }

        /// <summary>
        /// Executes an action for each item in the collection
        /// Note: Internally uses a for loop.
        /// </summary>
        /// <param name="action">Action to execute</param>
        public void ForEach(Action<T> action)
        {
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1)
                    action(Items[i]);
            }
        }

        /// <summary>
        /// Executes an action for each item in the collection
        /// Note: Internally uses a for loop.
        /// </summary>
        /// <param name="action">Action to execute, passed by reference</param>
        public void ForEach(RefAction<T> action)
        {
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1)
                    action(ref Items[i]);
            }
        }

        /// <summary>
        /// Executes an action for each item in the collection and passes to the action
        /// the handle of the element
        /// </summary>
        /// <param name="action">Action to execute</param>
        public void ForEachWithHandle(Action<T, int> action)
        {
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1)
                    action(Items[i], i);
            }
        }

        /// <summary>
        /// Executes an action for each item in the collection and passes to the action
        /// the handle of the element
        /// </summary>
        /// <param name="action">Action to execute</param>
        public void ForEachWithHandle(RefAction<T, int> action)
        {
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1)
                    action(ref Items[i], i);
            }
        }

        /// <summary>
        /// Executes an action for each item in the collection and passes to the action
        /// the index of the element
        /// </summary>
        /// <param name="action">Action to execute</param>
        public void ForEachWithIndex(Action<T, int> action)
        {
            int index = 0;
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1)
                    action(Items[i], index++);
            }
        }

        /// <summary>
        /// Executes an action for each item in the collection and passes to the action
        /// the index of the element
        /// </summary>
        /// <param name="action">Action to execute</param>
        public void ForEachWithIndex(RefAction<T, int> action)
        {
            int index = 0;
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1)
                    action(ref Items[i], index++);
            }
        }

        /// <summary>
        /// Removes an item from a collection by its reference.
        /// </summary>
        /// <param name="element"></param>
        public void Remove(T element)
        {
            var count = this.RawCount;
            var handle = -1;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1 && this.Items[i].Equals(element))
                {
                    handle = i;
                    break;
                }
            }

            if (handle != -1)
                this.Remove(handle);
        }

        /// <summary>
        /// Removes an item from a collection by its handle
        /// </summary>
        /// <param name="handle">Handle of the item to remove</param>
        public void Remove(int handle)
        {
            if (this.UsedList[handle] == 0)
                return; // the spot is already free

            this.Items[handle] = default(T); // if it's a class, mark as null (GC will collect it)
            this.UsedList[handle] = 0;
            this.Count--;

            if (this.FreeListCount < this.FreeList.Length)
            {
                this.FreeList[this.FreeListCount++] = handle;
                return;
            }

            // have to reallocate
            int newSize = this.FreeListCount;
            newSize |= newSize >> 1;
            newSize |= newSize >> 2;
            newSize |= newSize >> 4;
            newSize |= newSize >> 8;
            newSize |= newSize >> 16;
            newSize++;

            Array.Resize(ref this.FreeList, newSize);
            this.FreeList[this.FreeListCount] = handle;
            this.FreeListCount++;
        }

        /// <summary>
        /// Checks whether there is an element in use at a given position
        /// </summary>
        /// <param name="handle">The handle to check</param>
        /// <returns>True if there is an element in this position, otherwise false</returns>
        public bool HasElementAt(int handle)
        {
            return this.UsedList[handle] == 1;
        }

        /// <summary>
        /// Gets the element at the specifed handle, or the default value if no element was found.
        /// </summary>
        /// <param name="handle">The handle to get the element for.</param>
        /// <returns>The element found or the default value.</returns>
        public T Get(int handle)
        {
            if (this.UsedList[handle] == 1)
                return Items[handle];
            return default(T);
        }

        /// <summary>
        /// Clears the list
        /// </summary>
        public void Clear()
        {
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1)
                {
                    this.Items[i] = default(T);
                    this.UsedList[i] = 0;
                }
            }

            this.RawCount = 0;
            this.FreeListCount = 0;
            this.Count = 0;
        }

        #region IDisposable Members

        /// <summary>
        /// Disposes the object
        /// </summary>
        public void Dispose()
        {
            Clear();
        }

        #endregion IDisposable Members

        #region IEnumerable<T> Members

        public IEnumerator<T> GetEnumerator()
        {
            return new ArrayListCursor(this, 0, ArrayListCursorBehavior.Stop);
        }

        System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator()
        {
            return new ArrayListCursor(this, 0, ArrayListCursorBehavior.Stop);
        }

        #endregion IEnumerable<T> Members

        #region Nested Iterator

        public enum ArrayListCursorBehavior
        {
            /// <summary>
            /// The cursor will stop when it has reached the end (beginning)
            /// </summary>
            Stop,

            /// <summary>
            /// The cursor will loop when it has reached the end (beginning)
            /// </summary>
            Loop
        }

        public struct ArrayListCursor : IEnumerator<T>
        {
            /// <summary>
            /// Underlying items array
            /// </summary>
            public ArrayList<T> ArrayRef;

            /// <summary>
            /// Currently selected handle
            /// </summary>
            public int SelectedHandle;

            /// <summary>
            /// Behavior for this cursor
            /// </summary>
            public ArrayListCursorBehavior Behavior;

            #region Constructors

            public ArrayListCursor(ArrayList<T> arrayRef)
            {
                this.ArrayRef = arrayRef;
                this.SelectedHandle = 0;
                Behavior = ArrayListCursorBehavior.Stop;
            }

            public ArrayListCursor(ArrayList<T> arrayRef, ArrayListCursorBehavior behavior)
            {
                this.ArrayRef = arrayRef;
                this.SelectedHandle = 0;
                Behavior = behavior;
            }

            public ArrayListCursor(ArrayList<T> arrayRef, int startingPosition)
            {
                this.ArrayRef = arrayRef;
                this.SelectedHandle = startingPosition;
                Behavior = ArrayListCursorBehavior.Stop;
            }

            public ArrayListCursor(ArrayList<T> arrayRef, int startingPosition, ArrayListCursorBehavior behavior)
            {
                this.ArrayRef = arrayRef;
                this.SelectedHandle = startingPosition;
                Behavior = behavior;
            }

            #endregion Constructors

            #region IEnumerator<T> Members

            public T Current
            {
                get
                {
                    if (this.ArrayRef.Count == 0) // No elements, return default
                        return default(T);
                    return this.ArrayRef.Items[this.SelectedHandle];
                }
            }

            object System.Collections.IEnumerator.Current
            {
                get
                {
                    if (this.ArrayRef.Count == 0) // No elements, return default
                        return default(T);
                    return this.ArrayRef.Items[this.SelectedHandle];
                }
            }

            public bool MoveNext()
            {
                int elems = this.ArrayRef.Count;
                if (elems == 0) // No elements, can not move
                    return false;

                int count = ArrayRef.RawCount;
                T[] items = ArrayRef.Items;
                if (elems == 1 && this.ArrayRef.HasElementAt(this.SelectedHandle)) // One element and it's already selected
                    return false;

                if (elems == 1) // One element, but the selected is dirty
                {
                    for (int i = 0; i < count; ++i)
                    {
                        if (!this.ArrayRef.HasElementAt(i))
                            continue;
                        this.SelectedHandle = i;
                        return true; // we found the element
                    }
                    return false;
                }

                // Standard move next
                for (int i = this.SelectedHandle + 1; i < count; ++i)
                {
                    if (!this.ArrayRef.HasElementAt(i))
                        continue;
                    this.SelectedHandle = i;
                    return true;
                }

                if (this.Behavior == ArrayListCursorBehavior.Stop)
                    return false;
                else // Loop from the beginning
                {
                    for (int i = 0; i < this.SelectedHandle; ++i)
                    {
                        if (!this.ArrayRef.HasElementAt(i))
                            continue;
                        this.SelectedHandle = i;
                        return true;
                    }
                    return false;
                }
            }

            public void Reset()
            {
                int elems = this.ArrayRef.Count;
                if (elems == 0)
                {
                    // No elements in the array, can't reset
                    SelectedHandle = 0;
                }
                else
                {
                    // Find first element
                    int count = this.ArrayRef.RawCount;
                    T[] items = ArrayRef.Items;
                    for (int i = 0; i < count; ++i)
                    {
                        if (!this.ArrayRef.HasElementAt(i))
                            continue;
                        SelectedHandle = i;
                        return;
                    }
                }
            }

            public void Dispose()
            {
                // why there's always a dispose for an interator? crappy thing
            }

            #endregion IEnumerator<T> Members
        }

        #endregion Nested Iterator
    }
}