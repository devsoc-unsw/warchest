"use client";
import { CatchBoundary, ErrorComponent } from "./CatchBoundary.js";
import { matchContext } from "./matchContext.js";
import { useRouter } from "./useRouter.js";
import { useStructuralSharing } from "./useMatch.js";
import { Transitioner } from "./Transitioner.js";
import { SafeFragment } from "./SafeFragment.js";
import { Match } from "./Match.js";
import { rootRouteId } from "@tanstack/router-core";
import * as React$1 from "react";
import { jsx, jsxs } from "react/jsx-runtime";
import { useStore } from "@tanstack/react-store";
import { isServer } from "@tanstack/router-core/isServer";
//#region src/Matches.tsx
/**
* Internal component that renders the router's active match tree with
* suspense, error, and not-found boundaries. Rendered by `RouterProvider`.
*/
function Matches() {
	const router = useRouter();
	const PendingComponent = router.routesById[rootRouteId].options.pendingComponent ?? router.options.defaultPendingComponent;
	const pendingElement = PendingComponent ? /* @__PURE__ */ jsx(PendingComponent, {}) : null;
	const inner = /* @__PURE__ */ jsxs((isServer ?? router.isServer) || typeof document !== "undefined" && router.ssr ? SafeFragment : React$1.Suspense, {
		fallback: pendingElement,
		children: [!(isServer ?? router.isServer) && /* @__PURE__ */ jsx(Transitioner, {}), /* @__PURE__ */ jsx(MatchesInner, {})]
	});
	return router.options.InnerWrap ? /* @__PURE__ */ jsx(router.options.InnerWrap, { children: inner }) : inner;
}
function MatchesInner() {
	const router = useRouter();
	const _isServer = isServer ?? router.isServer;
	const matchId = _isServer ? router.stores.firstId.get() : useStore(router.stores.firstId, (id) => id);
	const resetKey = _isServer ? router.stores.loadedAt.get() : useStore(router.stores.loadedAt, (loadedAt) => loadedAt);
	const matchComponent = matchId ? /* @__PURE__ */ jsx(Match, { matchId }) : null;
	return /* @__PURE__ */ jsx(matchContext.Provider, {
		value: matchId,
		children: router.options.disableGlobalCatchBoundary ? matchComponent : /* @__PURE__ */ jsx(CatchBoundary, {
			getResetKey: () => resetKey,
			errorComponent: ErrorComponent,
			onCatch: process.env.NODE_ENV !== "production" ? (error) => {
				console.warn(`Warning: The following error wasn't caught by any route! At the very least, consider setting an 'errorComponent' in your RootRoute!`);
				console.warn(`Warning: ${error.message || error.toString()}`);
			} : void 0,
			children: matchComponent
		})
	});
}
/**
* Create a matcher function for testing locations against route definitions.
*
* The returned function accepts standard navigation options (`to`, `params`,
* `search`, etc.) and returns either `false` (no match) or the matched params
* object when the route matches the current or pending location.
*
* Useful for conditional rendering and active UI states.
*
* @returns A `matchRoute(options)` function that returns `false` or params.
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/useMatchRouteHook
*/
function useMatchRoute() {
	const router = useRouter();
	if (!(isServer ?? router.isServer)) useStore(router.stores.matchRouteDeps, (d) => d);
	return React$1.useCallback((opts) => {
		const { pending, caseSensitive, fuzzy, includeSearch, ...rest } = opts;
		return router.matchRoute(rest, {
			pending,
			caseSensitive,
			fuzzy,
			includeSearch
		});
	}, [router]);
}
/**
* Component that conditionally renders its children based on whether a route
* matches the provided `from`/`to` options. If `children` is a function, it
* receives the matched params object.
*
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/matchRouteComponent
*/
function MatchRoute(props) {
	const params = useMatchRoute()(props);
	if (typeof props.children === "function") return props.children(params);
	return params ? props.children : null;
}
function useMatches(opts) {
	const router = useRouter();
	if (isServer ?? router.isServer) {
		const matches = router.stores.matches.get();
		return opts?.select ? opts.select(matches) : matches;
	}
	return useStore(router.stores.matches, useStructuralSharing(opts, router));
}
/**
* Read the full array of active route matches or select a derived subset.
*
* Useful for debugging, breadcrumbs, or aggregating metadata across matches.
*
* @returns The array of matches (or the selected value).
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/useMatchesHook
*/
/**
* Read the full array of active route matches or select a derived subset.
*
* Useful for debugging, breadcrumbs, or aggregating metadata across matches.
*
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/useMatchesHook
*/
function useParentMatches(opts) {
	const contextMatchId = React$1.useContext(matchContext);
	return useMatches({
		select: (matches) => {
			matches = matches.slice(0, matches.findIndex((d) => d.id === contextMatchId));
			return opts?.select ? opts.select(matches) : matches;
		},
		structuralSharing: opts?.structuralSharing
	});
}
/**
* Read the array of active route matches that are children of the current
* match (or selected parent) in the match tree.
*/
function useChildMatches(opts) {
	const contextMatchId = React$1.useContext(matchContext);
	return useMatches({
		select: (matches) => {
			matches = matches.slice(matches.findIndex((d) => d.id === contextMatchId) + 1);
			return opts?.select ? opts.select(matches) : matches;
		},
		structuralSharing: opts?.structuralSharing
	});
}
//#endregion
export { MatchRoute, Matches, useChildMatches, useMatchRoute, useMatches, useParentMatches };

//# sourceMappingURL=Matches.js.map