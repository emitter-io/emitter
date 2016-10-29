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

namespace Emitter.Providers
{
    /// <summary>
    /// Represents an instance creator that is used to create instances by a registration key.
    /// </summary>
    public interface IInstanceCreator
    {
        /// <summary>
        /// Gets the registration key.
        /// </summary>
        string Key { get; }

        /// <summary>
        /// Constructs a new instance of an object.
        /// </summary>
        /// <param name="containerCache">The container caching policy to use for instance creation.</param>
        /// <returns>A constructed instance.</returns>
        object CreateInstance(ContainerCaching containerCache);
    }

    /// <summary>
    /// A container caching policy.
    /// </summary>
    public enum ContainerCaching
    {
        /// <summary>
        /// The instance will be cached in the container once it's created.
        /// </summary>
        InstanceCachedInContainer,

        /// <summary>
        /// The instance will not be cached in the container once it's created.
        /// </summary>
        InstanceNotCachedInContainer
    }

    /// <summary>
    /// Represents an IOC (inversion of control) container.
    /// </summary>
    public interface IIocContainer : IDisposable
    {
        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container.
        /// </summary>
        /// <param name="type">The type being registered.</param>
        /// <param name="name">The name of the Registration for this type.  Use to distinguish between different Registrations.</param>
        /// <param name="func">The function that creates the type. The function takes a single parameter of type Container.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        IRegistration Register(string name, Type type, Func<IIocContainer, object> func);

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container.
        /// </summary>
        /// <param name="type">The type being registered.</param>
        /// <param name="func">The function that creates the type. The function takes a single parameter of type Container.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        IRegistration Register(Type type, Func<IIocContainer, object> func);

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container.
        /// </summary>
        /// <typeparam name="TType">The type being registered.</typeparam>
        /// <param name="func">The function that creates the type. The function takes a single parameter of type Container.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        IRegistration Register<TType>(Func<IIocContainer, TType> func) where TType : class;

        /// <summary>
        /// Adds the function to resolve the unnamed registration of the specified type to the container.
        /// </summary>
        /// <typeparam name="TType">The type being registered.</typeparam>
        /// <param name="name">The name of the Registration for this type.  Use to distinguish between different Registrations.</param>
        /// <param name="func">The function that creates the type. The function takes a single parameter of type Container.</param>
        /// <returns>An IRegistration that can be used to configure the behavior of the registration.</returns>
        IRegistration Register<TType>(string name, Func<IIocContainer, TType> func) where TType : class;

        /// <summary>
        /// Registers an instance that will be returned whenever the IocContainer resolves the specified type.
        /// </summary>
        /// <param name="name">The name this registration will be registered under.</param>
        /// <param name="type">The type that is being registered for resolution.</param>
        /// <param name="instance">The instance that will alway be returned when type is resolved.</param>
        /// <returns>An instance of IRegistration that can be used to configure how the get information about
        /// the registration, or change the lifetime manager.</returns>
        IRegistration RegisterInstance(string name, Type type, object instance);

        /// <summary>
        /// Registers an instance that will be returned whenever the IocContainer resolves the specified type.
        /// </summary>
        /// <param name="type">The type that is being registered for resolution.</param>
        /// <param name="instance">The instance that will alway be returned when type is resolved.</param>
        /// <returns>An instance of IRegistration that can be used to configure how the get information about
        /// the registration, or change the lifetime manager.</returns>
        IRegistration RegisterInstance(Type type, object instance);

        /// <summary>
        /// Registers an instance that will be returned whenever the IocContainer resolves the specified type.
        /// </summary>
        /// <typeparam name="TType">The type that is being registered for resolution.</typeparam>
        /// <param name="name">The name this registration will be registered under.</param>
        /// <param name="instance">The instance that will alway be returned when type is resolved.</param>
        /// <returns>An instance of IRegistration that can be used to configure how the get information about
        /// the registration, or change the lifetime manager.</returns>
        IRegistration RegisterInstance<TType>(string name, TType instance) where TType : class;

        /// <summary>
        /// Registers an instance that will be returned whenever the IocContainer resolves the specified type.
        /// </summary>
        /// <typeparam name="TType">The type that is being registered for resolution.</typeparam>
        /// <param name="instance">The instance that will alway be returned when type is resolved.</param>
        /// <returns>An instance of IRegistration that can be used to configure how the get information about
        /// the registration, or change the lifetime manager.</returns>
        IRegistration RegisterInstance<TType>(TType instance) where TType : class;

        /// <summary>
        /// Removes a registration from the container.
        /// </summary>
        /// <param name="ireg">The registration to remove from the container.</param>
        void Unregister(IRegistration ireg);

        /// <summary>
        /// Returns an instance of a registered type.
        /// </summary>
        /// <param name="type">The type to resolve.</param>
        /// <param name="name">The name to resolve.</param>
        /// <returns>An instance of the type.  Throws a KeyNoFoundException if not registered.</returns>
        object Resolve(string name, Type type);

        /// <summary>
        /// Returns an instance of a registered type.
        /// </summary>
        /// <param name="type">The type to resolve.</param>
        /// <returns>An instance of the type.  Throws a KeyNoFoundException if not registered.</returns>
        object Resolve(Type type);

        /// <summary>
        /// Returns an instance of a registered type.
        /// </summary>
        /// <typeparam name="TType">The type to resolve</typeparam>
        /// <returns>An instance of the type.  Throws a KeyNoFoundException if not registered.</returns>
        TType Resolve<TType>() where TType : class;

        /// <summary>
        /// Returns an instance of a registered type.
        /// </summary>
        /// <typeparam name="TType">The type to resolve.</typeparam>
        /// <param name="name">The name to resolve.</param>
        /// <returns>An instance of the type.  Throws a KeyNoFoundException if not registered.</returns>
        TType Resolve<TType>(string name) where TType : class;

        /// <summary>
        /// Returns an Registration of a registered type.
        /// </summary>
        /// <param name="type">The type to get the Registration for.</param>
        /// <param name="name">The name associated with the named registration to get.</param>
        /// <returns>An Registration for the type. Throws a KeyNoFoundException if not registered.</returns>
        IRegistration GetRegistration(string name, Type type);

        /// <summary>
        /// Returns an Registration of a registered type.
        /// </summary>
        /// <param name="type">The type to get the Registration for.</param>
        /// <returns>An Registration for the type. Throws a KeyNoFoundException if not registered.</returns>
        IRegistration GetRegistration(Type type);

        /// <summary>
        /// Returns an Registration of a registered type.
        /// </summary>
        /// <typeparam name="TType">The type to get the Registration for.</typeparam>
        /// <returns>An Registration for the type. Throws a KeyNoFoundException if not registered.</returns>
        IRegistration GetRegistration<TType>() where TType : class;

        /// <summary>
        /// Returns an Registration of a registered type.
        /// </summary>
        /// <typeparam name="TType">The type to get the Registration for.</typeparam>
        /// <param name="name">The name associated with the named registration to get.</param>
        /// <returns>An Registration for the type. Throws a KeyNoFoundException if not registered.</returns>
        IRegistration GetRegistration<TType>(string name) where TType : class;

        /// <summary>
        /// Gets all Registrations for the specified type.
        /// </summary>
        /// <param name="type">The type for which the Registrations are required.</param>
        /// <returns>A list of the registration of the specified type.</returns>
        List<IRegistration> GetRegistrations(Type type);

        /// <summary>
        /// Gets all Registrations for the specified type.
        /// </summary>
        /// <typeparam name="TType">The type for which the Registrations are required.</typeparam>
        /// <returns>A list of the registration of the specified type.</returns>
        List<IRegistration> GetRegistrations<TType>() where TType : class;

        /// <summary>
        /// Gets the <see cref="ILifetimeManager"/> associated with this container.
        /// </summary>
        ILifetimeManager LifeTimeManager { get; }

        /// <summary>
        /// Specifies a default lifetime manager to use with this container.
        /// </summary>
        /// <param name="lifetimeManager">The lifetime manager to use by default.</param>
        /// <returns>Returns a reference to this container.</returns>
        IIocContainer UsesDefaultLifetimeManagerOf(ILifetimeManager lifetimeManager);
    }

    /// <summary>
    /// Defines the functionality for Lifetime Managers.  Implementation should instantiate an
    /// instance store and use the Registration's Key property to index the data in the store.
    /// This allows one lifetime manager to service multiple Registrations.
    /// </summary>
    public interface ILifetimeManager
    {
        /// <summary>
        /// Get an instance for the registration, using the lifetime manager to cache instance
        /// as required by the scope of the lifetime manager.
        /// </summary>
        /// <param name="creator">
        /// The instance creator which is used to supply the storage key and create a new instance if
        /// required.
        /// </param>
        /// <returns>The cached or new instance.</returns>
        object GetInstance(IInstanceCreator creator);

        /// <summary>
        /// Invalidate the instance in whatever storage is used by the lifetime manager.
        /// </summary>
        /// <param name="registration">
        /// The registration which is used to supply the storage key and create a new instance if
        /// required.
        /// </param>
        void InvalidateInstanceCache(IRegistration registration);
    }

    /// <summary>
    /// This is the result of registering a type in the container.
    /// </summary>
    public interface IRegistration
    {
        /// <summary>
        /// Gets the name of the registration.
        /// </summary>
        string Name { get; }

        /// <summary>
        /// Gets the key that is used to identify cached values.
        /// </summary>
        string Key { get; }

        /// <summary>
        /// Gets the type the contain will Resolve to when this Registration is used.
        /// </summary>
        Type ResolvesTo { get; }

        /// <summary>
        /// Sets the lifetime manager to be used by this Registration.
        /// </summary>
        /// <param name="manager">The ILifetimeManager to use.</param>
        /// <returns>'this', or the Registration.</returns>
        IRegistration WithLifetimeManager(ILifetimeManager manager);

        /// <summary>
        /// Invalidates any cached value so that a new instance will be created on
        /// the next Resolve call.
        /// </summary>
        void InvalidateInstanceCache();
    }

    /// <summary>
    /// This interface is used internally to identify registrations in the type registry.
    /// </summary>
    internal interface IRegistrationKey
    {
        /// <summary>
        /// Gets the type that this key identifies.
        /// </summary>
        /// <returns>Returns the type of the registration.</returns>
        Type GetInstanceType();

        bool Equals(object obj);

        int GetHashCode();
    }
}