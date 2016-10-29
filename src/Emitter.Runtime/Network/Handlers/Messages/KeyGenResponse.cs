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
using Emitter.Text.Json;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a key generation response.
    /// </summary>
    internal class KeyGenResponse : EmitterResponse
    {
        /// <summary>
        /// The generated key.
        /// </summary>
        [JsonProperty("key")]
        public string Key;

        /// <summary>
        /// The target channel for the key.
        /// </summary>
        [JsonProperty("channel")]
        public string Channel;
    }
}