const require_runtime = require("../_virtual/_rolldown/runtime.cjs");
let _tanstack_router_core_ssr_server = require("@tanstack/router-core/ssr/server");
let react_dom_server = require("react-dom/server");
react_dom_server = require_runtime.__toESM(react_dom_server, 1);
let node_stream = require("node:stream");
let isbot = require("isbot");
//#region src/ssr/renderRouterToStream.tsx
var noop = () => {};
async function waitForReadyOrAbort(ready, signal) {
	let cleanup = noop;
	try {
		await Promise.race([ready, new Promise((resolve) => {
			const onAbort = () => resolve();
			cleanup = () => signal.removeEventListener("abort", onAbort);
			signal.addEventListener("abort", onAbort, { once: true });
			if (signal.aborted) resolve();
		})]);
	} finally {
		cleanup();
	}
}
var isAbortError = (request, error) => request.signal.aborted && error === request.signal.reason || error instanceof Error && error.name === "AbortError";
var renderRouterToStream = async ({ request, router, responseHeaders, children }) => {
	if (typeof react_dom_server.default.renderToReadableStream === "function") {
		const stream = await react_dom_server.default.renderToReadableStream(children, {
			signal: request.signal,
			nonce: router.options.ssr?.nonce,
			progressiveChunkSize: Number.POSITIVE_INFINITY,
			onError: (error, info) => {
				if (!isAbortError(request, error)) console.error("Error in renderToReadableStream:", error, info);
			}
		});
		if ((0, isbot.isbot)(request.headers.get("User-Agent"))) await waitForReadyOrAbort(stream.allReady, request.signal);
		const responseStream = (0, _tanstack_router_core_ssr_server.transformReadableStreamWithRouter)(router, stream, { onAbort: () => stream.cancel().catch(() => {}) });
		return (0, _tanstack_router_core_ssr_server.createSsrStreamResponse)(router, new Response(responseStream, {
			status: router.stores.statusCode.get(),
			headers: responseHeaders
		}));
	}
	if (typeof react_dom_server.default.renderToPipeableStream === "function") {
		const reactAppPassthrough = new node_stream.PassThrough();
		let pipeable;
		let responseAttached = false;
		let aborted = false;
		let endedBeforeAttach = false;
		let pendingAbortReason;
		const toError = (reason) => reason instanceof Error ? reason : new Error(String(reason ?? "SSR aborted"));
		const destroyError = (reason) => reason === void 0 ? void 0 : toError(reason);
		const pendingDestroyError = () => pendingAbortReason === void 0 ? toError(pendingAbortReason) : destroyError(pendingAbortReason);
		const finishPassThrough = (reason, opts) => {
			if (reactAppPassthrough.destroyed) return;
			if (responseAttached) reactAppPassthrough.destroy(opts?.defaultError ? toError(reason) : destroyError(reason));
			else endedBeforeAttach = true;
		};
		const abortPipeable = (reason, opts) => {
			if (aborted) return;
			aborted = true;
			pendingAbortReason = reason;
			const err = toError(reason);
			try {
				pipeable?.abort(err);
			} catch {}
			finishPassThrough(reason, opts);
		};
		if (request.signal.aborted) abortPipeable(request.signal.reason);
		else {
			const onRequestAbort = () => abortPipeable(request.signal.reason);
			request.signal.addEventListener("abort", onRequestAbort, { once: true });
			router.serverSsr?.onCleanup(() => {
				request.signal.removeEventListener("abort", onRequestAbort);
			});
		}
		try {
			pipeable = react_dom_server.default.renderToPipeableStream(children, {
				nonce: router.options.ssr?.nonce,
				progressiveChunkSize: Number.POSITIVE_INFINITY,
				...(0, isbot.isbot)(request.headers.get("User-Agent")) ? { onAllReady() {
					pipeable.pipe(reactAppPassthrough);
				} } : { onShellReady() {
					pipeable.pipe(reactAppPassthrough);
				} },
				onError: (error, info) => {
					if (!isAbortError(request, error)) console.error("Error in renderToPipeableStream:", error, info);
					abortPipeable(error, { defaultError: true });
				}
			});
		} catch (e) {
			console.error("Error in renderToPipeableStream:", e);
			router.serverSsr?.cleanup();
			throw e;
		}
		const responseStream = (0, _tanstack_router_core_ssr_server.transformPipeableStreamWithRouter)(router, reactAppPassthrough, { onAbort: abortPipeable });
		responseAttached = true;
		if (endedBeforeAttach) reactAppPassthrough.destroy(pendingDestroyError());
		if (aborted && pipeable) try {
			pipeable.abort(toError(pendingAbortReason));
		} catch {}
		return (0, _tanstack_router_core_ssr_server.createSsrStreamResponse)(router, new Response(responseStream, {
			status: router.stores.statusCode.get(),
			headers: responseHeaders
		}));
	}
	throw new Error("No renderToReadableStream or renderToPipeableStream found in react-dom/server. Ensure you are using a version of react-dom that supports streaming.");
};
//#endregion
exports.renderRouterToStream = renderRouterToStream;

//# sourceMappingURL=renderRouterToStream.cjs.map