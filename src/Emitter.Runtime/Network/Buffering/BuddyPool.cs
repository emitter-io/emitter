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
using System.Linq;
using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;

namespace Emitter.Network
{
    /// <summary>
    /// Represents the status of the pool.
    /// </summary>
    internal enum BuddyBlockStatus : byte
    {
        /// <summary>
        /// The memory is free.
        /// </summary>
        Available = 0,

        /// <summary>
        /// The memory is reserved
        /// </summary>
        Reserved = 1
    }

    internal unsafe class BuddyBlock
    {
        /// <summary>
        /// The size of the block in memory
        /// </summary>
        public const int Size = 64;

        public const int DefaultMinMemory = 16; // 2^16 = 65 KB
        public const int DefaultMaxMemory = 22; // 2^22 = 4 MB

        /// <summary>
        /// Constructs a new <see cref="BuddyBlock"/>;
        /// </summary>
        public BuddyBlock(int power, int memoryOffset)
        {
            this.Status = BuddyBlockStatus.Available;
            this.Power = power;
            this.Offset = memoryOffset;
            this.Next = null;
            this.Previous = null;
        }

        /// <summary>
        /// Whether the block is available or not.
        /// </summary>
        public BuddyBlockStatus Status;

        /// <summary>
        /// The power of the available block.
        /// </summary>
        public int Power;

        /// <summary>
        /// The memory offset for a reserved block.
        /// </summary>
        public int Offset;

        /// <summary>
        /// The next link in the freelist.
        /// </summary>
        public BuddyBlock Next;

        /// <summary>
        /// The previous link in the freelist.
        /// </summary>
        public BuddyBlock Previous;

        /// <summary>
        /// Gets the length of the block.
        /// </summary>
        public int Length
        {
            get { return 1 << Power; }
        }

        /// <summary>
        /// Handy debugging toString method.
        /// </summary>
        public override string ToString()
        {
            if (Next == null && Previous == null)
                return "(unallocated)";

            return String.Format("[{0}] Buddy:{1}, Len:{2}, Link: ({3}, {4})",
                Status,
                this.Offset,
                Length,
                Next == null
                    ? "(null)"
                    : Next == this
                        ? "(self)"
                        : Next.Offset.ToString(),
                Previous == null
                    ? "(null)"
                    : Previous == this
                        ? "(self)"
                        : Previous.Offset.ToString());
        }
    }

    internal unsafe class BuddyPool : IDisposable
    {
        #region Constructor

        // The internal array, containing
        protected bool Disposed = false;

        protected byte[] Memory = new byte[0];
        private GCHandle hMemory;

        // Size of the storage pool
        private int NumberOfPowers;

        private int NumberOfBlocks;

        // Limits of the storage pool
        private int MaxMemoryBase = BuddyBlock.DefaultMaxMemory;

        // Metadata of every block
        private BuddyBlock[] Metadata = new BuddyBlock[0];

        // Sentinels for free lists
        private BuddyBlock[] Sentinel = new BuddyBlock[0];

        // Locking
        protected object Lock = new object();

        /// <summary>
        /// Construct a buddy pool for initial size buffer.
        /// </summary>
        public BuddyPool() : this(BuddyBlock.DefaultMinMemory)
        {
        }

        /// <summary>
        /// Construct a buddy pool for initial size buffer.
        /// </summary>
        /// <param name="initialPower">The minimum initial power of the pool.</param>
        public BuddyPool(int initialPower)
        {
            // Initialize the first block
            lock (this.Lock)
            {
                // Allocate the pool
                this.Resize(initialPower);

                this.Metadata[0].Status = BuddyBlockStatus.Available;
                this.Metadata[0].Power = this.NumberOfPowers;
                InsertAfter(this.Sentinel[this.NumberOfPowers], this.Metadata[0]);
            }
        }

        /// <summary>
        /// Gets or sets the maximum memory limit, in bytes.
        /// </summary>
        public int MaxMemory
        {
            get { return 1 << MaxMemoryBase; }
            set
            {
                if (value <= 0)
                    throw new ArgumentOutOfRangeException("MaxMemory", "The limit should be larger than zero bytes.");
                if ((value & (value - 1)) != 0)
                    throw new ArgumentException("MaxMemory", "The limit should be a power of two.");

                // Get the base-2 logarithm
                this.MaxMemoryBase = (int)Math.Log(value, 2);
            }
        }

        #endregion Constructor

        #region Public Members

        /// <summary>
        /// Returns a pointer to the region of memory that is allocated.
        /// </summary>
        /// <param name="size">An integer-valued argument which specifies the size of storage area required.</param>
        /// <returns>A pointer to the region of memory that is allocated.</returns>
        public int Acquire(int size)
        {
            lock (this.Lock)
            {
                int kPrime = Log2Ceil(size > BuddyBlock.Size ? size : BuddyBlock.Size);
                int i = kPrime;

                while (i <= this.NumberOfPowers && (this.Sentinel[i]).Next == (this.Sentinel[i]))
                    ++i;

                if (i > this.NumberOfPowers)
                {
                    int powers = this.NumberOfPowers;
                    int blocks = this.NumberOfBlocks;

                    // Resize instead, with a maximum check
                    this.Resize(++this.NumberOfPowers);

                    // Set the sentinel to the newly allocated space
                    BuddyBlock head = this.Metadata[blocks];
                    head.Status = BuddyBlockStatus.Available;
                    head.Power = powers;
                    InsertAfter(this.Sentinel[powers], head);

                    // Search again
                    return Acquire(size);
                }

                BuddyBlock block = this.Sentinel[i].Next;
                Unlink(block);
                while (block.Power > kPrime)
                {
                    block.Power--;

                    // Instead of hitting this.GetBuddy(block), we simply manually inline it
                    var buddy = this.Metadata[((block.Offset ^ (1 << (block.Power))) / BuddyBlock.Size)];
                    buddy.Status = BuddyBlockStatus.Available;
                    buddy.Power = block.Power;

                    InsertAfter(this.Sentinel[buddy.Power], buddy);
                }

                block.Status = BuddyBlockStatus.Reserved;

                // Return the offset of the allocated block
                return block.Offset;
            }
        }

        /// <summary>
        /// Releases the memory back in the buddy pool.
        /// </summary>
        /// <param name="offset"></param>
        public void Release(int offset)
        {
            // Do not do anything if disposed
            if (this.Disposed)
                return;

            // Convert the offset
            int index = offset / BuddyBlock.Size;
            if (index < 0 || index >= this.NumberOfBlocks)
                throw new ArgumentOutOfRangeException("The specified offset was not found in the pool.");

            if (this.Metadata == null)
                return;

            // Smaller lock, starting where we actually touch the pool
            lock (this.Lock)
            {
                // Get the block address
                BuddyBlock block = this.Metadata[index];

                // Decrement the segments
                if (block.Status > BuddyBlockStatus.Reserved)
                {
                    block.Status--;
                    return;
                }

                block.Status = BuddyBlockStatus.Available;

                //BuddyBlock* ptr;
                //for (ptr = block; ptr->Power < this.NumberOfPowers; (ptr->Power)++)
                BuddyBlock ptr = block;
                for (int i = ptr.Power; i < this.NumberOfPowers; i++)
                {
                    // Instead of hitting  this.GetBuddy(ptr), we simply manually inline this
                    BuddyBlock buddy = this.Metadata[((ptr.Offset ^ (1 << (ptr.Power))) / BuddyBlock.Size)];
                    if (buddy.Status == BuddyBlockStatus.Reserved || buddy.Power != ptr.Power)
                        break;

                    Unlink(buddy);

                    //if (buddy < ptr)
                    if (buddy.Offset < ptr.Offset)
                        ptr = buddy;
                }
                InsertAfter(this.Sentinel[ptr.Power], ptr);
            }
        }

        #endregion Public Members

        #region Private Members

        /// <summary>
        /// Resizes the pool to a minimum size.
        /// </summary>
        /// <param name="power">The minimum amount of bytes required to be in the pool.</param>
        private void Resize(int power)
        {
            lock (this.Lock)
            {
                // Do not do anything if disposed
                if (this.Disposed)
                    return;

                if (power < 6 || power > this.MaxMemoryBase)
                    throw new OutOfMemoryException("Unable to allocate more than " + this.MaxMemory + " bytes of memory.");

                //Console.WriteLine("Expanded to " + Math.Pow(2, power));

                this.NumberOfPowers = power;
                this.NumberOfBlocks = ((1 << NumberOfPowers) / BuddyBlock.Size);

                // Keep old arrays of pointers to update
                var memory = this.Memory;
                var hMemoryOld = hMemory;

                // Allocate the new memory block
                this.Memory = new byte[BuddyBlock.Size * this.NumberOfBlocks];
                hMemory = GCHandle.Alloc(this.Memory, GCHandleType.Pinned);
                Buffer.BlockCopy(memory, 0, this.Memory, 0, memory.Length);

                // Allocate the metadata pool
                Array.Resize<BuddyBlock>(ref this.Metadata, this.NumberOfBlocks);
                Array.Resize<BuddyBlock>(ref this.Sentinel, this.NumberOfPowers + 1);

                // Set the memory offset on every block
                for (int i = 0; i < this.NumberOfBlocks; ++i)
                {
                    if (this.Metadata[i] == null)
                        this.Metadata[i] = new BuddyBlock(0, i * BuddyBlock.Size);
                }

                // Initialize all the sentinels
                for (int i = 0; i < this.NumberOfPowers + 1; ++i)
                {
                    // Set the next and previous to the block itself, meaning it is unlinked
                    if (this.Sentinel[i] == null)
                    {
                        this.Sentinel[i] = new BuddyBlock(0, i);
                        this.Sentinel[i].Next = this.Sentinel[i];
                        this.Sentinel[i].Previous = this.Sentinel[i];
                    }
                }

                // Free the old memory if needed
                if (hMemoryOld != default(GCHandle) && hMemoryOld.IsAllocated)
                    hMemoryOld.Free();
            }
        }

        /// <summary>
        /// Gets logarithm of two ceiling.
        /// </summary>
        /// <param name="x">The number to compute the logarithm of two ceiling.</param>
        /// <returns>The computed logarithm of two ceiling.</returns>
        private int Log2Ceil(int x)
        {
            // The code below does: (int)Math.Ceiling(Math.Log((double)x, 2));
            uint v = (uint)(x - 1); // 32-bit word to find the log base 2 of
            uint r = 0; // r will be lg(v)
            while (v > 0) // unroll for more speed...
            {
                v >>= 1;
                r++;
            }
            return (int)r;
        }

        /// <summary>
        /// Gets a buddy of a given block.
        /// </summary>
        /// <param name="block">The block to get a buddy for.</param>
        /// <returns>The pointer to the buddy.</returns>
        private BuddyBlock GetBuddy(BuddyBlock block)
        {
            return this.Metadata[((block.Offset ^ (1 << (block.Power))) / BuddyBlock.Size)];
        }

        #endregion Private Members

        #region InsertAfter and Unlink

        /// <summary>
        /// Unlinks the block from the freelist.
        /// </summary>
        /// <param name="block"></param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private void Unlink(BuddyBlock block)
        {
            if (block.Next != null)
                block.Next.Previous = block.Previous;
            if (block.Previous != null)
                block.Previous.Next = block.Next;
            block.Next = block.Previous = block;
        }

        /// <summary>
        /// Inserts a buddy block after the specified block.
        /// </summary>
        /// <param name="src">The source block to insert after.</param>
        /// <param name="block">The target block to insert after the source one.</param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private void InsertAfter(BuddyBlock src, BuddyBlock block)
        {
            block.Next = src.Next;
            block.Previous = src;
            if (src.Next != null)
                src.Next.Previous = block;
            src.Next = block;
        }

        #endregion InsertAfter and Unlink

        #region IDisposable Members

        /// <summary>
        /// Called by the GC when the object is finalized.
        /// </summary>
        ~BuddyPool()
        {
            // Forward to OnDispose
            this.OnDispose(false);
        }

        /// <summary>
        /// Invoked when the object pool is disposing.
        /// </summary>
        /// <param name="isDisposing">Whether the OnDispose was called by a finalizer or a dispose method</param>
        protected void OnDispose(bool isDisposing)
        {
            lock (this.Lock)
            {
                // Do not do anything if disposed
                if (this.Disposed)
                    return;

                // Free the memory
                if (hMemory != default(GCHandle) && hMemory.IsAllocated)
                    hMemory.Free();

                // Mark as disposed
                this.Disposed = true;
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