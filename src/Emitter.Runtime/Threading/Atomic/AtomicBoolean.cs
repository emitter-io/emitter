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
    /// Represents an atomic boolean.
    /// </summary>
    [StructLayout(LayoutKind.Sequential, Pack = 4)]
    public struct AtomicBoolean : IAtomicValue<bool>
    {
        private int fValue;

        private AtomicBoolean(bool value)
        {
            fValue = value ? 1 : 0;
        }

        /// <summary>
        /// Gets the value of the atomic structure
        /// </summary>
        public bool Value
        {
            get { return fValue == 0 ? false : true; }
        }

        /// <summary>
        /// Atomically pefroms an assignment operation to the atomic value
        /// </summary>
        /// <param name="value">Value to set</param>
        public void Assign(bool value)
        {
            int oldValue;
            int newValue;

            do
            {
                oldValue = fValue;
                newValue = value ? 1 : 0;
            } while (Interlocked.CompareExchange(ref fValue, newValue, oldValue) != oldValue);
        }

        /// <summary>
        /// Atomically assignes a true value to this <see cref="AtomicBoolean"/> reference.
        /// </summary>
        public void SetTrue()
        {
            int oldValue;
            do
            {
                oldValue = fValue;
                if (oldValue == 1)
                    return;
            } while (Interlocked.CompareExchange(ref fValue, 1, oldValue) != oldValue);
        }

        /// <summary>
        /// Atomically assignes a false value to this <see cref="AtomicBoolean"/> reference.
        /// </summary>
        public void SetFalse()
        {
            int oldValue;
            do
            {
                oldValue = fValue;
                if (oldValue == 0)
                    return;
            } while (Interlocked.CompareExchange(ref fValue, 0, oldValue) != oldValue);
        }

        /// <summary>
        /// Atomically pefroms a computation and assigns it to the atomic value
        /// </summary>
        /// <param name="computation">Computation to execute atomically</param>
        public void Assign(Func<bool, bool> computation)
        {
            int oldValue;
            int newValue;

            do
            {
                oldValue = fValue;
                newValue = computation(oldValue == 0 ? false : true) ? 1 : 0;
            } while (Interlocked.CompareExchange(ref fValue, newValue, oldValue) != oldValue);
        }

        /// <summary>
        /// If the current value == the expected value of Atomically set the value to the given updated value.
        /// </summary>
        /// <param name="value1"></param>
        /// <param name="value2"></param>
        /// <returns></returns>
        public bool CompareExchange(bool value1, bool value2)
        {
            int iSetValue = value1 ? 1 : 0;
            int iCompareValue = value2 ? 1 : 0;

            return Interlocked.CompareExchange(ref fValue, iSetValue, iCompareValue) == iCompareValue;
        }

        /// <summary>
        /// Performs an atomic negation of the value (Unary operator "!")
        /// </summary>
        public void Invert()
        {
            int oldValue;
            int newValue;

            do
            {
                oldValue = fValue;
                newValue = fValue == 0 ? 1 : 0;
            } while (Interlocked.CompareExchange(ref fValue, newValue, oldValue) != oldValue);
        }

        /// <summary>
        /// Compares the equality of two values.
        /// </summary>
        /// <param name="left">Left parameter to compare.</param>
        /// <param name="right">Right parameter to compare.</param>
        /// <returns>Returns whether left and right parameters are equals or not.</returns>
        public static bool operator ==(AtomicBoolean left, AtomicBoolean right)
        {
            return left.fValue == right.fValue;
        }

        /// <summary>
        /// Compares the equality of two values.
        /// </summary>
        /// <param name="left">Left parameter to compare.</param>
        /// <param name="right">Right parameter to compare.</param>
        /// <returns>Returns whether left and right parameters are equals or not.</returns>
        public static bool operator ==(AtomicBoolean left, bool right)
        {
            return left.Value == right;
        }

        /// <summary>
        /// Checks for inequality of two values.
        /// </summary>
        /// <param name="left">Left parameter to compare.</param>
        /// <param name="right">Right parameter to compare.</param>
        /// <returns>Returns true whether left and right parameters are not equals, otherwise false.</returns>
        public static bool operator !=(AtomicBoolean left, AtomicBoolean right)
        {
            return left.fValue != right.fValue;
        }

        /// <summary>
        /// Checks for inequality of two values.
        /// </summary>
        /// <param name="left">Left parameter to compare.</param>
        /// <param name="right">Right parameter to compare.</param>
        /// <returns>Returns true whether left and right parameters are not equals, otherwise false.</returns>
        public static bool operator !=(AtomicBoolean left, bool right)
        {
            return left.Value != right;
        }

        /// <summary>
        /// Converts the atomic value to a non-atomic one.
        /// </summary>
        /// <param name="value">The value to convert.</param>
        /// <returns>The converted value.</returns>
        public static explicit operator bool(AtomicBoolean value)
        {
            return value.Value;
        }

        /// <summary>
        /// Converts the non-atomic value to an atomic one.
        /// </summary>
        /// <param name="value">The value to convert.</param>
        /// <returns>The converted value.</returns>
        public static implicit operator AtomicBoolean(bool value)
        {
            return new AtomicBoolean(value);
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
            AtomicBoolean atom = (AtomicBoolean)obj;
            if (atom == default(AtomicBoolean))
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