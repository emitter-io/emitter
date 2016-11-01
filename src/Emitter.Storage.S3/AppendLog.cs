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
using System.IO;
using System.Linq;
using System.Threading.Tasks;

namespace Emitter.Storage
{
    /// <summary>
    /// Represents a buffered stream for a log writer.
    /// </summary>
    internal class AppendLog : IDisposable
    {
        private string Path;
        private BufferedStream Stream = null;
        private DateTime Expiration;

        /// <summary>
        /// Gets the current output path.
        /// </summary>
        public string OutputPath
        {
            get { return this.Path; }
        }

        /// <summary>
        /// Gets the date/time of the expiration.
        /// </summary>
        public DateTime ExpireDate
        {
            get { return this.Expiration; }
        }

        /// <summary>
        /// Gets the current output stream.
        /// </summary>
        public Stream OutputStream
        {
            get { return this.Stream; }
        }

        /// <summary>
        /// Attempts to redirect the stream into another file.
        /// </summary>
        /// <param name="path">The file to redirect the output into.</param>
        /// <returns>Whether it was successfully redirected or not.</returns>
        public bool TryRedirect(string path)
        {
            lock (this)
            {
                // Do not redirect if nothing has changed
                if (this.Path == path)
                    return false;

                // Ensure we have the directory
                new FileInfo(path)?.Directory.Create();

                // Now set the properties
                this.Path = path;
                this.Expiration = DateTime.MinValue;
                this.Stream = new BufferedStream(File.Open(path, FileMode.Append, FileAccess.Write, FileShare.Read));
                return true;
            }
        }

        /// <summary>
        /// Appends a binary content to the end of a buffered file stream.
        /// </summary>
        /// <param name="content">The content to append.</param>
        /// <param name="expire">The expiration date of the message.</param>
        /// <returns></returns>
        public async Task<int> AppendAsync(ArraySegment<byte> content, DateTime expire)
        {
            // Get the offset & schedule the write
            var offset = (int)this.Stream.Position;
            await Stream.WriteAsync(content.Array, content.Offset, content.Count);

            // Add the expiration date, always max
            if (this.Expiration < expire)
                this.Expiration = expire;

            // Return the log entry once we're finished writing
            return offset;
        }

        #region IDisposable Members

        /// <summary>
        /// Disposes this instance.
        /// </summary>
        public void Dispose()
        {
            this.OnDispose(true);
            GC.SuppressFinalize(this);
        }

        /// <summary>
        /// Occurs when the stream is disposing or finalizing.
        /// </summary>
        /// <param name="disposing">Whether we are disposing or finalizing.</param>
        private void OnDispose(bool disposing)
        {
            if (this.Stream != null)
            {
                this.Stream.Flush();
                this.Stream.Dispose();
            }
        }

        /// <summary>
        /// Finalizes the stream
        /// </summary>
        ~AppendLog()
        {
            this.OnDispose(false);
        }

        #endregion IDisposable Members
    }
}