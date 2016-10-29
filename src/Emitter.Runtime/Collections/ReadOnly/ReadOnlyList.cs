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
    public struct ReadOnlyList<T> : IViewCollection<T>
    {
        private IList<T> fTarget;

        /// <summary>
        /// Constructs a new read-only collection view.
        /// </summary>
        /// <param name="target">The target collection to construct the view for.</param>
        public ReadOnlyList(IList<T> target)
        {
            if (target == null)
                throw new ArgumentNullException("target");
            fTarget = target;
        }

        #region IView<T> Members

        /// <summary>
        /// Gets the element count in the collection.
        /// </summary>
        public int Count
        {
            get { return fTarget.Count; }
        }

        /// <summary>
        /// Executes an action for each element of the collection.
        /// </summary>
        /// <param name="action">The action to execute.</param>
        public void ForEach(Action<T> action)
        {
            if (action == null)
                throw new ArgumentNullException("action");
            for (int i = 0; i < fTarget.Count; ++i)
                action(fTarget[i]);
        }

        #endregion IView<T> Members

        #region IEnumerable<T> Members

        /// <summary>
        /// Gets the enumerator for this collection
        /// </summary>
        /// <returns>The enumerator.</returns>
        public IEnumerator<T> GetEnumerator()
        {
            return fTarget.GetEnumerator();
        }

        System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator()
        {
            return fTarget.GetEnumerator();
        }

        #endregion IEnumerable<T> Members
    }
}