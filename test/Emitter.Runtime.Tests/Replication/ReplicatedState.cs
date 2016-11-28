using System;
using System.Linq;
using Emitter.Network;
using Emitter.Replication;

namespace Emitter.Tests.Replication
{
    /// <summary>
    /// Represents a test state.
    /// </summary>
    public sealed class ReplicatedState : ReplicatedObject, IComparable<ReplicatedState>
    {
        /// <summary>
        /// Gets the replicated data.
        /// </summary>
        public string Data;

        /// <summary>
        /// Deserialization constructor.
        /// </summary>
        public ReplicatedState()
        {
        }

        /// <summary>
        /// Creates an instance of an object.
        /// </summary>
        /// <param name="data"></param>
        public ReplicatedState(string data)
        {
            this.Data = data;
        }

        /// <summary>
        /// Compares the gossip member to another one.
        /// </summary>
        /// <param name="other"></param>
        /// <returns></returns>
        public int CompareTo(ReplicatedState other)
        {
            return this.Data.CompareTo(other.Data.ToString());
        }

        /// <summary>
        /// Serializes this packet to a binary stream.
        /// </summary>
        /// <param name="reader">PacketReader used to serialize the packet.</param>
        public override void Read(PacketReader reader)
        {
            this.Data = reader.ReadString();
        }

        /// <summary>
        /// Deserializes this packet from a binary stream.
        /// </summary>
        /// <param name="writer">PacketWriter used to deserialize the packet.</param>
        public override void Write(PacketWriter writer, ReplicatedVersion since)
        {
            writer.Write(this.Data);
        }
    }
}