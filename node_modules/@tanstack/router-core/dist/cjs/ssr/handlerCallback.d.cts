import { AnyRouter } from '../router.cjs';
export type SsrResponse = {
    response: Response;
    serverSsrCleanup: 'none';
} | {
    response: Response;
    serverSsrCleanup: 'stream';
    dispose: (reason?: unknown) => Promise<void>;
};
export type HandlerCallbackResult = Response | SsrResponse;
export declare function isSsrResponse(value: unknown): value is SsrResponse;
export declare function normalizeSsrResponse(result: HandlerCallbackResult): SsrResponse;
export declare function createSsrStreamResponse<TRouter extends AnyRouter>(router: TRouter, response: Response): SsrResponse;
export declare function replaceSsrResponse(result: HandlerCallbackResult, response: Response, reason?: unknown): Promise<SsrResponse>;
export declare function stripSsrResponseBody(result: HandlerCallbackResult, reason?: unknown): Promise<SsrResponse>;
export interface HandlerCallback<TRouter extends AnyRouter> {
    (ctx: {
        request: Request;
        router: TRouter;
        responseHeaders: Headers;
    }): HandlerCallbackResult | Promise<HandlerCallbackResult>;
}
export declare function defineHandlerCallback<TRouter extends AnyRouter>(handler: HandlerCallback<TRouter>): HandlerCallback<TRouter>;
