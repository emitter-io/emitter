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
using System.Collections.Concurrent;
using System.IO;
using System.Threading.Tasks;
using Emitter.Providers;
using Emitter.Threading;

namespace Emitter.Storage
{
    /// <summary>
    /// Represents a cache for append logs, avoiding querying an external storage service.
    /// </summary>
    internal sealed class AppendLogCache : ThreadQueue<AppendLogEntry>
    {
        /// <summary>
        /// Constructs a new append log cache.
        /// </summary>
        public AppendLogCache() : base(1000) { }

        /// <summary>
        /// Currently active append logs.
        /// </summary>
        private readonly ConcurrentDictionary<int, AppendLog> Registry =
            new ConcurrentDictionary<int, AppendLog>();

        /// <summary>
        /// Occurs when the work needs to be processed.
        /// </summary>
        protected override void OnProcess()
        {
        }

        /// <summary>
        /// Gets a currently used append log.
        /// </summary>
        /// <param name="contract">The contract to retrieve the append log for.</param>
        /// <returns>The append log retrieved.</returns>
        public AppendLog GetCurrent(int contract)
        {
            return this.Registry.GetOrAdd(contract, (k) => new AppendLog());
        }

        /// <summary>
        /// Reads a part of the file from the cache.
        /// </summary>
        /// <param name="key">The key file to download.</param>
        /// <param name="offset">The offset of the message.</param>
        /// <param name="length">The lenght of the message to download.</param>
        /// <returns>The stream to the body.</returns>
        public async Task<ArraySegment<byte>> GetFromCacheAsync(string key, int offset, int length)
        {
            try
            {
                var path = Path.GetFullPath("data/" + key);
                if (!File.Exists(path))
                    return StorageProvider.EmptyMessage;

                var buffer = new byte[length];
                using (var fs = File.Open(path, FileMode.Open, FileAccess.Read, FileShare.ReadWrite))
                {
                    fs.Seek(offset, SeekOrigin.Begin);
                    await fs.ReadAsync(buffer, 0, length);
                    return new ArraySegment<byte>(buffer);
                }
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                return StorageProvider.EmptyMessage;
            }
        }
    }

    public struct AppendLogEntry
    {
        public string Source;
    }
}