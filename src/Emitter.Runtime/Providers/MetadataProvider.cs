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
using Emitter.Collections;
using Emitter.Text.Json;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a MetaData provider which provides Assembies and Types metadata
    /// on which the server operates
    /// </summary>
    public abstract class MetadataProvider : Provider
    {
        /// <summary>
        /// Registers a module assembly to the provider.
        /// </summary>
        /// <param name="assembly">The assembly to register.</param>
        public abstract void RegisterAssembly(Assembly assembly);

        /// <summary>
        /// Removes the module assembly from the provider.
        /// </summary>
        /// <param name="assembly">The assembly to unregister.</param>
        public abstract void UnregisterAssembly(Assembly assembly);

        /// <summary>
        /// Gets the list of all registered assemblies in the provider.
        /// </summary>
        public abstract IViewCollection<Assembly> Assemblies { get; }

        /// <summary>
        /// Gets the list of all types in the provider.
        /// </summary>
        public abstract IViewCollection<Type> Types { get; }

        #region Built-in: Query Methods

        /// <summary>
        /// Gets all types containing a given attribute
        /// </summary>
        /// <param name="name">The name to lookup for</param>
        /// <returns>Returns a read-only collection of found types</returns>
        public virtual IEnumerable<Type> GetTypesByName(string name)
        {
            if (name == null)
                throw new ArgumentNullException("attribute");

            var result = new List<Type>();
            Types.ForEach(type =>
            {
                if (type.Name == name)
                    result.Add(type);
            });
            return result;
        }

        /// <summary>
        /// Gets all types implementing a given interface
        /// </summary>
        /// <param name="interfaceType">The interface to lookup for</param>
        /// <returns>Returns a read-only collection of found types</returns>
        public virtual IEnumerable<Type> GetTypesByInterface(Type interfaceType)
        {
            if (interfaceType == null)
                throw new ArgumentNullException("interface");

            var result = new List<Type>();
            Types.ForEach(type =>
            {
                var interfaces = type.GetInterfaces();
                for (int i = 0; i < interfaces.Length; ++i)
                {
                    if (interfaceType.Equals(interfaces[i]))
                    {
                        result.Add(type);
                        break;
                    }
                }
            });
            return result;
        }

        /// <summary>
        /// Gets all types implementing a given interface
        /// </summary>
        /// <param name="interfaceName">The interface name to lookup for</param>
        /// <returns>Returns a read-only collection of found types</returns>
        public virtual IEnumerable<Type> GetTypesByInterface(string interfaceName)
        {
            if (interfaceName == null)
                throw new ArgumentNullException("interfaceName");

            var result = new List<Type>();
            Types.ForEach(type =>
            {
                if (type.GetInterfaces().Where(i => i.Name == interfaceName).Any())
                    result.Add(type);
            });
            return result;
        }

        #endregion Built-in: Query Methods

        #region Built-in: Populate Members

        /// <summary>
        /// Attempts to populate the instance through the security provider configured.
        /// </summary>
        /// <param name="prefix">The prefix for the path.</param>
        /// <param name="instance">The instance to populate.</param>
        public virtual void PopulateFromSecurityProvider(string prefix, object instance)
        {
            // Check the environment variables for some things such as seed and keys
            var properties = instance.GetType()
                .GetFields(BindingFlags.Public | BindingFlags.Instance);

            // Load the provider
            var provider = Service.Providers.Resolve<SecurityProvider>();

            // Reflect each property
            foreach (var prop in properties)
            {
                var attrib = prop
                    .GetCustomAttributes(typeof(JsonPropertyAttribute), true)
                    .FirstOrDefault();
                if (attrib == null)
                    continue;

                var json = attrib as JsonPropertyAttribute;
                var path = prefix.ToLower() + "/" + json.PropertyName.ToLower();
                if (prop.FieldType.GetTypeInfo().IsPrimitive || prop.FieldType == typeof(string))
                {
                    var value = provider.GetSecret(path);
                    if (value != null)
                    {
                        try
                        {
                            prop.SetValue(instance, Convert.ChangeType(value, prop.FieldType));
                            Service.Logger.Log(LogLevel.Info, string.Format("{0}: Using '{1}'", provider.Name, path));
                        }
                        catch (Exception ex)
                        {
                            Service.Logger.Log(LogLevel.Info, string.Format("{0}: Unable to use {1}. {2}", provider.Name, path, ex.Message));
                        }
                    }
                }
                else if (prop.FieldType.GetTypeInfo().IsClass)
                {
                    PopulateFromSecurityProvider(path, prop.GetValue(instance));
                }
            }
        }

        #endregion Built-in: Populate Members
    }

    /// <summary>
    /// Represents a default implementation of a <see cref="MetadataProvider"/>.
    /// </summary>
    public sealed class DefaultMetadataProvider : MetadataProvider
    {
        private readonly ArrayList<Type> TypeMap = new ArrayList<Type>(8192);
        private readonly List<Assembly> RegisteredAssemblies = new List<Assembly>(16);

        /// <summary>
        /// Registers an assembly to the provider
        /// </summary>
        /// <param name="assembly">The assembly to register</param>
        public override void RegisterAssembly(Assembly assembly)
        {
            Type[] types = null;
            lock (RegisteredAssemblies)
            {
                if (RegisteredAssemblies.Contains(assembly))
                    return;

                try
                {
                    types = assembly.GetTypes();
                    TypeMap.AddRange(types.ToArray());
                }
                catch (ReflectionTypeLoadException ex)
                {
                    try
                    {
                        types = ex.Types;
                    }
                    catch (Exception exp)
                    {
                        Service.Logger.Log(LogLevel.Error,
                            String.Format("Error: Loading of assembly {0} has thrown an exception: {1}.", assembly.FullName, exp.Message));
                    }
                }
                finally
                {
                    RegisteredAssemblies.Add(assembly);
                }
            }
        }

        /// <summary>
        /// Removes an assembly from the provider
        /// </summary>
        /// <param name="assembly">The assembly to unregister</param>
        public override void UnregisterAssembly(Assembly assembly)
        {
            lock (RegisteredAssemblies)
            {
                if (RegisteredAssemblies.Contains(assembly))
                {
                    RegisteredAssemblies.Remove(assembly);
                    Type[] types = assembly.GetTypes();
                    TypeMap.ForEachWithHandle((type, handle) =>
                    {
                        if (type.AssemblyQualifiedName.EndsWith(assembly.FullName))
                            TypeMap.Remove(handle);
                    });
                }
            }
        }

        /// <summary>
        /// Gets the list of all registered assemblies in the provider
        /// </summary>
        public override IViewCollection<Assembly> Assemblies
        {
            get { return new ReadOnlyList<Assembly>(RegisteredAssemblies); }
        }

        /// <summary>
        /// Gets the list of all types in the provider
        /// </summary>
        public override IViewCollection<Type> Types
        {
            get { return new ReadOnlyArrayList<Type>(TypeMap); }
        }

        #region AssemblyNameComparer

        private class AssemblyNameComparer : IEqualityComparer<AssemblyName>
        {
            public bool Equals(AssemblyName x, AssemblyName y)
            {
                if (Object.ReferenceEquals(x, y)) return true;
                if (Object.ReferenceEquals(x, null) || Object.ReferenceEquals(y, null))
                    return false;
                return x.FullName == y.FullName;
            }

            public int GetHashCode(AssemblyName name)
            {
                if (Object.ReferenceEquals(name, null)) return 0;
                return name.FullName == null ? 0 : name.FullName.GetHashCode();
            }
        }

        #endregion AssemblyNameComparer

        #region IDisposable Members

        /// <summary>
        /// Frees the metadata.
        /// </summary>
        protected override void Dispose(bool disposing)
        {
            base.Dispose(disposing);
            if (TypeMap != null)
                TypeMap.Dispose();
        }

        #endregion IDisposable Members
    }
}