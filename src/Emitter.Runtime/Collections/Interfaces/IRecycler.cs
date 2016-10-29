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

namespace Emitter.Collections
{
    /// <summary>
    /// Defines a contract for an object pool.
    /// </summary>
    public interface IRecycler
    {
        /// <summary>
        /// Acquires an instance of a recyclable object
        /// </summary>
        /// <returns></returns>
        IRecyclable Acquire();

        /// <summary>
        /// Releases an instance of a recyclable object back to the pool.
        /// </summary>
        /// <param name="instance">The instance of IRecyclable to release.</param>
        void Release(IRecyclable instance);

        /// <summary>
        /// Gets the overall number of elements managed by this pool.
        /// </summary>
        int Count { get; }

        /// <summary>
        /// Gets the number of available elements currently contained in the pool.
        /// </summary>
        int AvailableCount { get; }

        /// <summary>
        /// Gets the number of elements currently in use and not available in this pool.
        /// </summary>
        int InUseCount { get; }
    }

    /// <summary>
    /// Defines a contract for an object pool.
    /// </summary>
    public interface IRecycler<T> : IRecycler
        where T : class, IRecyclable
    {
        /// <summary>
        /// Acquires an instance of a recyclable object
        /// </summary>
        /// <returns></returns>
        new T Acquire();

        /// <summary>
        /// Releases an instance of a recyclable object back to the pool.
        /// </summary>
        /// <param name="instance">The instance of IRecyclable to release.</param>
        void Release(T instance);
    }
}