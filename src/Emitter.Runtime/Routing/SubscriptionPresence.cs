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
using System.Linq;
using Emitter.Replication;

namespace Emitter
{
    /// <summary>
    /// Represents a subscription presence list that is syncronized across the cluster.
    /// </summary>
    public sealed class SubscriptionPresence : ReplicatedDictionary<SubscriptionPresenceInfo>, IReplicated
    {
        /// <summary>
        /// Constructs a subscription presence list instance.
        /// </summary>
        public SubscriptionPresence() : base()
        {
        }
    }
}