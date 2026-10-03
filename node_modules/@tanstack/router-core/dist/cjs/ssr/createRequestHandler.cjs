const require_headers = require("./headers.cjs");
const require_ssr_server = require("./ssr-server.cjs");
const require_handlerCallback = require("./handlerCallback.cjs");
let _tanstack_history = require("@tanstack/history");
//#region src/ssr/createRequestHandler.ts
function createRequestHandler({ createRouter, request, getRouterManifest }) {
	return async (cb) => {
		const router = createRouter();
		let responseOwnsCleanup = false;
		try {
			require_ssr_server.attachRouterServerSsrUtils({
				router,
				manifest: await getRouterManifest?.()
			});
			const { url } = require_ssr_server.getNormalizedURL(request.url, "http://localhost");
			const origin = require_ssr_server.getOrigin(request);
			const history = (0, _tanstack_history.createMemoryHistory)({ initialEntries: [url.href.replace(url.origin, "")] });
			router.update({
				history,
				origin: router.options.origin ?? origin
			});
			await router.load();
			await router.serverSsr?.dehydrate();
			const ssrResponse = require_handlerCallback.normalizeSsrResponse(await cb({
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
	return require_headers.mergeHeaders({ "Content-Type": "text/html; charset=UTF-8" }, ...matchHeaders);
}
//#endregion
exports.createRequestHandler = createRequestHandler;

//# sourceMappingURL=createRequestHandler.cjs.map