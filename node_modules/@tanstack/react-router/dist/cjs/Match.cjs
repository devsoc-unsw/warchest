"use client";
const require_runtime = require("./_virtual/_rolldown/runtime.cjs");
const require_utils = require("./utils.cjs");
const require_CatchBoundary = require("./CatchBoundary.cjs");
const require_ClientOnly = require("./ClientOnly.cjs");
const require_matchContext = require("./matchContext.cjs");
const require_useRouter = require("./useRouter.cjs");
const require_not_found = require("./not-found.cjs");
const require_SafeFragment = require("./SafeFragment.cjs");
const require_renderRouteNotFound = require("./renderRouteNotFound.cjs");
const require_scroll_restoration = require("./scroll-restoration.cjs");
let _tanstack_router_core = require("@tanstack/router-core");
let react = require("react");
react = require_runtime.__toESM(react, 1);
let react_jsx_runtime = require("react/jsx-runtime");
let _tanstack_react_store = require("@tanstack/react-store");
let _tanstack_router_core_isServer = require("@tanstack/router-core/isServer");
//#region src/Match.tsx
var matchViewFieldsEqual = (a, b) => a.routeId === b.routeId && a._displayPending === b._displayPending;
var outletMatchSelectionEqual = (a, b) => a[0] === b[0] && a[1] === b[1];
var Match = react.memo(function MatchImpl({ matchId }) {
	const router = require_useRouter.useRouter();
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) {
		const match = router.stores.matchStores.get(matchId)?.get();
		if (!match) {
			if (process.env.NODE_ENV !== "production") throw new Error(`Invariant failed: Could not find match for matchId "${matchId}". Please file an issue!`);
			(0, _tanstack_router_core.invariant)();
		}
		const routeId = match.routeId;
		const parentRouteId = router.routesById[routeId].parentRoute?.id;
		return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(MatchView, {
			router,
			matchId,
			resetKey: router.stores.loadedAt.get(),
			matchState: {
				routeId,
				ssr: match.ssr,
				_displayPending: match._displayPending,
				parentRouteId
			}
		});
	}
	const matchStore = router.stores.matchStores.get(matchId);
	if (!matchStore) {
		if (process.env.NODE_ENV !== "production") throw new Error(`Invariant failed: Could not find match for matchId "${matchId}". Please file an issue!`);
		(0, _tanstack_router_core.invariant)();
	}
	const resetKey = (0, _tanstack_react_store.useStore)(router.stores.loadedAt, (loadedAt) => loadedAt);
	const match = (0, _tanstack_react_store.useStore)(matchStore, (value) => value, matchViewFieldsEqual);
	return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(MatchView, {
		router,
		matchId,
		resetKey,
		matchState: react.useMemo(() => {
			const routeId = match.routeId;
			const parentRouteId = router.routesById[routeId].parentRoute?.id;
			return {
				routeId,
				ssr: match.ssr,
				_displayPending: match._displayPending,
				parentRouteId
			};
		}, [
			match._displayPending,
			match.routeId,
			match.ssr,
			router.routesById
		])
	});
});
function MatchView({ router, matchId, resetKey, matchState }) {
	const route = router.routesById[matchState.routeId];
	const PendingComponent = route.options.pendingComponent ?? router.options.defaultPendingComponent;
	const pendingElement = PendingComponent ? /* @__PURE__ */ (0, react_jsx_runtime.jsx)(PendingComponent, {}) : null;
	const routeErrorComponent = route.options.errorComponent ?? router.options.defaultErrorComponent;
	const routeOnCatch = route.options.onCatch ?? router.options.defaultOnCatch;
	const routeNotFoundComponent = route.isRoot ? route.options.notFoundComponent ?? router.options.notFoundRoute?.options.component : route.options.notFoundComponent;
	const resolvedNoSsr = matchState.ssr === false || matchState.ssr === "data-only";
	const ResolvedSuspenseBoundary = (!route.isRoot || route.options.wrapInSuspense || resolvedNoSsr) && (route.options.wrapInSuspense ?? PendingComponent ?? (route.options.errorComponent?.preload || resolvedNoSsr)) ? react.Suspense : require_SafeFragment.SafeFragment;
	const ResolvedCatchBoundary = routeErrorComponent ? require_CatchBoundary.CatchBoundary : require_SafeFragment.SafeFragment;
	const ResolvedNotFoundBoundary = routeNotFoundComponent ? require_not_found.CatchNotFound : require_SafeFragment.SafeFragment;
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(route.isRoot ? route.options.shellComponent ?? require_SafeFragment.SafeFragment : require_SafeFragment.SafeFragment, { children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_matchContext.matchContext.Provider, {
		value: matchId,
		children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(ResolvedSuspenseBoundary, {
			fallback: pendingElement,
			children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(ResolvedCatchBoundary, {
				getResetKey: () => resetKey,
				errorComponent: routeErrorComponent || require_CatchBoundary.ErrorComponent,
				onCatch: (error, errorInfo) => {
					if ((0, _tanstack_router_core.isNotFound)(error)) {
						error.routeId ??= matchState.routeId;
						throw error;
					}
					if (process.env.NODE_ENV !== "production") console.warn(`Warning: Error in route match: ${matchId}`);
					routeOnCatch?.(error, errorInfo);
				},
				children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(ResolvedNotFoundBoundary, {
					fallback: (error) => {
						error.routeId ??= matchState.routeId;
						if (!routeNotFoundComponent || error.routeId && error.routeId !== matchState.routeId || !error.routeId && !route.isRoot) throw error;
						return react.createElement(routeNotFoundComponent, error);
					},
					children: resolvedNoSsr || matchState._displayPending ? /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_ClientOnly.ClientOnly, {
						fallback: pendingElement,
						children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(MatchInner, { matchId })
					}) : /* @__PURE__ */ (0, react_jsx_runtime.jsx)(MatchInner, { matchId })
				})
			})
		})
	}), matchState.parentRouteId === _tanstack_router_core.rootRouteId ? /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(react_jsx_runtime.Fragment, { children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(OnRendered, {}), router.options.scrollRestoration && (_tanstack_router_core_isServer.isServer ?? router.isServer) ? /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_scroll_restoration.ScrollRestoration, {}) : null] }) : null] });
}
function OnRendered() {
	const router = require_useRouter.useRouter();
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) return null;
	const prevResolvedLocationRef = react.useRef();
	require_utils.useLayoutEffect(() => {
		const currentResolvedLocation = router.stores.resolvedLocation.get();
		const previousResolvedLocation = prevResolvedLocationRef.current;
		if (currentResolvedLocation && (!previousResolvedLocation || previousResolvedLocation.href !== currentResolvedLocation.href)) router.emit({
			type: "onRendered",
			...(0, _tanstack_router_core.getLocationChangeInfo)(router.stores.location.get(), previousResolvedLocation ?? currentResolvedLocation)
		});
		prevResolvedLocationRef.current = currentResolvedLocation;
	}, [(0, _tanstack_react_store.useStore)(router.stores.resolvedLocation, (resolvedLocation) => resolvedLocation?.state.__TSR_key), router]);
	return null;
}
var MatchInner = react.memo(function MatchInnerImpl({ matchId }) {
	const router = require_useRouter.useRouter();
	const getMatchPromise = (match, key) => {
		return router.getMatch(match.id)?._nonReactive[key] ?? match._nonReactive[key];
	};
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) {
		const match = router.stores.matchStores.get(matchId)?.get();
		if (!match) {
			if (process.env.NODE_ENV !== "production") throw new Error(`Invariant failed: Could not find match for matchId "${matchId}". Please file an issue!`);
			(0, _tanstack_router_core.invariant)();
		}
		const routeId = match.routeId;
		const route = router.routesById[routeId];
		const remountDeps = (router.routesById[routeId].options.remountDeps ?? router.options.defaultRemountDeps)?.({
			routeId,
			loaderDeps: match.loaderDeps,
			params: match._strictParams,
			search: match._strictSearch
		});
		const key = remountDeps ? JSON.stringify(remountDeps) : void 0;
		const Comp = route.options.component ?? router.options.defaultComponent;
		const out = Comp ? /* @__PURE__ */ (0, react_jsx_runtime.jsx)(Comp, {}, key) : /* @__PURE__ */ (0, react_jsx_runtime.jsx)(Outlet, {});
		if (match._displayPending) throw getMatchPromise(match, "displayPendingPromise");
		if (match._forcePending) throw getMatchPromise(match, "minPendingPromise");
		if (match.status === "pending") throw getMatchPromise(match, "loadPromise");
		if (match.status === "notFound") {
			if (!(0, _tanstack_router_core.isNotFound)(match.error)) {
				if (process.env.NODE_ENV !== "production") throw new Error("Invariant failed: Expected a notFound error");
				(0, _tanstack_router_core.invariant)();
			}
			return require_renderRouteNotFound.renderRouteNotFound(router, route, match.error);
		}
		if (match.status === "redirected") {
			if (!(0, _tanstack_router_core.isRedirect)(match.error)) {
				if (process.env.NODE_ENV !== "production") throw new Error("Invariant failed: Expected a redirect error");
				(0, _tanstack_router_core.invariant)();
			}
			throw getMatchPromise(match, "loadPromise");
		}
		if (match.status === "error") return /* @__PURE__ */ (0, react_jsx_runtime.jsx)((route.options.errorComponent ?? router.options.defaultErrorComponent) || require_CatchBoundary.ErrorComponent, {
			error: match.error,
			reset: void 0,
			info: { componentStack: "" }
		});
		return out;
	}
	const matchStore = router.stores.matchStores.get(matchId);
	if (!matchStore) {
		if (process.env.NODE_ENV !== "production") throw new Error(`Invariant failed: Could not find match for matchId "${matchId}". Please file an issue!`);
		(0, _tanstack_router_core.invariant)();
	}
	const match = (0, _tanstack_react_store.useStore)(matchStore, (value) => value);
	const routeId = match.routeId;
	const route = router.routesById[routeId];
	const key = react.useMemo(() => {
		const remountDeps = (router.routesById[routeId].options.remountDeps ?? router.options.defaultRemountDeps)?.({
			routeId,
			loaderDeps: match.loaderDeps,
			params: match._strictParams,
			search: match._strictSearch
		});
		return remountDeps ? JSON.stringify(remountDeps) : void 0;
	}, [
		routeId,
		match.loaderDeps,
		match._strictParams,
		match._strictSearch,
		router.options.defaultRemountDeps,
		router.routesById
	]);
	const out = react.useMemo(() => {
		const Comp = route.options.component ?? router.options.defaultComponent;
		if (Comp) return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(Comp, {}, key);
		return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(Outlet, {});
	}, [
		key,
		route.options.component,
		router.options.defaultComponent
	]);
	if (match._displayPending) throw getMatchPromise(match, "displayPendingPromise");
	if (match._forcePending) throw getMatchPromise(match, "minPendingPromise");
	if (match.status === "pending") {
		const pendingMinMs = route.options.pendingMinMs ?? router.options.defaultPendingMinMs;
		if (pendingMinMs) {
			const routerMatch = router.getMatch(match.id);
			if (routerMatch && !routerMatch._nonReactive.minPendingPromise) {
				if (!(_tanstack_router_core_isServer.isServer ?? router.isServer)) {
					const minPendingPromise = (0, _tanstack_router_core.createControlledPromise)();
					routerMatch._nonReactive.minPendingPromise = minPendingPromise;
					setTimeout(() => {
						minPendingPromise.resolve();
						routerMatch._nonReactive.minPendingPromise = void 0;
					}, pendingMinMs);
				}
			}
		}
		throw getMatchPromise(match, "loadPromise");
	}
	if (match.status === "notFound") {
		if (!(0, _tanstack_router_core.isNotFound)(match.error)) {
			if (process.env.NODE_ENV !== "production") throw new Error("Invariant failed: Expected a notFound error");
			(0, _tanstack_router_core.invariant)();
		}
		return require_renderRouteNotFound.renderRouteNotFound(router, route, match.error);
	}
	if (match.status === "redirected") {
		if (!(0, _tanstack_router_core.isRedirect)(match.error)) {
			if (process.env.NODE_ENV !== "production") throw new Error("Invariant failed: Expected a redirect error");
			(0, _tanstack_router_core.invariant)();
		}
		throw getMatchPromise(match, "loadPromise");
	}
	if (match.status === "error") {
		if (_tanstack_router_core_isServer.isServer ?? router.isServer) return /* @__PURE__ */ (0, react_jsx_runtime.jsx)((route.options.errorComponent ?? router.options.defaultErrorComponent) || require_CatchBoundary.ErrorComponent, {
			error: match.error,
			reset: void 0,
			info: { componentStack: "" }
		});
		throw match.error;
	}
	return out;
});
/**
* Render the next child match in the route tree. Typically used inside
* a route component to render nested routes.
*
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/outletComponent
*/
var Outlet = react.memo(function OutletImpl() {
	const router = require_useRouter.useRouter();
	const matchId = react.useContext(require_matchContext.matchContext);
	let routeId;
	let parentGlobalNotFound = false;
	let childMatchId;
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) {
		const matches = router.stores.matches.get();
		const parentIndex = matchId ? matches.findIndex((match) => match.id === matchId) : -1;
		const parentMatch = parentIndex >= 0 ? matches[parentIndex] : void 0;
		routeId = parentMatch?.routeId;
		parentGlobalNotFound = parentMatch?.globalNotFound ?? false;
		childMatchId = parentIndex >= 0 ? matches[parentIndex + 1]?.id : void 0;
	} else {
		const parentMatchStore = matchId ? router.stores.matchStores.get(matchId) : void 0;
		[routeId, parentGlobalNotFound] = (0, _tanstack_react_store.useStore)(parentMatchStore, (match) => [match?.routeId, match?.globalNotFound ?? false], outletMatchSelectionEqual);
		childMatchId = (0, _tanstack_react_store.useStore)(router.stores.matchesId, (ids) => {
			return ids[ids.findIndex((id) => id === matchId) + 1];
		});
	}
	const route = routeId ? router.routesById[routeId] : void 0;
	const pendingElement = router.options.defaultPendingComponent ? /* @__PURE__ */ (0, react_jsx_runtime.jsx)(router.options.defaultPendingComponent, {}) : null;
	if (parentGlobalNotFound) {
		if (!route) {
			if (process.env.NODE_ENV !== "production") throw new Error("Invariant failed: Could not resolve route for Outlet render");
			(0, _tanstack_router_core.invariant)();
		}
		return require_renderRouteNotFound.renderRouteNotFound(router, route, void 0);
	}
	if (!childMatchId) return null;
	const nextMatch = /* @__PURE__ */ (0, react_jsx_runtime.jsx)(Match, { matchId: childMatchId });
	if (routeId === _tanstack_router_core.rootRouteId) return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(react.Suspense, {
		fallback: pendingElement,
		children: nextMatch
	});
	return nextMatch;
});
//#endregion
exports.Match = Match;
exports.Outlet = Outlet;

//# sourceMappingURL=Match.cjs.map