// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.IO;
using System.Threading.Tasks;

using Emitter.Network.Threading;

namespace Emitter.Network.Filter
{
    public class FilteredStreamAdapter : IDisposable
    {
        private readonly ConnectionId _connectionId;
        private readonly Stream _filteredStream;
        private readonly MemoryPool _memory;
        private MemoryPoolBlock _block;
        private bool _aborted = false;

        public FilteredStreamAdapter(
            ConnectionId connectionId,
            Stream filteredStream,
            MemoryPool memory,
            IThreadPool threadPool,
            IBufferSizeControl bufferSizeControl)
        {
            SocketInput = new SocketInput(memory, threadPool, bufferSizeControl);
            SocketOutput = new StreamSocketOutput(connectionId, filteredStream, memory);

            _connectionId = connectionId;
            _filteredStream = filteredStream;
            _memory = memory;
        }

        public SocketInput SocketInput { get; private set; }

        public ISocketOutput SocketOutput { get; private set; }

        public Task ReadInputAsync()
        {
            _block = _memory.Lease();
            // Use pooled block for copy
            return FilterInputAsync(_block).ContinueWith((task, state) =>
            {
                ((FilteredStreamAdapter)state).OnStreamClose(task);
            }, this);
        }

        public void Abort()
        {
            _aborted = true;
        }

        public void Dispose()
        {
            SocketInput.Dispose();
        }

        private async Task FilterInputAsync(MemoryPoolBlock block)
        {
            int bytesRead;
            while ((bytesRead = await _filteredStream.ReadAsync(block.Array, block.Data.Offset, block.Data.Count)) != 0)
            {
                SocketInput.IncomingData(block.Array, block.Data.Offset, bytesRead);
            }
        }

        private void OnStreamClose(Task copyAsyncTask)
        {
            _memory.Return(_block);

            if (copyAsyncTask.IsFaulted)
            {
                SocketInput.AbortAwaiting();
                Service.Logger.Log(copyAsyncTask.Exception);
            }
            else if (copyAsyncTask.IsCanceled)
            {
                SocketInput.AbortAwaiting();
                Service.Logger.Log(LogLevel.Error, "FilteredStreamAdapter.CopyToAsync canceled.");
            }
            else if (_aborted)
            {
                SocketInput.AbortAwaiting();
            }

            try
            {
                SocketInput.IncomingFin();
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
        }
    }

    internal static class StreamExtensions
    {
        public static async Task CopyToAsync(this Stream source, Stream destination, MemoryPoolBlock block)
        {
            int bytesRead;
            while ((bytesRead = await source.ReadAsync(block.Array, block.Data.Offset, block.Data.Count)) != 0)
            {
                await destination.WriteAsync(block.Array, block.Data.Offset, bytesRead);
            }
        }
    }
}