import { ReadableStream } from 'node:stream/web';
import { Readable } from 'node:stream';
import { AnyRouter } from '../router.cjs';
export type TransformStreamWithRouterOptions = {
    /** Timeout for serialization to complete after app render finishes (default: 60000ms) */
    timeoutMs?: number;
    /** Maximum lifetime of the stream transform (default: 120000ms). Safety net for cleanup. */
    lifetimeMs?: number;
    /**
     * Called exactly once when the stream is torn down due to abort/error/
     * cancel/timeout — NOT on natural successful completion. Use this to
     * abort a hidden producer upstream of any PassThrough you passed in
     * (e.g. React `renderToPipeableStream`'s `abort()`).
     * Errors thrown from this callback are swallowed.
     */
    onAbort?: (reason?: unknown) => void;
};
export declare function transformReadableStreamWithRouter(router: AnyRouter, routerStream: ReadableStream, opts?: TransformStreamWithRouterOptions): ReadableStream<Uint8Array<ArrayBufferLike>>;
export declare function transformPipeableStreamWithRouter(router: AnyRouter, routerStream: Readable, opts?: TransformStreamWithRouterOptions): Readable;
export declare function transformStreamWithRouter(router: AnyRouter, appStream: ReadableStream, opts?: TransformStreamWithRouterOptions): ReadableStream<Uint8Array<ArrayBufferLike>>;
