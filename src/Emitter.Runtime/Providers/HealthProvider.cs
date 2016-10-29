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

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider of health status.
    /// </summary>
    public abstract class HealthProvider : Provider
    {
        /// <summary>
        /// Gets whether the service is healthy or not.
        /// </summary>
        /// <returns>Whether the service is healthy or not.</returns>
        public abstract bool IsHealthy();
    }

    /// <summary>
    /// Represents a provider of health status.
    /// </summary>
    public class DefaultHealthProvider : HealthProvider
    {
        /// <summary>
        /// Gets whether the service is healthy or not.
        /// </summary>
        /// <returns>Whether the service is healthy or not.</returns>
        public override bool IsHealthy()
        {
            return true;
        }
    }
}