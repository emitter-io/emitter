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
using System.Threading.Tasks;

namespace Emitter
{
    internal static class Enumerations
    {
        internal static void ForEach<TKey, TValue>(this IDictionary<TKey, TValue> dictionary, Action<TKey, TValue> action)
        {
            if (dictionary == null)
                throw new ArgumentNullException("dictionary");
            if (action == null)
                throw new ArgumentNullException("action");

            foreach (TKey key in dictionary.Keys)
            {
                action(key, dictionary[key]);
            }
        }

        internal static void ForEach<T>(this IList<T> list, Action<T> action)
        {
            if (list == null)
                throw new ArgumentNullException("list");
            if (action == null)
                throw new ArgumentNullException("action");

            for (int i = 0; i < list.Count; ++i)
            {
                action(list[i]);
            }
        }

        /// <summary>
        /// This function specifies that the task should be forgotten as we do
        /// not care for its result.
        /// </summary>
        /// <param name="task">The task to forget.</param>
        public static void Forget(this Task task)
        {
        }
    }
}