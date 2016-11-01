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

namespace Emitter
{
    /// <summary>
    /// Represents a helper function for option retrieval.
    /// </summary>
    internal static class EmitterOption
    {
        /// <summary>
        /// The time-to-live option
        /// </summary>
        public const string TimeToLive = "ttl";

        /// <summary>
        /// The last history option
        /// </summary>
        public const string LastHistory = "last";

        /// <summary>
        /// Attempts to get an option.
        /// </summary>
        public static bool TryGet(EmitterChannel channel, string option, int defaultValue, out int value)
        {
            // Do we have the option?
            value = defaultValue;
            if (channel.Options == null || !channel.Options.ContainsKey(option))
                return false;

            // Do we have a TTL with the message?
            return int.TryParse(channel.Options[option], out value);
        }
    }
}