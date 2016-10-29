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
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using Emitter.Network.Http;
using Emitter.Security;
using Emitter.Text.Json;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider of <see cref="EmitterContract"/> information.
    /// </summary>
    public abstract class EmitterContractProvider : ContractProvider
    {
        /// <summary>
        /// Creates a new instance of a <see cref="IContract"/> in the underlying data storage.
        /// </summary>
        /// <param name="tag">The tag of the contract for further loosely-coupled linking.</param>
        /// <returns>The newly created instance of a <see cref="IContract"/>.</returns>
        public override IContract Create(string tag)
        {
            throw new NotSupportedException();
        }

        /// <summary>
        /// Attempts to fetch a contract by its tag.
        /// </summary>
        /// <param name="tag">The tag to search by.</param>
        /// <returns>Whether the key was able to fetch.</returns>
        public override IContract GetByTag(string tag)
        {
            throw new NotSupportedException();
        }
    }

    /// <summary>
    /// Represents a provider of <see cref="EmitterContract"/> information.
    /// </summary>
    public sealed class SingleContractProvider : EmitterContractProvider
    {
        /// <summary>
        /// The single contract provided by this provider.
        /// </summary>
        private readonly EmitterContract Contract;

        /// <summary>
        /// Constructs a new provider for the single, embedded contract.
        /// </summary>
        public SingleContractProvider()
        {
            var contract = new EmitterContract();
            contract.Master = 1;
            contract.Oid = SecurityLicense.Current.Contract;
            contract.Signature = SecurityLicense.Current.Signature;
            this.Contract = contract;
        }

        /// <summary>
        /// Attempts to fetch a contract by its id.
        /// </summary>
        /// <param name="contractKey">The key to search by.</param>
        /// <returns>Whether the key was able to fetch.</returns>
        public override IContract GetByKey(int key)
        {
            if (this.Contract == null || this.Contract.Oid != key)
                return null;
            return this.Contract;
        }

        /// <summary>
        /// Iterates through all the contracts in the provider.
        /// </summary>
        /// <returns></returns>
        public override IEnumerator<IContract> GetEnumerator()
        {
            yield return this.Contract;
        }
    }

    /// <summary>
    /// Represents a provider of <see cref="EmitterContract"/> information.
    /// </summary>
    public sealed class HttpContractProvider : EmitterContractProvider
    {
        private readonly ConcurrentDictionary<int, EmitterContract> Cache =
            new ConcurrentDictionary<int, EmitterContract>(Environment.ProcessorCount, 64);

        private readonly HttpQuery Query = new HttpQuery("http://meta.emitter.io");

        /// <summary>
        /// Constructs a new contract provider.
        /// </summary>
        public HttpContractProvider()
        {
            // TODO: Expire from time to time
        }

        /// <summary>
        /// Attempts to fetch a contract by its id.
        /// </summary>
        /// <param name="contractKey">The key to search by.</param>
        /// <returns>Whether the key was able to fetch.</returns>
        public override IContract GetByKey(int contractKey)
        {
            try
            {
                return this.Cache.GetOrAdd(contractKey, FetchContract);
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                return null;
            }
        }

        /// <summary>
        /// Gets the contract from the meta service.
        /// </summary>
        /// <param name="key"></param>
        /// <returns></returns>
        private EmitterContract FetchContract(int key)
        {
            try
            {
                Service.Logger.Log("Fetch contract #" + key);

                // Query the meta service
                var response = Query.Get("/v1/contract/" + key, 10000);
                if (response != null)
                {
                    // Deserialize the response as JSON
                    return JsonConvert.DeserializeObject<EmitterContract>(response);
                }

                Service.Logger.Log(LogLevel.Error, "Unable to fetch contract #" + key);
                return null;
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
                return null;
            }
        }

        /// <summary>
        /// Iterates through all the contracts in the provider.
        /// </summary>
        /// <returns></returns>
        public override IEnumerator<IContract> GetEnumerator()
        {
            // The enumerator returned from the dictionary is safe to use concurrently with reads and writes to the
            // dictionary, however it does not represent a moment-in-time snapshot of the dictionary. The contents
            // exposed through the enumerator may contain modifications made to the dictionary after GetEnumerator
            // was called.
            foreach (var kvp in this.Cache)
                yield return kvp.Value;
        }
    }
}