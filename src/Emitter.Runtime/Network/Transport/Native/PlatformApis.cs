// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;
using System.Runtime.InteropServices;

namespace Emitter.Network.Native
{
    /// <summary>
    /// Represents runtime plaltform information.
    /// </summary>
    internal static class Platform
    {
        static Platform()
        {
#if DOTNET
            var win = RuntimeInformation.IsOSPlatform(OSPlatform.Windows);
            var nux = RuntimeInformation.IsOSPlatform(OSPlatform.Linux);

            if (win) Platform.System = OS.Windows;
            if (nux) Platform.System = OS.Linux;
#else
            Platform.System = OS.Windows;
#endif
        }

        /// <summary>
        /// Gets the underlying operating system.
        /// </summary>
        public static readonly OS System;
    }

    public enum OS
    {
        Unknown = 0,
        Windows = 1,
        Linux = 2,
        Darwin = 3
    }
}