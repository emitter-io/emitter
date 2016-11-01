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
using Amazon;
using Amazon.S3;
using Amazon.S3.Model;
using Emitter.Configuration;
using Emitter.Providers;
using Emitter.Threading;

namespace Emitter.Storage
{
    /// <summary>
    /// Represents a buffered stream for an indexer.
    /// </summary>
    internal sealed class ObjectStorage
    {
        #region Stream Reading

        [ThreadStatic]
        private static byte[] ReadBuffer;

        /// <summary>
        /// Fully reads the stream.
        /// </summary>
        private static byte[] ReadStream(Stream input)
        {
            if (ReadBuffer == null)
                ReadBuffer = new byte[8 * 1024];

            using (var ms = new MemoryStream())
            {
                int read;
                while ((read = input.Read(ReadBuffer, 0, ReadBuffer.Length)) > 0)
                    ms.Write(ReadBuffer, 0, read);
                return ms.ToArray();
            }
        }

        #endregion Stream Reading

        #region Constructor

        /// <summary>
        /// The indexer to use
        /// </summary>
        private readonly ObjectIndex Index;

        /// <summary>
        /// The scheduler to use for the uploader
        /// </summary>
        private readonly ThreadPool WorkPool = new ThreadPool(64);

        /// <summary>
        /// Create a new client
        /// </summary>
        private AmazonS3Client Client;

        /// <summary>
        /// Constructs the storage.
        /// </summary>
        /// <param name="index">The indexer to use</param>
        public ObjectStorage(ObjectIndex index)
        {
            this.Index = index;
            this.RenewCredentials();
            Timer.PeriodicCall(TimeSpan.FromMinutes(30), RenewCredentials);
        }

        /// <summary>
        /// Periodically renews the client.
        /// </summary>
        private async void RenewCredentials()
        {
            try
            {
                var client = await GetClient();
                if (client != null)
                    this.Client = client;
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
        }

        /// <summary>
        /// Gets a client with new credentials.
        /// </summary>
        /// <returns></returns>
        private async Task<AmazonS3Client> GetClient()
        {
            try
            {
                var credentials = await Service.Providers
                    .Resolve<SecurityProvider>()
                    .GetCredentialsAsync("storage-emitter");
                if (credentials == null)
                    return null;

                Service.Logger.Log("Storage: Requested new token until " + credentials.Expires);
                return new AmazonS3Client(credentials.AccessKey, credentials.SecretKey, credentials.Token, new AmazonS3Config()
                {
                    UseHttp = true,
                    BufferSize = 8192,
                    RegionEndpoint = RegionEndpoint.EUWest1
                });
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                return null;
            }
        }

        #endregion Constructor

        #region Upload Members

        /// <summary>
        /// Uploads the log file to S3 storage.
        /// </summary>
        /// <param name="key">The key of the file to upload to S3.</param>
        /// <param name="expire">The expiration date to set.</param>
        public void UploadAsync(string key, DateTime expire)
        {
            var work = new UploadWork();
            work.File = key;
            work.Expires = expire;
            WorkPool.QueueUserWorkItem(UploadWork, work);
        }

        /// <summary>
        /// Uploads the work.
        /// </summary>
        /// <param name="task">The task to upload.</param>
        private async void UploadWork(object task)
        {
            try
            {
                var work = (UploadWork)task;
                if (work.File == null)
                    return;

                // Prepare props
                var key = work.File.Replace("data/", "");
                var ttl = (int)((work.Expires.Ticks - ObjectStorageProvider.TimeOffset) / TimeSpan.TicksPerSecond);

                // Prepare the request
                var request = new PutObjectRequest();
                request.FilePath = Path.GetFullPath(work.File);
                request.BucketName = EmitterConfig.Default.Storage.Location;
                request.Key = key;
                request.Headers.Expires = work.Expires;

                // Put the object
                var upload = await this.Client.PutObjectAsync(request);

                // Now that the thing is uploaded, we can safely delete the file from disk
                File.Delete(work.File);

                // Now we can index the bucket as well
                this.Index.IndexBucket(key, ttl);

                // Debug
#if DEBUG
                Console.WriteLine("Upload: {0} ({1})", request.Key, upload.HttpStatusCode);
#endif
            }
            catch (Exception ex)
            {
                // Log the error
                Service.Logger.Log(ex);

                // TODO: retry
            }
        }

        #endregion Upload Members

        #region Download Members

        /// <summary>
        /// Download a part of the file.
        /// </summary>
        /// <param name="key">The key file to download.</param>
        /// <param name="offset">The offset of the message.</param>
        /// <param name="length">The lenght of the message to download.</param>
        /// <returns>The stream to the body.</returns>
        public async Task<ArraySegment<byte>> DownloadAsync(string key, int offset, int length)
        {
            try
            {
                var request = new GetObjectRequest();
                request.Key = key;
                request.BucketName = EmitterConfig.Default.Storage.Location;
                request.ByteRange = new ByteRange(offset, offset + length - 1);

                // Get the message and return the stream to the content
                var response = await this.Client.GetObjectAsync(request);

                // If we have a message
                if (response.HttpStatusCode != System.Net.HttpStatusCode.PartialContent)
                    return StorageProvider.EmptyMessage;

                // Read the stream to a buffer
                using (var responseStream = response.ResponseStream)
                {
                    return new ArraySegment<byte>(ReadStream(responseStream));
                }
            }
            catch (AmazonS3Exception ex)
            {
                // Forbidden means no message actually found yet.
                if (ex.StatusCode != System.Net.HttpStatusCode.Forbidden)
                    Service.Logger.Log(ex);

                return StorageProvider.EmptyMessage;
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                return StorageProvider.EmptyMessage;
            }
        }

        #endregion Download Members
    }

    /// <summary>
    /// Represents an upload work.
    /// </summary>
    internal struct UploadWork
    {
        /// <summary>
        /// The filename to upload.
        /// </summary>
        public string File;

        /// <summary>
        /// The expiration date to set.
        /// </summary>
        public DateTime Expires;
    }
}