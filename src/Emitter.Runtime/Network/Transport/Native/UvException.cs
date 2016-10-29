// Copyright (c) .NET Foundation. All rights reserved.
// Licensed under the Apache License, Version 2.0. See License.txt in the project root for license information.

using System;

namespace Emitter.Network.Native
{
    internal class UvException : Exception
    {
        private string Caller;

        public UvException(string message, string source) : base(message)
        {
            this.Caller = source;
        }

        public override string Source
        {
            get
            {
                return this.Caller;
            }
        }
    }
}