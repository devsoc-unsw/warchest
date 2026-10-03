import { PluginInfo, SerovalNode } from 'seroval';
/**
 * Hint for RawStream encoding strategy during SSR serialization.
 * - 'binary': Always use base64 encoding (best for binary data like files, images)
 * - 'text': Try UTF-8 first, fallback to base64 (best for text-heavy data like RSC payloads)
 */
export type RawStreamHint = 'binary' | 'text';
/**
 * Options for RawStream configuration.
 */
export interface RawStreamOptions {
    /**
     * Encoding hint for SSR serialization.
     * - 'binary' (default): Always use base64 encoding
     * - 'text': Try UTF-8 first, fallback to base64 for invalid UTF-8 chunks
     */
    hint?: RawStreamHint;
}
/**
 * Marker class for ReadableStream<Uint8Array> that should be serialized
 * with base64 encoding (SSR) or binary framing (server functions).
 *
 * Wrap your binary streams with this to get efficient serialization:
 * ```ts
 * // For binary data (files, images, etc.)
 * return { data: new RawStream(file.stream()) }
 *
 * // For text-heavy data (RSC payloads, etc.)
 * return { data: new RawStream(rscStream, { hint: 'text' }) }
 * ```
 */
export declare class RawStream {
    readonly stream: ReadableStream<Uint8Array>;
    readonly hint: RawStreamHint;
    constructor(stream: ReadableStream<Uint8Array>, options?: RawStreamOptions);
}
/**
 * Callback type for RPC plugin to register raw streams with multiplexer
 */
export type OnRawStreamCallback = (streamId: number, stream: ReadableStream<Uint8Array>) => void;
export interface RawStreamSSRNode extends PluginInfo {
    hint: SerovalNode;
    factory: SerovalNode;
    stream: SerovalNode;
}
export interface RawStreamRPCNode extends PluginInfo {
    streamId: SerovalNode;
}
/**
 * SSR Plugin - uses base64 or UTF-8+base64 encoding for chunks, delegates to seroval's stream mechanism.
 * Used during SSR when serializing to JavaScript code for HTML injection.
 *
 * Supports two modes based on RawStream hint:
 * - 'binary': Always base64 encode (default)
 * - 'text': Try UTF-8 first, fallback to base64 for invalid UTF-8
 */
export declare const RawStreamSSRPlugin: import('seroval').Plugin<RawStream, RawStreamSSRNode>;
/**
 * Creates an RPC plugin instance that registers raw streams with a multiplexer.
 * Used for server function responses where we want binary framing.
 * Note: RPC always uses binary framing regardless of hint.
 *
 * @param onRawStream Callback invoked when a RawStream is encountered during serialization
 */
export declare function createRawStreamRPCPlugin(onRawStream: OnRawStreamCallback): import('seroval').Plugin<RawStream, RawStreamRPCNode>;
/**
 * Creates a deserialize-only plugin for client-side stream reconstruction.
 * Used in serverFnFetcher to wire up streams from frame decoder.
 *
 * @param getOrCreateStream Function to get/create a stream by ID from frame decoder
 */
export declare function createRawStreamDeserializePlugin(getOrCreateStream: (id: number) => ReadableStream<Uint8Array>): import('seroval').Plugin<any, RawStreamRPCNode>;
