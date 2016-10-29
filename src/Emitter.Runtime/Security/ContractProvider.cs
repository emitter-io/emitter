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
using System.Collections;
using System.Collections.Generic;

namespace Emitter.Security
{
    /// <summary>
    /// Represents a provider of <see cref="IContract"/> information.
    /// </summary>
    public abstract class ContractProvider : Provider, IEnumerable<IContract>
    {
        /// <summary>
        /// Randomly generates a 32-bit key.
        /// </summary>
        /// <returns></returns>
        protected int GenerateRandom()
        {
            return SecurityRandom.Default.Next();
        }

        /// <summary>
        /// Attempts to fetch a contract by its id.
        /// </summary>
        /// <param name="key">The key to search by.</param>
        /// <returns>Whether the key was able to fetch.</returns>
        public abstract IContract GetByKey(int key);

        /// <summary>
        /// Attempts to fetch a contract by its tag.
        /// </summary>
        /// <param name="tag">The tag to search by.</param>
        /// <returns>Whether the key was able to fetch.</returns>
        public abstract IContract GetByTag(string tag);

        /// <summary>
        /// Creates a new instance of a <see cref="IContract"/> in the underlying data storage.
        /// </summary>
        /// <param name="tag">The tag of the contract for further loosely-coupled linking.</param>
        /// <returns>The newly created instance of a <see cref="IContract"/>.</returns>
        public abstract IContract Create(string tag);

        /// <summary>
        /// Creates a master key for this contract.
        /// </summary>
        /// <param name="contract">The contract to write into the key.</param>
        /// <param name="index">The index to write to the key.</param>
        /// <param name="signature">The signature.</param>
        /// <returns>The newly generated master key.</returns>
        public SecurityKey CreateMasterKey(int contract, int signature, int index)
        {
            var key = SecurityKey.Create();
            key.Master = (ushort)index;               // Also store the reference to itself
            key.Contract = contract;                  // Store the contract id
            key.Signature = signature;                // The signature of the contract
            key.Permissions = SecurityAccess.Master;  // Permission of 1 means it's a master key
            key.Target = 0;                           // Master key does not have a target
            return key;
        }

        /// <summary>
        /// Iterates through all the contracts in the provider.
        /// </summary>
        /// <returns></returns>
        public abstract IEnumerator<IContract> GetEnumerator();

        /// <summary>
        /// Iterates through all the contracts in the provider.
        /// </summary>
        /// <returns></returns>
        IEnumerator IEnumerable.GetEnumerator()
        {
            return this.GetEnumerator();
        }
    }

    /// <summary>
    /// Represents the status of the key.
    /// </summary>
    public enum GetStatus
    {
        /// <summary>
        /// The key is valid.
        /// </summary>
        Success = 0,

        /// <summary>
        /// The key has expired.
        /// </summary>
        KeyExpired = 1,

        /// <summary>
        /// The key was revoked.
        /// </summary>
        KeyRevoked = 2,

        /// <summary>
        /// The key was not associated with a contract.
        /// </summary>
        NotFound = 3,

        /// <summary>
        /// The key was not parsed properly.
        /// </summary>
        BadKey = 4,

        /// <summary>
        /// The key was not authorized.
        /// </summary>
        Unauthorized = 5
    }
}