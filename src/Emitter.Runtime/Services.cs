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

using Emitter.Providers;
using Emitter.Security;

namespace Emitter
{
    /// <summary>
    /// Represents a static helper class for providers resolution.
    /// </summary>
    internal static class Services
    {
        private static ContractProvider ContractProvider;
        private static StorageProvider StorageProvider;

        /// <summary>
        /// Gets the contracts provider.
        /// </summary>
        public static ContractProvider Contract
        {
            get
            {
                // Resolve the provider
                if (ContractProvider == null)
                    ContractProvider = Service.Providers.Resolve<ContractProvider>();

                // Return the cached provider
                return ContractProvider;
            }
        }

        /// <summary>
        /// Gets the storage provider.
        /// </summary>
        public static StorageProvider Storage
        {
            get
            {
                // Resolve the provider
                if (StorageProvider == null)
                    StorageProvider = Service.Providers.Resolve<StorageProvider>();

                // Return the cached provider
                return StorageProvider;
            }
        }

        /// <summary>
        /// Gets the ip address provider.
        /// </summary>
        public static AddressProvider Address
        {
            get { return Service.Providers.Resolve<AddressProvider>(); }
        }
    }
}