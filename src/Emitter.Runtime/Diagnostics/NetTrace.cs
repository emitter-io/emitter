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
using System.Diagnostics;
using System.Linq;

namespace Emitter.Diagnostics
{
    /// <summary>
    /// Provides a set of methods and properties that help you trace the execution
    /// of your networking code. This class cannot be inherited. This class is configured
    /// via <see cref="System.Diagnostics.Trace"/> class static properties.
    /// </summary>
    public sealed class NetTrace
    {
        #region Properties

        /// <summary>
        /// Gets the collection of listeners that is monitoring the trace output.
        /// This is a shortcut to <see cref="System.Diagnostics.Trace.Listeners"/> property.
        /// </summary>
        public static TraceListenerCollection Listeners
        {
            get { return Trace.Listeners; }
        }

        /// <summary>
        /// Gets or sets whether spike network tracing is enabled or not.
        /// </summary>
        public static bool Enabled
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether basic (tcp/udp) channel tracer should be enabled or not.
        /// </summary>
        public static bool TraceChannel
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether Emitter Tracer should be enabled or not.
        /// </summary>
        public static bool TraceEmitter
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether HTTP Tracer should be enabled or not.
        /// </summary>
        public static bool TraceHttp
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether web sockets tracer should be enabled or not.
        /// </summary>
        public static bool TraceWebSocket
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether SSL/TLS tracer should be enabled or not.
        /// </summary>
        public static bool TraceSsl
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether HTTP REST tracer should be enabled or not.
        /// </summary>
        public static bool TraceRest
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether socket.io tracer should be enabled or not.
        /// </summary>
        public static bool TraceFabric
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether cluster tracer should be enabled or not.
        /// </summary>
        public static bool TraceMesh
        {
            get;
            set;
        }

        /// <summary>
        /// Gets or sets whether MQTT tracer should be enabled or not.
        /// </summary>
        public static bool TraceMqtt
        {
            get;
            set;
        }

        #endregion Properties

        #region Constructor

        static NetTrace()
        {
        }

        #endregion Constructor

        #region Public Members

        /// <summary>
        /// Checks for a condition; if the condition is false, displays a message box
        /// that shows the call stack.
        /// </summary>
        /// <param name="condition">The conditional expression to evaluate. If the condition is true, a failure
        /// message is not sent and the message box is not displayed.</param>
        [Conditional("TRACE")]
        public static void Assert(bool condition)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Assert(condition);
        }

        /// <summary>
        /// Checks for a condition; if the condition is false, displays a message box
        /// that shows the call stack.
        /// </summary>
        /// <param name="condition">The conditional expression to evaluate. If the condition is true, a failure
        /// message is not sent and the message box is not displayed.</param>
        /// <param name="message">The message to send to the System.Diagnostics.Trace.Listeners collection.</param>
        [Conditional("TRACE")]
        public static void Assert(bool condition, string message)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Assert(condition, message);
        }

        /// <summary>
        /// Checks for a condition; if the condition is false, displays a message box
        /// that shows the call stack.
        /// </summary>
        /// <param name="condition">The conditional expression to evaluate. If the condition is true, a failure
        /// message is not sent and the message box is not displayed.</param>
        /// <param name="message">The message to send to the System.Diagnostics.Trace.Listeners collection.</param>
        /// <param name="detailMessage">The detailed message to send to the System.Diagnostics.Trace.Listeners collection.</param>
        [Conditional("TRACE")]
        public static void Assert(bool condition, string message, string detailMessage)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Assert(condition, message, detailMessage);
        }

        /// <summary>
        /// Flushes the output buffer, and then closes the System.Diagnostics.Trace.Listeners.
        /// </summary>
        [Conditional("TRACE")]
        public static void Close()
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Close();
        }

        /// <summary>
        /// Emits the specified error message.
        /// </summary>
        /// <param name="message">A message to emit.</param>
        [Conditional("TRACE")]
        public static void Fail(string message)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Fail(message);
        }

        /// <summary>
        /// Emits the specified error message.
        /// </summary>
        /// <param name="message">A message to emit.</param>
        /// <param name="detailMessage">A detailed message to emit.</param>
        [Conditional("TRACE")]
        public static void Fail(string message, string detailMessage)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Fail(message, detailMessage);
        }

        /// <summary>
        /// Flushes the output buffer, and causes buffered data to be written to the
        /// System.Diagnostics.Trace.Listeners.
        /// </summary>
        [Conditional("TRACE")]
        public static void Flush()
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Flush();
        }

        /// <summary>
        /// Increases the current System.Diagnostics.Trace.IndentLevel by one.
        /// </summary>
        [Conditional("TRACE")]
        public static void Indent()
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Indent();
        }

        /// <summary>
        /// Refreshes the trace configuration data.
        /// </summary>
        public static void Refresh()
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Refresh();
        }

        /// <summary>
        /// Writes an error message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection using the specified array of objects and formatting information.
        /// </summary>
        /// <param name="message">The error message to write.</param>
        [Conditional("TRACE")]
        public static void TraceError(string message)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.TraceError(message);
        }

        /// <summary>
        /// Writes an error message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection using the specified array of objects and formatting information.
        /// </summary>
        /// <param name="format">A format string that contains zero or more format items, which correspond
        /// to objects in the args array.</param>
        /// <param name="args">An object array containing zero or more objects to format.</param>
        [Conditional("TRACE")]
        public static void TraceError(string format, params object[] args)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.TraceError(format, args);
        }

        /// <summary>
        /// Writes an informational message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection using the specified array of objects and formatting information.
        /// </summary>
        /// <param name="message">The informative message to write.</param>
        [Conditional("TRACE")]
        public static void TraceInformation(string message)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.TraceInformation(message);
        }

        /// <summary>
        /// Writes an informational message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection using the specified array of objects and formatting information.
        /// </summary>
        /// <param name="format">A format string that contains zero or more format items, which correspond
        /// to objects in the args array.
        ///</param>
        /// <param name="args">An object array containing zero or more objects to format.</param>
        [Conditional("TRACE")]
        public static void TraceInformation(string format, params object[] args)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.TraceInformation(format, args);
        }

        /// <summary>
        /// Writes a warning message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection using the specified array of objects and formatting information.
        /// </summary>
        /// <param name="message">The warning message to write.</param>
        [Conditional("TRACE")]
        public static void TraceWarning(string message)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.TraceWarning(message);
        }

        /// <summary>
        /// Writes a warning message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection using the specified array of objects and formatting information.
        /// </summary>
        /// <param name="format">A format string that contains zero or more format items, which correspond
        /// to objects in the args array.
        ///</param>
        /// <param name="args">An object array containing zero or more objects to format.</param>
        [Conditional("TRACE")]
        public static void TraceWarning(string format, params object[] args)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.TraceWarning(format, args);
        }

        /// <summary>
        ///  Decreases the current System.Diagnostics.Trace.IndentLevel by one.
        /// </summary>
        [Conditional("TRACE")]
        public static void Unindent()
        {
            if (!NetTrace.Enabled)
                return;

            Trace.Unindent();
        }

        /// <summary>
        /// Writes a category name and a message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection.
        /// </summary>
        /// <param name="channel">The channel which is currently used.</param>
        /// <param name="message">A message to write.</param>
        /// <param name="category">A category name used to organize the output.</param>
        [Conditional("TRACE")]
        public static void WriteLine(string message, Connection channel, NetTraceCategory category)
        {
            //Console.WriteLine(NetTrace.GetMessage(message, channel));
            if (!NetTrace.Enabled)
                return;

            if (!NetTrace.ValidateCategory(category))
                return;

            Trace.WriteLine(
                NetTrace.GetMessage(message, channel),
                NetTrace.GetCategoryName(category)
                );
        }

        /// <summary>
        /// Writes a category name and a message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection.
        /// </summary>
        /// <param name="channel">The channel which is currently used.</param>
        /// <param name="message">A message to write.</param>
        /// <param name="category">A category name used to organize the output.</param>
        [Conditional("TRACE")]
        public static void WriteLine(string message, Connection channel, string category)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.WriteLine(
                NetTrace.GetMessage(message, channel),
                category
                );
        }

        /// <summary>
        /// Writes a category name and message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection if a condition is true.
        /// </summary>
        /// <param name="condition">true to cause a message to be written; otherwise, false.</param>
        /// <param name="message">A message to write.</param>
        /// <param name="channel">The channel which is currently used.</param>
        /// <param name="category">A category name used to organize the output.</param>
        [Conditional("TRACE")]
        public static void WriteLineIf(bool condition, string message, Connection channel, string category)
        {
            if (!NetTrace.Enabled)
                return;

            Trace.WriteLineIf(
                condition,
                NetTrace.GetMessage(message, channel),
                category
                );
        }

        /// <summary>
        /// Writes a category name and message to the trace listeners in the System.Diagnostics.Trace.Listeners
        /// collection if a condition is true.
        /// </summary>
        /// <param name="channel">The channel which is currently used.</param>
        /// <param name="condition">true to cause a message to be written; otherwise, false.</param>
        /// <param name="message">A message to write.</param>
        /// <param name="category">A category name used to organize the output.</param>
        [Conditional("TRACE")]
        public static void WriteLineIf(bool condition, string message, Connection channel, NetTraceCategory category)
        {
            if (!NetTrace.Enabled)
                return;

            if (!NetTrace.ValidateCategory(category))
                return;

            Trace.WriteLineIf(
                condition,
                NetTrace.GetMessage(message, channel),
                NetTrace.GetCategoryName(category)
                );
        }

        #endregion Public Members

        #region Private Members

        /// <summary>
        /// Gets the category string to print in the message.
        /// </summary>
        private static string GetCategoryName(NetTraceCategory category)
        {
            return Enum.GetName(typeof(NetTraceCategory), category).ToUpper();
        }

        /// <summary>
        /// Gets whether the specified category is enabled or not
        /// </summary>
        private static bool ValidateCategory(NetTraceCategory category)
        {
            switch (category)
            {
                case NetTraceCategory.Channel: return NetTrace.TraceChannel;
                case NetTraceCategory.Http: return NetTrace.TraceHttp;
                case NetTraceCategory.Emitter: return NetTrace.TraceEmitter;
                case NetTraceCategory.WebSocket: return NetTrace.TraceWebSocket;
                case NetTraceCategory.Ssl: return NetTrace.TraceSsl;
                case NetTraceCategory.Rest: return NetTrace.TraceRest;
                case NetTraceCategory.Mesh: return NetTrace.TraceMesh;
                case NetTraceCategory.Mqtt: return NetTrace.TraceMqtt;
                default: return true;
            }
        }

        /// <summary>
        /// Gets the message format.
        /// </summary>
        private static string GetMessage(string message, Connection channel)
        {
            if (channel == null)
                return message;

            return String.Format("[{1}] {0}", message, channel.ConnectionId.DebugView);
        }

        #endregion Private Members
    }

    /// <summary>
    /// The enumeration for predefined internal spike net trace categories.
    /// </summary>
    public enum NetTraceCategory
    {
        /// <summary>
        /// Basic networking channel tracer category
        /// </summary>
        Channel,

        /// <summary>
        /// Emitter message protocol tracer category
        /// </summary>
        Emitter,

        /// <summary>
        /// HTTP tracer category
        /// </summary>
        Http,

        /// <summary>
        /// HTTP WebSocket protocols tracer category
        /// </summary>
        WebSocket,

        /// <summary>
        /// SSL/TLS security protocol tracer category
        /// </summary>
        Ssl,

        /// <summary>
        /// HTTP REST protocol tracer category
        /// </summary>
        Rest,

        /// <summary>
        /// Mesh protocol tracer category
        /// </summary>
        Mesh,

        /// <summary>
        /// Mqtt protocol tracer category
        /// </summary>
        Mqtt
    }
}