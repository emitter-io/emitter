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

namespace Emitter
{
    /// <summary>
    /// Represents a subscription interest.
    /// </summary>
    [Flags]
    public enum SubscriptionInterest
    {
        /// <summary>
        /// Represents no interest.
        /// </summary>
        None = 0,

        /// <summary>
        /// Represents an interest in subscription messages.
        /// </summary>
        Messages = 1 << 0,

        /// <summary>
        /// Represents an interest in subscription presence
        /// </summary>
        Presence = 1 << 1,

        /// <summary>
        /// Represents an interest in both messages and presence.
        /// </summary>
        Everything = Messages | Presence,
    }
}