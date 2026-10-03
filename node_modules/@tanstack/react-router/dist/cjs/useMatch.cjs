"use client";
const require_runtime = require("./_virtual/_rolldown/runtime.cjs");
const require_matchContext = require("./matchContext.cjs");
const require_useRouter = require("./useRouter.cjs");
let _tanstack_router_core = require("@tanstack/router-core");
let react = require("react");
react = require_runtime.__toESM(react, 1);
let _tanstack_react_store = require("@tanstack/react-store");
let _tanstack_router_core_isServer = require("@tanstack/router-core/isServer");
//#region src/useMatch.tsx
var dummyStore = {
	get() {},
	subscribe() {
		return { unsubscribe() {} };
	}
};
function useStructuralSharing(opts, router) {
	const previousResult = react.useRef();
	return (slice) => {
		const selected = opts?.select ? opts.select(slice) : slice;
		if (opts?.structuralSharing ?? router.options.defaultStructuralSharing) return previousResult.current = (0, _tanstack_router_core.replaceEqualDeep)(previousResult.current, selected);
		return selected;
	};
}
/**
* Read and select the nearest or targeted route match.
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/useMatchHook
*/
function useMatch(opts) {
	const router = require_useRouter.useRouter();
	const nearestMatchId = react.useContext(opts.from ? require_matchContext.dummyMatchContext : require_matchContext.matchContext);
	const matchStore = opts.from ? router.stores.getRouteMatchStore(opts.from) : router.stores.matchStores.get(nearestMatchId);
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) {
		const match = matchStore?.get();
		if (!match) {
			if (opts.shouldThrow ?? true) {
				if (process.env.NODE_ENV !== "production") throw new Error(`Invariant failed: Could not find ${opts.from ? `an active match from "${opts.from}"` : "a nearest match!"}`);
				(0, _tanstack_router_core.invariant)();
			}
			return;
		}
		return opts.select ? opts.select(match) : match;
	}
	const selector = useStructuralSharing(opts, router);
	const matchSelection = (0, _tanstack_react_store.useStore)(matchStore ?? dummyStore, (match) => match ? selector(match) : dummyStore);
	if (matchSelection !== dummyStore) return matchSelection;
	if (opts.shouldThrow ?? true) {
		if (process.env.NODE_ENV !== "production") throw new Error(`Invariant failed: Could not find ${opts.from ? `an active match from "${opts.from}"` : "a nearest match!"}`);
		(0, _tanstack_router_core.invariant)();
	}
}
//#endregion
exports.useMatch = useMatch;
exports.useStructuralSharing = useStructuralSharing;

//# sourceMappingURL=useMatch.cjs.map