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
using Emitter.Diagnostics;

namespace Emitter.Collections
{
    /// <summary>
    /// Class that represents a pool of byte[] available on request.
    /// </summary>
    public sealed class BufferPool : IDisposable
    {
        private int InitialCapacity;
        private int BufferSize;
        private int RawCount;
        private int Count;
        private byte[][] Items;
        private int FreeListCount;
        private int[] FreeList;
        private byte[] UsedList;

        /// <summary>
        /// Constructs a <see cref="BufferPool"/> object with the specified initial capacity and buffer sizes.
        /// </summary>
        /// <param name="initialCapacity">The initial capacity of the pool.</param>
        /// <param name="bufferSize">The sizes of the buffers to allocate initially.</param>
        public BufferPool(int initialCapacity, int bufferSize)
        {
            this.InitialCapacity = initialCapacity;
            this.BufferSize = bufferSize;

            this.Items = new byte[initialCapacity][];
            this.FreeList = new int[initialCapacity];
            this.UsedList = new byte[initialCapacity];

            for (int i = 0; i < initialCapacity; ++i)
            {
                this.Items[i] = new byte[bufferSize];
                this.FreeList[i] = i;
                this.UsedList[i] = 0;
            }

            this.RawCount = initialCapacity;
            this.FreeListCount = initialCapacity;
            this.Count = 0;
        }

        /// <summary>
        /// Acquires an instance of byte[] and returns the reference to the instance
        /// as well as the handle in the Pool
        /// </summary>
        /// <param name="item">Reference to the recently acquired instance of byte[]</param>
        /// <returns>The handle in the Pool (index in the internal array)</returns>
        public int Acquire(out byte[] item)
        {
            lock (this.Items)
            {
                NetStat.UsedByteArrays.Increment();

                if (this.FreeListCount > 0)
                {
                    this.FreeListCount--;
                    int handle = this.FreeList[this.FreeListCount];
                    this.UsedList[handle] = 1;
                    this.Count++;

                    if (this.Items[handle] == null)
                    {
                        item = new byte[this.BufferSize];
                        this.Items[handle] = item;
                        return handle;
                    }

                    item = this.Items[handle];
                    return handle;
                }
                else
                {
                    int handle = this.RawCount;
                    if (handle < this.Items.Length)
                    {
                        this.UsedList[handle] = 1;
                        this.Count++;

                        if (this.Items[handle] == null)
                        {
                            item = new byte[this.BufferSize];
                            this.Items[handle] = item;
                            this.RawCount++;
                            return handle;
                        }

                        item = this.Items[handle];
                        this.RawCount++;
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

                        item = new byte[this.BufferSize];
                        this.Items[handle] = item;
                        this.UsedList[handle] = 1;
                        this.Count++;
                        this.RawCount++;
                        return handle;
                    }
                }
            }
        }

        /// <summary>
        /// Acquires an instance of byte[] and returns the reference to the instance
        /// as well as the handle in the Pool.
        /// </summary>
        /// <param name="buffer">Reference to the recently acquired instance of byte[].</param>
        /// <param name="minimumLength">Minimum lenght of the buffer to acquire.</param>
        /// <returns>The handle in the Pool (index in the internal array)</returns>
        public int Acquire(out byte[] buffer, int minimumLength)
        {
            lock (this.Items)
            {
                NetStat.UsedByteArrays.Increment();

                buffer = null;
                int handle = 0;

                if (this.FreeListCount > 0) // Check the free list
                {
                    // goes through freed items
                    for (int i = 0; i < this.FreeListCount; ++i)
                    {
                        // satisfies the length condition?
                        handle = this.FreeList[i];
                        if (this.Items[handle].Length >= minimumLength)
                        {
                            // Overwrite the current number with the last one
                            this.FreeList[i] = this.FreeList[--this.FreeListCount];
                            this.UsedList[handle] = 1;
                            this.Count++;

                            // Return the buffer
                            buffer = this.Items[handle];
                            return handle;
                        }
                    }
                }

                // Not found in free buffers, allocate
                handle = this.RawCount;
                if (handle < this.Items.Length)
                {
                    // Return a new buffer
                    buffer = AllocateBuffer(handle, minimumLength);
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

                    // Return a new buffer
                    buffer = AllocateBuffer(handle, minimumLength);
                    return handle;
                }
            }
        }

        private byte[] AllocateBuffer(int handle, int minimumLength)
        {
            // Compute the size of the new buffer
            int newBufferSize = minimumLength;
            newBufferSize |= newBufferSize >> 1;
            newBufferSize |= newBufferSize >> 2;
            newBufferSize |= newBufferSize >> 4;
            newBufferSize |= newBufferSize >> 8;
            newBufferSize |= newBufferSize >> 16;
            newBufferSize++;

            // Allocate the buffer
            byte[] buffer = new byte[newBufferSize];
            this.Items[handle] = buffer;

            // Set the used list index & increment counters
            this.UsedList[handle] = 1;
            this.RawCount++;
            this.Count++;

            return buffer;
        }

        /// <summary>
        /// Releases the instance of T specified by the handle
        /// </summary>
        /// <param name="handle"></param>
        public void Release(int handle)
        {
            lock (this.Items)
            {
                if (this.UsedList[handle] == 0)
                    return; // the spot is already free

                NetStat.UsedByteArrays.Decrement();

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
        }

        /// <summary>
        /// Gets the buffer by a buffer handle.
        /// </summary>
        /// <param name="handle">The handle of the buffer to get.</param>
        /// <returns>Returns the buffer if one can be used, null otherwise.</returns>
        public byte[] GetBuffer(int handle)
        {
            if (this.UsedList[handle] == 0)
                return null; // the spot is free
            return this.Items[handle];
        }

        /// <summary>
        /// Clears the buffer pool without deallocating the buffers.
        /// </summary>
        public void Clear()
        {
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
            {
                if (this.UsedList[i] == 1)
                    Release(i);
            }
        }

        #region IDisposable Members

        /// <summary>
        /// Frees the pool, deallocates the reserved memory
        /// </summary>
        private void Dispose(bool disposing)
        {
            int count = this.RawCount;
            for (int i = 0; i < count; ++i)
                this.Items[i] = default(byte[]); // set to default

            this.RawCount = 0;
            this.FreeListCount = 0;
        }

        /// <summary>
        /// Disposes the BufferPool object.
        /// </summary>
        public void Dispose()
        {
            Dispose(true);
            GC.SuppressFinalize(this);
        }

        /// <summary>
        /// Finalizes the BufferPool object.
        /// </summary>
        ~BufferPool()
        {
            Dispose(false);
        }

        #endregion IDisposable Members
    }
}