// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Net;
using System.Runtime.InteropServices;

namespace Emitter.Network.Native
{
    /// <summary>
    /// Represents a TCP handle.
    /// </summary>
    internal class UvTcpHandle : UvStreamHandle
    {
        public UvTcpHandle() : base()
        {
        }

        public void Init(UvLoopHandle loop, Action<Action<IntPtr>, IntPtr> queueCloseHandle)
        {
            CreateHandle(
                loop.Libuv,
                loop.ThreadId,
                loop.Libuv.handle_size(Libuv.HandleType.TCP), queueCloseHandle);

            _uv.tcp_init(loop, this);
        }

        public void Bind(ServiceAddress address)
        {
            var addr = GetAddress(address);
            _uv.tcp_bind(this, ref addr, 0);
        }

        public void Connect(UvLoopHandle loop, ServiceAddress address, Action<UvConnectRequest, int, Exception, object> callback, object state)
        {
            var addr = GetAddress(address);

            var request = new UvConnectRequest();
            request.Init(loop);
            request.Connect(this, ref addr, callback, state);
        }

        public IPEndPoint GetPeerIPEndPoint()
        {
            SockAddr socketAddress;
            int namelen = Marshal.SizeOf<SockAddr>();
            _uv.tcp_getpeername(this, out socketAddress, ref namelen);

            return socketAddress.GetIPEndPoint();
        }

        public IPEndPoint GetSockIPEndPoint()
        {
            SockAddr socketAddress;
            int namelen = Marshal.SizeOf<SockAddr>();
            _uv.tcp_getsockname(this, out socketAddress, ref namelen);

            return socketAddress.GetIPEndPoint();
        }

        public void Open(IntPtr hSocket)
        {
            _uv.tcp_open(this, hSocket);
        }

        /// <summary>
        /// Disables or enables Nagle on the socket.
        /// </summary>
        /// <param name="enable">Whether to disable or enable the nagle.</param>
        public void NoDelay(bool enable)
        {
            _uv.tcp_nodelay(this, enable);
        }

        /// <summary>
        /// Disables or enables TCP KeepAlive on the socket.
        /// </summary>
        /// <param name="enable">Whether to disable or enable the keepalive.</param>
        /// <param name="delay">The number of seconds to configure for the keepalive packet.</param>
        public void KeepAlive(bool enable, uint delay = 45)
        {
            _uv.tcp_keepalive(this, enable, delay);
        }

        /// <summary>
        /// Gets a native address from a service address.
        /// </summary>
        /// <param name="address">The service address to convert.</param>
        /// <returns>The libuv address.</returns>
        private SockAddr GetAddress(ServiceAddress address)
        {
            var endpoint = address.EndPoint;

            SockAddr addr;
            var addressText = endpoint.Address.ToString();

            Exception error1;
            _uv.ip4_addr(addressText, endpoint.Port, out addr, out error1);

            if (error1 != null)
            {
                Exception error2;
                _uv.ip6_addr(addressText, endpoint.Port, out addr, out error2);
                if (error2 != null)
                {
                    throw error1;
                }
            }

            return addr;
        }
    }
}