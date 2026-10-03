import { HandlerCallback } from './handlerCallback.js';
import { AnyRouter } from '../router.js';
import { ServerManifest } from '../manifest.js';
export type RequestHandler<TRouter extends AnyRouter> = (cb: HandlerCallback<TRouter>) => Promise<Response>;
export declare function createRequestHandler<TRouter extends AnyRouter>({ createRouter, request, getRouterManifest, }: {
    createRouter: () => TRouter;
    request: Request;
    getRouterManifest?: () => ServerManifest | Promise<ServerManifest>;
}): RequestHandler<TRouter>;
