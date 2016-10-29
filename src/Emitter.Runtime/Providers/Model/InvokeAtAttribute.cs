using System;
using System.Linq;

namespace Emitter
{
    /// <summary>
    /// Enumeration that defines a state during which the target annotated method should be invoked.
    /// </summary>
    public enum InvokeAtType
    {
        /// <summary>
        /// Specifies that the target method should be invoked during the server configuration state.
        /// </summary>
        Configure,

        /// <summary>
        /// Specifies that the target method should be invoked during the server initialization state.
        /// </summary>
        Initialize,

        /// <summary>
        /// Specifies that the target method should be invoked during the server termination state.
        /// </summary>
        Terminate
    }

    /// <summary>
    /// Attribute that allows a target method to be annotated for automatic invoke.
    /// </summary>
    [AttributeUsage(AttributeTargets.Method)]
    public class InvokeAtAttribute : Attribute
    {
        private int fPriority;
        private InvokeAtType fType;

        /// <summary>
        /// Gets or sets the priority of the function to invoke
        /// </summary>
        public int Priority
        {
            get { return fPriority; }
            set { fPriority = value; }
        }

        /// <summary>
        /// Gets or sets when the function should be invoked
        /// </summary>
        public InvokeAtType Type
        {
            get { return fType; }
            set { fType = value; }
        }

        /// <summary>
        /// Attribute that allows a target method to be annotated for automatic invoke.
        /// </summary>
        /// <param name="type">Specifies a state during which the target annotated method should be invoked.</param>
        public InvokeAtAttribute(InvokeAtType type)
        {
            fType = type;
        }

        /// <summary>
        /// Attribute that allows a target method to be annotated for automatic invoke.
        /// </summary>
        /// <param name="type">Specifies a state during which the target annotated method should be invoked.</param>
        /// <param name="priority">Specifies the priority with which this method should be invoked.</param>
        public InvokeAtAttribute(InvokeAtType type, int priority)
        {
            fType = type;
            fPriority = priority;
        }
    }
}