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
using System.Text;
using Emitter.Replication;

namespace Emitter.Network
{
    /// <summary>
    /// Represents a packet reader used for Emitter deserialization.
    /// </summary>
    public sealed class PacketReader : RecyclableObject
    {
        #region Properties

        // Static encoding (perf tweak)
        private static Encoding UTF8 = Encoding.UTF8;

        // Binary data
        private byte[] fTemp = new byte[8];

        private byte[] fData;
        private int fIndexMax;
        private int fIndex;

        #endregion Properties

        #region Constructors

        internal PacketReader()
        {
        }

        #endregion Constructors

        #region Internal Members

        /// <summary>
        /// Gets or sets the byte array on which this <see cref="PacketReader"/> operates on.
        /// </summary>
        internal byte[] Data
        {
            get { return fData; }
            set { fData = value; }
        }

        /// <summary>
        /// Gets or sets the lenght of the byte array chunk on which this <see cref="PacketReader"/> operates on.
        /// </summary>
        internal int IndexMax
        {
            get { return fIndexMax; }
            set { fIndexMax = value; }
        }

        /// <summary>
        /// Gets or sets the current position in the byte array.
        /// </summary>
        internal int Index
        {
            get { return fIndex; }
            set { fIndex = value; }
        }

        #endregion Internal Members

        #region Reading Methods

        /// <summary>
        /// Performs a seek in the underlying stream.
        /// </summary>
        /// <param name="offset">The offset to seek.</param>
        /// <param name="origin">The origin type to seek.</param>
        /// <returns>The current position.</returns>
        public int Seek(int offset, SeekOrigin origin)
        {
            switch (origin)
            {
                case SeekOrigin.Begin: fIndex = offset; break;
                case SeekOrigin.Current: fIndex += offset; break;
                case SeekOrigin.End: fIndex = fIndexMax - offset; break;
            }

            return fIndex;
        }

        /// <summary>
        /// Performs a peek from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public byte PeekByte()
        {
            if ((fIndex + 1) > fIndexMax)
                return 0;

            return fData[fIndex];
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public long ReadInt64()
        {
            if ((fIndex + 8) > fIndexMax)
                return 0;

            return ((long)fData[fIndex++] << 56)
                 | ((long)fData[fIndex++] << 48)
                 | ((long)fData[fIndex++] << 40)
                 | ((long)fData[fIndex++] << 32)
                 | ((long)fData[fIndex++] << 24)
                 | ((long)fData[fIndex++] << 16)
                 | ((long)fData[fIndex++] << 8)
                 | ((long)fData[fIndex++]);
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public ulong ReadUInt64()
        {
            if ((fIndex + 8) > fIndexMax)
                return 0;

            return (ulong)
                   ((ulong)fData[fIndex++] << 56)
                 | ((ulong)fData[fIndex++] << 48)
                 | ((ulong)fData[fIndex++] << 40)
                 | ((ulong)fData[fIndex++] << 32)
                 | ((ulong)fData[fIndex++] << 24)
                 | ((ulong)fData[fIndex++] << 16)
                 | ((ulong)fData[fIndex++] << 8)
                 | ((ulong)fData[fIndex++]);
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public int ReadInt32()
        {
            if ((fIndex + 4) > fIndexMax)
                return 0;

            return fData[fIndex++] << 24
                 | (fData[fIndex++] << 16)
                 | (fData[fIndex++] << 8)
                 | (fData[fIndex++]);
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public short ReadInt16()
        {
            if ((fIndex + 2) > fIndexMax)
                return 0;

            return (short)((fData[fIndex++] << 8) | fData[fIndex++]);
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public byte ReadByte()
        {
            if ((fIndex + 1) > fIndexMax)
                return 0;

            return fData[fIndex++];
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public uint ReadUInt32()
        {
            if ((fIndex + 4) > fIndexMax)
                return 0;

            return (uint)((fData[fIndex++] << 24) | (fData[fIndex++] << 16) | (fData[fIndex++] << 8) | fData[fIndex++]);
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public ushort ReadUInt16()
        {
            if ((fIndex + 2) > fIndexMax)
                return 0;

            return (ushort)((fData[fIndex++] << 8) | fData[fIndex++]);
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public sbyte ReadSByte()
        {
            if ((fIndex + 1) > fIndexMax)
                return 0;

            return (sbyte)fData[fIndex++];
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public bool ReadBoolean()
        {
            if ((fIndex + 1) > fIndexMax)
                return false;

            return (fData[fIndex++] != 0);
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public string ReadString()
        {
            var length = ReadInt32();
            if (length == 0)
                return String.Empty;
            var result = UTF8.GetString(fData, fIndex, length);
            fIndex += length;
            return result;
        }

        /// <summary>
        /// Performs a read from the underlying stream.
        /// </summary>
        /// <returns>Returns value read.</returns>
        public DateTime ReadDateTime()
        {
            return new DateTime(ReadInt64());
        }

        /// <summary>
        /// Reads an IEEE 754 single-precision (32-bit) floating-point number from the buffer
        /// </summary>
        public unsafe float ReadSingle()
        {
            if ((fIndex + 4) > fIndexMax)
                return 0;

            // Which way should we write?
            if (BitConverter.IsLittleEndian)
            {
                // We have to reverse
                fTemp[3] = fData[fIndex];
                fTemp[2] = fData[fIndex + 1];
                fTemp[1] = fData[fIndex + 2];
                fTemp[0] = fData[fIndex + 3];
            }
            else
            {
                // We have it in order
                fTemp[0] = fData[fIndex];
                fTemp[1] = fData[fIndex + 1];
                fTemp[2] = fData[fIndex + 2];
                fTemp[3] = fData[fIndex + 3];
            }

            unchecked
            {
                // Read the value from the buffer
                fIndex += 4;
                fixed (byte* pBuffer = fTemp)
                    return *((float*)(pBuffer));
            }
        }

        /// <summary>
        /// Reads an IEEE 754 double-precision (64-bit) floating-point number from the buffer
        /// </summary>
        public unsafe double ReadDouble()
        {
            if ((fIndex + 8) > fIndexMax)
                return 0;

            // Which way should we write?
            if (BitConverter.IsLittleEndian)
            {
                // We have to reverse
                fTemp[7] = fData[fIndex];
                fTemp[6] = fData[fIndex + 1];
                fTemp[5] = fData[fIndex + 2];
                fTemp[4] = fData[fIndex + 3];
                fTemp[3] = fData[fIndex + 4];
                fTemp[2] = fData[fIndex + 5];
                fTemp[1] = fData[fIndex + 6];
                fTemp[0] = fData[fIndex + 7];
            }
            else
            {
                // We have it in order
                fTemp[0] = fData[fIndex];
                fTemp[1] = fData[fIndex + 1];
                fTemp[2] = fData[fIndex + 2];
                fTemp[3] = fData[fIndex + 3];
                fTemp[4] = fData[fIndex + 4];
                fTemp[5] = fData[fIndex + 5];
                fTemp[6] = fData[fIndex + 6];
                fTemp[7] = fData[fIndex + 7];
            }

            unchecked
            {
                fIndex += 8;
                fixed (byte* pBuffer = fTemp)
                    return *((double*)(pBuffer));
            }
        }

        /// <summary>
        /// Reads a list of packets
        /// </summary>
        public T ReadSerializable<T>()
            where T : ISerializable, new()
        {
            var packet = new T();
            packet.Read(this);
            return packet;
        }

        /// <summary>
        /// Reads a list of packets
        /// </summary>
        public T ReadReplicated<T>()
            where T : IReplicated, new()
        {
            var packet = new T();
            packet.Read(this);
            return packet;
        }

        /// <summary>
        /// Reads a byte array
        /// </summary>
        public byte[] ReadByteArray()
        {
            var length = ReadInt32();
            var result = new byte[length];
            Memory.Copy(fData, fIndex, result, 0, length);
            fIndex += length;
            return result;
        }

        /// <summary>
        /// Reads a byte array
        /// </summary>
        public byte[] ReadBytes(int length)
        {
            var result = new byte[length];
            Memory.Copy(fData, fIndex, result, 0, length);
            fIndex += length;
            return result;
        }

        #region Read ValueType lists

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<bool> ReadListOfBoolean()
        {
            var length = ReadInt32();
            var result = new List<bool>();
            for (int i = 0; i < length; ++i)
                result.Add((bool)ReadBoolean());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<byte> ReadListOfByte()
        {
            var length = ReadInt32();
            var result = new List<byte>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadByte());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<sbyte> ReadListOfSByte()
        {
            var length = ReadInt32();
            var result = new List<sbyte>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadSByte());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<short> ReadListOfInt16()
        {
            var length = ReadInt32();
            var result = new List<short>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadInt16());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<ushort> ReadListOfUInt16()
        {
            var length = ReadInt32();
            var result = new List<ushort>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadUInt16());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<int> ReadListOfInt32()
        {
            var length = ReadInt32();
            var result = new List<int>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadInt32());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<uint> ReadListOfUInt32()
        {
            var length = ReadInt32();
            var result = new List<uint>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadUInt32());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<long> ReadListOfInt64()
        {
            var length = ReadInt32();
            var result = new List<long>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadInt64());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<ulong> ReadListOfUInt64()
        {
            var length = ReadInt32();
            var result = new List<ulong>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadUInt64());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<double> ReadListOfDouble()
        {
            var length = ReadInt32();
            var result = new List<double>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadDouble());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<float> ReadListOfSingle()
        {
            var length = ReadInt32();
            var result = new List<float>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadSingle());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<DateTime> ReadListOfDateTime()
        {
            var length = ReadInt32();
            var result = new List<DateTime>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadDateTime());
            return result;
        }

        /// <summary>
        /// Reads a list
        /// </summary>
        public List<string> ReadListOfString()
        {
            var length = ReadInt32();
            var result = new List<string>();
            for (int i = 0; i < length; ++i)
                result.Add(ReadString());
            return result;
        }

        /// <summary>
        /// Reads a list of packets
        /// </summary>
        public List<T> ReadListOfSerializable<T>()
            where T : ISerializable, new()
        {
            var length = ReadInt32();
            var result = new List<T>();

            for (int i = 0; i < length; ++i)
            {
                var packet = new T();
                packet.Read(this);
                result.Add(packet);
            }
            return result;
        }

        #endregion Read ValueType lists

        #endregion Reading Methods

        /// <summary>
        /// Recycles (resets) the object to the original state.
        /// </summary>
        public override void Recycle()
        {
            fData = null;
            fTemp = new byte[8];
            fIndexMax = 0;
            fIndex = 0;
        }
    }
}