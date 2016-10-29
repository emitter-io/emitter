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

using Emitter.Network.Http;
using Emitter.Network.Mesh;

namespace Emitter.Network
{
    /// <summary>
    /// Represents all the decoders that are built-in.
    /// </summary>
    public static class Decode
    {
        /// <summary>
        /// A processor for websockets using 'draft76' protocol.
        /// </summary>
        public static readonly Processor WebSocketDraft76 = WebSocketDraft76Decoder.Process;

        /// <summary>
        /// A processor for websockets using 'hybi13' protocol.
        /// </summary>
        public static readonly Processor WebSocketHybi13 = WebSocketHybi13Decoder.Process;

        /// <summary>
        /// A processor for Snappy Compression.
        /// </summary>
        public static readonly Processor Snappy = SnappyDecoder.Process;

        /// <summary>
        /// A processor for HTTP requets.
        /// </summary>
        public static readonly Processor Http = HttpDecoder.Process;

        /// <summary>
        /// A processor for Mesh protocol.
        /// </summary>
        public static readonly Processor Mesh = MeshDecoder.Process;

        /// <summary>
        /// A processor for MQTT Protocol.
        /// </summary>
        public static readonly Processor Mqtt = MqttDecoder.Process;

        /// <summary>
        /// A default custom processor
        /// </summary>
        public static readonly Processor Null = NullProcessor.Process;
    }

    /// <summary>
    /// Represents all the handlers that are built-in.
    /// </summary>
    public static class Handle
    {
        /// <summary>
        /// A processor for HTTP requests.
        /// </summary>
        public static readonly Processor Http = HttpHandler.Process;

        /// <summary>
        /// A processor for websocket upgrade requests.
        /// </summary>
        public static readonly Processor WebSocketUpgrade = WebSocketUpgradeHandler.Process;

        /// <summary>
        /// A default custom processor
        /// </summary>
        public static readonly Processor Null = NullProcessor.Process;
    }

    /// <summary>
    /// Represents all the encoders that are built-in.
    /// </summary>
    public static class Encode
    {
        /// <summary>
        /// A processor for websockets using 'draft76' protocol.
        /// </summary>
        public static readonly Processor WebSocketDraft76 = WebSocketDraft76Encoder.Process;

        /// <summary>
        /// A processor for websockets using 'hybi13' protocol.
        /// </summary>
        public static readonly Processor WebSocketHybi13 = WebSocketHybi13Encoder.Process;

        /// <summary>
        /// A processor for string packets.
        /// </summary>
        public static readonly Processor String = StringEncoder.Process;

        /// <summary>
        /// A processor for byte packets.
        /// </summary>
        public static readonly Processor Byte = ByteEncoder.Process;

        /// <summary>
        /// A processor for Mesh Protocol.
        /// </summary>
        public static readonly Processor Mesh = MeshEncoder.Process;

        /// <summary>
        /// A processor for HTTP responses.
        /// </summary>
        public static readonly Processor Http = HttpEncoder.Process;

        /// <summary>
        /// A processor for MQTT Protocol.
        /// </summary>
        public static readonly Processor Mqtt = MqttEncoder.Process;

        /// <summary>
        /// A default custom processor
        /// </summary>
        public static readonly Processor Null = NullProcessor.Process;
    }
}