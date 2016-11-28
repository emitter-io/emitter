using System;
using Emitter.Replication;
using Xunit;

namespace Emitter.Tests.Replication
{
    public class TestOfReplicatedDictionary
    {
        [Fact]
        public void AddRemove()
        {
            // Create a registry to replicate
            var max = 5;
            var registry = new ReplicatedHybridDictionary(1);

            for (int i = 0; i < max; ++i)
            {
                var data = i.ToHex();
                registry.Add(data, new ReplicatedState(data));
            }

            Assert.True(registry.Count == max);

            for (int i = 0; i < max; ++i)
            {
                registry.Remove(i.ToHex());
            }

            Assert.True(registry.Count == 0);
        }
    }
}