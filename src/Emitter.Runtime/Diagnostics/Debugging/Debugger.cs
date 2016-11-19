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
using System.Text;
using Emitter.Security;
using Emitter.Text.Json;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Represents a root context used for inspecting.
    /// </summary>
    public sealed class Debugger : Instrument
    {
        #region Static Members

        /// <summary>
        /// Gets the default profiler.
        /// </summary>
        public readonly static Debugger Default = new Debugger();

        /// <summary>
        /// Starts the default profiler.
        /// </summary>
        [InvokeAt(InvokeAtType.Initialize)]
        public static void Initialize()
        {
            Default.Start();
        }

        #endregion Static Members

        #region Constructor

        /// <summary>
        /// The map of tracked objects.
        /// </summary>
        public readonly ConcurrentDictionary<string, DebugHandle> TrackedObjects
            = new ConcurrentDictionary<string, DebugHandle>();

        /// <summary>
        /// Creates a new debugger.
        /// </summary>
        private Debugger() : base(
            "debug/" + EmitterStatus.Address + "/",
            TimeSpan.FromSeconds(1))
        { }

        #endregion Constructor

        #region Instrument Members

        /// <summary>
        /// Occurs when the instrument is starting.
        /// </summary>
        protected override void OnStart()
        {
        }

        /// <summary>
        /// Gets the message to publish to emitter channel.
        /// </summary>
        /// <returns></returns>
        protected override void OnExecute()
        {
            // First, find all dead handles
            DebugHandle.Cleanup();

            try
            {
                // Send some objects
                var handles = new List<DebugHandle>();
                //foreach (var mesh in Service.Mesh.Members)
                //    handles.Add(DebugHandle.Inspect(mesh));
                handles.Add(
                    DebugHandle.Inspect(Service.Registry.Collection)
                    );

                /*foreach (var client in Service.Clients)
                {
                    if (client.Channel?.Binding is MeshBinding)
                        continue;
                    handles.Add(DebugHandle.Inspect(client));
                }*/

                if (handles.Count == 0)
                    return;

                // Return the serialized object
                var message = Encoding.UTF8
                    .GetBytes(JsonConvert.SerializeObject(handles))
                    .AsSegment();

                // Publish the message in the cluster, don't need to store it as it might slow everything down
                Dispatcher.Publish(SecurityLicense.Current.Contract, Info.Target, this.Channel + "root/", message, 60);
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
        }

        #endregion Instrument Members

        public void Inspect(string id)
        {
            try
            {
                // Get the handle
                var handle = DebugHandle.Get(id);
                if (handle == null)
                    return;

                // Get the measurements
                var message = Encoding.UTF8
                    .GetBytes(handle.ToString())
                    .AsSegment();

                // Publish the message in the cluster, don't need to store it as it might slow everything down
                Dispatcher.Publish(SecurityLicense.Current.Contract, Info.Target, this.Channel + "result/", message, 60);
            }
            catch (Exception ex)
            {
                Service.Logger.Log(ex);
            }
        }
    }
}