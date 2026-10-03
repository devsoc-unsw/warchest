import { mergeHeaders } from "./headers.js";
import { attachRouterServerSsrUtils, getNormalizedURL, getOrigin } from "./ssr-server.js";
import { normalizeSsrResponse } from "./handlerCallback.js";
import { createMemoryHistory } from "@tanstack/history";
//#region src/ssr/createRequestHandler.ts
function createRequestHandler({ createRouter, request, getRouterManifest }) {
	return async (cb) => {
		const router = createRouter();
		let responseOwnsCleanup = false;
		try {
			attachRouterServerSsrUtils({
				router,
				manifest: await getRouterManifest?.()
			});
			const { url } = getNormalizedURL(request.url, "http://localhost");
			const origin = getOrigin(request);
			const history = createMemoryHistory({ initialEntries: [url.href.replace(url.origin, "")] });
			router.update({
				history,
				origin: router.options.origin ?? origin
			});
			await router.load();
			await router.serverSsr?.dehydrate();
			const ssrResponse = normalizeSsrResponse(await cb({
				request,
				router,
				responseHeaders: getRequestHeaders({ router })
			}));
			responseOwnsCleanup = ssrResponse.serverSsrCleanup === "stream";
			return ssrResponse.response;
		} finally {
			if (!responseOwnsCleanup) router.serverSsr?.cleanup();
		}
	};
}
function getRequestHeaders(opts) {
	const matchHeaders = [];
	for (const match of opts.router.stores.matches.get()) matchHeaders.push(match.headers);
	const redirect = opts.router.stores.redirect.get();
	if (redirect) matchHeaders.push(redirect.headers);
	return mergeHeaders({ "Content-Type": "text/html; charset=UTF-8" }, ...matchHeaders);
}
//#endregion
export { createRequestHandler };

//# sourceMappingURL=createRequestHandler.js.map