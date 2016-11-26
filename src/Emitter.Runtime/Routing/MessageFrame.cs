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
using System.Runtime.CompilerServices;
using Emitter.Collections;

namespace Emitter
{
    /// <summary>
    /// Represents a message frame that contains multiple messages.
    /// </summary>
    public sealed class MessageFrame : RecyclableObject
    {
        #region Static Members

        [ThreadStatic]
        private static ParseInfo Parser;

        /// <summary>
        /// Scratch wrappers.
        /// </summary>
        private static readonly ConcurrentPool<MessageFrame> Pool =
            new ConcurrentPool<MessageFrame>("Message Frames", (c) => new MessageFrame());

        /// <summary>
        /// Acquires a new frame from the pool.
        /// </summary>
        /// <returns></returns>
        public static MessageFrame Acquire()
        {
            return Pool.Acquire();
        }

        #endregion Static Members

        #region Constructors

        /// <summary>
        /// The internal buffer.
        /// </summary>
        private byte[] Buffer;

        /// <summary>
        /// The current offset in the buffer.
        /// </summary>
        private int Index = 0;

        /// <summary>
        /// How many messages are appended in the frame.
        /// </summary>
        private int AppendCount = 0;

        /// <summary>
        /// Constructs a new instance of the builder.
        /// </summary>
        public MessageFrame()
        {
            this.Buffer = new byte[EmitterConst.FrameSize];
        }

        /// <summary>
        /// Recycles the frame so it can be used once again.
        /// </summary>
        public override void Recycle()
        {
            this.Index = 0;
            this.AppendCount = 0;
        }

        /// <summary>
        /// Gets the underlying buffer.
        /// </summary>
        public byte[] Array
        {
            get { return this.Buffer; }
        }

        /// <summary>
        /// Gets the current size of the buffer.
        /// </summary>
        public int Length
        {
            get { return this.Index; }
        }

        /// <summary>
        /// Gets the current amount of appended messages.
        /// </summary>
        public int Messages
        {
            get { return this.AppendCount; }
        }

        #endregion Constructors

        #region Public Members

        /// <summary>
        /// Appends a message to the current frame.
        /// </summary>
        /// <param name="contract"></param>
        /// <param name="channel"></param>
        /// <param name="message"></param>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public unsafe bool TryAppendMessage(int contract, string channel, ArraySegment<byte> message)
        {
            // How big is the message?
            var msgLen = message.Count;
            var length = 1 + 4 + channel.Length + 1 + 4 + msgLen;
            if (this.Index + length > this.Buffer.Length)
                return false;

            // Append the contract bytes first
            this.Buffer[this.Index++] = 1;
            this.Buffer[this.Index++] = (byte)(contract >> 24);
            this.Buffer[this.Index++] = (byte)(contract >> 16);
            this.Buffer[this.Index++] = (byte)(contract >> 8);
            this.Buffer[this.Index++] = (byte)contract;

            // Append the channel string now
            for (int i = 0; i < channel.Length; ++i)
                this.Buffer[this.Index++] = (byte)channel[i];
            this.Buffer[this.Index++] = 0; // Terminate by /0

            // Append the message length
            this.Buffer[this.Index++] = (byte)(msgLen >> 24);
            this.Buffer[this.Index++] = (byte)(msgLen >> 16);
            this.Buffer[this.Index++] = (byte)(msgLen >> 8);
            this.Buffer[this.Index++] = (byte)msgLen;

            // Copy the message in
            Memory.Copy(message.Array, message.Offset, this.Buffer, this.Index, message.Count);
            Index += message.Count;

            // Increase the append count
            AppendCount++;
            return true;
        }

        /// <summary>
        /// Writes a native buffer to the frame.
        /// </summary>
        /// <param name="buffer"></param>
        /// <param name="length"></param>
        public static unsafe void TryParse(ArraySegment<byte> frame)
        {
            // Allocate if we need to
            if (Parser == null)
                Parser = new ParseInfo();

            // Since we're having some nice assurances by nanomsg, just set to begin
            Parser.State = ParseState.OP_HEADER;

            // Set the length
            var length = frame.Count;

            // Pin the buffer so the GC won't move it
            fixed (byte* pArray = frame.Array)
            {
                byte* pBuffer = (pArray + frame.Offset);
                byte b; int i;
                for (i = 0; i < length; ++i)
                {
                    // Get the current byte
                    b = *(pBuffer + i);

                    // Switch over the current parser state
                    switch (Parser.State)
                    {
                        // We need to read the contract number, simply an int.
                        case ParseState.OP_HEADER:
                            // Read the first byte, which tells us whether this is a full message or a reference
                            Parser.State = ParseState.OP_CONTRACT;
                            break;

                        // We need to read the contract number, simply an int.
                        case ParseState.OP_CONTRACT:

                            // Read the contract first
                            Parser.Contract = (*(pBuffer + i)) << 24
                                | (*(pBuffer + i + 1) << 16)
                                | (*(pBuffer + i + 2) << 8)
                                | (*(pBuffer + i + 3));

                            i += 4;
                            Parser.State = ParseState.OP_CHANNEL;
                            break;

                        // Start reading the channel, remember the offset
                        case ParseState.OP_CHANNEL:
                            Parser.Offset = i;
                            Parser.State = ParseState.OP_CHANNEL_CONTENT;
                            break;

                        // Channel is delimited by \0, so we need to wait until we have it
                        case ParseState.OP_CHANNEL_CONTENT:
                            if (b == (byte)0)
                            {
                                //Parser.Prefix = new string((sbyte*)pBuffer, Parser.Offset - 5, i - Parser.Offset + 5);
                                //Parser.Channel = new string((sbyte*)pBuffer, Parser.Offset - 1, i - Parser.Offset);

                                Parser.Prefix = Memory.CopyString(pBuffer + Parser.Offset - 5, i - Parser.Offset + 5);
                                Parser.Channel = Memory.CopyString(pBuffer + Parser.Offset - 1, i - Parser.Offset);
                                Parser.State = ParseState.OP_MESSAGE_LENGTH;
                            }
                            break;

                        // Reading the message length
                        case ParseState.OP_MESSAGE_LENGTH:

                            // Read the contract first
                            Parser.MessageLength = (*(pBuffer + i)) << 24
                                | (*(pBuffer + i + 1) << 16)
                                | (*(pBuffer + i + 2) << 8)
                                | (*(pBuffer + i + 3));

                            i += 4;
                            Parser.State = ParseState.OP_MESSAGE;
                            break;

                        // Reading the message body
                        case ParseState.OP_MESSAGE:

                            // Skip the bytes, simply return the segment
                            var contract = Parser.Contract;
                            var channel = Parser.Channel;
                            var message = new ArraySegment<byte>(
                                frame.Array, frame.Offset + i - 1, Parser.MessageLength
                                );

                            // Invoke the callback
                            i += Parser.MessageLength;

                            // Send the message to current subscriptions
                            Dispatcher.ForwardToClients(contract, channel, message);
                            Parser.State = ParseState.OP_HEADER;

                            // Console.WriteLine("user={0} channel={1} msg={2}", Parser.Contract, Parser.Channel, Parser.Message);
                            break;

                        default:
                            goto ParseError;
                    }
                }
            }

            return;

            ParseError:
            Console.WriteLine("Parse error");
        }

        /// <summary>
        /// Returns the underlying byte segment.
        /// </summary>
        /// <returns>The byte segment</returns>
        public ArraySegment<byte> AsSegment()
        {
            return new ArraySegment<byte>(this.Buffer, 0, this.Length);
        }

        #endregion Public Members

        #region Parser Structures

        internal sealed class ParseInfo
        {
            /// <summary>
            /// Gets the current state of the parser.
            /// </summary>
            public ParseState State;

            /// <summary>
            /// Gets the currently parsed contract
            /// </summary>
            public int Contract;

            /// <summary>
            /// Gets the currently parsed channel.
            /// </summary>
            public string Channel;

            /// <summary>
            /// Gets the currently parsed prefix (contract + channel).
            /// </summary>
            public string Prefix;

            /// <summary>
            /// Gets the message length.
            /// </summary>
            public int MessageLength;

            /// <summary>
            /// Gets the offset
            /// </summary>
            public int Offset;
        }

        internal enum ParseState
        {
            OP_HEADER = 0,
            OP_CONTRACT = 1,
            OP_CHANNEL = 2,
            OP_CHANNEL_CONTENT = 3,
            OP_MESSAGE_LENGTH = 4,
            OP_MESSAGE = 5
        }

        #endregion Parser Structures
    }
}