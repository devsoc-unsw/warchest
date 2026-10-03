import { HandlerCallback } from './handlerCallback.cjs';
import { AnyRouter } from '../router.cjs';
import { ServerManifest } from '../manifest.cjs';
export type RequestHandler<TRouter extends AnyRouter> = (cb: HandlerCallback<TRouter>) => Promise<Response>;
export declare function createRequestHandler<TRouter extends AnyRouter>({ createRouter, request, getRouterManifest, }: {
    createRouter: () => TRouter;
    request: Request;
    getRouterManifest?: () => ServerManifest | Promise<ServerManifest>;
}): RequestHandler<TRouter>;
