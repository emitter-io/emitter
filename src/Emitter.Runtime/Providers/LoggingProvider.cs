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
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Text;
using Emitter.Network;
using Emitter.Text.Json;

namespace Emitter.Providers
{
    /// <summary>
    /// Provides the logging functionnality
    /// </summary>
    public abstract class LoggingProvider : Provider
    {
        /// <summary>
        /// Logs an info or an error.
        /// </summary>
        /// <param name="level">Level of the log entry (info, warning or error).</param>
        /// <param name="message">Message of the log entry.</param>
        public abstract void Log(LogLevel level, string message);

        /// <summary>
        /// Logs an info.
        /// </summary>
        /// <param name="message">Message of the log entry (info).</param>
        public abstract void Log(string message);

        /// <summary>
        /// Logs an exception as a warning.
        /// </summary>
        /// <param name="exception">The exception object that contains the error.</param>
        public virtual void Log(Exception exception)
        {
            Log(LogLevel.Error, exception, default(ArraySegment<byte>));
        }

        /// <summary>
        /// Logs an exception as a warning.
        /// </summary>
        /// <param name="exception">The exception object that contains the error.</param>
        /// <param name="buffer">The buffer for the log entry.</param>
        public virtual void Log(Exception exception, BufferSegment buffer)
        {
            Log(LogLevel.Error, exception, buffer.AsSegment());
        }

        /// <summary>
        /// Logs an exception with a specified error level.
        /// </summary>
        /// <param name="exception">The exception object that contains the error.</param>
        /// <param name="level">Level of the log entry (info, warning or error).</param>
        /// <param name="buffer">The buffer for the log entry.</param>
        public abstract void Log(LogLevel level, Exception exception, ArraySegment<byte> buffer);
    }

    /// <summary>
    /// Default, console multi text logging
    /// </summary>
    public class MultiTextLoggingProvider : LoggingProvider
    {
        internal static MultiTextWriter fOut;

        static MultiTextLoggingProvider()
        {
            Console.SetOut(fOut = new MultiTextWriter(Console.Out));
        }

        #region LoggingProvider Implementation

        /// <summary>
        /// Logs an info or an error.
        /// </summary>
        public override void Log(LogLevel level, string message)
        {
            if (level == LogLevel.Error) fOut.WriteLine("Error: " + message);
            else if (level == LogLevel.Warning) fOut.WriteLine("Warning: " + message);
            else fOut.WriteLine(message);
        }

        /// <summary>
        /// Logs an info or an error.
        /// </summary>
        public override void Log(string message)
        {
            Log(LogLevel.Info, message);
        }

        /// <summary>
        /// Logs an exception with a specified error level.
        /// </summary>
        public override void Log(LogLevel level, Exception exception, ArraySegment<byte> buffer)
        {
            if (exception is TargetInvocationException)
                exception = exception.InnerException;
            if (exception == null)
                return;

            fOut.WriteLine(new ExceptionObject(level, exception, buffer));
        }

        #endregion LoggingProvider Implementation
    }

    #region Text Writers

    /// <summary>
    /// Represents a <see cref="TextWriter"/> which outputs to several child TextWriters.
    /// </summary>
    public class MultiTextWriter : TextWriter
    {
        private List<TextWriter> fStreams;

        /// <summary>
        /// Constructs a new <see cref="MultiTextWriter"/> instance.
        /// </summary>
        /// <param name="streams">The streams to write the output into.</param>
        public MultiTextWriter(params TextWriter[] streams)
        {
            fStreams = new List<TextWriter>(streams);

            if (fStreams.Count < 0)
                throw new ArgumentException("You must specify at least one stream.");
        }

        /// <summary>
        /// Adds a new <see cref="TextWriter"/> to the list of writers.
        /// </summary>
        /// <param name="tw">The <see cref="TextWriter"/> to add.</param>
        public void Add(TextWriter tw)
        {
            fStreams.Add(tw);
        }

        /// <summary>
        /// Removes a <see cref="TextWriter"/> from the list of writers.
        /// </summary>
        /// <param name="tw">The <see cref="TextWriter"/> to remove.</param>
        public void Remove(TextWriter tw)
        {
            fStreams.Remove(tw);
        }

        /// <summary>
        /// Writes a character.
        /// </summary>
        /// <param name="ch">Character to write.</param>
        public override void Write(char ch)
        {
            for (int i = 0; i < fStreams.Count; i++)
                fStreams[i].Write(ch);
        }

        /// <summary>
        /// Writes a new line.
        /// </summary>
        /// <param name="line">A new line to write.</param>
        public override void WriteLine(string line)
        {
            for (int i = 0; i < fStreams.Count; i++)
                fStreams[i].WriteLine(line);
        }

        /// <summary>
        /// Writes a new line and applies a String.Format prior to writing.
        /// </summary>
        /// <param name="line">A new line to write.</param>
        /// <param name="args">Format arguments.</param>
        public override void WriteLine(string line, params object[] args)
        {
            WriteLine(String.Format(line, args));
        }

        /// <summary>
        /// Gets the encoding used by the <see cref="MultiTextWriter"/>.
        /// </summary>
        public override Encoding Encoding
        {
            get { return Encoding.UTF8; }
        }
    }

    #endregion Text Writers

    /// <summary>
    /// Represents a serializable exception object.
    /// </summary>
    public class ExceptionObject
    {
        public ExceptionObject(LogLevel level, Exception ex, ArraySegment<byte> buffer)
        {
            this.Level = level;
            this.Exception = ex.GetType()?.FullName;
            this.Message = ex.Message;
            this.Source = ex.Source;
            this.Stack = ex.StackTrace;

            if (buffer != default(ArraySegment<byte>))
                this.Buffer = Encoding.ASCII.GetString(buffer.Array, buffer.Offset, buffer.Count);
        }

        public LogLevel Level;

        [JsonProperty(NullValueHandling = NullValueHandling.Ignore)]
        public string Source;

        [JsonProperty(NullValueHandling = NullValueHandling.Ignore)]
        public string Exception;

        [JsonProperty(NullValueHandling = NullValueHandling.Ignore)]
        public string Message;

        [JsonProperty(NullValueHandling = NullValueHandling.Ignore)]
        public string Stack;

        [JsonProperty(NullValueHandling = NullValueHandling.Ignore)]
        public string Buffer;

        public override string ToString()
        {
            return "Exception : " + this.Exception + Environment.NewLine +
                   "Message   : " + this.Message + Environment.NewLine +
                   (string.IsNullOrWhiteSpace(this.Stack) ? ("Source    : " + this.Source) : ("Stack     : " + this.Stack));
        }

        public string ToJson()
        {
            return JsonConvert.SerializeObject(this);
        }

        public byte[] ToBytes()
        {
            return Encoding.UTF8.GetBytes(this.ToJson());
        }
    }
}