using System;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;

namespace Emitter
{
    /// <summary>
    /// Performs priority comparison of the Invoke.
    /// </summary>
    internal class InvokePriorityComparer : IComparer<MethodInfo>
    {
        public int Compare(MethodInfo x, MethodInfo y)
        {
            if (x == null && y == null)
                return 0;

            if (x == null)
                return 1;

            if (y == null)
                return -1;

            return GetPriority(x) - GetPriority(y);
        }

        private int GetPriority(MethodInfo mi)
        {
            var attr = mi.GetCustomAttribute<InvokeAtAttribute>();
            if (attr == null)
                return 0;

            return attr.Priority;
        }
    }
}