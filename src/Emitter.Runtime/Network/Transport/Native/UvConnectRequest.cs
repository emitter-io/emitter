// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;

namespace Emitter.Network.Native
{
    /// <summary>
    /// Summary description for UvWriteRequest
    /// </summary>
    internal class UvConnectRequest : UvRequest
    {
        private readonly static Libuv.uv_connect_cb _uv_connect_cb = (req, status) => UvConnectCb(req, status);

        private Action<UvConnectRequest, int, Exception, object> _callback;
        private object _state;

        public UvConnectRequest() : base()
        {
        }

        public void Init(UvLoopHandle loop)
        {
            var requestSize = loop.Libuv.req_size(Libuv.RequestType.CONNECT);
            CreateMemory(
                loop.Libuv,
                loop.ThreadId,
                requestSize);
        }

        public void Connect(
            UvPipeHandle pipe,
            string name,
            Action<UvConnectRequest, int, Exception, object> callback,
            object state)
        {
            _callback = callback;
            _state = state;

            Pin();
            Libuv.pipe_connect(this, pipe, name, _uv_connect_cb);
        }

        public void Connect(
             UvTcpHandle tcp,
             ref SockAddr address,
             Action<UvConnectRequest, int, Exception, object> callback,
             object state)
        {
            _callback = callback;
            _state = state;

            Pin();
            _uv.tcp_connect(this, tcp, ref address, _uv_connect_cb);
        }

        private static void UvConnectCb(IntPtr ptr, int status)
        {
            var req = FromIntPtr<UvConnectRequest>(ptr);
            req.Unpin();

            var callback = req._callback;
            req._callback = null;

            var state = req._state;
            req._state = null;

            Exception error = null;
            if (status < 0)
            {
                req.Libuv.Check(status, out error);
            }

            try
            {
                callback(req, status, error, state);
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                throw;
            }
        }
    }
}