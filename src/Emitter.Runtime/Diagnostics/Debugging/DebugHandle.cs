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
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Reflection;
using Emitter.Replication;
using Emitter.Text.Json;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a handle which can be used for inspecting.
    /// </summary>
    [JsonConverter(typeof(DebugHandleSerializer))]
    public sealed class DebugHandle
    {
        #region Static Members

        // Currently tracked handles (for querying)
        private readonly static ConcurrentDictionary<string, DebugHandle> Tracked
            = new ConcurrentDictionary<string, DebugHandle>();

        /// <summary>
        /// Inspects a target and returns a handle.
        /// </summary>
        /// <param name="target">The target to wrap.</param>
        /// <returns></returns>
        public static DebugHandle Inspect(object target)
        {
            // Fetch from the cache first
            var key = String.Format("0x{0:X}", target.GetHashCode());
            return Tracked.GetOrAdd(key, (k) =>
            {
                var handle = new DebugHandle();
                handle.Reference = new WeakReference(target);
                return handle;
            });
        }

        /// <summary>
        /// Retrieves a handle by the key
        /// </summary>
        public static DebugHandle Get(string key)
        {
            DebugHandle handle;
            if (Tracked.TryGetValue(key, out handle))
                return handle;
            return null;
        }

        /// <summary>
        /// Removes dead handles from the cache.
        /// </summary>
        public static void Cleanup()
        {
            // First, find all dead handles
            var dead = new Queue<string>();
            foreach (var kvp in Tracked)
            {
                var handle = kvp.Value;
                if (!handle.Reference.IsAlive)
                    dead.Enqueue(kvp.Key);
            }

            // Clean up the dead handles now
            DebugHandle dummy;
            foreach (var d in dead)
                Tracked.TryRemove(d, out dummy);
        }

        #endregion Static Members

        #region Constructors

        /// <summary>
        /// Gets or sets the reference of the entry.
        /// </summary>
        public WeakReference Reference;

        /// <summary>
        /// Gets the list of properties, fields & (maybe) methods.
        /// </summary>
        private readonly Dictionary<string, object> Members =
            new Dictionary<string, object>();

        /// <summary>
        /// Private constructor.
        /// </summary>
        private DebugHandle() { }

        #endregion Constructors

        #region Populate Members

        /// <summary>
        /// Inspects itself and fills the 'Members' list.
        /// </summary>
        private void Populate()
        {
            try
            {
                // Reset the members
                this.Members.Clear();

                // Check if it's alive
                var target = this.Reference.Target;
                if (!this.Reference.IsAlive || target == null)
                    return;

                // Get the type
                var type = target.GetType();
                var info = type.GetTypeInfo();

                // Is it a replicated dictionary?
                var replicated = target as IReplicatedCollection;
                if (replicated != null)
                {
                    foreach (var entry in replicated.GetEntries())
                    {
                        var entryType = entry.GetType();
                        var entryKey = entryType.GetField("Key").GetValue(entry);
                        var entryValue = entryType.GetField("Value").GetValue(entry);
                        var entryVersion = entryType.GetField("Version").GetValue(entry);
                        var entryDeleted = (bool)entryType.GetField("Deleted").GetValue(entry);

                        // Format the key
                        var key = string.Format("{0}: v{1}{2}", entryKey.ToString(), entryVersion.ToString(), entryDeleted ? "(-)" : "");

                        // Inspect the value
                        this.Members[key] = PopulateValue(
                            entryValue
                            );
                    }
                    return;
                }

                // Get the properties
                var properties = info
                    .GetProperties(BindingFlags.Instance | BindingFlags.Public | BindingFlags.NonPublic)
                    .Where((property) => property.GetIndexParameters().Length == 0 && property.CanRead)
                    .ToArray();

                // Get the fields
                var fields = info
                    .GetFields(BindingFlags.Instance | BindingFlags.Public | BindingFlags.NonPublic)
                    .ToArray();

                // Populate the id
                this.Members["$id"] = String.Format("0x{0:X}", target.GetHashCode());
                this.Members["$type"] = type.FullName;

                // Populate the properties
                if (properties.Length > 0)
                {
                    foreach (var member in properties)
                    {
                        try
                        {
                            // Inspect the value
                            this.Members[member.Name] = PopulateValue(
                                member.GetValue(target, null)
                                );
                        }
                        catch (Exception ex)
                        {
                            this.Members[member.Name] = "(" + ex.Message + ")";
                        }
                    }
                }

                // Populate the fields
                if (fields.Length > 0)
                {
                    foreach (var member in fields)
                    {
                        try
                        {
                            // Inspect the value
                            this.Members[member.Name] = PopulateValue(
                                member.GetValue(target)
                                );
                        }
                        catch (Exception ex)
                        {
                            this.Members[member.Name] = "(" + ex.Message + ")";
                        }
                    }
                }
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
        }

        /// <summary>
        /// Prepares a value to inspect.
        /// </summary>
        private object PopulateValue(object value)
        {
            try
            {
                if (value == null)
                    return "(null)";

                // Get the type
                var type = value.GetType();
                var info = type.GetTypeInfo();

                // If it's a Boolean, Byte, SByte, Int16, UInt16, Int32, UInt32, Int64, UInt64, IntPtr, UIntPtr, Char, Double, or Single.
                if (info.IsPrimitive)
                    return value;

                // If it's a string, return a max of 100 characters
                if (value is string)
                {
                    var strValue = value as string;
                    if (strValue.Length < 100)
                        return strValue;
                    return strValue.Substring(0, 100);
                }

                // If it's some 'almost' primitive, ToString() it
                if (value is IPEndPoint)
                    return value.ToString();
                if (value is TimeSpan)
                    return value.ToString();
                if (value is DateTime)
                    return value.ToString();

                // If it's an enum, return the name
                if (info.IsEnum)
                    return Enum.GetName(type, value);

                // If it's a value type, it's safe to copy
                if (info.IsValueType)
                    return value;

                // Anything else, return the handle wrapped
                if (info.IsClass)
                    return Inspect(value);
                return "(unknown)";
            }
            catch (Exception ex)
            {
                return "(debugger error: " + ex.Message + ")";
            }
        }

        #endregion Populate Members

        #region Serialization Members

        /// <summary>
        /// Converts the debug handle to a string representation.
        /// </summary>
        /// <returns></returns>
        public override string ToString()
        {
            // Check if it's alive
            var target = this.Reference.Target;
            if (!this.Reference.IsAlive || target == null)
                return "{dead}";

            // Populate now
            this.Populate();

            // Do we have data?
            if (this.Members.Count == 0)
                return "{empty}";

            // Serialize the members
            return JsonConvert.SerializeObject(this.Members, Text.Json.Formatting.Indented);
        }

        #endregion Serialization Members
    }
}