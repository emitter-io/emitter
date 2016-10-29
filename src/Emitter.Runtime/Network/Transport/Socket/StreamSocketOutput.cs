// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.IO;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace Emitter.Network
{
    public class StreamSocketOutput : ISocketOutput
    {
        private static readonly byte[] _endChunkBytes = Encoding.ASCII.GetBytes("\r\n");
        private static readonly byte[] _nullBuffer = new byte[0];

        private readonly ConnectionId _connectionId;
        private readonly Stream _outputStream;
        private readonly MemoryPool _memory;
        private MemoryPoolBlock _producingBlock;

        private bool _canWrite = true;

        private object _writeLock = new object();

        public StreamSocketOutput(ConnectionId connectionId, Stream outputStream, MemoryPool memory)
        {
            _connectionId = connectionId;
            _outputStream = outputStream;
            _memory = memory;
        }

        public void Write(ArraySegment<byte> buffer)
        {
            lock (_writeLock)
            {
                if (buffer.Count == 0)
                    return;

                try
                {
                    if (!_canWrite)
                        return;

                    _outputStream.Write(buffer.Array ?? _nullBuffer, buffer.Offset, buffer.Count);
                }
                catch (Exception)
                {
                    _canWrite = false;
                }
            }
        }

        public Task WriteAsync(ArraySegment<byte> buffer, CancellationToken cancellationToken)
        {
            return _outputStream.WriteAsync(buffer.Array ?? _nullBuffer, buffer.Offset, buffer.Count, cancellationToken);
        }

        public MemoryPoolIterator ProducingStart()
        {
            _producingBlock = _memory.Lease();
            return new MemoryPoolIterator(_producingBlock);
        }

        public void ProducingComplete(MemoryPoolIterator end)
        {
            var block = _producingBlock;
            while (block != end.Block)
            {
                // If we don't handle an exception from _outputStream.Write() here, we'll leak memory blocks.
                if (_canWrite)
                {
                    try
                    {
                        _outputStream.Write(block.Data.Array, block.Data.Offset, block.Data.Count);
                    }
                    catch (Exception)
                    {
                        _canWrite = false;
                    }
                }

                var returnBlock = block;
                block = block.Next;
                returnBlock.Pool.Return(returnBlock);
            }

            if (_canWrite)
            {
                try
                {
                    _outputStream.Write(end.Block.Array, end.Block.Data.Offset, end.Index - end.Block.Data.Offset);
                }
                catch (Exception)
                {
                    _canWrite = false;
                }
            }

            end.Block.Pool.Return(end.Block);
        }
    }
}