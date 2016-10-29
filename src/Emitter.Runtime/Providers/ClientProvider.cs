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
using System.Linq;
using Emitter.Network;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider for the <see cref="IClient"/> related functionnality.
    /// </summary>
    public abstract class ClientProvider : Provider
    {
        /// <summary>
        /// Gets a default client for a newly created channel.
        /// </summary>
        /// <param name="channel">The channel that have been just created.</param>
        /// <returns>The client to bind to the channel</returns>
        public abstract IClient GetDefaultClient(Connection channel);

        /// <summary>
        /// Gets the default upper memory limit for a particular type of processing.
        /// </summary>
        /// <param name="channel">The channel to set the limit for.</param>
        /// <param name="processingType">The processing type.</param>
        /// <returns>The amount of memory in bytes. This should be a power of two.</returns>
        public virtual int GetMaxMemoryFor(Connection channel, ProcessingType processingType)
        {
            // If it's a mesh binding
            if (channel.Binding is MeshBinding)
                return (1 << (BuddyBlock.DefaultMaxMemory + 4)); // 2^26 = 64 MB

            // If it's a normal binding
            return (1 << BuddyBlock.DefaultMaxMemory); // 2 ^ 22 = 4 MB
        }
    }

    /// <summary>
    /// Represents a provider for the <see cref="IClient"/> related functionnality.
    /// </summary>
    public sealed class DefaultClientProvider : ClientProvider
    {
        /// <summary>
        /// Gets a default client for a newly created channel.
        /// </summary>
        /// <param name="channel">The channel that have been just created.</param>
        /// <returns>The client to bind to the channel</returns>
        public override IClient GetDefaultClient(Connection channel)
        {
            return new Client();
        }
    }
}