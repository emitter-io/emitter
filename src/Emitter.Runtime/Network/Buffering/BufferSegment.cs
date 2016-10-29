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

#define POOL_BUFFERSEGMENT
//#define TRACE_BUFFERSEGMENT

using System;
using System.Linq;
using System.Text;
using Emitter.Collections;
using Emitter.Network.Http;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a segment of a buffer from a <see cref="BufferProvider"/>.
    /// </summary>
    public unsafe sealed class BufferSegment : RecyclableObject
    {
        #region Constructors

        /// <summary>
        /// THe main pool for buffer segments
        /// </summary>
        internal static readonly ConcurrentPool<BufferSegment> SegmentPool =
            new ConcurrentPool<BufferSegment>("Buffers", (c) =>
            {
#if TRACE_BUFFERSEGMENT
                var segment = new BufferSegment();
                if (BufferSegment.All == null)
                    BufferSegment.All = new ConcurrentBag<BufferSegment>();
                BufferSegment.All.Add(segment);
                return segment;
#else
                return new BufferSegment();
#endif
            }, 64);

#if TRACE_BUFFERSEGMENT
        /// <summary>
        /// Tracing bag.
        /// </summary>
        internal static ConcurrentBag<BufferSegment> All = new ConcurrentBag<BufferSegment>();
#endif

        /// <summary>
        /// Constructs a new <see cref="BufferSegment"/> structure.
        /// </summary>
        private BufferSegment() { }

        /// <summary>
        /// This creates a buffer segment in a non-pooled mode, with an internal buffer attached to it.
        /// </summary>
        /// <param name="size"></param>
        public BufferSegment(int size)
        {
            if (size <= 0)
            {
                this.Offset = 0;
                this.Handle = -1;
            }
            else
            {
                this.Array = new byte[size];
                this.Offset = 0;
                this.Length = size;
                this.Handle = -1;
            }
        }

        /// <summary>
        /// Constructs a new <see cref="BufferSegment"/> structure.
        /// </summary>
        /// <param name="owner">The owner of this segment.</param>
        /// <param name="origin">The origin buffer.</param>
        /// <param name="offset">The starting position.</param>
        /// <param name="length">The number of bytes.</param>
        /// <param name="handle">The offset in the original array.</param>
        internal static BufferSegment Acquire(BufferProvider owner, byte[] origin, int offset, int length, int handle)
        {
            var segment = SegmentPool.Acquire();

            segment.Owner = owner;
            segment.Array = origin;
            segment.Offset = offset;
            segment.Length = length;
            segment.Handle = handle;

#if TRACE_BUFFERSEGMENT
            segment.AcquiredAt = Environment.StackTrace;
            segment.ReleasedAt = String.Empty;
#endif

            return segment;
        }

        /// <summary>
        /// Wraps an existing buffer.
        /// </summary>
        /// <param name="buffer">The buffer to wrap with this segment.</param>
        /// <param name="offset">The offset to set.</param>
        /// <param name="length">The length to set.</param>
        public void Wrap(byte[] buffer, int offset, int length)
        {
            this.Array = buffer;
            this.Offset = offset;
            this.Length = length;
        }

        #endregion Constructors

        /// <summary>
        /// Gets the owner buffer. We keep this in order to be able to release the segment back.
        /// </summary>
        internal BufferProvider Owner;

        /// <summary>
        /// Gets the reference to the byte array. This is done in order to prevent the memory being
        /// collected by the GC until the BufferSegment is disposed. Therefore, we can perform a byte*
        /// read on an old copy of the data, and write back to the new memory table.
        /// </summary>
        internal byte[] Array;

        /// <summary>
        /// Gets the offset in the managed buffer.
        /// </summary>
        internal int Handle;

        /// <summary>
        /// Gets the offset in the original array.
        /// </summary>
        internal int Offset;

        /// <summary>
        /// Gets the number of elements in the range delimited by the array segment.
        /// </summary>
        internal int Length;

        /// <summary>
        /// Gets or sets the number of elements in the range delimited by the array segment.
        /// </summary>
        public int Size
        {
            get { return this.Length; }
            set { this.Length = value; }
        }

        /// <summary>
        /// Gets whether this buffer is acquired or not.
        /// </summary>
        public bool IsAcquired
        {
            get { return this.ObjectAcquired; }
        }

        /// <summary>
        /// Gets the segment as a ASCII string, for debug view.
        /// </summary>
        internal string ViewAsString
        {
            get { return Encoding.ASCII.GetString(this.Array, this.Offset, this.Length); }
        }

        /// <summary>
        /// Gets the segment as a ASCII string, for debug view.
        /// </summary>
        internal string ViewAsShortString
        {
            get
            {
                try
                {
                    return Encoding.ASCII.GetString(this.Array, this.Offset, Math.Min(this.Length, 50));
                }
                catch (Exception ex)
                {
                    return "(error: " + ex.Message + ")";
                }
            }
        }

        /// <summary>
        /// Gets a byte fiew for the first 10 bytes.
        /// </summary>
        internal string ViewAsBytes
        {
            get
            {
                try
                {
                    const int size = 10;
                    var view = "[";
                    for (int i = 0; i < this.Length && i < size; ++i)
                    {
                        view += this.Array[this.Offset + i];
                        if (i < size - 1)
                            view += ", ";
                    }
                    return view + "]";
                }
                catch (Exception ex)
                {
                    return "(error: " + ex.Message + ")";
                }
            }
        }

        /// <summary>
        /// Gets a byte fiew for the first 20 bytes.
        /// </summary>
        internal string ViewAsSsl
        {
            get
            {
                var b0 = this.Array[this.Offset];
                var b1 = this.Array[this.Offset + 1];
                var b2 = this.Array[this.Offset + 2];
                var b3 = this.Array[this.Offset + 3];
                var b4 = this.Array[this.Offset + 4];
                var b5 = this.Array[this.Offset + 5];

                string content = null;
                string version = null;
                string message = null;

                switch (b0)
                {
                    case 0x14: content = "ChangeCipherSpec"; break;
                    case 0x15: content = "Alert"; break;
                    case 0x16: content = "Handshake"; break;
                    case 0x17: content = "Application"; break;
                    case 0x18: content = "Heartbeat"; break;
                }

                if (b1 == 3 && b2 == 0) version = "SSL 3.0";
                if (b1 == 3 && b2 == 1) version = "TLS 1.0";
                if (b1 == 3 && b2 == 2) version = "TLS 1.1";
                if (b1 == 3 && b2 == 3) version = "TLS 1.2";

                if (b0 == 0x16)
                {
                    switch (b5)
                    {
                        case 0: message = "HelloRequest"; break;
                        case 1: message = "ClientHello"; break;
                        case 2: message = "ServerHello"; break;
                        case 4: message = "NewSessionTicket"; break;
                        case 11: message = "Certificate"; break;
                        case 12: message = "ServerKeyExchange"; break;
                        case 13: message = "CertificateRequest"; break;
                        case 14: message = "ServerHelloDone"; break;
                        case 15: message = "CertificateVerify"; break;
                        case 16: message = "ClientKeyExchange"; break;
                        case 20: message = "Finished"; break;
                    }
                }

                var length = (ushort)((b3 << 8) | b4);

                if (content == null)
                    return String.Format("Content: {0}", this.ViewAsString);
                return String.Format("{0}: {1}{2} of {3} bytes, buffer of {4} bytes", version, content, message != null ? ", " + message : String.Empty, length, this.Length);
            }
        }

        /// <summary>
        /// Gets the hybi13 header.
        /// </summary>
        internal string ViewAsHybi13
        {
            get
            {
                var b0 = this.Array[this.Offset];
                var b1 = this.Array[this.Offset + 1];

                // Get the first part of the header
                var frameType = (WebSocketFrameType)(b0 & 15);
                var isFinal = (b0 & 128) != 0;

                // Get the second part of the header
                var isMasked = (b1 & 128) != 0;
                var dataLength = (b1 & 127);

                return isMasked ? String.Format("Hybi13 {0} of {1} bytes ({2})", frameType, dataLength, isFinal ? "complete" : "incomplete") : "(Invalid: Hybi13)";
            }
        }

        /// <summary>
        /// View the packet as MQTT
        /// </summary>
        internal unsafe string ViewAsMQTT
        {
            get
            {
                var pBuffer = this.AsBytePointer();
                var type = (MqttPacketType)((*pBuffer & MqttPacket.MSG_TYPE_MASK) >> MqttPacket.MSG_TYPE_OFFSET);

                int multiplier = 1;
                int length = 0;
                int digit = 0;
                int headerSize = 1;
                do
                {
                    digit = *(pBuffer++);
                    length += ((digit & 127) * multiplier);
                    multiplier *= 128;
                    ++headerSize;
                } while ((digit & 128) != 0);

                return string.Format("{0}: size={1}", type.ToString().ToUpper(), length);
            }
        }

        /// <summary>
        /// Converts this <see cref="BufferSegment"/> to an ArraySegment.
        /// </summary>
        /// <returns>The byte array segment corresponding to this <see cref="BufferSegment"/>.</returns>
        public ArraySegment<byte> AsSegment()
        {
            if (!this.ObjectAcquired)
                throw new ObjectDisposedException("BufferSegment");

            return new ArraySegment<byte>(this.Array, this.Offset, this.Length);
        }

        /// <summary>
        /// Converts the buffer segment to a UTF-8 string.
        /// </summary>
        /// <returns>The string representation of the buffer.</returns>
        public string AsString()
        {
            if (!this.ObjectAcquired)
                throw new ObjectDisposedException("BufferSegment");

            return Encoding.UTF8.GetString(this.Array, this.Offset, this.Length);
        }

        /// <summary>
        /// Converts this <see cref="BufferSegment"/> to an array of bytes.
        /// </summary>
        /// <returns>The byte array segment corresponding to this <see cref="BufferSegment"/>.</returns>
        public byte[] AsArray()
        {
            if (!this.ObjectAcquired)
                throw new ObjectDisposedException("BufferSegment");

            var array = new byte[this.Length];
            Memory.Copy(this.Array, this.Offset, array, 0, this.Length);
            return array;
        }

        /// <summary>
        /// Gets a pointer <see cref="BufferSegment"/> to the buffer.
        /// </summary>
        /// <returns>The byte pointer corresponding to the beginning of this <see cref="BufferSegment"/>.</returns>
        public byte* AsBytePointer()
        {
            if (!this.ObjectAcquired)
                throw new ObjectDisposedException("BufferSegment");

            fixed (byte* pOrigin = this.Array)
                return pOrigin + this.Offset;
        }

        /// <summary>
        /// Splits the current segment in two and returns the tailing segment.
        /// </summary>
        /// <param name="offset">The starting offset of the subsegment.</param>
        /// <returns>Returns a tailing segment.</returns>
        public BufferSegment Split(int offset)
        {
            if (!this.ObjectAcquired)
                throw new ObjectDisposedException("BufferSegment");
            if (this.Owner == null)
                throw new InvalidOperationException("This operation requires the BufferSegment to be bound to a valid BufferProvider.");

            // Get a new sub-segment
            var splitLen = this.Length - offset;
            if (splitLen == 0)
                return null;

            var newSegment = this.Owner.Reserve(splitLen);

            // Copy the tail to the new segment
            Memory.Copy(this.Array, this.Offset + offset, newSegment.Array, newSegment.Offset, splitLen);

            // Change the length of the curent segment
            this.Length = offset;

            // Notify that we have two segments now instead of one
            //this.Owner.Increment(this.Handle);

            // Return the tailing segment
            return newSegment;
        }

        /// <summary>
        /// Joins the current segment with another segment to produce a new one, merged segment.
        /// </summary>
        /// <param name="segmentRight">A segment to merge with.</param>
        /// <returns>Returns a merged segment.</returns>
        public BufferSegment Join(BufferSegment segmentRight)
        {
            if (!this.ObjectAcquired)
                throw new ObjectDisposedException("BufferSegment");
            if (this.Owner == null)
                throw new InvalidOperationException("This operation requires the BufferSegment to be bound to a valid BufferProvider.");

            var newLength = this.Length + segmentRight.Length;
            var newSegment = this.Owner.Reserve(newLength);

            // Copy the segments into a new one
            Memory.Copy(this.Array, this.Offset, newSegment.Array, newSegment.Offset, this.Length);
            Memory.Copy(segmentRight.Array, segmentRight.Offset, newSegment.Array, newSegment.Offset + this.Length, segmentRight.Length);

            return newSegment;
        }

        #region Debug Traces
#if TRACE_BUFFERSEGMENT
        private string ReleasedAt = String.Empty;
        private string AcquiredAt = String.Empty;

        /// <summary>
        /// Gets the acquire trace.
        /// </summary>
        public string[] AcquireTrace
        {
            get
            {
                return this.AcquiredAt.Split(new string[] { "\r\n" }, StringSplitOptions.RemoveEmptyEntries)
                    .Select(s => s.Trim())
                    .ToArray();
            }
        }

        /// <summary>
        /// Gets the release trace.
        /// </summary>
        public string[] ReleaseTrace
        {
            get
            {
                return this.ReleasedAt.Split(new string[]{"\r\n"},StringSplitOptions.RemoveEmptyEntries)
                    .Select(s => s.Trim())
                    .ToArray();
            }
        }
#endif
        #endregion Debug Traces

        #region IDisposable Members

        /// <summary>
        /// Recycles the object.
        /// </summary>
        public override void Recycle()
        {
#if TRACE_BUFFERSEGMENT
            this.ReleasedAt = Environment.StackTrace;
#endif

            // If this is not bound to a buffer provider, do not recycle
            if (this.Owner == null)
                return;

            // Release the owner
            this.Owner.Release(this);

            // Reset the properties
            this.Owner = null;
            this.Array = null;
            this.Offset = -1;
            this.Handle = -1;
            this.Length = 0;
        }

        #endregion IDisposable Members
    }
}