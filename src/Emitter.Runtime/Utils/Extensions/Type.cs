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

namespace Emitter
{
    internal static class TypeExtensions
    {
        /// <summary>
        /// Gets invokable methods by type
        /// </summary>
        public static MethodInfo[] GetInvokeAt(this Type source, InvokeAtType type)
        {
            var result = new List<MethodInfo>();
            var methods = source.GetMethods(BindingFlags.Static | BindingFlags.Public);

            for (int i = 0; i < methods.Length; ++i)
            {
                var info = methods[i];
                var attr = info.GetCustomAttributes(typeof(InvokeAtAttribute), false)
                    .FirstOrDefault();
                if (attr == null)
                    continue;

                if ((attr as InvokeAtAttribute).Type != type)
                    continue;

                if (info.ReturnType != typeof(void))
                    throw new TargetInvocationException("A function is marked with InvokeAt attribute, but the return type is not System.Void. Source: " + source.ToString(), null);

                if (info.GetParameters().Length != 0)
                    throw new TargetInvocationException("A function is marked with InvokeAt attribute, but contains parameters. Source: " + source.ToString(), null);

                result.Add(info);
            }

            return result.ToArray();
        }

        /// <summary>
        /// Checks whether a type is a subclass of another type.
        /// </summary>
        /// <param name="source"></param>
        /// <param name="other"></param>
        /// <returns></returns>
        public static bool IsSubclassOf(this Type source, Type other)
        {
            return source.GetTypeInfo().IsSubclassOf(other);
        }

        /// <summary>
        /// Gets the murmur hash code for the type.
        /// </summary>
        /// <param name="source"></param>
        /// <returns></returns>
        public static int ToIdentifier(this Type source)
        {
            return Murmur32.GetHash(source.FullName);
        }
    }
}