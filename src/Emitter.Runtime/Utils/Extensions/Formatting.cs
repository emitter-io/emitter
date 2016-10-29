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
    public static class FormattingEx
    {
        /// <summary>
        /// Unix epoch offset.
        /// </summary>
        private static readonly DateTime UnixOffset
            = new DateTime(1970, 1, 1, 0, 0, 0, 0, DateTimeKind.Utc);

        public static string ToFormatedNumberString(this double source)
        {
            return String.Format("{0:N2}", source);
        }

        public static string ToFormatedNumberString(this float source)
        {
            return String.Format("{0:N2}", source);
        }

        public static string ToFormatedPercentageString(this double source)
        {
            return String.Format("{0:N2}%", source);
        }

        public static string ToFormatedSizeString(this double source)
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

        public static string ToFormatedSizeString(this long source)
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

        public static string ToFormatedSpeedString(this double source)
        {
            if (source < 1024)
                return String.Format("{0:N} bytes/sec", source);
            double KoSize = (double)source / 1024;
            if (KoSize < 1024)
                return String.Format("{0:N2} KB/sec", KoSize);
            double MoSize = (double)KoSize / 1024;
            return String.Format("{0:N2} MB/sec", MoSize);
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
            /*string formatted = string.Format("{0}{1}{2}{3}",
                span.Days > 0 ? string.Format("{0:0} days, ", span.Days) : string.Empty,
                span.Hours > 0 ? string.Format("{0:0} h., ", span.Hours) : string.Empty,
                span.Minutes > 0 ? string.Format("{0:0} m., ", span.Minutes) : string.Empty,
                span.Seconds > 0 ? string.Format("{0:0} s.", span.Seconds) : string.Empty);

            if (formatted.EndsWith(", ")) formatted = formatted.Substring(0, formatted.Length - 2);
            return formatted;
             */

            return span.Days > 0 ? span.ToString(@"d\.hh\:mm\:ss") : span.ToString(@"hh\:mm\:ss");
        }

        /// <summary>
        /// Converts a date time to a unix timestamp.
        /// </summary>
        /// <param name="dateTime">Date time (should be UTC).</param>
        /// <returns>The UNIX timestamp in seconds</returns>
        public static long ToUnixTimestamp(this DateTime dateTime)
        {
            return (long)(dateTime - UnixOffset).TotalSeconds;
        }

        public static byte[] ToBigEndianBytes<T>(this int source)
        {
            byte[] bytes;

            var type = typeof(T);
            if (type == typeof(ushort))
                bytes = BitConverter.GetBytes((ushort)source);
            else if (type == typeof(ulong))
                bytes = BitConverter.GetBytes((ulong)source);
            else if (type == typeof(int))
                bytes = BitConverter.GetBytes(source);
            else
                throw new InvalidCastException("Cannot be cast to T");

            if (BitConverter.IsLittleEndian)
                Array.Reverse(bytes);
            return bytes;
        }

        public static int ToLittleEndianInt(this byte[] source)
        {
            if (BitConverter.IsLittleEndian)
                Array.Reverse(source);

            if (source.Length == 2)
                return BitConverter.ToUInt16(source, 0);

            if (source.Length == 8)
                return (int)BitConverter.ToUInt64(source, 0);

            throw new ArgumentException("Unsupported Size");
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
    }
}