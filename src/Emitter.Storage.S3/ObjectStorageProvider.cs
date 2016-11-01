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
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Threading.Tasks;
using Emitter.Providers;

namespace Emitter.Storage
{
    /// <summary>
    /// The log writer that manages multiple file streams.
    /// </summary>
    public sealed class ObjectStorageProvider : StorageProvider
    {
        #region Private Properties

        /// <summary>
        /// Gets the time offset (1st jan 2010);
        /// </summary>
        internal const long TimeOffset = 633979008000000000;

        /// <summary>
        /// Quick encoding table for the time.
        /// </summary>
        private readonly char[] TimeEncode = new char[] {
            '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
            'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
            'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
            'U', 'V', 'W', 'X', 'Y', 'Z',
            'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
            'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
            'u', 'v', 'w', 'x', 'y', 'z',
            '+', '/'};

        // The indexer and the storage
        private readonly ObjectIndex Index;

        private readonly ObjectStorage Store;
        private readonly AppendLogCache Append;

        /// <summary>
        /// The ip address bytes.
        /// </summary>
        private readonly char[] Address;

        #endregion Private Properties

        #region Constructors

        /// <summary>
        /// Constructs the new provider.
        /// </summary>
        public ObjectStorageProvider()
        {
            this.Index = new ObjectIndex();
            this.Store = new ObjectStorage(this.Index);
            this.Append = new AppendLogCache();

            // Get the external address
            var addr = Service.Providers
                .Resolve<AddressProvider>()
                .GetExternal();
            this.Address = Convert.ToBase64String(
                addr.GetAddressBytes()
                ).ToCharArray();
        }

        #endregion Constructors

        #region StorageProvider Members

        /// <summary>
        /// Asynchronously stores a message in the store.
        /// </summary>
        /// <param name="contract">The contract.</param>
        /// <param name="ssid">The subscription id.</param>
        /// <param name="message">The message payload to store.</param>
        /// <param name="ttl">The time to live for this message, in seconds.</param>
        /// <returns>The task for completion notification.</returns>
        public override async Task AppendAsync(int contract, uint[] ssid, int ttl, ArraySegment<byte> message)
        {
            // Get the path for the current write
            var time = DateTime.UtcNow;
            var channel = ssid[1];

            // Get the stream we operate on
            var current = this.Append.GetCurrent(contract);

            // Do we have a previous stream?
            var oldStream = current.OutputStream;
            var oldExpire = current.ExpireDate;
            var oldPath = current.OutputPath;

            // Construct a new filename
            var filename = GetSource(ref time);
            var key = GetKey(contract, time, filename);

            // Set the filename, if no changes were made to this, it will do the write on the same
            // underlying file stream. Otherwise, it will create a new file stream.
            if (current.TryRedirect("data/" + key))
            {
                // Flush the stream and dispose it
                oldStream?.FlushAsync()
                    .ContinueWith(t => oldStream.Dispose())
                    .ContinueWith(t => this.Store.UploadAsync(oldPath, oldExpire));
            }

            // Forward the write and check where the message was written
            var offset = await current.AppendAsync(message, ttl <= 0 ? DateTime.MinValue : time.AddSeconds(ttl));

            // Store to the index log
            this.Index.IndexAsync(contract, ssid, filename, ttl, offset, message.Count);
        }

        /// <summary>
        /// Constructs the S3 key.
        /// </summary>
        private static string GetKey(int contract, DateTime time, string source)
        {
            return contract.ToHex() + "/" + time.ToString("yyyyMMdd") + "/" + source + ".bin";
        }

        /// <summary>
        /// Gets the filename source for the stream.
        /// </summary>
        /// <param name="time">The time to use.</param>
        /// <returns>The source string</returns>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private unsafe string GetSource(ref DateTime time)
        {
            var name = stackalloc char[10];
            name[0] = Address[0];
            name[1] = Address[1];
            name[2] = Address[2];
            name[3] = Address[3];
            name[4] = Address[4];
            name[5] = Address[5];
            name[6] = TimeEncode[time.Hour & 63];
            name[7] = TimeEncode[time.Minute & 63];
            name[8] = TimeEncode[(time.Second / 10) & 63];
            name[9] = '\0';
            return new string(name);
        }

        /// <summary>
        /// Asynchronously retrieves last x messages.
        /// </summary>
        /// <param name="contract">The contract.</param>
        /// <param name="ssid">The subscription id to use.</param>
        /// <param name="limit">The amount of messages to retrieve.</param>
        /// <returns></returns>
        public override Task<IMessageStream> GetLastAsync(int contract, uint[] ssid, int limit)
        {
            // Do this asynchronously
            return Task.Run(() =>
            {
                // First, we need to query the index, we need to finish the read so we don't
                // unnecessarily block the connection.
                var queryIndex = this.Index
                    .QueryLast(contract, ssid, limit);

                // Return as message stream
                return (IMessageStream)(new MessageStream(this, contract, queryIndex));
            });
        }

        #endregion StorageProvider Members

        #region MessageStream

        /// <summary>
        /// Represents a stream that allows asynchronous download of messages.
        /// </summary>
        private class MessageStream : IMessageStream
        {
            private readonly List<IndexEntry> Index;
            private readonly ObjectStorageProvider Owner;
            private readonly int Contract;
            private int Current;

            /// <summary>
            /// Constructs a new stream
            /// </summary>
            /// <param name="index">The index to use.</param>
            /// <param name="contract">The user contract.</param>
            public MessageStream(ObjectStorageProvider owner, int contract, List<IndexEntry> index)
            {
                this.Owner = owner;
                this.Index = index;
                this.Current = index.Count;
                this.Contract = contract;
            }

            /// <summary>
            /// Whether the stream has a next element or not.
            /// </summary>
            public bool HasNext
            {
                get { return this.Current > 0; }
            }

            /// <summary>
            /// Gets the next message, asynchronously.
            /// </summary>
            /// <returns></returns>
            public async Task<ArraySegment<byte>> GetNext()
            {
                // If there's no next, don't download anything
                if (!this.HasNext)
                    return default(ArraySegment<byte>);

                // Figure out the index and download asynchronously
                var idx = this.Index[--this.Current];
                var key = GetKey(this.Contract, idx.Time, idx.Source);

                // Attempt to get from the cache first, if the message is recent
                if (idx.Time + TimeSpan.FromSeconds(60) > Timer.UtcNow)
                {
                    var message = await this.Owner.Append.GetFromCacheAsync(key, idx.Offset, idx.Length);
                    if (message != EmptyMessage)
                        return message;
                }

                // Attempt to retrieve the message from the store
                return await this.Owner.Store.DownloadAsync(key, idx.Offset, idx.Length);
            }
        }

        #endregion MessageStream
    }
}