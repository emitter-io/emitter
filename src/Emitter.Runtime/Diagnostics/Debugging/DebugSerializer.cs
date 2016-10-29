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

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Custom JSON serializer for a debug handle.
    /// </summary>
    public class DebugHandleSerializer : JsonConverter
    {
        /// <summary>
        /// Writes the JSON.
        /// </summary>
        public override void WriteJson(JsonWriter writer, object value, JsonSerializer serializer)
        {
            var handle = (DebugHandle)value;
            if (!handle.Reference.IsAlive)
            {
                writer.WriteValue("(dead)");
            }
            else
            {
                var target = handle.Reference.Target;
                writer.WriteValue(String.Format("ref:0x{0:X} {1}",
                    target.GetHashCode(),
                    target.GetType().Name
                    ));
            }
        }

        /// <summary>
        /// Reads the JSON and returns the object.
        /// </summary>
        public override object ReadJson(JsonReader reader, Type objectType, object existingValue, JsonSerializer serializer)
        {
            throw new NotImplementedException();
        }

        /// <summary>
        /// Checks what types it can convert.
        /// </summary>
        public override bool CanConvert(Type objectType)
        {
            return typeof(DebugHandle) == objectType;
        }
    }
}