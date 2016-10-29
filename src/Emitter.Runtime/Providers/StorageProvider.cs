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
using System.Threading.Tasks;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a storage module.
    /// </summary>
    public abstract class StorageProvider : Provider
    {
        /// <summary>
        /// An empty message
        /// </summary>
        public readonly static ArraySegment<byte> EmptyMessage = default(ArraySegment<byte>);

        /// <summary>
        /// Asynchronously stores a message in the store.
        /// </summary>
        /// <param name="contract">The contract.</param>
        /// <param name="ssid">The subscription id .</param>
        /// <param name="message">The message payload to store.</param>
        /// <param name="ttl">The time to live for this message, in seconds.</param>
        /// <returns>The task for completion notification.</returns>
        public abstract Task AppendAsync(int contract, uint[] ssid, int ttl, ArraySegment<byte> message);

        /// <summary>
        /// Asynchronously retrieves last x messages.
        /// </summary>
        /// <param name="contract">The contract.</param>
        /// <param name="ssid">The subscription id .</param>
        /// <param name="limit">The amount of messages to retrieve.</param>
        /// <returns></returns>
        public abstract Task<IMessageStream> GetLastAsync(int contract, uint[] ssid, int limit);
    }

    /// <summary>
    /// Represents a stream that allows asynchronous download of messages.
    /// </summary>
    public interface IMessageStream
    {
        /// <summary>
        /// Checks whether the stream has a next element or not.
        /// </summary>
        bool HasNext { get; }

        /// <summary>
        /// Gets the next message from the stream.
        /// </summary>
        /// <returns>The message payload.</returns>
        Task<ArraySegment<byte>> GetNext();
    }
}