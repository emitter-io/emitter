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

namespace Emitter.Collections
{
    internal interface IArrayList<T> : IEnumerable<T>
    {
        /// <summary>
        /// Adds the item into a free space in the Items array and returns the handle to the item
        /// </summary>
        /// <param name="item">Item to add to Items array</param>
        /// <returns>Handle to the added item</returns>
        int Add(T item);

        /// <summary>
        /// Adds the item into a free space in the Items array and returns the handle to the item
        /// </summary>
        /// <param name="item">Item to add to Items array</param>
        /// <returns>Handle to the added item</returns>
        int Add(ref T item);

        /// <summary>
        ///  Adds a range of items into free spaces in the Items array
        /// </summary>
        /// <param name="items">Range of items to add</param>
        void AddRange(T[] items);

        /// <summary>
        /// Executes an action for each item in the collection
        /// Note: Internally uses a for loop.
        /// </summary>
        /// <param name="action">Action to execute</param>
        void ForEach(Action<T> action);

        /// <summary>
        /// Removes an item from a collection by its handle
        /// </summary>
        /// <param name="handle">Handle of the item to remove</param>
        void Remove(int handle);

        /// <summary>
        /// Checks whether there is an element in use at a given position
        /// </summary>
        /// <param name="handle">The handle to check</param>
        /// <returns>True if there is an element in this position, otherwise false</returns>
        bool HasElementAt(int handle);

        /// <summary>
        /// Clears the list
        /// </summary>
        void Clear();
    }
}