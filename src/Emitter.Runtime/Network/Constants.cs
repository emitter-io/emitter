// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;

using Emitter.Network.Native;

namespace Emitter.Network
{
    internal class Constants
    {
        public const int ListenBacklog = 128;
        public const int ReceiveBufferSize = 8192;

        public const int EOF = -4095;
        public static readonly int ECONNRESET = GetECONNRESET();

        /// <summary>
        /// Prefix of host name used to specify Unix sockets in the configuration.
        /// </summary>
        public const string UnixPipeHostPrefix = "unix:/";

        /// <summary>
        /// DateTime format string for RFC1123. See  https://msdn.microsoft.com/en-us/library/az4se3k1(v=vs.110).aspx#RFC1123
        /// for info on the format.
        /// </summary>
        public const string RFC1123DateFormat = "r";

        private static int GetECONNRESET()
        {
            switch (Platform.System)
            {
                case OS.Windows:
                    return -4077;

                case OS.Linux:
                    return -104;

                case OS.Darwin:
                    return -54;

                default:
                    throw new ArgumentException("Unable to determine the operating system.");
            }
        }
    }
}