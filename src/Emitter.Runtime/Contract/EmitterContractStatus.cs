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

namespace Emitter
{
    /// <summary>
    /// Represents a state that is communicated by emitter meta.
    /// </summary>
    public enum EmitterContractStatus
    {
        /// <summary>
        /// The client is allowed (paying customer or under the free tier limit).
        /// </summary>
        Allowed = 0,

        /// <summary>
        /// The client is refused by the meta.
        /// </summary>
        Refused = 1
    }

    /// <summary>
    /// Represents the type of the channel string.
    /// </summary>
    public enum ChannelType : byte
    {
        /// <summary>
        /// Unknown channel string.
        /// </summary>
        Unknown = 0,

        /// <summary>
        /// Invalid channel string.
        /// </summary>
        Invalid,

        /// <summary>
        /// Static channel string (eg: 'foo.bar.baz')
        /// </summary>
        Static,

        /// <summary>
        /// Wildcard channel string (eg: 'foo.*.baz');
        /// </summary>
        Wildcard
    }
}