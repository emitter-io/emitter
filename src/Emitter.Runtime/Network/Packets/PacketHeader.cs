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

namespace Emitter.Network
{
    internal static class PacketHeader
    {
        /// <summary>
        /// Total size of the header for each packet
        /// </summary>
        public const int TotalSize = PacketLengthSize + PacketKeySize;

        /// <summary>
        /// Size of the header bytes that determine Lenght of the packet
        /// </summary>
        public const int PacketLengthSize = 4;

        /// <summary>
        /// Size of the header data to add to the message length
        /// </summary>
        public const int PacketKeySize = 4;
    }
}