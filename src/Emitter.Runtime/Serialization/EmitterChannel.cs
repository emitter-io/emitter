using System;
using System.Collections.Generic;
using System.Runtime.CompilerServices;
using Emitter.Serialization;

namespace Emitter
{
    /// <summary>
    /// Set of extension methods for various packets.
    /// </summary>
    public sealed class EmitterChannel
    {
        /// <summary>
        /// Gets or sets the API key of the channel.
        /// </summary>
        public string Key;

        /// <summary>
        /// Gets or sets the channel string.
        /// </summary>
        public string Channel;

        /// <summary>
        /// Gets or sets the root hash.
        /// </summary>
        public uint Target;

        /// <summary>
        /// Gets or sets the channel type provided.
        /// </summary>
        public ChannelType Type;

        /// <summary>
        /// Gets or sets the options.
        /// </summary>
        public Dictionary<string, string> Options;

        /// <summary>
        /// Represents the parsing state.
        /// </summary>
        public enum ParseState
        {
            ReadingKey,
            ReadingChannel,
            ReadingOptionKey,
            ReadingOptionValue
        }

        #region Option Parsing

        /// <summary>
        /// Attempts to get an option.
        /// </summary>
        /// <param name="value">The parsed value.</param>
        public bool HasTimeToLive(out int value)
        {
            return EmitterOption.TryGet(this, EmitterOption.TimeToLive, EmitterConst.Transient, out value);
        }

        /// <summary>
        /// Attempts to get an option.
        /// </summary>
        /// <param name="value">The parsed value.</param>
        public bool RequestedLast(out int value)
        {
            return EmitterOption.TryGet(this, EmitterOption.LastHistory, 0, out value);
        }

        #endregion Option Parsing

        /// <summary>
        /// Splits the key from the channel for MQTT.
        /// </summary>
        /// <param name="channel">The channel to split</param>
        /// <param name="key">The key extracted.</param>
        /// <returns></returns>
        public static unsafe bool TrySplit(ref string channel, out string key)
        {
            var length = channel.Length;
            fixed (char* pChar = channel)
            {
                for (int i = 0; i < length; ++i)
                {
                    if (*(pChar + i) != EmitterConst.Separator)
                        continue;

                    key = new string(pChar, 0, i);
                    channel = new string(pChar, i + 1, length - i - 1);
                    return true;
                }

                key = null;
                return false;
            }
        }

        /// <summary>
        /// Attempts to parse the channel.
        /// </summary>
        /// <param name="channel">The channel string to parse.</param>
        /// <param name="info">The channel information parsed.</param>
        /// <param name="withKey">Whether we should parse the key first or not.</param>
        /// <returns>Whether we parsed or not.</returns>
        public static unsafe bool TryParse(string channel, bool withKey, out EmitterChannel info)
        {
            // The result
            info = new EmitterChannel();
            info.Type = ChannelType.Invalid;
            info.Target = 0;

            // Variables and state
            char symbol;
            string key = null;
            int begin = 0;
            var last = false;
            var state = withKey ? ParseState.ReadingKey : ParseState.ReadingChannel;
            var length = channel.Length;

            // Pin the string
            fixed (char* pChar = channel)
            {
                // Start the single-pass parsing
                for (int i = 0; i < length; ++i)
                {
                    // Is this the last character?
                    last = (i + 1 == length);
                    symbol = *(pChar + i);
                    switch (state)
                    {
                        // We're trying to read the provided API key, this should be the 32-character long
                        // key or 'emitter' string for custom API requests.
                        case ParseState.ReadingKey:
                            // We only need to wait until the '/' sign
                            if (symbol != EmitterConst.Separator)
                                break;

                            // Copy the key
                            info.Key = new string(pChar, 0, i);
                            begin = i + 1;

                            // Change the state and continue
                            state = ParseState.ReadingChannel;
                            continue;

                        // We're trying to read and validate the channel and infer the type of the channel.
                        case ParseState.ReadingChannel:

                            // If we're reading the first separator and haven't set the channel hash yet, we should
                            // compute the Murmur32 hash on our root.
                            if (info.Target == 0 && (last || symbol == EmitterConst.Separator))
                                info.Target = GetHash(pChar + begin, i - begin + (symbol == EmitterConst.Separator ? 0 : 1));

                            // If this symbol is a wildcard symbol
                            if (symbol == '+' || symbol == '*')
                            {
                                // Check if previous character is a '/' and whether we have another '/', '?' or nothing after.
                                if (*(pChar + i - 1) == EmitterConst.Separator &&
                                   (*(pChar + i + 1) == EmitterConst.Separator || *(pChar + i + 1) == '?' || last))
                                {
                                    // This is a valid wildcard
                                    info.Type = ChannelType.Wildcard;
                                    continue;
                                }

                                // Channel is invalid, we should stop parsing
                                return false;
                            }
                            // If it's the last character in the string, assign the whole thing to the channel. Similarly,
                            // if this is the '?' sign that means we have finished parsing the channel and can assign it.
                            else if (last || symbol == '?')
                            {
                                // We can now set the channel string
                                info.Channel = new string(pChar, begin, i - begin + (symbol == '?' ? 0 : 1));
                                begin = i + 1;

                                // Make sure channel ends with trailing slash
                                if (!(info.Channel[info.Channel.Length - 1] == EmitterConst.Separator))
                                    info.Channel += EmitterConst.Separator;

                                // If we haven't detected that the channel is a wildcard, this means this is the
                                // static channel or root. Set the type now.
                                if (info.Type != ChannelType.Wildcard)
                                    info.Type = ChannelType.Static;

                                // We need to create the options dictionary, since we do have options
                                if (symbol == '?')
                                    info.Options = new Dictionary<string, string>();

                                // Change the state and continue
                                state = ParseState.ReadingOptionKey;
                                continue;
                            }
                            // Is this a valid character?
                            else if ((symbol >= Code.DIGIT_0 && symbol <= Code.DIGIT_9) ||
                                (symbol == Code.DOT) ||
                                (symbol == Code.MINUS) ||
                                (symbol == Code.COLON) ||
                                (symbol == EmitterConst.Separator) ||
                                (symbol >= Code.A && symbol <= Code.Z) ||
                                (symbol >= Code.a && symbol <= Code.z))
                            {
                                // That's ok, continue
                                continue;
                            }
                            else
                            {
                                // Channel is invalid, we should stop parsing
                                return false;
                            }

                        // We're now reading the options, using the query-string format. This is a simple set of key/value pairs
                        // separated from the channel by '?', from each other by '&' and key from value by '='.
                        case ParseState.ReadingOptionKey:

                            // This is the case where we do not have a value, we assign an empty value
                            if (last || symbol == '=')
                            {
                                // Get the key
                                key = new string(pChar, begin, i - begin + (symbol == '=' ? 0 : 1));

                                // Change the state and continue
                                begin = i + 1;
                                state = ParseState.ReadingOptionValue;
                                continue;
                            }
                            // If we don't have the value
                            else if (last || symbol == '&')
                            {
                                // Write the key we have for this value and the value itself
                                key = new string(pChar, begin, i - begin + (symbol == '&' ? 0 : 1));
                                if (!info.Options.ContainsKey(key))
                                {
                                    info.Options.Add(key, string.Empty);
                                    key = null;
                                }

                                // Keep the state and continue
                                begin = i + 1;
                                continue;
                            }
                            // Is this a valid character?
                            else if ((symbol >= Code.DIGIT_0 && symbol <= Code.DIGIT_9) ||
                                (symbol == Code.MINUS) ||
                                (symbol >= Code.A && symbol <= Code.Z) ||
                                (symbol >= Code.a && symbol <= Code.z))
                            {
                                // That's ok, continue
                                continue;
                            }
                            else
                            {
                                // Channel is invalid, we should stop parsing
                                return false;
                            }

                        // Similarly, we parse the value of the option. In some cases, options could have no value and serve as
                        // simple flags (ie: true if present, false if not present).
                        case ParseState.ReadingOptionValue:

                            // We've finished reading the value, put it to the options dictionary now
                            if (last || symbol == '&')
                            {
                                // Write the key we have for this value and the value itself
                                if (!info.Options.ContainsKey(key))
                                {
                                    info.Options.Add(key, new string(pChar, begin, i - begin + (symbol == '&' ? 0 : 1)));
                                    key = null;
                                }

                                // Change the state and continue
                                begin = i + 1;
                                state = ParseState.ReadingOptionKey;
                                continue;
                            }
                            // Is this a valid character?
                            else if ((symbol >= Code.DIGIT_0 && symbol <= Code.DIGIT_9) ||
                                (symbol == Code.MINUS) ||
                                (symbol >= Code.A && symbol <= Code.Z) ||
                                (symbol >= Code.a && symbol <= Code.z))
                            {
                                // That's ok, continue
                                continue;
                            }
                            else
                            {
                                // Channel is invalid, we should stop parsing
                                return false;
                            }
                    }
                }

                // We're done parsing, return the structure
                return true;
            }
        }

        /// <summary>
        /// Gets the ssid array, 2-pass through the string.
        /// </summary>
        /// <param name="contract">The contract to add to the query key.</param>
        /// <param name="channel">The channel to add to the query key.</param>
        /// <returns></returns>
        internal unsafe static uint[] Ssid(int contract, string channel)
        {
            const char separator = EmitterConst.Separator;
            int length = channel.Length;
            int qSize = 1;
            int i = 0;
            var offset = 0;
            int minSkip = 100;

            fixed (char* pChar = channel)
            {
                // In this first pass we only need to calculate the length of the
                // query (how many separators we have).
                for (i = 0; i < length; ++i)
                {
                    if (*(pChar + i) == separator)
                    {
                        // Increment the size
                        ++qSize;

                        // skip one, we never should have two separators together
                        ++i;

                        // Calculate the minimum safe skip interval for the second
                        // pass (e.g.: tweet/fr/ would have minSkip = 2).
                        minSkip = Math.Min(i - offset, minSkip);
                        offset = i + 1;
                    }
                }

                // Prepare a query array
                var query = new uint[qSize];
                var section = 1;

                // Push the contract as a first element, and into unsigned format
                query[0] = (uint)contract;
                offset = 0;

                // Do a second pass to compute the hashes for every sub-section of
                // the query (e.g.: tweet/fr/ should result in int[2]
                for (i = 0; i < length; ++i)
                {
                    if (*(pChar + i) == separator)
                    {
                        // Get the hash until the current segment
                        query[section++] = GetHash(pChar + offset, i - offset);
                        offset = i + 1;

                        // skip one, we never should have two separators together
                        i += minSkip;
                    }
                }

                // We're done, return the query
                return query;
            }
        }

        /// <summary>
        /// Murmur32: seed.
        /// </summary>
        private const uint Seed = 37;

        /// <summary>
        /// Computes MurmurHash3 on this set of bytes and returns the calculated hash value.
        /// </summary>
        /// <param name="data">The data to compute the hash of.</param>
        /// <returns>A 32bit hash value.</returns>
        private static unsafe uint GetHash(char* pStr, int length)
        {
            const uint c1 = 0xcc9e2d51;
            const uint c2 = 0x1b873593;

            int curLength = length; // Current position in byte array
            uint h1 = Seed;
            uint k1 = 0;

            // body, eat stream a 32-bit int at a time
            while (curLength >= 4)
            {
                // Get four bytes from the input into an uint
                k1 = (uint)(*(pStr++)
                  | *(pStr++) << 8
                  | *(pStr++) << 16
                  | *(pStr++) << 24);

                // bitmagic hash
                k1 *= c1;
                k1 = Rotl32(k1, 15);
                k1 *= c2;

                h1 ^= k1;
                h1 = Rotl32(h1, 13);
                h1 = h1 * 5 + 0xe6546b64;
                curLength -= 4;
            }

            /* tail, the reminder bytes that did not make it to a full int */
            /* (this switch is slightly more ugly than the C++ implementation
             * because we can't fall through) */
            switch (curLength)
            {
                case 3:
                    k1 = (uint)(*(pStr++)
                      | *(pStr++) << 8
                      | *(pStr++) << 16);
                    k1 *= c1;
                    k1 = Rotl32(k1, 15);
                    k1 *= c2;
                    h1 ^= k1;
                    break;

                case 2:
                    k1 = (uint)(*(pStr++)
                      | *(pStr++) << 8);
                    k1 *= c1;
                    k1 = Rotl32(k1, 15);
                    k1 *= c2;
                    h1 ^= k1;
                    break;

                case 1:
                    k1 = (uint)(*(pStr++));
                    k1 *= c1;
                    k1 = Rotl32(k1, 15);
                    k1 *= c2;
                    h1 ^= k1;
                    break;
            };

            // finalization, magic chants to wrap it all up
            h1 ^= (uint)length;
            h1 ^= h1 >> 16;
            h1 *= 0x85ebca6b;
            h1 ^= h1 >> 13;
            h1 *= 0xc2b2ae35;
            h1 ^= h1 >> 16;

            return (uint)
                   ((byte)(h1) << 24
                 | ((byte)(h1 >> 8) << 16)
                 | ((byte)(h1 >> 16) << 8)
                 | ((byte)(h1 >> 24)));
        }

        /// <summary>
        /// Murmur32: rolling hash function.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static uint Rotl32(uint x, byte r)
        {
            return (x << r) | (x >> (32 - r));
        }
    }
}