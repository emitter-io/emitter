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

using System.Diagnostics;
using Emitter.Network.Threading;

namespace Emitter.Network
{
    public class BufferSizeControl : IBufferSizeControl
    {
        private readonly long _maxSize;
        private readonly Connection _connectionControl;
        private readonly EventThread _connectionThread;

        private readonly object _lock = new object();

        private long _size;
        private bool _connectionPaused;

        public BufferSizeControl(long maxSize, Connection connectionControl, EventThread connectionThread)
        {
            _maxSize = maxSize;
            _connectionControl = connectionControl;
            _connectionThread = connectionThread;
        }

        private long Size
        {
            get
            {
                return _size;
            }
            set
            {
                // Caller should ensure that bytes are never consumed before the producer has called Add()
                Debug.Assert(value >= 0);
                _size = value;
            }
        }

        public void Add(int count)
        {
            Debug.Assert(count >= 0);

            if (count == 0)
            {
                // No-op and avoid taking lock to reduce contention
                return;
            }

            lock (_lock)
            {
                Size += count;
                if (!_connectionPaused && Size >= _maxSize)
                {
                    _connectionPaused = true;
                    _connectionThread.Post(
                        (connectionControl) => ((Connection)connectionControl).Pause(),
                        _connectionControl);
                }
            }
        }

        public void Subtract(int count)
        {
            Debug.Assert(count >= 0);

            if (count == 0)
            {
                // No-op and avoid taking lock to reduce contention
                return;
            }

            lock (_lock)
            {
                Size -= count;
                if (_connectionPaused && Size < _maxSize)
                {
                    _connectionPaused = false;
                    _connectionThread.Post(
                        (connectionControl) => ((Connection)connectionControl).Resume(),
                        _connectionControl);
                }
            }
        }
    }
}