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
using System.Runtime.InteropServices;
using System.Threading;
using Emitter.Text;

namespace Emitter
{
    /// <summary>
    /// Represents a connection id, globally unique.
    /// </summary>
    [StructLayout(LayoutKind.Sequential)]
    public struct ConnectionId : IComparable, IComparable<ConnectionId>, IEquatable<ConnectionId>
    {
        #region Static Members

        /// <summary>
        /// Represents an empty connection id.
        /// </summary>
        public readonly static ConnectionId Empty = new ConnectionId(0, 0);

        /// <summary>
        /// The first connection id, with restart tolerance.
        /// </summary>
        private static readonly long FirstConnectionId = DateTime.UtcNow.Ticks;

        /// <summary>
        /// The current connection id.
        /// </summary>
        private static long LastConnectionId = FirstConnectionId;

        /// <summary>
        /// This creates a new connection identifier and returns it atomically.
        /// </summary>
        /// <returns></returns>
        public static ConnectionId NewConnectionId()
        {
            return new ConnectionId(
                Service.Mesh.Identifier,
                Interlocked.Increment(ref LastConnectionId)
                );
        }

        #endregion Static Members

        #region Member Variables

        /// <summary>
        /// Gets or sets the node id of this connection id.
        /// </summary>
        private int Global;

        /// <summary>
        /// Gets or sets the index of this connection within the node.
        /// </summary>
        private long Local;

        #endregion Member Variables

        #region Calculated Properties

        /// <summary>
        /// Gets the debug id.
        /// </summary>
        internal int DebugView
        {
            get { return (int)(this.Local - FirstConnectionId); }
        }

        #endregion Calculated Properties

        #region Constructors

        /// <summary>
        /// Constructs a new connection id.
        /// </summary>
        /// <param name="global">The globaly unique part.</param>
        /// <param name="local">The locally unique part.</param>
        public ConnectionId(int global, long local)
        {
            this.Global = global;
            this.Local = local;
        }

        #endregion Constructors

        #region Comparable Members and Operators

        /// <summary>
        /// Defines the equality operator.
        /// </summary>
        /// <param name="a"></param>
        /// <param name="b"></param>
        /// <returns></returns>
        public static bool operator ==(ConnectionId a, ConnectionId b)
        {
            // Now compare each of the elements
            if (a.Global != b.Global)
                return false;
            if (a.Local != b.Local)
                return false;
            return true;
        }

        /// <summary>
        /// Defines the inequality operator.
        /// </summary>
        /// <param name="a"></param>
        /// <param name="b"></param>
        /// <returns></returns>
        public static bool operator !=(ConnectionId a, ConnectionId b)
        {
            return !(a == b);
        }

        /// <summary>
        /// Compares two connection ids.
        /// </summary>
        /// <param name="other"></param>
        /// <returns></returns>
        public int CompareTo(ConnectionId other)
        {
            if (this.Global != other.Global)
                return (this.Global < other.Global) ? -1 : 1;
            if (this.Local != other.Local)
                return (this.Local < other.Local) ? -1 : 1;
            return 0;
        }

        /// <summary>
        /// Compares two connection ids.
        /// </summary>
        /// <param name="other"></param>
        /// <returns></returns>
        public int CompareTo(object value)
        {
            if (value == null)
                return 1;
            if (!(value is ConnectionId))
                throw new ArgumentException("value");

            return this.CompareTo((ConnectionId)value);
        }

        /// <summary>
        /// Checks whether two connection ids are equal or not.
        /// </summary>
        /// <param name="other"></param>
        /// <returns></returns>
        public bool Equals(ConnectionId other)
        {
            if (other == null)
                return false;

            return this == other;
        }

        /// <summary>
        /// Checks whether two connection ids are equal or not.
        /// </summary>
        /// <param name="other"></param>
        /// <returns></returns>
        public override bool Equals(object other)
        {
            if (other == null)
                return false;
            if (!(other is ConnectionId))
                throw new ArgumentException("value");

            return this == (ConnectionId)other;
        }

        /// <summary>
        /// Overrides the hash code.
        /// </summary>
        /// <returns></returns>
        public override int GetHashCode()
        {
            // Simply XOR all the bits
            return (int)(this.Global ^ this.Local);
        }

        #endregion Comparable Members and Operators

        #region ToString Members

        /// <summary>
        /// Converts the connection id to a string representation
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            //return this.Global.ToHex() + "-" + this.Local.ToHex();
            unsafe
            {
                fixed (int* ptr = &Global)
                {
                    // Get base64 representation of the struct
                    return Base64.Default.ToBase((byte*)ptr, 12);
                }
            }
        }

        #endregion ToString Members
    }
}