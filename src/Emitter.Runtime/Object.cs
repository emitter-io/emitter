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

namespace Emitter
{
    /// <summary>
    /// Represents an object that implements IDisposable contract.
    /// </summary>
    public abstract class DisposableObject : IDisposable
    {
        #region IDisposable Members

        /// <summary>
        /// Performs application-defined tasks associated with freeing, releasing, or
        /// resetting unmanaged resources.
        /// </summary>
        public void Dispose()
        {
            // the object is actually going to die.
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
        }

        /// <summary>
        /// Finalizer for the recyclable object.
        /// </summary>
        ~DisposableObject()
        {
            // The object is actually going to die.
            Dispose(false);
        }

        #endregion IDisposable Members
    }

    /// <summary>
    /// Represents an object that implements IRecyclable contract, allowing the object instance to be reused.
    /// </summary>
    public abstract class RecyclableObject : IRecyclable, IDisposable
    {
        #region IRecyclable Members
        private ReleaseInstanceDelegate Release = null;

        /// <summary>
        /// A fild that contains the value representing whether the object is acquired or not.
        /// </summary>
        protected bool ObjectAcquired = false;

        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        public abstract void Recycle();

        /// <summary>
        /// Binds an <see cref="ReleaseInstanceDelegate"/> delegate which releases the <see cref="IRecyclable"/> object
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
            // Flag this as acquired
            this.ObjectAcquired = true;
        }

        /// <summary>
        /// Gets whether this <see cref="RecyclableObject"/> object is pooled or not.
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
        public void Dispose()
        {
            // Object is still registered for finalization
            if (Release != null && this.ObjectAcquired)
            {
                // Release back to the pool.
                this.ObjectAcquired = false;
                this.Release(this);
            }
            else
            {
                // Otherwise, the object is actually going to die.
                Dispose(true);
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
        /// Releases the unmanaged resources used by the ByteSTream class and optionally releases the managed resources.
        /// </summary>
        /// <param name="disposing">
        /// If set to true, release both managed and unmanaged resources, othewise release only unmanaged resources.
        /// </param>
        protected virtual void Dispose(bool disposing)
        {
        }

        /// <summary>
        /// Finalizer for the recyclable object.
        /// </summary>
        ~RecyclableObject()
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

    /// <summary>
    /// Defines the the IRecyclable contract, allowing the object instance to be reused.
    /// </summary>
    public interface IRecyclable : IDisposable
    {
        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        void Recycle();

        /// <summary>
        /// Binds an <see cref="ReleaseInstanceDelegate"/> which releases the <see cref="IRecyclable"/> object
        /// instance back to the pool.
        /// </summary>
        /// <param name="releaser">The <see cref="ReleaseInstanceDelegate"/> delegate to bind.</param>
        void Bind(ReleaseInstanceDelegate releaser);

        /// <summary>
        /// Invoked when <see cref="IRecyclable"/> object is about to be acquired.
        /// </summary>
        void OnAcquire();
    }

    /// <summary>
    /// The <see cref="ReleaseInstanceDelegate"/> delegate is used as a callback to return an object back to the pool.
    /// </summary>
    public delegate void ReleaseInstanceDelegate(IRecyclable @object);

    /// <summary>
    /// Represents generic, aws-compatible credentials container.
    /// </summary>
    public interface ICredentials
    {
        /// <summary>
        /// The access key.
        /// </summary>
        string AccessKey { get; }

        /// <summary>
        /// The secret key.
        /// </summary>
        string SecretKey { get; }

        /// <summary>
        /// The token.
        /// </summary>
        string Token { get; }

        /// <summary>
        /// The duration of the credentials.
        /// </summary>
        TimeSpan Duration { get; }

        /// <summary>
        /// The expiration date of the credentials.
        /// </summary>
        DateTime Expires { get; }
    }
}