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
using Emitter.Replication;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider of a registry that should be replicated across the cluster.
    /// </summary>
    public abstract class RegistryProvider : Provider
    {
        /// <summary>
        /// Gets the version of the replicated
        /// </summary>
        public abstract ReplicatedVersion Version
        {
            get;
        }

        /// <summary>
        /// Gets the replicated collection
        /// </summary>
        public abstract ReplicatedHybridDictionary Collection
        {
            get;
        }

        /// <summary>
        /// Retrieves the value from the registry. The value should be able to replicate and is created
        /// if no value is found.
        /// </summary>
        /// <typeparam name="T">The type of the value to retrieve.</typeparam>
        /// <param name="key">The key to retrieve.</param>
        /// <returns>The value if found, null otherwise.</returns>
        public abstract T Get<T>(string key)
            where T : IReplicated;
    }

    /// <summary>
    /// Represents a provider of a state replicated across the cluster.
    /// </summary>
    public sealed class ReplicatedRegistryProvider : RegistryProvider
    {
        /// <summary>
        /// The underlying replicated state
        /// </summary>
        private readonly ReplicatedHybridDictionary Registry;

        /// <summary>
        /// Constructs a new replicated registry provider.
        /// </summary>
        /// <param name="nodeIdentifier">The node identifier to use for the replication.</param>
        public ReplicatedRegistryProvider(int nodeIdentifier)
        {
            // Create a new registry
            this.Registry = new ReplicatedHybridDictionary(nodeIdentifier);
        }

        /// <summary>
        /// Gets the version of the replicated
        /// </summary>
        public override ReplicatedVersion Version
        {
            get { return this.Registry.Version; }
        }

        /// <summary>
        /// Gets the replicated collection
        /// </summary>
        public override ReplicatedHybridDictionary Collection
        {
            get { return this.Registry; }
        }

        /// <summary>
        /// Retrieves the value from the registry. The value should be able to replicate and is created
        /// if no value is found.
        /// </summary>
        /// <typeparam name="T">The type of the value to retrieve.</typeparam>
        /// <param name="key">The key to retrieve.</param>
        /// <returns>The value if found, null otherwise.</returns>
        public override T Get<T>(string key)
        {
            return this.Registry.GetOrCreate<T>(key);
        }
    }
}