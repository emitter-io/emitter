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
using System.IO;
using System.Linq;
using System.Text;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a stream of bytes.
    /// </summary>
    public sealed class ByteStream : MemoryStream, IRecyclable
    {
        #region Constructors

        /// <summary>
        /// Initializes a new instance of the <see cref="ByteStream"/> class with an expandable
        /// capacity initialized to zero.
        /// </summary>
        public ByteStream() : base() { }

        /// <summary>
        /// Initializes a new instance of the <see cref="ByteStream"/> class with an expandable
        /// capacity initialized as specified.
        /// </summary>
        /// <param name="capacity">The initial size of the internal array in bytes.</param>
        public ByteStream(int capacity) : base(capacity) { }

        /// <summary>
        /// Initializes a new instance of the <see cref="ByteStream"/> class and writes a string data to the underlying buffer.
        /// </summary>
        /// <param name="text">String data to write.</param>
        /// <param name="encoding">Encoding to use for writing the string data.</param>
        public ByteStream(string text, Encoding encoding) : base() { Write(encoding.GetBytes(text)); }

        /// <summary>
        /// Initializes a new non-resizable instance of the <see cref="ByteStream"/> class
        /// based on the specified region of a byte array, with the <see cref="ByteStream"/> CanWrite
        /// property set as specified.
        /// </summary>
        /// <param name="bytes">The array of unsigned bytes from which to create this stream.</param>
        /// <param name="writable">The setting of the <see cref="ByteStream"/> CanWrite property, which determines whether the stream supports writing.</param>
        public ByteStream(byte[] bytes, bool writable) : base(bytes, 0, bytes.Length, writable, true) { }

        /// <summary>
        /// Initializes a new resizable instance of the <see cref="ByteStream"/> class
        /// based on the specified byte array.
        /// </summary>
        /// <param name="bytes">The array of unsigned bytes from which to create the current stream.</param>
        public ByteStream(byte[] bytes) : base(bytes, 0, bytes.Length, true, true) { }

        #endregion Constructors

        #region Public Members

        /// <summary>
        /// Writes a block of bytes to the current stream using data read from buffer.
        /// </summary>
        /// <param name="bytes">The buffer to write data from.</param>
        public void Write(byte[] bytes)
        {
            Write(bytes, 0, bytes.Length);
        }

        #endregion Public Members

        #region IRecyclable Members
        private ReleaseInstanceDelegate Release = null;
        internal bool ObjectAcquired = false;

        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        public void Recycle()
        {
            // Set the memory stream length to zero.
            this.SetLength(0);
        }

        /// <summary>
        /// Binds an <see cref="ReleaseInstanceDelegate"/> which releases the <see cref="IRecyclable"/> object
        /// instance back to the pool.
        /// </summary>
        /// <param name="releaser">The <see cref="ReleaseInstanceDelegate"/> delegate to bind.</param>
        public void Bind(ReleaseInstanceDelegate releaser)
        {
            this.Release = releaser;
        }

        /// <summary>
        /// Invoked when a pool acquires the instance.
        /// </summary>
        public void OnAcquire()
        {
            this.ObjectAcquired = true;
        }

        /// <summary>
        /// Gets whether this <see cref="ByteStream"/> object is pooled or not.
        /// </summary>
        public bool IsPooled
        {
            get { return Release != null; }
        }

        #endregion IRecyclable Members

        #region IDisposable Members

        /// <summary>
        /// Performs application-defined tasks associated with freeing, releasing, or
        /// resetting unmanaged resources.
        /// </summary>
        new public void Dispose()
        {
            // If the buffer is pooled we must release it.
            // It is possible the buffer is not pooled, in which case this will fail and no damage will be done, that's
            // why we hide the .Dispose() method here.
            if (Release != null && this.ObjectAcquired)
            {
                // Release back to the pool.
                this.ObjectAcquired = false;
                this.Release(this);
            }
            else
            {
                // Otherwise, the object is actually going to die.
                base.Dispose();
            }
        }

        /// <summary>
        /// Attempts to release this instance back to the pool. If the instance is not pooled, nothing will be done.
        /// </summary>
        public void TryRelease()
        {
            // Release back to the pool.
            if (Release != null && this.ObjectAcquired)
            {
                this.ObjectAcquired = false;
                this.Release(this);
            }
        }

        /// <summary>
        /// Finalizer for the recyclable object.
        /// </summary>
        ~ByteStream()
        {
            if (Release != null && this.ObjectAcquired)
            {
                // Release back to the pool and register back to the finalizer thread.
                this.ObjectAcquired = false;
                this.Release(this);
                GC.ReRegisterForFinalize(this);
            }
            else
            {
                // Otherwise, the object is actually going to die.
                Dispose(false);
            }
        }

        #endregion IDisposable Members
    }
}