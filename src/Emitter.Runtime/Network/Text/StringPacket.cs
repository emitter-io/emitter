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
using System.Text;

namespace Emitter.Network
{
    /// <summary>
    /// Defines a custom packet compiled with a StringBuilder.
    /// </summary>
    public class StringPacket : Packet
    {
        private StringBuilder fBuilder = new StringBuilder();
        private Encoding fEncoding;
        private string fInitialValue = null;

        /// <summary>
        /// Constructs a new <see cref="StringPacket"/> instance.
        /// </summary>
        /// <param name="encoding">The encoding used to write the string values.</param>
        public StringPacket(Encoding encoding) : base()
        {
            fBuilder = new StringBuilder();
            fEncoding = encoding;
        }

        /// <summary>
        /// Constructs a new <see cref="StringPacket"/> instance.
        /// </summary>
        /// <param name="encoding">The encoding used to write the string values.</param>
        /// <param name="initialValue">Initial string value to write in this packet.</param>
        public StringPacket(Encoding encoding, string initialValue) : base()
        {
            fBuilder = new StringBuilder();
            fBuilder.Append(initialValue);
            fInitialValue = initialValue;
            fEncoding = encoding;
        }

        /// <summary>
        /// Gets the string builder used to compile this packet.
        /// </summary>
        public StringBuilder StringBuilder
        {
            get { return fBuilder; }
            protected set { fBuilder = value; }
        }

        /// <summary>
        /// Gets the encoding used to compile the packet.
        /// </summary>
        public Encoding Encoding
        {
            get { return fEncoding; }
            protected set { fEncoding = value; }
        }

        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        public override void Recycle()
        {
            base.Recycle();
            fBuilder.Length = 0;

            if (fInitialValue != null)
                fBuilder.Append(fInitialValue);
        }
    }
}