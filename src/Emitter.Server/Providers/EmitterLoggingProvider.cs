using System;
using System.Diagnostics;
using System.Reflection;
using Emitter.Providers;

namespace Emitter
{
    /// <summary>
    /// Logginf provider for emitter.
    /// </summary>
    public class EmitterLoggingProvider : LoggingProvider
    {
        internal static MultiTextWriter Out;

        static EmitterLoggingProvider()
        {
            Console.SetOut(Out = new MultiTextWriter(Console.Out));
        }

        #region LoggingProvider Implementation

        /// <summary>
        /// Logs an info or an error.
        /// </summary>
        public override void Log(LogLevel level, string message)
        {
            if (level == LogLevel.Error) Out.WriteLine("Error: " + message);
            else if (level == LogLevel.Warning) Out.WriteLine("Warning: " + message);
            else Out.WriteLine(message);
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

            // Prepare an exception object.
            var ex = new ExceptionObject(level, exception, buffer);

            // Write to loggly and forget
            /*HttpUtility.PostAsync(
                "http://logs-01.loggly.com/inputs/186be958-25d8-4290-bc51-36e5f9502482/tag/http/",
                ex.ToBytes(),
                5000
                ).Forget();*/

            // Write to console
            Out.WriteLine(ex);
        }

        #endregion LoggingProvider Implementation
    }

    // Below taken from Reference Source
    // Outputs trace messages to the console.
    public class ConsoleTraceListener : TextWriterTraceListener
    {
        public ConsoleTraceListener()
            : base(Console.Out)
        { }

        public ConsoleTraceListener(bool useErrorStream)
            : base(useErrorStream ? Console.Error : Console.Out)
        { }
    }
}