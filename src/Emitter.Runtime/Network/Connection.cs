// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Net;
using System.Threading;
using System.Threading.Tasks;
using Emitter.Diagnostics;
using Emitter.Network;
using Emitter.Network.Filter;
using Emitter.Network.Native;
using Emitter.Providers;

namespace Emitter
{
    /// <summary>
    /// Represents a connection.
    /// </summary>
    public class Connection : ConnectionContext, IConnection
    {
        #region Static Members
        private static readonly ArraySegment<byte> EmptyBuffer = new ArraySegment<byte>(new byte[0]);

        private static readonly Action<UvStreamHandle, int, object> _readCallback =
            (handle, status, state) => ReadCallback(handle, status, state);

        private static readonly Func<UvStreamHandle, int, object, Libuv.uv_buf_t> _allocCallback =
            (handle, suggestedsize, state) => AllocCallback(handle, suggestedsize, state);

        #endregion Static Members

        #region Constructors
        private readonly UvStreamHandle _socket;
        private ConnectionFilterContext _filterContext;
        private LibuvStream _libuvStream;
        private readonly SocketInput _rawSocketInput;
        private readonly SocketOutput _rawSocketOutput;
        private readonly object _stateLock = new object();
        private ConnectionState State;
        private TaskCompletionSource<object> _socketClosedTcs;
        private FilteredStreamAdapter _filteredStreamAdapter;
        private Task _readInputTask;
        private BufferSizeControl _bufferSizeControl = null;

        // Processing stuff
        protected volatile bool ProcessingStarted;

        protected volatile bool ProcessingStopping;

        private Task ProcessingTask;
        private TimeSpan ConnectionTimeout = TimeSpan.FromSeconds(60);
        protected int _requestAborted;
        protected CancellationTokenSource _abortedCts;
        protected CancellationToken? _manuallySetRequestAbortToken;

        // Flags
        private int Disposed = 0;

        /// <summary>
        /// Creates a new connection from the listener context.
        /// </summary>
        /// <param name="context">The listener context to use.</param>
        /// <param name="socket">The socket handle.</param>
        internal Connection(ListenerContext context, UvStreamHandle socket) : base(context)
        {
            _socket = socket;
            socket.Connection = this;

            ConnectionId = ConnectionId.NewConnectionId();

            _rawSocketInput = new SocketInput(Memory, ThreadPool);
            _rawSocketOutput = new SocketOutput(Thread, _socket, Memory, this, ConnectionId, ThreadPool, WriteReqPool);

            this.Expires = Timer.Now + TimeSpan.FromMinutes(1);
            this.ConnectionManager.Register(this);
        }

        /// <summary>
        /// Creates a connection for a remote endpoint.
        /// </summary>
        /// <param name="context">The listener context to use.</param>
        /// <param name="remote">The endpoint to connect to.</param>
        public static Task<Connection> ConnectAsync(ListenerContext context, IPEndPoint remote)
        {
            var tcs = new TaskCompletionSource<Connection>();
            context.Thread.PostAsync((state) =>
            {
                var socket = new UvTcpHandle();
                try
                {
                    socket.Init(context.Thread.Loop, context.Thread.QueueCloseHandle);
                    socket.NoDelay(true);
                    socket.Connect(context.Thread.Loop, ServiceAddress.FromIPEndPoint(remote), (request, status, ex, sender) =>
                    {
                        request.Dispose();

                        if (ex != null)
                        {
                            // Error has occured, set the exception
                            socket.Dispose();
                            tcs.SetException(ex);
                            return;
                        }

                        // Create the connection and notify
                        var connection = new Connection(context, socket);
                        tcs.SetResult(connection.Start() ? connection : null);
                    }, tcs);
                }
                catch (UvException ex)
                {
                    //Service.Logger.Log(ex);
                    socket.Dispose();
                    tcs.SetException(ex);
                }
            }, tcs);
            return tcs.Task;
        }

        #endregion Constructors

        #region Public Properties

        /// <summary>
        /// Event that is issued when a channel was disconnected and about to be disposed.
        /// </summary>
        public event ChannelEvent Disconnect;

        /// <summary>
        /// Gets whether the connection is running.
        /// </summary>
        public bool IsRunning
        {
            get
            {
                return this.State == ConnectionState.Open
                    && this.ProcessingStarted
                    && !this.ProcessingStopping
                    && this.Disposed == 0
                    && this._socket != null;
            }
        }

        /// <summary>
        /// Gets or sets the timeout of this connection.
        /// </summary>
        public TimeSpan Timeout
        {
            get { return this.ConnectionTimeout; }
            set
            {
                this.ConnectionTimeout = value;
                this.Expires = Timer.Now + value;
            }
        }

        /// <summary>
        /// Gets or sets when the next activity check should be performed.
        /// </summary>
        public DateTime Expires
        {
            get;
            set;
        }

        #endregion Public Properties

        #region Start/Stop Members

        /// <summary>
        /// Starts listening on this <see cref="Connection"/>.
        /// </summary>
        /// <returns>Whether the channel was started successfully or not.</returns>
        public bool Start()
        {
            if (this.State != ConnectionState.Creating)
                return false;

            //NetTrace.WriteLine("Connection Starting", this, NetTraceCategory.Channel);
            this.OnAfterConstruct();

            // Start socket prior to applying the ConnectionFilter
            _socket.ReadStart(_allocCallback, _readCallback, this);

            var tcpHandle = _socket as UvTcpHandle;
            if (tcpHandle != null)
            {
                RemoteEndPoint = tcpHandle.GetPeerIPEndPoint();
                LocalEndPoint = tcpHandle.GetSockIPEndPoint();
            }

            // Don't initialize _frame until SocketInput and SocketOutput are set to their final values.
            if (Options.ConnectionFilter == null)
            {
                lock (_stateLock)
                {
                    if (State != ConnectionState.Creating)
                    {
                        throw new InvalidOperationException("Invalid connection state: " + State);
                    }

                    State = ConnectionState.Open;
                    SocketInput = _rawSocketInput;
                    SocketOutput = _rawSocketOutput;

                    this.OnConnect();
                    return true;
                }
            }
            else
            {
                _libuvStream = new LibuvStream(_rawSocketInput, _rawSocketOutput);
                _filterContext = new ConnectionFilterContext
                {
                    Connection = _libuvStream,
                    Address = ServerAddress
                };

                try
                {
                    Options.ConnectionFilter.OnConnectionAsync(_filterContext).ContinueWith((task, state) =>
                    {
                        var connection = (Connection)state;

                        if (task.IsFaulted)
                        {
                            Service.Logger.Log(task.Exception);
                            connection.Close(CloseType.SocketDisconnect);
                        }
                        else if (task.IsCanceled)
                        {
                            //connection.Log.LogError("ConnectionFilter.OnConnection Canceled");
                            connection.Close(CloseType.SocketDisconnect);
                        }
                        else
                        {
                            connection.ApplyConnectionFilter();
                        }
                    }, this);
                    return true;
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                    this.Close(CloseType.SocketDisconnect);
                    return false;
                }
            }
        }

        /// <summary>
        /// Should be called when the server wants to initiate a shutdown. The Task returned will
        /// become complete when the RequestProcessingAsync function has exited. It is expected that
        /// Stop will be called on all active connections, and Task.WaitAll() will be called on every
        /// return value.
        /// </summary>
        private Task StopProcessing()
        {
            if (!this.ProcessingStopping)
                this.ProcessingStopping = true;

            return this.ProcessingTask ?? TaskUtilities.CompletedTask;
        }

        public CancellationToken RequestAborted
        {
            get
            {
                // If a request abort token was previously explicitly set, return it.
                if (_manuallySetRequestAbortToken.HasValue)
                {
                    return _manuallySetRequestAbortToken.Value;
                }
                // Otherwise, get the abort CTS.  If we have one, which would mean that someone previously
                // asked for the RequestAborted token, simply return its token.  If we don't,
                // check to see whether we've already aborted, in which case just return an
                // already canceled token.  Finally, force a source into existence if we still
                // don't have one, and return its token.
                var cts = _abortedCts;
                return
                    cts != null ? cts.Token :
                    (Volatile.Read(ref _requestAborted) == 1) ? new CancellationToken(true) :
                    RequestAbortedSource.Token;
            }
            set
            {
                // Set an abort token, overriding one we create internally.  This setter and associated
                // field exist purely to support IHttpRequestLifetimeFeature.set_RequestAborted.
                _manuallySetRequestAbortToken = value;
            }
        }

        private CancellationTokenSource RequestAbortedSource
        {
            get
            {
                // Get the abort token, lazily-initializing it if necessary.
                // Make sure it's canceled if an abort request already came in.
                var cts = LazyInitializer.EnsureInitialized(ref _abortedCts, () => new CancellationTokenSource());
                if (Volatile.Read(ref _requestAborted) == 1)
                {
                    cts.Cancel();
                }
                return cts;
            }
        }

        /// <summary>
        /// Immediate kill the connection and poison the request and response streams.
        /// </summary>
        private void AbortProcessing()
        {
            if (Interlocked.CompareExchange(ref _requestAborted, 1, 0) == 0)
            {
                this.ProcessingStopping = true;
                try
                {
                    this.Close(CloseType.SocketDisconnect);
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                }

                try
                {
                    RequestAbortedSource.Cancel();
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                }
                _abortedCts = null;
            }
        }

        #endregion Start/Stop Members

        #region OnConnect/OnReceive... Members

        /// <summary>
        /// Called once by Connection class to begin the OnReceive loop.
        /// </summary>
        private void OnConnect()
        {
            if (!this.ProcessingStarted)
            {
                this.ProcessingStarted = true;
                this.Expires = Timer.Now + Timeout;
                this.ProcessingTask =
                    Task.Factory.StartNew(
                        (o) => ((Connection)o).OnReceive(),
                        this,
                        default(CancellationToken),
                        TaskCreationOptions.DenyChildAttach,
                        TaskScheduler.Default);
            }
        }

        /// <summary>
        /// Invoked when the channel is fully constructed and is ready to be used.
        /// The base implementation binds the client.
        /// </summary>
        protected void OnAfterConstruct()
        {
            // Bind to a single connection client by default
            var provider = Service.Providers.Resolve<ClientProvider>();
            if (provider != null)
            {
                // Default case, use the provider
                this.Client = provider.GetDefaultClient(this);
                this.Client.BindChannel(this);
            }
            else
            {
                // Fallback mechanism in case the provider was not specified
                this.Client = new Client();
                this.Client.BindChannel(this);
            }

            // Apply the configuration now
            this.DecodingSettings.ConfigureFor(this);
            this.EncodingSettings.ConfigureFor(this);
        }

        /// <summary>
        /// Primary loop which consumes socket input, parses it for protocol framing, and invokes the
        /// application delegate for as long as the socket is intended to remain open. The resulting
        /// Task from this loop is preserved in a field which is used when the server needs to drain
        /// and close all currently active connections.
        /// </summary>
        public async Task OnReceive()
        {
            try
            {
                while (!this.ProcessingStopping)
                {
                    // Update the expires timer
                    this.Expires = Timer.Now + Timeout;
                    this._abortedCts = null;

                    // If _requestAbort is set, the connection has already been closed.
                    if (Volatile.Read(ref _requestAborted) != 0)
                        return;

                    // If we're disconnecting, do not process
                    if (this.State == ConnectionState.Disconnecting || this.State == ConnectionState.SocketClosed || this.Decoding == null)
                        return;

                    // Book a buffer
                    var incoming = this.DecodingSettings.BufferProvider.Reserve(Constants.ReceiveBufferSize);

                    // Read from the socket
                    incoming.Length = await this.SocketInput.ReadAsync(incoming.Array, incoming.Offset, incoming.Length);

                    // Process whatever we just read
                    using (var context = ProcessingContext.Acquire(this, this.Decoding))
                    {
                        // Forward to the receive
                        await context.OnReceive(incoming);
                    }
                }
            }
            catch (Exception ex)
            {
                // Dig out the inner exception
                if (ex.InnerException != null)
                    ex = ex.InnerException;

                // If the task was canceled, do not log
                if (ex is TaskCanceledException)
                    return;

                // Log everything else
                Service.Logger.Log(ex);
                //Service.Logger.Log(LogLevel.Warning, "Connection processing ended abnormally. Reason: " + ex.Message);
            }
            finally
            {
                try
                {
                    // Reset the task source
                    this._abortedCts = null;

                    // If _requestAborted is set, the connection has already been closed.
                    if (Volatile.Read(ref _requestAborted) == 0)
                    {
                        this.Close(CloseType.SocketShutdown);
                    }

                    // Also dispose
                    this.OnDispose();
                }
                catch (Exception ex)
                {
                    Service.Logger.Log(ex);
                    //Service.Logger.Log(LogLevel.Warning,  "Connection shutdown abnormally. Reason: " + ex.Message));
                }
            }
        }

        /// <summary>
        /// Applies a connection filter to the stream.
        /// </summary>
        private void ApplyConnectionFilter()
        {
            lock (_stateLock)
            {
                if (State == ConnectionState.Creating)
                {
                    State = ConnectionState.Open;
                    if (_filterContext.Connection != _libuvStream)
                    {
                        _filteredStreamAdapter = new FilteredStreamAdapter(ConnectionId, _filterContext.Connection, Memory, ThreadPool, _bufferSizeControl);
                        SocketInput = _filteredStreamAdapter.SocketInput;
                        SocketOutput = _filteredStreamAdapter.SocketOutput;
                        _readInputTask = _filteredStreamAdapter.ReadInputAsync();
                    }
                    else
                    {
                        SocketInput = _rawSocketInput;
                        SocketOutput = _rawSocketOutput;
                    }

                    this.OnConnect();
                }
                else
                {
                    this.Close(CloseType.SocketDisconnect);
                }
            }
        }

        /// <summary>
        /// Performs application-defined tasks associated with freeing, releasing, or
        /// resetting unmanaged resources.
        /// </summary>
        private void OnDispose()
        {
            // If we've not yet disposed
            if (Interlocked.CompareExchange(ref Disposed, 1, 0) == 0)
            {
                // Debug write
                //NetTrace.WriteLine("Disposing channel: " + this.RemoteEndPoint.ToString() + ", Binding: " + this.Binding.GetType(), this, NetTraceCategory.Channel);
                //NetTrace.WriteLine("Disposing channel", this, NetTraceCategory.Channel);

                // If we have a close handler, invoke it
                if (this.Disconnect != null)
                {
                    // Make sure we never throw here
                    try { this.Disconnect(this); }
                    catch { }
                }

                // Unbind the client from the connection
                try { this.Client?.UnbindChannel(this); } catch { }

                // Make sure we release the memory associated with the encoding part
                if (this.EncodingSettings != null)
                {
                    this.EncodingSettings.Dispose();
                    this.EncodingSettings = null;
                }

                // Make sure we release the memory associated with the decoding part
                if (this.DecodingSettings != null)
                {
                    this.DecodingSettings.Dispose();
                    this.DecodingSettings = null;
                }

                // Unregister the connection from the manager
                try { this.ConnectionManager.Unregister(this); }
                catch { }
            }
        }

        /// <summary>
        /// Invoked by the GC when the connection needs to be finalized.
        /// </summary>
        ~Connection()
        {
            // Make sure we disposed this properly
            this.OnDispose();
        }

        #endregion OnConnect/OnReceive... Members

        #region Send Method

        /// <summary>
        /// Sends a <see cref="Packet"/> through this channel.
        /// </summary>
        /// <param name="packet">The packet to send through the channel.</param>
        /// <returns>Whether the send was successful or not.</returns>
        public void Send(Packet packet)
        {
            // Check if we are processing
            if (!this.IsRunning || packet == null)
                return;

            // Get a context for this send
            using (var context = ProcessingContext.Acquire(this, this.Encoding))
            {
                try
                {
                    // Encode the packet, then queue the buffer
                    var buffer = context.OnSend(packet);
                    if (buffer != null)
                    {
                        if (buffer.Length <= 0)
                        {
                            // We have to release the buffer since we are not going to send it and
                            // the packet was already compiled.
                            buffer.TryRelease();
                            return;
                        }

                        // We don't need to await here, since WriteAsync() will actually copy the buffer
                        // segment into a tail, we can dispose the thing straight away.
                        this.WriteAsync(buffer.AsSegment(), default(CancellationToken)).Forget();

                        // Release the buffer once we sent the thing
                        buffer.TryRelease();
                    }
                }
                /*catch (CapacityExceededException)
                {
                    Service.Logger.Log(LogLevel.Info, String.Format("{0} disconnected. Too much data pending.", this));
                    OnDispose(false);
                    return;
                }*/
#if DEBUG
                catch (Exception ex) { Service.Logger.Log(ex); return; }
#else
                catch (Exception) { return; }
#endif
            }
        }

        #endregion Send Method

        #region CheckAlive

        /// <summary>
        /// Checks whether the underlying network mechanism is alive or not.
        /// </summary>
        /// <returns>Whether the underlying network mechanism is alive or not.</returns>
        public bool CheckAlive()
        {
            if (!this.IsRunning)
                return false;

            if (this.MeshIdentifier != 0)
                return true;

            if (Timer.Now < this.Expires)
                return true;

            // Trace the inactivity timeout
            NetTrace.WriteLine(this + " disconnected due to inactivity timeout", this, NetTraceCategory.Channel);
            return false;
        }

        #endregion CheckAlive

        #region IConnection Members

        /// <summary>
        /// Pauses the connection (stops the read from it).
        /// </summary>
        public void Pause()
        {
            _socket.ReadStop();
        }

        /// <summary>
        /// Resumes the paused connection (starts reading from it again).
        /// </summary>
        public void Resume()
        {
            try
            {
                _socket.ReadStart(_allocCallback, _readCallback, this);
            }
            catch (UvException)
            {
                // ReadStart() can throw a UvException in some cases (e.g. socket is no longer connected).
                // This should be treated the same as OnRead() seeing a "normalDone" condition.
                _rawSocketInput.IncomingComplete(0, null);
            }
        }

        /// <summary>
        /// Terminates the connection.
        /// </summary>
        /// <param name="endType">The type of a connection close.</param>
        private void Close(CloseType endType)
        {
            lock (_stateLock)
            {
                switch (endType)
                {
                    case CloseType.ConnectionKeepAlive:
                        if (State != ConnectionState.Open)
                        {
                            return;
                        }

                        //Log.ConnectionKeepAlive(ConnectionId);
                        break;

                    case CloseType.SocketShutdown:
                    case CloseType.SocketDisconnect:
                        if (State == ConnectionState.Disconnecting ||
                            State == ConnectionState.SocketClosed)
                        {
                            return;
                        }
                        State = ConnectionState.Disconnecting;

                        //Log.ConnectionDisconnect(ConnectionId);
                        _rawSocketOutput.End(endType);
                        break;
                }
            }
        }

        /// <summary>
        /// Called externally to close the socket.
        /// </summary>
        public void Close()
        {
            this.Close(CloseType.SocketShutdown);
        }

        /// <summary>
        /// Writes some data to the connection.
        /// </summary>
        public void Write(ArraySegment<byte> data)
        {
            // Update the expires timer
            this.Expires = Timer.Now + Timeout;

            // Increment the counters
            NetStat.BytesOutgoing.Add(data.Count);

            // Write through the socket
            SocketOutput.Write(data);
        }

        /// <summary>
        /// Writes some data to the connection.
        /// </summary>
        public Task WriteAsync(ArraySegment<byte> data, CancellationToken cancellationToken)
        {
            // Update the expires timer
            this.Expires = Timer.Now + Timeout;

            // Increment the counters
            NetStat.BytesOutgoing.Add(data.Count);

            // Write through the socket
            return SocketOutput.WriteAsync(data, cancellationToken: cancellationToken);
        }

        /// <summary>
        /// Flushes the connection.
        /// </summary>
        public void Flush()
        {
            SocketOutput.Write(EmptyBuffer);
        }

        /// <summary>
        /// Flushes the connection.
        /// </summary>
        public async Task FlushAsync(CancellationToken cancellationToken)
        {
            await SocketOutput.WriteAsync(EmptyBuffer, cancellationToken: cancellationToken);
        }

        #endregion IConnection Members

        #region Cleanup Members

        /// <summary>
        /// Stops the connection asynchronously.
        /// </summary>
        internal Task StopAsync()
        {
            lock (_stateLock)
            {
                switch (State)
                {
                    case ConnectionState.SocketClosed:
                        return TaskUtilities.CompletedTask;

                    case ConnectionState.Creating:
                        State = ConnectionState.ToDisconnect;
                        break;

                    case ConnectionState.Open:
                        this.StopProcessing();
                        this.SocketInput.CompleteAwaiting();
                        break;
                }

                _socketClosedTcs = new TaskCompletionSource<object>();
                return _socketClosedTcs.Task;
            }
        }

        /// <summary>
        /// Aborts the connection.
        /// </summary>
        internal virtual void Abort()
        {
            // Frame.Abort calls user code while this method is always
            // called from a libuv thread.
            ThreadPool.Run(() =>
            {
                var connection = this;

                lock (connection._stateLock)
                {
                    if (connection.State == ConnectionState.Creating)
                    {
                        connection.State = ConnectionState.ToDisconnect;
                    }
                    else
                    {
                        connection.AbortProcessing();
                    }
                }
            });
        }

        /// <summary>
        /// Called on libuv thread.
        /// </summary>
        internal virtual void OnSocketClosed()
        {
            if (_filteredStreamAdapter != null)
            {
                _filteredStreamAdapter.Abort();
                _rawSocketInput.IncomingFin();
                _readInputTask.ContinueWith((task, state) =>
                {
                    ((Connection)state)._filterContext.Connection.Dispose();
                    ((Connection)state)._filteredStreamAdapter.Dispose();
                    ((Connection)state)._rawSocketInput.Dispose();
                }, this);
            }
            else
            {
                _rawSocketInput.Dispose();
            }

            lock (_stateLock)
            {
                State = ConnectionState.SocketClosed;

                if (_socketClosedTcs != null)
                {
                    // This is always waited on synchronously, so it's safe to
                    // call on the libuv thread.
                    _socketClosedTcs.TrySetResult(null);
                }
            }

            // We're already in the libuv thread, simply abort the processing
            AbortProcessing();
        }

        #endregion Cleanup Members

        #region Alloc/Read Members

        /// <summary>
        /// Occurs when libuv needs to allocate.
        /// </summary>
        private static Libuv.uv_buf_t AllocCallback(UvStreamHandle handle, int suggestedSize, object state)
        {
            return ((Connection)state).OnAlloc(handle, suggestedSize);
        }

        /// <summary>
        /// Occurs when libuv needs to allocate.
        /// </summary>
        private Libuv.uv_buf_t OnAlloc(UvStreamHandle handle, int suggestedSize)
        {
            var result = _rawSocketInput.IncomingStart();

            return handle.Libuv.buf_init(
                result.DataArrayPtr + result.End,
                result.Data.Offset + result.Data.Count - result.End);
        }

        /// <summary>
        /// Occurs when a read is performed on the connection.
        /// </summary>
        private static void ReadCallback(UvStreamHandle handle, int status, object state)
        {
            ((Connection)state).OnRead(handle, status);
        }

        /// <summary>
        /// Occurs when a read is performed on the connection.
        /// </summary>
        private void OnRead(UvStreamHandle handle, int status)
        {
            if (status == 0)
            {
                // A zero status does not indicate an error or connection end. It indicates
                // there is no data to be read right now.
                // See the note at http://docs.libuv.org/en/v1.x/stream.html#c.uv_read_cb.
                // We need to clean up whatever was allocated by OnAlloc.
                _rawSocketInput.IncomingDeferred();
                return;
            }

            var normalRead = status > 0;
            var normalDone = status == Constants.ECONNRESET || status == Constants.EOF;
            var errorDone = !(normalDone || normalRead);
            var readCount = normalRead ? status : 0;
            if (!normalRead)
            {
                _socket.ReadStop();
                this.AbortProcessing();
            }

            Exception error = null;
            if (errorDone)
            {
                handle.Libuv.Check(status, out error);
            }

            _rawSocketInput.IncomingComplete(readCount, error);
            if (errorDone)
            {
                Abort();
            }
        }

        private enum ConnectionState
        {
            Creating,
            ToDisconnect,
            Open,
            Disconnecting,
            SocketClosed
        }

        #endregion Alloc/Read Members
    }

    /// <summary>
    /// The type of the connection close.
    /// </summary>
    public enum CloseType
    {
        SocketShutdown,
        SocketDisconnect,
        ConnectionKeepAlive,
    }
}