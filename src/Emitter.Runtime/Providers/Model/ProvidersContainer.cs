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
using System.Collections.Specialized;
using System.Linq;

namespace Emitter.Providers
{
    /// <summary>
    /// The implementation of the IOC container.
    /// </summary>
    public class ProvidersContainer : IIocContainer, IEnumerable
    {
        private HybridDictionary typeRegistry = new HybridDictionary();

        // Track whether Dispose has been called.
        private bool disposed;

        // null for the lifetime manager is the same as AlwaysNew, but slightly faster.
        private ILifetimeManager defaultLifetimeManager = null;

        /// <summary>
        /// Gets the <see cref="ILifetimeManager"/> associated with this container.
        /// </summary>
        public ILifetimeManager LifeTimeManager
        {
            get { return defaultLifetimeManager; }
        }

        internal ProvidersContainer(Dictionary<Type, Type> defaultRegistrations)
        {
            defaultRegistrations.ForEach((key, value) =>
            {
                if (this.Resolve(key) == null)
                {
                    if (key.IsSubclassOf(typeof(Provider)))
                        this.Register(key, defaultRegistrations[key]);
                    else
                        this.Register(key, c => Activator.CreateInstance(defaultRegistrations[key]));
                }
            });
        }

        /// <summary>
        /// Initializes a new instance of the ProvidersContainer class.
        /// </summary>
        public ProvidersContainer()
        {
        }

        #region Register Members

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container.
        /// </summary>
        /// <typeparam name="TType">The type being registered.</typeparam>
        /// <param name="func">The function that creates the type. The function takes a single parameter of type Container.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        public IRegistration Register<TType>(Func<IIocContainer, TType> func) where TType : class
        {
            if (func == null)
                throw new ArgumentNullException("func", "func is null.");
            return Register(typeof(TType), c => (func(c) as Object));
        }

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container.
        /// </summary>
        /// <typeparam name="TType">The type being registered.</typeparam>
        /// <param name="name">The name of the Registration for this type.  Use to distinguish between different Registrations.</param>
        /// <param name="func">The function that creates the type. The function takes a single parameter of type Container.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        public IRegistration Register<TType>(string name, Func<IIocContainer, TType> func) where TType : class
        {
            if (func == null)
                throw new ArgumentNullException("func", "func is null.");
            return Register(name, typeof(TType), c => (func(c) as Object));
        }

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container.
        /// </summary>
        /// <param name="type">The type being registered.</param>
        /// <param name="func">The function that creates the type. The function takes a single parameter of type Container.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        public IRegistration Register(Type type, Func<IIocContainer, object> func)
        {
            if (func == null)
                throw new ArgumentNullException("func");

            var entry = new Registration(this, null, type, func);
            entry.WithLifetimeManager(defaultLifetimeManager);

            typeRegistry[new UnNamedRegistrationKey(type)] = entry;

            return entry;
        }

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container.
        /// </summary>
        /// <param name="type">The type being registered.</param>
        /// <param name="name">The name of the Registration for this type.  Use to distinguish between different Registrations.</param>
        /// <param name="func">The function that creates the type. The function takes a single parameter of type Container.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        public IRegistration Register(string name, Type type, Func<IIocContainer, object> func)
        {
            if (func == null)
                throw new ArgumentNullException("func");

            var entry = new Registration(this, name ?? String.Empty, type, func);
            entry.WithLifetimeManager(defaultLifetimeManager);

            typeRegistry[new NamedRegistrationKey(name, type)] = entry;

            return entry;
        }

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container. This
        /// also performs provider initialization and instanciation.
        /// </summary>
        /// <param name="providerBaseType">The base type provider.</param>
        /// <param name="providerType">The type of the provider to instanciate. Must be a sub-class of base type provider.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        public IRegistration Register(Type providerBaseType, Type providerType)
        {
            var providerInstance = Activator.CreateInstance(providerType) as Provider;
            if (providerInstance == null)
                throw new ArgumentException(String.Format("Provider {0} could not be created. Make sure it has a public parameterless constructor and it's derived from Provider.", providerType.Name));

            // providerInstance.Initialize(providerType.Name, config ?? new NameValueCollection());
            providerInstance.Initialize(providerType.Name, null);

            // Make sure we have only one instance of the provider, that's why we are calling RegisterInstance
            IRegistration registration = RegisterInstance(providerBaseType, providerInstance);

            // Invoke the registration event
            Service.InvokeProviderRegistered(new ProviderRegisteredEventArgs(providerBaseType, providerInstance));
            return registration;
        }

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container. This
        /// also performs provider initialization and instanciation.
        /// </summary>
        /// <typeparam name="TProviderBaseType">The base type provider</typeparam>
        /// <param name="providerType">The type of the provider to instanciate. Must be a sub-class of base type provider.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        public IRegistration Register<TProviderBaseType>(Type providerType)
        {
            return Register(typeof(TProviderBaseType), providerType);
        }

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container. This
        /// also performs provider initialization and instanciation.
        /// </summary>
        /// <typeparam name="TProviderBaseType">The base type provider</typeparam>
        /// <param name="providerTypeName">The type of the provider to instanciate. Must be a sub-class of base type provider.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        public IRegistration Register<TProviderBaseType>(string providerTypeName)
        {
            if (string.IsNullOrWhiteSpace(providerTypeName))
                return null;

            // Get the first one
            var providerType = Service.MetadataProvider.Assemblies
                    .SelectMany(x => x.GetTypes())
                    .FirstOrDefault(x => x.Name == providerTypeName);
            if (providerType == null)
                throw new KeyNotFoundException("The provider '" + providerTypeName + "' was not found.");

            return Register(typeof(TProviderBaseType), providerType);
        }

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container. This
        /// also performs provider initialization and instanciation.
        /// </summary>
        /// <typeparam name="TProviderBaseType">The base type provider</typeparam>
        /// <param name="providerInstance">The instance of the provider. Must be a sub-class of base type provider.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        public IRegistration Register<TProviderBaseType>(Provider providerInstance)
        {
            var providerBaseType = typeof(TProviderBaseType);
            var providerType = providerInstance.GetType();
            providerInstance.Initialize(providerType.Name, null);

            // Make sure we have only one instance of the provider, that's why we are calling RegisterInstance
            IRegistration registration = RegisterInstance(providerBaseType, providerInstance);

            // Invoke the registration event
            Service.InvokeProviderRegistered(new ProviderRegisteredEventArgs(providerBaseType, providerInstance));
            return registration;
        }

        #endregion Register Members

        #region RegisterInstance Members

        /// <summary>
        /// Registers an instance that will be returned whenever the IocContainer resolves the specified type.
        /// </summary>
        /// <typeparam name="TType">The type that is being registered for resolution.</typeparam>
        /// <param name="instance">The instance that will alway be returned when type is resolved.</param>
        /// <returns>An instance of IRegistration that can be used to configure how the get information about
        /// the registration, or change the lifetime manager.</returns>
		public IRegistration RegisterInstance<TType>(TType instance) where TType : class
        {
            return Register<TType>(c => instance);
        }

        /// <summary>
        /// Registers an instance that will be returned whenever the IocContainer resolves the specified type.
        /// </summary>
        /// <typeparam name="TType">The type that is being registered for resolution.</typeparam>
        /// <param name="name">The name this registration will be registered under.</param>
        /// <param name="instance">The instance that will alway be returned when type is resolved.</param>
        /// <returns>An instance of IRegistration that can be used to configure how the get information about
        /// the registration, or change the lifetime manager.</returns>
        public IRegistration RegisterInstance<TType>(string name, TType instance) where TType : class
        {
            return Register<TType>(name, c => instance);
        }

        /// <summary>
        /// Registers an instance that will be returned whenever the IocContainer resolves the specified type.
        /// </summary>
        /// <param name="type">The type that is being registered for resolution.</param>
        /// <param name="instance">The instance that will alway be returned when type is resolved.</param>
        /// <returns>An instance of IRegistration that can be used to configure how the get information about
        /// the registration, or change the lifetime manager.</returns>
        public IRegistration RegisterInstance(Type type, object instance)
        {
            return Register(type, c => instance);
        }

        /// <summary>
        /// Registers an instance that will be returned whenever the IocContainer resolves the specified type.
        /// </summary>
        /// <param name="name">The name this registration will be registered under.</param>
        /// <param name="type">The type that is being registered for resolution.</param>
        /// <param name="instance">The instance that will alway be returned when type is resolved.</param>
        /// <returns>An instance of IRegistration that can be used to configure how the get information about
        /// the registration, or change the lifetime manager.</returns>
		public IRegistration RegisterInstance(string name, Type type, object instance)
        {
            return Register(name, type, c => instance);
        }

        #endregion RegisterInstance Members

        #region Unregister Members

        /// <summary>
        /// Removes a registration from the container.
        /// </summary>
        /// <param name="ireg">The registration to remove from the container.</param>
        public void Unregister(IRegistration ireg)
        {
            object key;
            if (ireg.Name == null)
                key = new UnNamedRegistrationKey(ireg.ResolvesTo);
            else
                key = new NamedRegistrationKey(ireg.Name, ireg.ResolvesTo);

            typeRegistry.Remove(key);

            ireg.InvalidateInstanceCache();
        }

        #endregion Unregister Members

        #region Resolve Members

        /// <summary>
        /// Returns an instance of a registered type.
        /// </summary>
        /// <typeparam name="TType">The type to resolve</typeparam>
        /// <returns>An instance of the type.  Throws a KeyNoFoundException if not registered.</returns>
        public TType Resolve<TType>() where TType : class
        {
            return Resolve(typeof(TType)) as TType;
        }

        /// <summary>
        /// Returns an instance of a registered type.
        /// </summary>
        /// <typeparam name="TType">The type to resolve.</typeparam>
        /// <param name="name">The name to resolve.</param>
        /// <returns>An instance of the type.  Throws a KeyNoFoundException if not registered.</returns>
        public TType Resolve<TType>(string name) where TType : class
        {
            return Resolve(name, typeof(TType)) as TType;
        }

        /// <summary>
        /// Returns an instance of a registered type.
        /// </summary>
        /// <param name="type">The type to resolve.</param>
        /// <returns>An instance of the type.  Throws a KeyNoFoundException if not registered.</returns>
        public object Resolve(Type type)
        {
            var entry = typeRegistry[new UnNamedRegistrationKey(type)] as Registration;

            if (entry != null)
                return entry.GetInstance();
            return null;
        }

        /// <summary>
        /// Returns an instance of a registered type.
        /// </summary>
        /// <param name="type">The type to resolve.</param>
        /// <param name="name">The name to resolve.</param>
        /// <returns>An instance of the type.  Throws a KeyNoFoundException if not registered.</returns>
        public object Resolve(string name, Type type)
        {
            var entry = typeRegistry[new NamedRegistrationKey(name, type)] as Registration;

            if (entry != null)
                return entry.GetInstance();
            return null;
        }

        #endregion Resolve Members

        #region GetRegistration Members

        /// <summary>
        /// Returns an Registration of a registered type.
        /// </summary>
        /// <typeparam name="TType">The type to get the Registration for.</typeparam>
        /// <returns>An Registration for the type. Throws a KeyNoFoundException if not registered.</returns>
        public IRegistration GetRegistration<TType>() where TType : class
        {
            return GetRegistration(typeof(TType));
        }

        /// <summary>
        /// Returns an Registration of a registered type.
        /// </summary>
        /// <typeparam name="TType">The type to get the Registration for.</typeparam>
        /// <param name="name">The name associated with the named registration to get.</param>
        /// <returns>An Registration for the type. Throws a KeyNoFoundException if not registered.</returns>
        public IRegistration GetRegistration<TType>(string name) where TType : class
        {
            return GetRegistration(name, typeof(TType));
        }

        /// <summary>
        /// Returns an Registration of a registered type.
        /// </summary>
        /// <param name="type">The type to get the Registration for.</param>
        /// <returns>An Registration for the type. Throws a KeyNoFoundException if not registered.</returns>
        public IRegistration GetRegistration(Type type)
        {
            var reg = typeRegistry[new UnNamedRegistrationKey(type)] as IRegistration;
            if (reg == null)
                throw new KeyNotFoundException();

            return reg;
        }

        /// <summary>
        /// Returns an Registration of a registered type.
        /// </summary>
        /// <param name="type">The type to get the Registration for.</param>
        /// <param name="name">The name associated with the named registration to get.</param>
        /// <returns>An Registration for the type. Throws a KeyNoFoundException if not registered.</returns>
        public IRegistration GetRegistration(string name, Type type)
        {
            var reg = typeRegistry[new NamedRegistrationKey(name, type)] as IRegistration;
            if (reg == null)
                throw new KeyNotFoundException();

            return reg;
        }

        /// <summary>
        /// Gets all Registrations for the specified type.
        /// </summary>
        /// <typeparam name="TType">The type for which the Registrations are required.</typeparam>
        /// <returns>A list of the registration of the specified type.</returns>
        public List<IRegistration> GetRegistrations<TType>() where TType : class
        {
            return GetRegistrations(typeof(TType));
        }

        /// <summary>
        /// Gets all Registrations for the specified type.
        /// </summary>
        /// <param name="type">The type for which the Registrations are required.</param>
        /// <returns>A list of the registration of the specified type.</returns>
        public List<IRegistration> GetRegistrations(Type type)
        {
            List<IRegistration> registrations = new List<IRegistration>();
            foreach (IRegistrationKey key in typeRegistry.Keys)
            {
                if (key.GetInstanceType() == type)
                    registrations.Add((IRegistration)typeRegistry[key]);
            }
            return registrations;
        }

        #endregion GetRegistration Members

        #region Fluent Interface Members

        /// <summary>
        /// Specifies a default lifetime manager to use with this container.
        /// </summary>
        /// <param name="lifetimeManager">The lifetime manager to use by default.</param>
        /// <returns>Returns a reference to this container.</returns>
        public IIocContainer UsesDefaultLifetimeManagerOf(ILifetimeManager lifetimeManager)
        {
            defaultLifetimeManager = lifetimeManager;
            return this;
        }

        #endregion Fluent Interface Members

        #region IDisposable Members

        /// <summary>
        /// Performs application-defined tasks associated with freeing, releasing, or resetting unmanaged resources.
        /// </summary>
        /// <remarks>
        /// Disposes of all Container scoped (ContainerLifetime) instances cached in the type registry, and
        /// disposes of the type registry itself.
        /// </remarks>
        public void Dispose()
        {
            Dispose(true);
            GC.SuppressFinalize(this);
        }

        /// <summary>
        /// Implements the Disposed(boolean disposing) method of Disposable pattern.
        /// </summary>
        /// <param name="disposing">True if disposing.</param>
        protected virtual void Dispose(bool disposing)
        {
            // Check to see if Dispose has already been called.
            if (!disposed)
            {
                // If disposing equals true, dispose all ContainerLifetime instances
                if (disposing)
                {
                    foreach (Registration reg in typeRegistry.Values)
                    {
                        var instance = reg.Instance as IDisposable;
                        if (instance != null)
                        {
                            instance.Dispose();
                            reg.Instance = null;
                        }
                    }
                }
            }
            disposed = true;
        }

        /// <summary>
        /// The finalizer just ensures the container is disposed.
        /// </summary>
        ~ProvidersContainer() { Dispose(false); }

        #endregion IDisposable Members

        #region IEnumerable Members

        /// <summary>
        /// Returns an enumerator that iterates through a collection.
        /// </summary>
        /// <returns>Returns an enumerator that iterates through a collection.</returns>
        public IEnumerator GetEnumerator()
        {
            return typeRegistry.GetEnumerator();
        }

        #endregion IEnumerable Members
    }
}