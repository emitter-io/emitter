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
    /// Represents typed array utilitiy.
    /// </summary>
    /// <typeparam name="T">The type of the array</typeparam>
    public static class ArrayUtils<T>
    {
        public static readonly T[] Empty = new T[0];
    }

    public static class ArrayUtils
    {
        public static void Add<T>(ref T[] array, T newItem)
        {
            if (array == null)
            {
                array = new T[1];
                array[0] = newItem;
                return;
            }

            int count = array.Length;
            T[] newArray = new T[count + 1];
            Array.Copy(array, newArray, count);
            newArray[count] = newItem;
            array = newArray;
        }

        public static void AddUnique<T>(ref T[] array, T newItem)
        {
            if (array == null)
            {
                array = new T[1];
                array[0] = newItem;
                return;
            }

            int count = array.Length;
            for (int i = 0; i < count; ++i)
            {
                if (array[i].Equals(newItem))
                    return;
            }

            T[] newArray = new T[count + 1];
            Array.Copy(array, newArray, count);
            newArray[count] = newItem;
            array = newArray;
        }

        public static int Remove<T>(ref T[] array, T removedItem)
        {
            int count = array.Length;

            if (count == 0)
                return -1;

            for (int i = 0; i < count; i++)
            {
                if (!Equals(array[i], removedItem))
                    continue;

                T[] newArray = new T[count - 1];
                Array.Copy(array, 0, newArray, 0, i);
                Array.Copy(array, i + 1, newArray, i, count - i - 1);
                array = newArray;

                return i;
            }

            return -1;
        }

        public static void Reserve<StorageType>(ref StorageType[] array, ref int count, int elementCount)
        {
            if (array == null)
            {
                array = new StorageType[elementCount];
                count = 0;
                return;
            }

            if (array.Length < elementCount)
            {
                int allocSize = HigherPower2(elementCount);

                Array.Resize(ref array, allocSize);
            }
        }

        public static void Reserve<StorageType>(ref StorageType[] array, int elementCount)
        {
            if (array == null)
            {
                array = new StorageType[elementCount];
                return;
            }

            if (array.Length < elementCount)
            {
                int allocSize = HigherPower2(elementCount);

                Array.Resize(ref array, allocSize);
            }
        }

        public static int HigherPower2(int value)
        {
            int result = value;
            result |= result >> 1;
            result |= result >> 2;
            result |= result >> 4;
            result |= result >> 8;
            result |= result >> 16;
            result++;
            return result;
        }

        public static void RemoveAt<T>(ref T[] array, int index)
        {
            int count = array.Length;

            T[] newArray = new T[count > 1 ? count - 1 : count];
            Array.Copy(array, 0, newArray, 0, index);
            Array.Copy(array, index + 1, newArray, index, count - index - 1);
            array = newArray;
        }

        public static void Add<T>(ref T[] array, ref int count, T newItem)
        {
            if (array == null)
            {
                array = new T[1];
                array[0] = newItem;
                count = 1;
                return;
            }

            if (count < array.Length)
            {
                array[count++] = newItem;
                return;
            }

            // have to reallocate
            int newSize = (int)HigherPower2(count + 1);
            Array.Resize(ref array, newSize);
            array[count] = newItem;

            count++;
        }

        public static void Add<T>(ref T[] array, ref int count, ref T newItem)
        {
            if (array == null)
            {
                array = new T[1];
                array[0] = newItem;
                count = 1;
                return;
            }

            if (count < array.Length)
            {
                array[count++] = newItem;
                return;
            }

            // have to reallocate
            int growth;
            if (count < 2)
                growth = 1;
            else
                growth = (int)Math.Log(count) + 1;

            T[] newArray = new T[count + growth];
            Array.Copy(array, newArray, count);
            newArray[count] = newItem;
            array = newArray;

            count++;
        }

        public static int CreateSlot<T>(ref T[] array, ref int count)
        {
            int entry;

            // assumption for this function is
            // the array passed in is never null
            /*if (array == null)
            {
                array = new T[1];
                count = 1;
                return entry;
            }*/

            if (count < array.Length)
            {
                entry = count;
                count++;
                return entry;
            }

            // have to reallocate
            int growth;
            if (count < 2)
                growth = 1;
            else
                growth = (int)Math.Log(count) + 1;

            T[] newArray = new T[count + growth];
            Array.Copy(array, newArray, count);
            array = newArray;

            entry = count;
            count++;

            return entry;
        }

        public static void AddRange<StorageType>(ref StorageType[] array, ref int count, StorageType[] newItems, int newItemCount)
        {
            if (array == null)
            {
                Array.Resize(ref array, newItemCount);
                count = newItemCount;
                return;
            }

            if ((count + newItemCount) < array.Length)
            {
                Array.Copy(newItems, 0, array, count, newItemCount);
                count += newItemCount;

                return;
            }

            // have to reallocate
            int allocSize = HigherPower2(count + newItemCount);
            StorageType[] newArray = new StorageType[allocSize];
            Array.Copy(array, newArray, count);
            Array.Copy(newItems, 0, newArray, count, newItemCount);
            count += newItemCount;
            array = newArray;
        }

        public static void AddRange<T>(ref T[] array, ref int count, T[] newItems)
        {
            int newItemCount = newItems.Length;

            if (newItemCount == 0)
                return;

            if (array == null)
            {
                array = new T[newItemCount];
                Array.Copy(newItems, array, newItemCount);
                count = newItemCount;
                return;
            }

            if ((count + newItemCount) < array.Length)
            {
                Array.Copy(newItems, 0, array, count, newItemCount);
                count += newItemCount;

                return;
            }

            // have to reallocate
            T[] newArray = new T[count + newItemCount];
            Array.Copy(array, newArray, count);
            Array.Copy(newItems, 0, newArray, count, newItemCount);
            count += newItemCount;
            array = newArray;
        }

        public static void Remove<T>(ref T[] array, ref int count, T removedItem)
        {
            for (int i = 0; i < count; i++)
            {
                if (!Equals(array[i], removedItem))
                    continue;

                RemoveAt(ref array, ref count, i);

                return;
            }
        }

        public static void Remove(ref int[] array, ref int count, int removedItem)
        {
            if (count - removedItem < removedItem)
            {
                for (int i = count - 1; i >= 0; --i)
                {
                    if (array[i] != removedItem)
                        continue;

                    RemoveAt(ref array, ref count, i);

                    return;
                }
            }
            else
            {
                for (int i = 0; i < count; i++)
                {
                    if (array[i] != removedItem)
                        continue;

                    RemoveAt(ref array, ref count, i);

                    return;
                }
            }
        }

        public static void RemoveAt<T>(ref T[] array, ref int count, int index)
        {
            Array.Copy(array, index + 1, array, index, count - index - 1);
            count--;
        }

        public static bool Contains<T>(T[] array, int count, T searchItem)
        {
            for (int i = 0; i < count; i++)
            {
                if (Equals(array[i], searchItem))
                    return true;
            }

            return false;
        }
    }
}