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

namespace Emitter.Providers
{
    internal class UnNamedRegistrationKey : IRegistrationKey
    {
        internal Type InstanceType;

        public UnNamedRegistrationKey(Type type)
        {
            InstanceType = type;
        }

        public Type GetInstanceType()
        {
            return InstanceType;
        }

        // comparison methods
        public override bool Equals(object obj)
        {
            var r = obj as UnNamedRegistrationKey;
            return (r != null) && (InstanceType == r.InstanceType);
        }

        public override int GetHashCode()
        {
            return InstanceType.GetHashCode();
        }
    }

    internal class NamedRegistrationKey : IRegistrationKey
    {
        internal Type InstanceType;
        internal string Name;

        public NamedRegistrationKey(string name, Type type)
        {
            Name = name ?? String.Empty;
            InstanceType = type;
        }

        public Type GetInstanceType()
        {
            return InstanceType;
        }

        // comparison methods
        public override bool Equals(object obj)
        {
            var r = obj as NamedRegistrationKey;
            return (r != null) &&
                (InstanceType == r.InstanceType) &&
                String.Compare(Name, r.Name, true) == 0; // ignore case
        }

        public override int GetHashCode()
        {
            return InstanceType.GetHashCode();
        }
    }

    internal class Registration : IRegistration, IInstanceCreator
    {
        internal ILifetimeManager LifetimeManager;
        internal Func<IIocContainer, object> Factory;
        internal Func<object> LazyFactory;
        private string _key;
        private Type _type;

        public object Instance;
        private IIocContainer Container;

        public Registration(IIocContainer container, string name, Type type, Func<IIocContainer, object> factory)
        {
            LifetimeManager = null;
            Container = container;
            Factory = factory;
            Name = name;
            _type = type;
            _key = "[" + (name ?? "null") + "]:" + type.Name;

            if (name == null)
                LazyFactory = () => container.Resolve(_type);
            else
                LazyFactory = () => container.Resolve(Name, _type);
        }

        public string Key
        {
            get
            {
                return _key;
            }
        }

        public Type ResolvesTo
        {
            get
            {
                return _type;
            }
        }

        public string Name { get; private set; }

        public IRegistration WithLifetimeManager(ILifetimeManager manager)
        {
            LifetimeManager = manager;
            return this;
        }

        public object CreateInstance(ContainerCaching containerCache)
        {
            if (containerCache == ContainerCaching.InstanceCachedInContainer)
            {
                if (Instance == null)
                    Instance = Factory(Container);
                return Instance;
            }
            else
                return Factory(Container);
        }

        public object GetInstance()
        {
            return (LifetimeManager != null)
                ? LifetimeManager.GetInstance(this)
                : Factory(Container);
        }

        public void InvalidateInstanceCache()
        {
            if (LifetimeManager != null)
                LifetimeManager.InvalidateInstanceCache(this);
        }
    }
}