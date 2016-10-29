namespace Emitter.Providers
{
    /// <summary>
    /// A Lifetime Manager that always returns a new instance.
    /// </summary>
    /// <remarks>
    /// <para>Setting the IocContainer's or Registration's lifetime manager to null is equivalent
    /// to setting it to an instance of AlwaysNewLifetime. This is the default lifetime manager
    /// for the IocContainer.</para>
    /// The one instance can be used for all registrations.
    /// </remarks>
    /// <example>
    /// The IocContainer default lifetime manager is set to Request lifetime, but one interface
    /// needs to alway be a new instance.
    /// <code>
    ///     var requestLifetime   = new RequestLifetime();
    ///     var alwaysNewLifetime = new AlwaysNewLifetime();
    ///     var container         = new ProvidersContainer();
    ///
    ///     container.UsesDefaultLifetimeManagerOf(requestLifetime);
    ///     container.Register&lt;IMyInterface, MyImplementation&gt;()
    ///			     .WithLifetimeManager(alwaysNewLifetime);
    ///     container.Register&lt;IMyInterface2, MyImplementation2&gt;()
    ///			     .WithLifetimeManager(alwaysNewLifetime);
    ///	    ...
    /// </code>
    /// </example>
    public class AlwaysNewLifetime : ILifetimeManager
    {
        #region ILifetimeManager Members

        /// <summary>
        /// Gets a new instance.
        /// </summary>
        /// <param name="creator">The creator (registration) that can create an instance</param>
        /// <returns>The new instance.</returns>
        public object GetInstance(IInstanceCreator creator)
        {
            return creator.CreateInstance(ContainerCaching.InstanceNotCachedInContainer);
        }

        /// <summary>
        /// Invalidates any cached instances.
        /// </summary>
        /// <param name="registration">The registration.</param>
        public void InvalidateInstanceCache(IRegistration registration)
        {
            // there is no instance cache ...
        }

        #endregion ILifetimeManager Members
    }
}