import { createNonReactiveMutableStore, createNonReactiveReadonlyStore } from "@tanstack/router-core";
import { batch, createAtom } from "@tanstack/react-store";
import { isServer } from "@tanstack/router-core/isServer";
//#region src/routerStores.ts
var getStoreFactory = (opts) => {
	if (isServer ?? opts.isServer) return {
		createMutableStore: createNonReactiveMutableStore,
		createReadonlyStore: createNonReactiveReadonlyStore,
		batch: (fn) => fn()
	};
	return {
		createMutableStore: createAtom,
		createReadonlyStore: createAtom,
		batch
	};
};
//#endregion
export { getStoreFactory };

//# sourceMappingURL=routerStores.js.map