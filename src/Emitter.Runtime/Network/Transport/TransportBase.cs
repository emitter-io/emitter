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
using Emitter.Threading;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a base class for an engine.io transport implementation.
    /// </summary>
    internal abstract class TransportBase : DisposableObject, ITransport
    {
        /// <summary>
        /// The current transport state
        /// </summary>
        private AtomicByte ReadyState = 0;

        /// <summary>
        /// Gets the encoder for this transport.
        /// </summary>
        public abstract Processor Encoder { get; }

        /// <summary>
        /// Gets the decoder for this transport.
        /// </summary>
        public abstract Processor Decoder { get; }

        /// <summary>
        /// Gets or sets a channel that can be used for sending the data back to the remote client.
        /// </summary>
        public virtual Connection Channel { get; set; }

        /// <summary>
        /// Gets or sets the CORS Origin.
        /// </summary>
        public string Origin { get; set; }

        /// <summary>
        /// Gets or sets the state of the transport.
        /// </summary>
        public TransportState State
        {
            get { return (TransportState)this.ReadyState.Value; }
            set { this.ReadyState.Assign((byte)value); }
        }

        /// <summary>
        /// Writes a packet to the transport.
        /// </summary>
        /// <param name="packet">The packet to send.</param>
        /// <returns>Whether the packet was written successfully or not.</returns>
        public abstract bool Write(Packet packet);

        /// <summary>
        /// Returns a string representation of the transport.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            return "["
                + this.GetType().Name +
                (this.Channel != null ? ": " + this.Channel.RemoteEndPoint.ToString() : "") +
                "]";
        }

        /// <summary>
        /// Occurs when the transport is about to be disposed.
        /// </summary>
        /// <param name="disposing">Whether we are disposing or finalizing.</param>
        protected override void Dispose(bool disposing)
        {
            // Close the bound transport channel as well
            //if (this.Channel != null && this.Channel.IsRunning)
            //    this.Channel.Close();

            this.State = TransportState.Closed;

            // Call the base
            base.Dispose(disposing);
        }
    }
}