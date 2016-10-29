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

namespace Emitter.Threading
{
    /// <summary>
    /// Represents a contract for a value that can be assigned in an atomic way.
    /// </summary>
    /// <typeparam name="T">The type of the underlying value.</typeparam>
    public interface IAtomicValue<T>
    {
        /// <summary>
        /// Gets the value of the atomic structure
        /// </summary>
        T Value { get; }

        /// <summary>
        /// Atomically assigns a value to the structure
        /// </summary>
        /// <param name="value">The value to assign atomically</param>
        void Assign(T value);

        /// <summary>
        /// Atomically pefroms a computation and assigns it to the atomic value
        /// </summary>
        /// <param name="computation">Computation to execute atomically</param>
        void Assign(Func<T, T> computation);
    }
}