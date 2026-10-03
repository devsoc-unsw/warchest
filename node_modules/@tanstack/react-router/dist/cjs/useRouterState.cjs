"use client";
const require_useRouter = require("./useRouter.cjs");
const require_useMatch = require("./useMatch.cjs");
let _tanstack_react_store = require("@tanstack/react-store");
let _tanstack_router_core_isServer = require("@tanstack/router-core/isServer");
//#region src/useRouterState.tsx
/**
* Subscribe to the router's state store with optional selection and
* structural sharing for render optimization.
*
* Options:
* - `select`: Project the full router state to a derived slice
* - `structuralSharing`: Replace-equal semantics for stable references
* - `router`: Read state from a specific router instance instead of context
*
* @returns The selected router state (or the full state by default).
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/useRouterStateHook
*/
function useRouterState(opts) {
	const contextRouter = require_useRouter.useRouter({ warn: opts?.router === void 0 });
	const router = opts?.router || contextRouter;
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) {
		const state = router.stores.__store.get();
		return opts?.select ? opts.select(state) : state;
	}
	return (0, _tanstack_react_store.useStore)(router.stores.__store, require_useMatch.useStructuralSharing(opts, router));
}
//#endregion
exports.useRouterState = useRouterState;

//# sourceMappingURL=useRouterState.cjs.map