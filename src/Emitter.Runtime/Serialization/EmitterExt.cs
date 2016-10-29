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
using System.Text;
using Emitter.Security;
using Emitter.Text.Json;

namespace Emitter
{
    /// <summary>
    /// Set of extension methods for various packets.
    /// </summary>
    internal static class EmitterExt
    {
        #region Convert

        /// <summary>
        /// Takes a segment over the whole bufer.
        /// </summary>
        /// <param name="buffer"></param>
        /// <returns></returns>
        public static ArraySegment<byte> AsSegment(this byte[] buffer)
        {
            return new ArraySegment<byte>(buffer, 0, buffer.Length);
        }

        /// <summary>
        /// Takes a segment over the whole bufer.
        /// </summary>
        /// <param name="buffer"></param>
        /// <returns></returns>
        public static byte[] AsArray(this ArraySegment<byte> buffer)
        {
            var array = new byte[buffer.Count];
            Buffer.BlockCopy(buffer.Array, buffer.Offset, array, 0, buffer.Count);
            return array;
        }

        /// <summary>
        /// Deserializes the segment as a specific JSON message.
        /// </summary>
        /// <typeparam name="T">The message type to deserialize as.</typeparam>
        /// <param name="buffer">The buffer to deserialize.</param>
        /// <returns></returns>
        public static T As<T>(this ArraySegment<byte> buffer)
        {
            return JsonConvert.DeserializeObject<T>(
                buffer.AsString()
                );
        }

        /// <summary>
        /// Takes a segment over the whole bufer.
        /// </summary>
        /// <param name="source"></param>
        /// <returns></returns>
        public static byte[] AsUTF8(this string source)
        {
            return Encoding.UTF8.GetBytes(source);
        }

        /// <summary>
        /// Takes a segment over the whole bufer.
        /// </summary>
        /// <param name="source"></param>
        /// <returns></returns>
        public static byte[] AsASCII(this string source)
        {
            return Encoding.ASCII.GetBytes(source);
        }

        /// <summary>
        /// Converts the segment to a UTF-8 string.
        /// </summary>
        /// <param name="buffer"></param>
        /// <returns></returns>
        public static string AsString(this ArraySegment<byte> buffer)
        {
            return Encoding.UTF8.GetString(buffer.Array, buffer.Offset, buffer.Count);
        }

        /// <summary>
        /// Checks whether the key contains a specific flag.
        /// </summary>
        /// <param name="key">The key to check.</param>
        /// <param name="flag">The flag we need to check.</param>
        /// <returns>Whether it contains a flag or not.</returns>
        public static bool HasPermission(this SecurityKey key, SecurityAccess flag)
        {
            var access = (SecurityAccess)key.Permissions;
            return (access & flag) == flag;
        }

        #endregion Convert

        #region Formatting

        public static string ToRatioString(this double source)
        {
            return String.Format("{0:N2}%", source);
        }

        public static string ToSizeString(this double source)
        {
            if (source < 1024)
                return String.Format("{0:N} bytes", source);

            double KoSize = source / 1024;
            if (KoSize < 1024)
                return String.Format("{0:N2} KB", KoSize);
            double MoSize = KoSize / 1024;
            if (MoSize < 1024)
                return String.Format("{0:N2} MB", MoSize);
            double GoSize = MoSize / 1024;
            return String.Format("{0:N2} GB", GoSize);
        }

        public static string ToSizeString(this long source)
        {
            if (source < 1024)
                return String.Format("{0:N} bytes", source);

            double KoSize = source / 1024;
            if (KoSize < 1024)
                return String.Format("{0:N2} KB", KoSize);
            double MoSize = KoSize / 1024;
            if (MoSize < 1024)
                return String.Format("{0:N2} MB", MoSize);
            double GoSize = MoSize / 1024;
            return String.Format("{0:N2} GB", GoSize);
        }

        public static string ToSpeedString(this double source)
        {
            if (source < 1024)
                return String.Format("{0:N} bytes/sec", source);
            double KoSize = (double)source / 1024;
            if (KoSize < 1024)
                return String.Format("{0:N2} KB/sec", KoSize);
            double MoSize = (double)KoSize / 1024;
            return String.Format("{0:N2} MB/sec", MoSize);
        }

        public static string ToFrequencyString(this double source)
        {
            if (source < 1024)
                return String.Format("{0:N} bit/sec", source * 8);
            double KoSize = (double)source / 1024;
            if (KoSize < 1024)
                return String.Format("{0:N2} KBit/sec", KoSize * 8);
            double MoSize = (double)KoSize / 1024;
            return String.Format("{0:N2} MBit/sec", MoSize * 8);
        }

        public static string ToFormatedTimeElapsedString(this double source)
        {
            UInt64 TotalNumSeconds = (UInt64)Math.Ceiling((double)source);
            if (TotalNumSeconds < 60)
                return String.Format("00:00:{0:00}", TotalNumSeconds);

            UInt16 NumMinutes = (UInt16)(TotalNumSeconds / 60);
            UInt16 NumSeconds = (UInt16)((TotalNumSeconds) - ((UInt64)NumMinutes * 60));

            if (NumMinutes < 60)
                return String.Format("00:{0:00}:{1:00}", NumMinutes, NumSeconds);

            UInt16 NumHours = (UInt16)(NumMinutes / 60);
            NumMinutes = (UInt16)(NumMinutes - (NumHours * 60));

            return String.Format("{0:00}:{1:00}:{2:00}", NumHours, NumMinutes, NumSeconds);
        }

        public static string ToReadableString(this TimeSpan span)
        {
            return span.Days > 0 ? span.ToString(@"d\.hh\:mm\:ss") : span.ToString(@"hh\:mm\:ss");
        }

        public static string ToTimeAgo(this DateTime source)
        {
            TimeSpan span = DateTime.Now - source;
            if (span.Days > 365)
            {
                int years = (span.Days / 365);
                if (span.Days % 365 != 0)
                    years += 1;
                return String.Format("about {0} {1} ago",
                years, years == 1 ? "year" : "years");
            }
            if (span.Days > 30)
            {
                int months = (span.Days / 30);
                if (span.Days % 31 != 0)
                    months += 1;
                return String.Format("about {0} {1} ago",
                months, months == 1 ? "month" : "months");
            }
            if (span.Days > 0)
                return String.Format("about {0} {1} ago",
                span.Days, span.Days == 1 ? "day" : "days");
            if (span.Hours > 0)
                return String.Format("about {0} {1} ago",
                span.Hours, span.Hours == 1 ? "hour" : "hours");
            if (span.Minutes > 0)
                return String.Format("about {0} {1} ago",
                span.Minutes, span.Minutes == 1 ? "minute" : "minutes");
            if (span.Seconds > 5)
                return String.Format("about {0} seconds ago", span.Seconds);
            if (span.Seconds <= 5)
                return "just now";
            return string.Empty;
        }

        #endregion Formatting
    }
}