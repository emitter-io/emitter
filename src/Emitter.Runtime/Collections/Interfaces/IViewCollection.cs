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
    /// <summary>
    /// A view allows to view a particular collection in a read-only mode
    /// </summary>
    public interface IViewCollection<T> : IEnumerable<T>
    {
        /// <summary>
        /// Gets the count of the elements in the collection
        /// </summary>
        int Count { get; }

        /// <summary>
        /// Executes an action for each item in the collection
        /// </summary>
        /// <param name="action">Action to execute</param>
        void ForEach(Action<T> action);
    }
}