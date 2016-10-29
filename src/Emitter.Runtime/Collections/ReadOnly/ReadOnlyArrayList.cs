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
    internal struct ReadOnlyArrayList<T> : IViewCollection<T>
        where T : class
    {
        private ArrayList<T> fTarget;

        public ReadOnlyArrayList(ArrayList<T> target)
        {
            if (target == null)
                throw new ArgumentNullException("target");
            fTarget = target;
        }

        #region IView<T> Members

        public int Count
        {
            get { return fTarget.Count; }
        }

        public void ForEach(Action<T> action)
        {
            if (action == null)
                throw new ArgumentNullException("action");
            fTarget.ForEach(action);
        }

        #endregion IView<T> Members

        #region IEnumerable<T> Members

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