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
using Emitter.Diagnostics;
using Emitter.Replication;
using Emitter.Security;
using Emitter.Text.Json;

namespace Emitter
{
    /// <summary>
    /// Represents a contract.
    /// </summary>
    public class EmitterContract : IContract
    {
        /// <summary>
        /// The cached info pointer.
        /// </summary>
        [JsonIgnore]
        private ReplicatedHybridDictionary CachedInfo = null;

        /// <summary>
        /// Gets the contract id.
        /// </summary>
        [JsonProperty("id")]
        public int Oid;

        /// <summary>
        /// Gets the signature of the key.
        /// </summary>
        [JsonProperty("sign")]
        public int Signature;

        /// <summary>
        /// Gets the master key id.
        /// </summary>
        [JsonProperty("master")]
        public int Master;

        /// <summary>
        /// Gets the state of the contract.
        /// </summary>
        [JsonProperty("state")]
        public EmitterContractStatus Status;

        /// <summary>
        /// Gets the monitor usage for the contract.
        /// </summary>
        [JsonIgnore]
        public MonitorUsage Usage = new MonitorUsage();

        /// <summary>
        /// Gets the state associated with the emitter contract.
        /// </summary>
        [JsonIgnore]
        public ReplicatedHybridDictionary Info
        {
            get
            {
                if (this.CachedInfo == null)
                    this.CachedInfo = Service.Registry.Get<ReplicatedHybridDictionary>(this.Oid.ToHex());
                return this.CachedInfo;
            }
        }

        /// <summary>
        /// Validates the security key.
        /// </summary>
        /// <param name="key">The key reference.</param>
        /// <returns>Whether the key is active or not.</returns>
        public bool Validate(ref SecurityKey key)
        {
            // Using the 'Key' property for lazy cache load
            return this.Master == key.Master &&
                this.Signature == key.Signature &&
                this.Oid == key.Contract;
        }
    }
}