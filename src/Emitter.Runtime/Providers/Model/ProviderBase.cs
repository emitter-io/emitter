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
using System.Collections.Specialized;
using System.Linq;

namespace Emitter
{
    /// <summary>
    /// This class provides a base implementation for the extensible provider model.
    /// </summary>
    public abstract class Provider : DisposableObject
    {
        // Fields
        private string fName;

        private string fDescription;
        private bool fInitialized;

        /// <summary>
        /// Initializes a new instance of the System.Configuration.Provider.ProviderBase
        /// class.
        /// </summary>
        protected Provider()
        {
        }

        /// <summary>
        /// Initializes the provider.
        /// </summary>
        /// <param name="name">The friendly name of the provider.</param>
        /// <param name="config">A collection of the name/value pairs representing the provider-specific attributes
        /// specified in the configuration for this provider.
        /// </param>
        public virtual void Initialize(string name, NameValueCollection config)
        {
            lock (this)
            {
                if (this.fInitialized)
                    throw new InvalidOperationException("Provider have already been initialized.");
                this.fInitialized = true;
            }

            if (name == null)
                throw new ArgumentNullException("name");

            if (name.Length == 0)
                throw new ArgumentException("The supplied provider name is null or empty.", "name");

            this.fName = name;
            if (config != null)
            {
                this.fDescription = config["description"];
                config.Remove("description");
            }
        }

        /// <summary>
        /// Gets a brief, friendly description suitable for display in administrative
        /// tools or other user interfaces.
        /// </summary>
        public virtual string Description
        {
            get
            {
                if (!string.IsNullOrEmpty(this.fDescription))
                    return this.fDescription;

                return this.Name;
            }
        }

        /// <summary>
        /// Gets the friendly name used to refer to the provider during configuration.
        /// </summary>
        public virtual string Name
        {
            get { return this.fName; }
        }
    }
}