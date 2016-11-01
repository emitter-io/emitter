----------------
STORING MESSAGES 
----------------

1. On write, put the message into LRU cache. This will be an agressive caching mechanism which caches
   everything in memory for the past X minutes. X can be configurable per customer. Each message we 
   want to store should be compressed with high-ration compression algorithm such as LZMA. It might 
   take much CPU power to compress, but we will gain time & money since we won't need to store nearly
   as much data and transfer times will be significantly shorter.

2. Write the messages to a file on disk, locally. Each file name will have the following format where
   the date is based on the server timestamp: [contract + channel id]/[yyyymmdd]/[hhmm]

3. For each message, write (channel, timestamp, offset, filename) to a /index/[contract + channel id] 
   channel. An indexer service will be listening to this and having a Elasticsearch database which 
   we could query. Since we're going to have a separate database and won't store actual data, we can 
   go ahead and index stuff there.

3. Have an uploader which will stream the files to a S3(-compatible) Object Store. The objects can be 
   uploaded as multi-part, multi-threaded transfer. The idea is that it will be hitting as many storage
   servers as possible, therefore maximizing throughput. S3 should be fast enough in the beginning. The
   object expiry should be set to the max(ttl) of the file. This will ensure that we don't keep the 
   data more than needed and still be able to retrieve it.

4. Once the file is written to S3, delete it from the disk. In case of lag (not able to write fast
   enough) we should be notified (email/sms, stuffs).

-------------------
RETRIEVING MESSAGES
-------------------

1. If the history requested is for the past X minutes, query all brokers for the cached objects. The
   brokers will reply if they have the data in cache, making the whole thing fast. For the data which
   is stale (more than X minutes old), go to step 2. Otherwise, go to step 4.

2. Query the indexer and request the list of offsets and sizes to read. The indexer will parse and 
   execute the query, but will only return the list of offsets and won't proxy any data.

3. Use this list of objects & offsets returned by the indexer and download the data from the object 
   storage. Issue S3 GetObject request and retrieve the parts of the files we need.

4. Send the history (in order) to the connected client requesting the history.
