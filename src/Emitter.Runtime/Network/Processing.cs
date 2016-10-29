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

namespace Emitter.Network
{
    /// <summary>
    /// Processes the packet within the context. Returns true whether the packet was processed or throttled.
    /// </summary>
    /// <param name="channel">The through which the packet is coming/going out.</param>
    /// <param name="context">The packet context for this operation.</param>
    /// <returns>True whether the packet was processed or throttled, false otherwise.</returns>
    public delegate ProcessingState Processor(Emitter.Connection channel, ProcessingContext context);

    /// <summary>
    /// Represents a process result for an <see cref="Processor"/>.
    /// </summary>
    public enum ProcessingState : short
    {
        /// <summary>
        /// A successfully decoded/encoded packet.
        /// </summary>
        Success,

        /// <summary>
        /// An unrecognized packet received or encoding failed.
        /// </summary>
        Failure,

        /// <summary>
        /// A pipeline have been successfully completed and should not be executed further.
        /// </summary>
        Stop,

        /// <summary>
        /// Insufficient data has been received and the buffer should be throttled.
        /// </summary>
        InsufficientData,

        /// <summary>
        /// The buffer segment was not processed and will be processed during the next iteration.
        /// </summary>
        HandleLater
    }

    /// <summary>
    /// Represents a processing type for an <see cref="Processor"/>.
    /// </summary>
    public enum ProcessingType : short
    {
        /// <summary>
        /// Specifies that the processor is used for encoding outgoing packets.
        /// </summary>
        Encoding,

        /// <summary>
        /// Specifies that the processor is used for decoding incoming packets.
        /// </summary>
        Decoding
    }

    /// <summary>
    /// Represents an exception which has occured during either decoding or encoding.
    /// </summary>
    public class ProcessingException : Exception
    {
        /// <summary>
        /// Constructs a new instance of <see cref="ProcessingException"/>.
        /// </summary>
        public ProcessingException() { }

        /// <summary>
        /// Constructs a new instance of <see cref="ProcessingException"/>.
        /// </summary>
        /// <param name="message">A message of the exception.</param>
        public ProcessingException(string message) : base(message) { }

        /// <summary>
        /// Constructs a new instance of <see cref="ProcessingException"/>.
        /// </summary>
        /// <param name="message">A message of the exception.</param>
        /// <param name="inner">An inner exception.</param>
        public ProcessingException(string message, Exception inner) : base(message, inner) { }
    }
}