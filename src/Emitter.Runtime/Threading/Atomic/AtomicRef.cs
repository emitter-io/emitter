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
using System.Runtime.InteropServices;
using System.Threading;

namespace Emitter.Threading
{
    /// <summary>
    /// Represents an atomic reference to an object integer value.
    /// </summary>
    [StructLayout(LayoutKind.Sequential, Pack = 4)]
    public struct AtomicRef<ObjectT> : IAtomicValue<ObjectT> where ObjectT : class
    {
        private ObjectT fValue;

        /// <summary>
        /// Constructs a new <see cref="AtomicRef{T}"/> instance.
        /// </summary>
        /// <param name="objectT">The target object reference.</param>
        public AtomicRef(ObjectT objectT)
        {
            fValue = objectT;
        }

        /// <summary>
        /// Gets the value of the atomic structure
        /// </summary>
        public ObjectT Value
        {
            get { return fValue; }
        }

        /// <summary>
        /// Atomically pefroms an assignment operation to the atomic value
        /// </summary>
        /// <param name="value">Value to set</param>
        public void Assign(ObjectT value)
        {
            ObjectT oldAtomicObject;
            ObjectT newAtomicObject;

            do
            {
                oldAtomicObject = fValue;
                newAtomicObject = value;
            } while (Interlocked.CompareExchange<ObjectT>(ref fValue, newAtomicObject, oldAtomicObject) != oldAtomicObject);
        }

        /// <summary>
        /// Atomically pefroms a computation and assigns it to the atomic value
        /// </summary>
        /// <param name="computation">Computation to execute atomically</param>
        public void Assign(Func<ObjectT, ObjectT> computation)
        {
            ObjectT oldValue;
            ObjectT newValue;

            do
            {
                oldValue = fValue;
                newValue = computation(oldValue);
            } while (Interlocked.CompareExchange<ObjectT>(ref fValue, newValue, oldValue) != oldValue);
        }

        /// <summary>
        /// Compares two values for equality and, if they are equal, replaces one of the values.
        /// </summary>
        /// <param name="value1">First value to compare.</param>
        /// <param name="value2">Second value to compare.</param>
        /// <returns>The original value in value1.</returns>
        public bool CompareExchange(ObjectT value1, ObjectT value2)
        {
            return Interlocked.CompareExchange<ObjectT>(ref fValue, value1, value2) == value2;
        }

        /// <summary>
        /// Compares the equality of two values.
        /// </summary>
        /// <param name="left">Left parameter to compare.</param>
        /// <param name="right">Right parameter to compare.</param>
        /// <returns>Returns whether left and right parameters are equals or not.</returns>
        public static bool operator ==(AtomicRef<ObjectT> left, AtomicRef<ObjectT> right)
        {
            return left.fValue == right.fValue;
        }

        /// <summary>
        /// Compares the equality of two values.
        /// </summary>
        /// <param name="left">Left parameter to compare.</param>
        /// <param name="right">Right parameter to compare.</param>
        /// <returns>Returns whether left and right parameters are equals or not.</returns>
        public static bool operator ==(AtomicRef<ObjectT> left, ObjectT right)
        {
            return left.fValue == right;
        }

        /// <summary>
        /// Checks for inequality of two values.
        /// </summary>
        /// <param name="left">Left parameter to compare.</param>
        /// <param name="right">Right parameter to compare.</param>
        /// <returns>Returns true whether left and right parameters are not equals, otherwise false.</returns>
        public static bool operator !=(AtomicRef<ObjectT> left, AtomicRef<ObjectT> right)
        {
            return left.fValue != right.fValue;
        }

        /// <summary>
        /// Checks for inequality of two values.
        /// </summary>
        /// <param name="left">Left parameter to compare.</param>
        /// <param name="right">Right parameter to compare.</param>
        /// <returns>Returns true whether left and right parameters are not equals, otherwise false.</returns>
        public static bool operator !=(AtomicRef<ObjectT> left, ObjectT right)
        {
            return left.fValue != right;
        }

        /// <summary>
        /// Converts the atomic value to a non-atomic one.
        /// </summary>
        /// <param name="value">The value to convert.</param>
        /// <returns>The converted value.</returns>
        public static explicit operator ObjectT(AtomicRef<ObjectT> value)
        {
            return value.fValue;
        }

        /// <summary>
        /// Converts the non-atomic value to an atomic one.
        /// </summary>
        /// <param name="value">The value to convert.</param>
        /// <returns>The converted value.</returns>
        public static implicit operator AtomicRef<ObjectT>(ObjectT value)
        {
            return new AtomicRef<ObjectT>(value);
        }

        /// <summary>
        /// Converts the value of this instance to its equivalent string representation.
        /// </summary>
        public override string ToString()
        {
            return fValue.ToString();
        }

        /// <summary>
        /// Returns a value indicating whether this instance is equal to a specified value
        /// </summary>
        public override bool Equals(object obj)
        {
            AtomicRef<ObjectT> atom = (AtomicRef<ObjectT>)obj;
            if (atom == null)
                return false;

            return atom.fValue == fValue;
        }

        /// <summary>
        /// Returns the hash code for this instance.
        /// </summary>
        /// <returns>A 32-bit signed integer hash code.</returns>
        public override int GetHashCode()
        {
            return fValue.GetHashCode();
        }
    }
}