"use client";
const require_useRouter = require("./useRouter.cjs");
const require_useMatch = require("./useMatch.cjs");
let _tanstack_react_store = require("@tanstack/react-store");
let _tanstack_router_core_isServer = require("@tanstack/router-core/isServer");
//#region src/useLocation.tsx
/**
* Read the current location from the router state with optional selection.
* Useful for subscribing to just the pieces of location you care about.
*
* Options:
* - `select`: Project the `location` object to a derived value
* - `structuralSharing`: Enable structural sharing for stable references
*
* @returns The current location (or selected value).
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/useLocationHook
*/
function useLocation(opts) {
	const router = require_useRouter.useRouter();
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) {
		const location = router.stores.location.get();
		return opts?.select ? opts.select(location) : location;
	}
	return (0, _tanstack_react_store.useStore)(router.stores.location, require_useMatch.useStructuralSharing(opts, router));
}
//#endregion
exports.useLocation = useLocation;

//# sourceMappingURL=useLocation.cjs.map