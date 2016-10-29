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
using System.Linq.Expressions;
using System.Reflection;

namespace Emitter.Replication
{
    /// <summary>
    /// Represents an object activator that can be used to create instances.
    /// </summary>
    /// <param name="args">The arguments</param>
    /// <returns></returns>
    internal delegate IReplicated ReplicatedObjectActivator(params object[] args);

    /// <summary>
    /// Represents a metadata service that can be used for activating replicated objects.
    /// </summary>
    public class ReplicatedActivator
    {
        private static readonly ConcurrentDictionary<int, ReplicatedObjectActivator> Activators
            = new ConcurrentDictionary<int, ReplicatedObjectActivator>();

        /// <summary>
        /// Configures the activator.
        /// </summary>
        [InvokeAt(InvokeAtType.Configure)]
        public static void Configure()
        {
            // Register all replicated types
            var replicated = Service.MetadataProvider.GetTypesByInterface(typeof(IReplicated));
            foreach (var type in replicated)
                Register(type);
        }

        /// <summary>
        /// Register a type to this metadata provider.
        /// </summary>
        /// <param name="type">The CLR type to register.</param>
        private static void Register(Type type)
        {
            // Ignore abstract types
            if (type.GetTypeInfo().IsAbstract || type.GetTypeInfo().IsGenericTypeDefinition)
                return;

            // Check if we already have this type
            var identity = type.ToIdentifier();
            if (Activators.ContainsKey(identity))
                return;

            // We must have a constructor
            var constructor = type.GetConstructor(Type.EmptyTypes);
            if (constructor == null)
                throw new TypeLoadException("Unable to find a default parameterless constructor for type " + type.FullName);

            // Compile and add an activator
            var activator = GetActivator(type, constructor);
            if (activator != default(ReplicatedObjectActivator))
                Activators.TryAdd(type.ToIdentifier(), activator);

            //Service.Logger.Log(LogLevel.Info, string.Format("Activator: {0} mapped to {1}", identity, type.FullName));
        }

        /// <summary>
        /// Creates an object instance by its identifier.
        /// </summary>
        /// <param name="identifier">The identifier of the object to create.</param>
        /// <returns>The created object instance.</returns>
        public static IReplicated CreateInstance(int identifier)
        {
            ReplicatedObjectActivator activator;
            if (Activators.TryGetValue(identifier, out activator))
                return activator();

            Service.Logger.Log(LogLevel.Error, "Unable to find an activator for type " + identifier);
            return null;
        }

        /// <summary>
        /// Creates an instance of an object.
        /// </summary>
        /// <typeparam name="T">The type to create.</typeparam>
        /// <returns>The created object instance.</returns>
        public static T CreateInstance<T>()
            where T : IReplicated
        {
            return (T)CreateInstance(typeof(T).ToIdentifier());
        }

        /// <summary>
        /// Creates an activator
        /// </summary>
        /// <param name="type">The type to compile an activator for.</param>
        /// <param name="ctor">The constructor to use.</param>
        /// <returns>The compiled lambda.</returns>
        private static ReplicatedObjectActivator GetActivator(Type type, ConstructorInfo ctor)
        {
            var paramsInfo = ctor.GetParameters();

            //create a single param of type object[]
            var param = Expression.Parameter(typeof(object[]), "args");
            var argsExp = new Expression[paramsInfo.Length];

            //pick each arg from the params array
            //and create a typed expression of them
            for (int i = 0; i < paramsInfo.Length; i++)
            {
                var index = Expression.Constant(i);
                var paramType = paramsInfo[i].ParameterType;
                var paramAccessorExp = Expression.ArrayIndex(param, index);
                var paramCastExp = Expression.Convert(paramAccessorExp, paramType);

                argsExp[i] = paramCastExp;
            }

            //make a NewExpression that calls the
            //ctor with the args we just created
            var newExp = Expression.New(ctor, argsExp);

            //create a lambda with the New
            //Expression as body and our param object[] as arg
            var lambda = Expression.Lambda(typeof(ReplicatedObjectActivator), newExp, param);

            //compile it
            return (ReplicatedObjectActivator)lambda.Compile();
        }
    }
}