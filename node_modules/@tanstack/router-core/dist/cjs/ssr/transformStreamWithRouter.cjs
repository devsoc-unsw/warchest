require("./constants.cjs");
let node_stream_web = require("node:stream/web");
let node_stream = require("node:stream");
//#region src/ssr/transformStreamWithRouter.ts
function transformReadableStreamWithRouter(router, routerStream, opts) {
	return transformStreamWithRouter(router, routerStream, opts);
}
function transformPipeableStreamWithRouter(router, routerStream, opts) {
	return node_stream.Readable.fromWeb(transformStreamWithRouter(router, node_stream.Readable.toWeb(routerStream), opts));
}
const MIN_CLOSING_TAG_LENGTH = 4;
const DEFAULT_SERIALIZATION_TIMEOUT_MS = 6e4;
const DEFAULT_LIFETIME_TIMEOUT_MS = DEFAULT_SERIALIZATION_TIMEOUT_MS * 2;
const MAX_LEFTOVER_CHARS = 2048;
const MAX_TAIL_CHARS = 64 * 1024;
const MAX_ROUTER_HTML_CHARS = 16 * 1024 * 1024;
const MAX_PENDING_WRITE_CHARS = 16 * 1024 * 1024;
const MergeState = {
	ReadingBody: 0,
	HoldingTail: 1,
	AppDone: 2,
	Draining: 3,
	Done: 4
};
const textEncoder = new TextEncoder();
const noop = () => {};
const resolvedPromise = Promise.resolve();
function findHtmlBoundary(str) {
	let lastClosingTagEnd = -1;
	let searchFrom = str.length - MIN_CLOSING_TAG_LENGTH;
	while (searchFrom >= 0) {
		const openSlash = str.lastIndexOf("</", searchFrom);
		if (openSlash === -1) break;
		if ((str.charCodeAt(openSlash + 2) | 32) === 98 && (str.charCodeAt(openSlash + 3) | 32) === 111 && (str.charCodeAt(openSlash + 4) | 32) === 100 && (str.charCodeAt(openSlash + 5) | 32) === 121 && str.charCodeAt(openSlash + 6) === 62) return -openSlash - 2;
		if (lastClosingTagEnd === -1) {
			let i = openSlash + 2;
			const startCode = str.charCodeAt(i);
			if (startCode >= 97 && startCode <= 122 || startCode >= 65 && startCode <= 90) {
				i++;
				while (i < str.length) {
					const code = str.charCodeAt(i);
					if (code >= 97 && code <= 122 || code >= 65 && code <= 90 || code >= 48 && code <= 57 || code === 95 || code === 58 || code === 46 || code === 45) i++;
					else break;
				}
				if (str.charCodeAt(i) === 62) lastClosingTagEnd = i + 1;
			}
		}
		searchFrom = openSlash - 1;
	}
	return lastClosingTagEnd;
}
function safeReleaseReader(reader) {
	try {
		reader.releaseLock();
		return true;
	} catch {
		return false;
	}
}
/**
* Cancel a reader without producing an unhandled rejection. `reader.cancel()`
* can reject (e.g. when the underlying source's cancel() throws), and
* downstream cancel() should still wait for upstream teardown when possible.
*/
function safeCancelReader(reader, reason) {
	let cancelPromise;
	try {
		cancelPromise = reader.cancel(reason);
	} catch {}
	if (!safeReleaseReader(reader) && cancelPromise) return cancelPromise.then(noop, noop).then(() => {
		safeReleaseReader(reader);
	});
	return cancelPromise ? cancelPromise.then(noop, noop) : resolvedPromise;
}
function createReaderState(appStream) {
	const reader = appStream.getReader();
	let released = false;
	return {
		reader,
		cancel: (reason) => {
			if (released) return resolvedPromise;
			released = true;
			return safeCancelReader(reader, reason);
		},
		release: () => {
			if (released) return;
			released = true;
			safeReleaseReader(reader);
		}
	};
}
function createAbortNotifier(opts) {
	let abortNotified = false;
	return (reason) => {
		if (abortNotified) return;
		abortNotified = true;
		try {
			opts?.onAbort?.(reason);
		} catch {}
	};
}
function transformStreamWithRouter(router, appStream, opts) {
	const serverSsr = router.serverSsr;
	if (!serverSsr) throw new Error("Invariant failed: router.serverSsr is required");
	if (serverSsr.reserveStreamFastPath()) return makeFastPathStream(appStream, opts, serverSsr);
	return makeMainStream(serverSsr, appStream, opts);
}
function makeFastPathStream(appStream, opts, serverSsr) {
	let cleanedUp = false;
	let controller;
	let state = MergeState.ReadingBody;
	let lifetimeTimeoutHandle;
	let stopListeningToInjectedHtml;
	const readerState = createReaderState(appStream);
	const notifyAbort = createAbortNotifier(opts);
	const isDone = () => state === MergeState.Done;
	let renderFinished = false;
	const finishSsrRendering = () => {
		if (!serverSsr || renderFinished) return true;
		renderFinished = true;
		try {
			serverSsr.setRenderFinished();
			return true;
		} catch (error) {
			safeError(error);
			cleanup(error);
			return false;
		}
	};
	const cleanup = (reason, cancelReader = true) => {
		if (cleanedUp) return resolvedPromise;
		cleanedUp = true;
		if (lifetimeTimeoutHandle !== void 0) {
			clearTimeout(lifetimeTimeoutHandle);
			lifetimeTimeoutHandle = void 0;
		}
		try {
			stopListeningToInjectedHtml?.();
		} catch {}
		stopListeningToInjectedHtml = void 0;
		if (cancelReader) notifyAbort(reason);
		const readerDone = cancelReader ? readerState.cancel(reason) : (readerState.release(), resolvedPromise);
		if (serverSsr) try {
			serverSsr.cleanup();
		} catch (error) {
			console.error("Error in SSR cleanup:", error);
		}
		return readerDone;
	};
	const safeClose = () => {
		if (isDone()) return;
		state = MergeState.Done;
		try {
			controller?.close();
		} catch {}
	};
	const safeError = (error) => {
		if (isDone()) return;
		state = MergeState.Done;
		try {
			controller?.error(error);
		} catch {}
	};
	if (serverSsr) stopListeningToInjectedHtml = serverSsr.onInjectedHtml(() => {
		const err = /* @__PURE__ */ new Error("SSR router HTML injected during fast path");
		safeError(err);
		cleanup(err);
	});
	const lifetimeMs = opts?.lifetimeMs ?? DEFAULT_LIFETIME_TIMEOUT_MS;
	lifetimeTimeoutHandle = setTimeout(() => {
		if (!cleanedUp && !isDone()) {
			const err = /* @__PURE__ */ new Error("Stream lifetime exceeded");
			console.warn(`SSR stream transform exceeded maximum lifetime (${lifetimeMs}ms), forcing cleanup`);
			safeError(err);
			cleanup(err);
		}
	}, lifetimeMs);
	return new node_stream_web.ReadableStream({
		start(c) {
			controller = c;
		},
		async pull(c) {
			if (cleanedUp || isDone()) return;
			try {
				const { done, value } = await readerState.reader.read();
				if (!done) {
					if (!cleanedUp && !isDone()) c.enqueue(value);
					return;
				}
				if (cleanedUp || isDone()) return;
				if (!finishSsrRendering()) return;
				safeClose();
				return cleanup(void 0, false);
			} catch (error) {
				if (cleanedUp) return;
				console.error("Error reading appStream:", error);
				if (state < MergeState.AppDone) try {
					serverSsr?.setRenderFinished();
				} catch {}
				safeError(error);
				return cleanup(error);
			} finally {
				if (cleanedUp || isDone()) readerState.release();
			}
		},
		cancel(reason) {
			state = MergeState.Done;
			return cleanup(reason);
		}
	});
}
function makeMainStream(serverSsr, appStream, opts) {
	let stopListeningToInjectedHtml;
	let stopListeningToSerializationFinished;
	let serializationTimeoutHandle;
	let lifetimeTimeoutHandle;
	let cleanedUp = false;
	let controller;
	let closeWhenDrained = false;
	let state = MergeState.ReadingBody;
	const readerState = createReaderState(appStream);
	const notifyAbort = createAbortNotifier(opts);
	const pendingWrites = [];
	let pendingWriteHead = 0;
	let pendingWriteChars = 0;
	function clearPending() {
		pendingWrites.length = 0;
		pendingWriteHead = 0;
		pendingWriteChars = 0;
	}
	let drainResolve = null;
	const waitForDrain = () => new Promise((r) => {
		drainResolve = r;
	});
	const signalDrain = () => {
		if (drainResolve) {
			const r = drainResolve;
			drainResolve = null;
			r();
		}
	};
	const isDone = () => state === MergeState.Done;
	function drainPending() {
		if (!controller || isDone()) return;
		while (pendingWriteHead < pendingWrites.length) {
			const ds = controller.desiredSize;
			if (ds !== null && ds <= 0) return;
			const next = pendingWrites[pendingWriteHead];
			pendingWrites[pendingWriteHead] = "";
			pendingWriteHead++;
			pendingWriteChars -= next.length;
			try {
				controller.enqueue(textEncoder.encode(next));
			} catch (error) {
				safeError(error);
				cleanup(error);
				return;
			}
		}
		if (pendingWriteHead >= pendingWrites.length) {
			pendingWrites.length = 0;
			pendingWriteHead = 0;
		}
		if (closeWhenDrained && pendingWriteHead >= pendingWrites.length) {
			closeWhenDrained = false;
			safeClose();
			cleanup(void 0, false);
		}
	}
	/**
	* Enqueue a string chunk through the backpressure queue. Stored as a
	* string and encoded only when the downstream actually accepts the chunk
	* — keeps native-memory pressure inside the controller's queue (which
	* honors desiredSize) rather than ours.
	*/
	function writeChunk(chunk) {
		if (cleanedUp || isDone()) return;
		if (!chunk.length) return;
		if (pendingWriteChars + chunk.length > MAX_PENDING_WRITE_CHARS) {
			const err = /* @__PURE__ */ new Error("SSR stream pending output exceeded maximum buffer");
			safeError(err);
			cleanup(err);
			return;
		}
		pendingWrites.push(chunk);
		pendingWriteChars += chunk.length;
		drainPending();
	}
	function safeClose() {
		if (isDone()) return;
		state = MergeState.Done;
		try {
			controller?.close();
		} catch {}
	}
	function safeError(error) {
		if (isDone()) return;
		state = MergeState.Done;
		try {
			controller?.error(error);
		} catch {}
	}
	/**
	* Cleanup with guards; must be idempotent.
	*/
	function cleanup(reason, cancelReader = true) {
		if (cleanedUp) return resolvedPromise;
		cleanedUp = true;
		try {
			stopListeningToInjectedHtml?.();
			stopListeningToSerializationFinished?.();
		} catch {}
		stopListeningToInjectedHtml = void 0;
		stopListeningToSerializationFinished = void 0;
		if (serializationTimeoutHandle !== void 0) {
			clearTimeout(serializationTimeoutHandle);
			serializationTimeoutHandle = void 0;
		}
		if (lifetimeTimeoutHandle !== void 0) {
			clearTimeout(lifetimeTimeoutHandle);
			lifetimeTimeoutHandle = void 0;
		}
		clearPendingRouterHtml();
		leftover = "";
		pendingTail = "";
		clearPending();
		if (cancelReader) notifyAbort(reason);
		const readerDone = cancelReader ? readerState.cancel(reason) : (readerState.release(), resolvedPromise);
		signalDrain();
		try {
			serverSsr.cleanup();
		} catch (error) {
			console.error("Error in SSR cleanup:", error);
		}
		return readerDone;
	}
	const textDecoder = new TextDecoder();
	const pendingRouterHtml = [];
	let pendingRouterHtmlChars = 0;
	let leftover = "";
	let pendingTail = "";
	let streamBarrierLifted = false;
	let streamBarrierMarkerSeen = false;
	let serializationFinished = false;
	function noteBarrierMarker(chunk) {
		if (streamBarrierMarkerSeen) return;
		if (chunk.includes("$tsr-stream-barrier")) streamBarrierMarkerSeen = true;
	}
	function liftBarrierAfterBoundary() {
		if (streamBarrierLifted) return;
		if (!streamBarrierMarkerSeen) return;
		streamBarrierLifted = true;
		serverSsr.liftScriptBarrier();
	}
	const stream = new node_stream_web.ReadableStream({
		start(c) {
			controller = c;
			drainPending();
		},
		pull() {
			drainPending();
			signalDrain();
		},
		cancel(reason) {
			state = MergeState.Done;
			return cleanup(reason);
		}
	});
	function drainRouterHtml() {
		if (cleanedUp || isDone()) return;
		let html;
		try {
			html = serverSsr.takeBufferedHtml();
		} catch (error) {
			safeError(error);
			cleanup(error);
			return;
		}
		if (!html) return;
		if (state >= MergeState.Draining) {
			const err = /* @__PURE__ */ new Error("SSR router HTML injected after stream finalization");
			safeError(err);
			cleanup(err);
			return;
		}
		if (state === MergeState.HoldingTail) {
			flushPendingRouterHtml();
			writeChunk(html);
		} else {
			if (pendingRouterHtmlChars + html.length > MAX_ROUTER_HTML_CHARS) {
				const err = /* @__PURE__ */ new Error("SSR router HTML exceeded maximum buffer");
				safeError(err);
				cleanup(err);
				return;
			}
			pendingRouterHtml.push(html);
			pendingRouterHtmlChars += html.length;
		}
	}
	function flushPendingRouterHtml() {
		if (!pendingRouterHtml.length) return;
		for (const html of pendingRouterHtml) writeChunk(html);
		clearPendingRouterHtml();
	}
	function clearPendingRouterHtml() {
		pendingRouterHtml.length = 0;
		pendingRouterHtmlChars = 0;
	}
	function appendTail(chunk) {
		pendingTail += chunk;
		if (pendingTail.length > MAX_TAIL_CHARS) throw new Error("SSR stream tail exceeded maximum buffer");
	}
	function waitForBackpressure() {
		return !!(controller && controller.desiredSize !== null && controller.desiredSize <= 0);
	}
	function startSerializationTimeout() {
		if (cleanedUp || isDone()) return;
		if (serializationTimeoutHandle !== void 0) return;
		const timeoutMs = opts?.timeoutMs ?? DEFAULT_SERIALIZATION_TIMEOUT_MS;
		serializationTimeoutHandle = setTimeout(() => {
			if (!cleanedUp && !isDone()) {
				const err = /* @__PURE__ */ new Error("Serialization timeout after app render finished");
				console.error("Serialization timeout after app render finished");
				safeError(err);
				cleanup(err);
			}
		}, timeoutMs);
	}
	/**
	* Finish only when app done and serialization complete. Queues final
	* output and requests close-when-drained so we don't close ahead of
	* pending writes still waiting on downstream capacity.
	*/
	function tryFinish() {
		if (state !== MergeState.AppDone || !serializationFinished) return;
		if (cleanedUp || isDone()) return;
		if (serializationTimeoutHandle !== void 0) {
			clearTimeout(serializationTimeoutHandle);
			serializationTimeoutHandle = void 0;
		}
		drainRouterHtml();
		if (cleanedUp || isDone()) return;
		const decoderRemainder = textDecoder.decode();
		if (leftover) writeChunk(leftover);
		if (cleanedUp || isDone()) return;
		if (decoderRemainder) writeChunk(decoderRemainder);
		if (cleanedUp || isDone()) return;
		flushPendingRouterHtml();
		if (cleanedUp || isDone()) return;
		if (pendingTail) writeChunk(pendingTail);
		if (cleanedUp || isDone()) return;
		leftover = "";
		pendingTail = "";
		state = MergeState.Draining;
		closeWhenDrained = true;
		drainPending();
	}
	function finishAppRendering() {
		if (state >= MergeState.AppDone) return;
		state = MergeState.AppDone;
		try {
			serverSsr.setRenderFinished();
		} catch (error) {
			safeError(error);
			cleanup(error);
			return;
		}
		drainRouterHtml();
		if (cleanedUp || isDone()) return;
		serializationFinished = serializationFinished || serverSsr.isSerializationFinished();
		if (serializationFinished) tryFinish();
		else startSerializationTimeout();
	}
	const timeoutMs = opts?.timeoutMs ?? DEFAULT_SERIALIZATION_TIMEOUT_MS;
	const lifetimeMs = opts?.lifetimeMs ?? timeoutMs * 2;
	lifetimeTimeoutHandle = setTimeout(() => {
		if (!cleanedUp && !isDone()) {
			const err = /* @__PURE__ */ new Error("Stream lifetime exceeded");
			console.warn(`SSR stream transform exceeded maximum lifetime (${lifetimeMs}ms), forcing cleanup`);
			safeError(err);
			cleanup(err);
		}
	}, lifetimeMs);
	stopListeningToInjectedHtml = serverSsr.onInjectedHtml(() => {
		drainRouterHtml();
	});
	stopListeningToSerializationFinished = serverSsr.onSerializationFinished(() => {
		serializationFinished = true;
		drainRouterHtml();
		tryFinish();
	});
	drainRouterHtml();
	if (cleanedUp || isDone()) return stream;
	serializationFinished = serializationFinished || serverSsr.isSerializationFinished();
	if (serializationFinished) {
		drainRouterHtml();
		if (cleanedUp || isDone()) return stream;
	}
	(async () => {
		try {
			while (true) {
				if (waitForBackpressure()) {
					await waitForDrain();
					if (cleanedUp || isDone()) return;
				}
				const { done, value } = await readerState.reader.read();
				if (done) break;
				if (cleanedUp || isDone()) return;
				const text = typeof value === "string" ? value : textDecoder.decode(value, { stream: true });
				const chunkString = leftover ? leftover + text : text;
				if (state >= MergeState.HoldingTail) {
					appendTail(chunkString);
					leftover = "";
					continue;
				}
				const boundary = findHtmlBoundary(chunkString);
				if (boundary < -1) {
					const bodyEndIndex = -boundary - 2;
					state = MergeState.HoldingTail;
					appendTail(chunkString.slice(bodyEndIndex));
					const bodyChunk = chunkString.slice(0, bodyEndIndex);
					writeChunk(bodyChunk);
					if (cleanedUp || isDone()) return;
					noteBarrierMarker(bodyChunk);
					liftBarrierAfterBoundary();
					if (cleanedUp || isDone()) return;
					flushPendingRouterHtml();
					leftover = "";
					continue;
				}
				const lastClosingTagEnd = boundary;
				if (lastClosingTagEnd > 0) {
					const safeChunk = chunkString.slice(0, lastClosingTagEnd);
					writeChunk(safeChunk);
					if (cleanedUp || isDone()) return;
					noteBarrierMarker(safeChunk);
					liftBarrierAfterBoundary();
					if (cleanedUp || isDone()) return;
					flushPendingRouterHtml();
					leftover = chunkString.slice(lastClosingTagEnd);
					if (leftover.length > MAX_LEFTOVER_CHARS) {
						noteBarrierMarker(leftover);
						writeChunk(leftover.slice(0, leftover.length - MAX_LEFTOVER_CHARS));
						leftover = leftover.slice(-2048);
					}
				} else {
					const combined = chunkString;
					if (combined.length > MAX_LEFTOVER_CHARS) {
						noteBarrierMarker(combined);
						const flushUpto = combined.length - MAX_LEFTOVER_CHARS;
						writeChunk(combined.slice(0, flushUpto));
						leftover = combined.slice(flushUpto);
					} else leftover = combined;
				}
			}
			if (cleanedUp || isDone()) return;
			finishAppRendering();
		} catch (error) {
			if (cleanedUp) return;
			console.error("Error reading appStream:", error);
			if (state < MergeState.AppDone) try {
				serverSsr.setRenderFinished();
			} catch {}
			safeError(error);
			cleanup(error);
		} finally {
			readerState.release();
		}
	})().catch((error) => {
		if (cleanedUp) return;
		console.error("Error in stream transform:", error);
		safeError(error);
		cleanup(error);
	});
	return stream;
}
//#endregion
exports.transformPipeableStreamWithRouter = transformPipeableStreamWithRouter;
exports.transformReadableStreamWithRouter = transformReadableStreamWithRouter;
exports.transformStreamWithRouter = transformStreamWithRouter;

//# sourceMappingURL=transformStreamWithRouter.cjs.map