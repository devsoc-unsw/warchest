"use client";
import { useRouter } from "./useRouter.js";
import { useStructuralSharing } from "./useMatch.js";
import { useStore } from "@tanstack/react-store";
import { isServer } from "@tanstack/router-core/isServer";
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
	const router = useRouter();
	if (isServer ?? router.isServer) {
		const location = router.stores.location.get();
		return opts?.select ? opts.select(location) : location;
	}
	return useStore(router.stores.location, useStructuralSharing(opts, router));
}
//#endregion
export { useLocation };

//# sourceMappingURL=useLocation.js.map