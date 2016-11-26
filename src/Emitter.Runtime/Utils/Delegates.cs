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

namespace Emitter
{
    /// <summary>
    /// Encapsulates a method that has two parameters passed by reference and does not return a value.
    /// </summary>
    /// <typeparam name="T1">The type of the first parameter of the method that this delegate encapsulates.</typeparam>
    /// <typeparam name="T2">The type of the second parameter of the method that this delegate encapsulates.</typeparam>
    /// <param name="arg1">The first parameter of the method that this delegate encapsulates.</param>
    /// <param name="arg2">The second parameter of the method that this delegate encapsulates.</param>
    public delegate void RefAction<T1, T2>(ref T1 arg1, T2 arg2);

    /// <summary>
    /// Encapsulates a method that has one parameter passed by reference and does not return a value.
    /// </summary>
    /// <typeparam name="T">The type of the first parameter of the method that this delegate encapsulates.</typeparam>
    /// <param name="value">The first parameter of the method that this delegate encapsulates.</param>
    public delegate void RefAction<T>(ref T value);

    /// <summary>
    /// Defines an event issued by <see cref="IClient"/> instance.
    /// </summary>
    public delegate void ClientEvent(IClient client);

    /// <summary>
    /// Defines an event issued by <see cref="Connection"/> instance.
    /// </summary>
    public delegate void ChannelEvent(Connection channel);
}