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
using Emitter.Configuration;
using Emitter.Threading;
using Npgsql;
using NpgsqlTypes;

namespace Emitter.Storage
{
    /// <summary>
    /// Represents a buffered stream for an indexer.
    /// </summary>
    internal sealed class ObjectIndex : ThreadQueue<IndexAppendEntry>
    {
        /// <summary>
        /// Constructs a new PostgreSQL indexer.
        /// </summary>
        public ObjectIndex() : base(1000) { }

        /// <summary>
        /// Indexes asynchronously.
        /// </summary>
        /// <param name="contract">The associated contract.</param>
        /// <param name="ssid">The channel ssid.</param>
        /// <param name="ttl">The time to live for this message, in seconds.</param>
        /// <param name="filename">The filename which contains the payload.</param>
        /// <param name="offset">The offset of the payload.</param>
        /// <param name="length">The length of the payload</param>
        /// <returns>The task for completion notification.</returns>
        public void IndexAsync(int contract, uint[] ssid, string filename, int ttl, int offset, int length)
        {
            var cid = ((long)contract << 32) + ssid[1];
            var loc = ((long)offset << 32) + length;
            var now = (int)((Timer.UtcNow.Ticks - ObjectStorageProvider.TimeOffset) / TimeSpan.TicksPerSecond);
            var exp = now + ttl;
            var sub = GetSubchannel(ssid);
            this.WorkQueue.Enqueue(new IndexAppendEntry()
            {
                Stream = cid,
                Time = now,
                Subchannel = sub,
                Source = filename,
                Expires = exp,
                Location = loc
            });
        }

        /// <summary>
        /// Indexes a bucket which contains multiple payloads.
        /// </summary>
        /// <param name="key">The key of the bucket.</param>
        /// <param name="ttl">The expiration date of the bucket.</param>
        public void IndexBucket(string key, int ttl)
        {
            using (var connection = GetConnection())
            using (var cmd = new NpgsqlCommand())
            {
                cmd.Connection = connection;
                cmd.CommandText = string.Format("INSERT INTO msg_bucket (oid, ttl) VALUES ('{0}', {1})", key, ttl);
                cmd.ExecuteNonQuery();
            }
        }

        /// <summary>
        /// Queries last x messages.
        /// </summary>
        /// <param name="contract">The associated contract.</param>
        /// <param name="ssid">channel ssid.</param>
        /// <param name="limit">The limit of messages to get.</param>
        /// <returns></returns>
        public List<IndexEntry> QueryLast(int contract, uint[] ssid, int limit)
        {
            // Cap
            if (limit > 100)
                limit = 100;

            // Prepare the output
            var list = new List<IndexEntry>();

            // Prepare the select query
            var cid = ((long)contract << 32) + ssid[1];
            var sub = GetSubchannel(ssid);
            var now = (int)((Timer.UtcNow.Ticks - ObjectStorageProvider.TimeOffset) / TimeSpan.TicksPerSecond);
            var query = string.Format("SELECT ts, src, loc FROM msg WHERE cid = {0} AND ts > 0 AND ts < {1} AND sub LIKE '{2}'  AND ttl > {3} ORDER by ts DESC LIMIT {4}",
                cid,
                now,
                string.IsNullOrWhiteSpace(sub) ? "%" : sub,
                now,
                limit);

            using (var connection = GetConnection())
            using (var cmd = new NpgsqlCommand())
            {
                cmd.Connection = connection;
                cmd.CommandText = query;
                using (var reader = cmd.ExecuteReader())
                {
                    while (reader.Read())
                    {
                        var ts = reader.GetInt32(0);
                        var src = reader.GetString(1);
                        var loc = reader.GetInt64(2);

                        // Create an entry
                        var entry = new IndexEntry();
                        entry.Time = new DateTime(ObjectStorageProvider.TimeOffset + (ts * TimeSpan.TicksPerSecond));
                        entry.Source = src;
                        entry.Offset = (int)(loc >> 32);
                        entry.Length = (int)(loc - ((long)entry.Offset << 32));
                        list.Add(entry);
                    }
                }
            }
            return list;
        }

        /// <summary>
        /// Gets a connection from the pool.
        /// </summary>
        /// <returns>The open connection.</returns>
        private NpgsqlConnection GetConnection()
        {
            // Open the connection
            var connection = new NpgsqlConnection(EmitterConfig.Default.Storage.Connection);
            connection.Open();
            return connection;
        }

        /// <summary>
        /// Occurs when the work needs to be processed.
        /// </summary>
        protected override void OnProcess()
        {
            using (var connection = GetConnection())
            using (var writer = connection.BeginBinaryImport("COPY msg (cid, ts, sub, src, loc, ttl) FROM STDIN (FORMAT BINARY)"))
            {
                // Dequeue and write in the batch
                IndexAppendEntry entry;
                while (this.WorkQueue.TryDequeue(out entry))
                {
                    writer.StartRow();
                    writer.Write(entry.Stream, NpgsqlDbType.Bigint);
                    writer.Write(entry.Time, NpgsqlDbType.Integer);
                    writer.Write(entry.Subchannel, NpgsqlDbType.Text);
                    writer.Write(entry.Source, NpgsqlDbType.Text);
                    writer.Write(entry.Location, NpgsqlDbType.Bigint);
                    writer.Write(entry.Expires, NpgsqlDbType.Integer);
                }
            }
        }

        /// <summary>
        /// Gets the subchannel from a full channel name.
        /// </summary>
        /// <param name="ssid">The ssid to parse.</param>
        /// <returns>The subchannel extracted.</returns>
        private static unsafe string GetSubchannel(uint[] ssid)
        {
            // Compute the length of the subchannel
            const int offset = 2;
            int length = ssid.Length - offset;
            if (length <= 0)
                return string.Empty;

            // Prepare a array in memory
            uint target;
            int value;
            var idx = 0;
            var buffer = stackalloc char[length * 8 + 1];
            buffer[length * 8] = '\0';
            for (int i = offset; i < ssid.Length; ++i)
            {
                target = ssid[i];
                if (target != 1815237614)
                {
                    value = (byte)(target >> 24) >> 4;
                    buffer[idx++] = (char)(55 + value + (((value - 10) >> 31) & -7));
                    value = ((byte)(target >> 24)) & 0xF;
                    buffer[idx++] = (char)(55 + value + (((value - 10) >> 31) & -7));

                    value = (byte)(target >> 16) >> 4;
                    buffer[idx++] = (char)(55 + value + (((value - 10) >> 31) & -7));
                    value = ((byte)(target >> 16)) & 0xF;
                    buffer[idx++] = (char)(55 + value + (((value - 10) >> 31) & -7));

                    value = (byte)(target >> 8) >> 4;
                    buffer[idx++] = (char)(55 + value + (((value - 10) >> 31) & -7));
                    value = ((byte)(target >> 8)) & 0xF;
                    buffer[idx++] = (char)(55 + value + (((value - 10) >> 31) & -7));

                    value = (byte)(target) >> 4;
                    buffer[idx++] = (char)(55 + value + (((value - 10) >> 31) & -7));
                    value = ((byte)(target)) & 0xF;
                    buffer[idx++] = (char)(55 + value + (((value - 10) >> 31) & -7));
                }
                else
                {
                    buffer[idx++] = '_';
                    buffer[idx++] = '_';
                    buffer[idx++] = '_';
                    buffer[idx++] = '_';
                    buffer[idx++] = '_';
                    buffer[idx++] = '_';
                    buffer[idx++] = '_';
                    buffer[idx++] = '_';
                }
            }

            return new string(buffer);
        }
    }

    public struct IndexAppendEntry
    {
        public long Stream;
        public int Time;
        public string Subchannel;
        public string Source;
        public long Location;
        public int Expires;
    }

    public struct IndexEntry
    {
        public DateTime Time;
        public string Source;
        public int Offset;
        public int Length;
    }
}