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
using System.Collections.Generic;
using System.Linq;
using System.Reflection;

namespace Emitter.Providers
{
    internal static class DefaultProviders
    {
        internal static Dictionary<Type, Type> fAssociations = new Dictionary<Type, Type>()
        {
            {typeof(ClientProvider), typeof(DefaultClientProvider)},
            {typeof(LoggingProvider), typeof(MultiTextLoggingProvider)},
            {typeof(HttpProvider), typeof(DefaultHttpProvider)},
            {typeof(SecurityProvider), typeof(EnvironmentSecurityProvider)},
            {typeof(CertificateProvider), typeof(FileCertificateProvider)},
            {typeof(CorsProvider), typeof(DefaultCorsProvider)},
            {typeof(AddressProvider), typeof(DefaultAddressProvider)},
            {typeof(HealthProvider), typeof(DefaultHealthProvider)},
            {typeof(AnalyticsProvider), typeof(EmitterAnalyticsProvider)},
        };

        /// <summary>
        /// Gets default built-in provider
        /// </summary>
        internal static Type GetDefaultProvider(Type baseProvider)
        {
            if (fAssociations.ContainsKey(baseProvider))
                return fAssociations[baseProvider];
            return null;
        }

        /// <summary>
        /// Gets all default associations
        /// </summary>
        internal static Dictionary<Type, Type> Associations
        {
            get { return fAssociations; }
        }

        /// <summary>
        /// Gets first or default provider for a particular base provider
        /// If a default provider does not exists, it will try to search for an implementation
        /// </summary>
        internal static Type GetDefaultOrFistProviderFor(Type baseProvider)
        {
            var first = Service.MetadataProvider
                .Types
                .Where(type => type.IsSubclassOf(baseProvider) && !type.GetTypeInfo().IsAbstract)
                .FirstOrDefault();

            if (first != null)
                return first;
            return GetDefaultProvider(baseProvider);
        }

        /// <summary>
        /// Registers all default providers, but does not overwrite the existing ones.
        /// Called when the configuration specified providers were already added.
        /// </summary>
        internal static void RegisterAllDefault()
        {
            fAssociations.ForEach((key, value) =>
            {
                if (Service.Providers.Resolve(key) == null)
                {
                    if (key.IsSubclassOf(typeof(Provider)))
                        Service.Providers.Register(key, fAssociations[key]);
                    else
                        Service.Providers.Register(key, c => Activator.CreateInstance(fAssociations[key]));
                }
            });
        }
    }
}