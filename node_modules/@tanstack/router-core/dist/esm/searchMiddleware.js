import { deepEqual, hasOwn } from "./utils.js";
//#region src/searchMiddleware.ts
/**
* Retain specified search params across navigations.
*
* If `keys` is `true`, retain all current params. Otherwise, copy only the
* listed keys from the current search into the next search.
*
* @param keys `true` to retain all, or a list of keys to retain.
* @returns A search middleware suitable for route `search.middlewares`.
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/retainSearchParamsFunction
*/
function retainSearchParams(keys) {
	return ({ search, next }) => {
		const { search: resultSearch, meta } = next(search, true);
		if (keys === true) {
			const copy = {
				...search,
				...resultSearch
			};
			const removed = meta.removed;
			const explicit = meta.explicit;
			for (const key of removed?.keys() || []) if (explicit && hasOwn.call(explicit, key) || deepEqual(search[key], removed.get(key))) delete copy[key];
			for (const key of meta.removedAny || []) delete copy[key];
			for (const key of meta.defaulted?.keys() || []) if (key in search && !meta.removedAny?.has(key) && !(meta.removed?.has(key) && (explicit && hasOwn.call(explicit, key) || deepEqual(search[key], meta.removed.get(key))))) copy[key] = search[key];
			return copy;
		}
		const copy = { ...resultSearch };
		const explicit = meta.explicit;
		for (const key of keys) {
			const stringKey = key;
			if (!(meta.removedAny?.has(stringKey) || meta.removed?.has(stringKey) && (explicit && hasOwn.call(explicit, stringKey) || deepEqual(search[key], meta.removed.get(stringKey)))) && (!(key in copy) || key in search && meta.defaulted?.has(stringKey) && deepEqual(copy[key], meta.defaulted.get(stringKey)))) copy[key] = search[key];
		}
		return copy;
	};
}
/**
* Remove optional or default-valued search params from navigations.
*
* - Pass `true` (only if there are no required search params) to strip all.
* - Pass an array to always remove those optional keys.
* - Pass an object of default values; keys equal (deeply) to the defaults are removed.
*
* @returns A search middleware suitable for route `search.middlewares`.
* @link https://tanstack.com/router/latest/docs/framework/react/api/router/stripSearchParamsFunction
*/
function stripSearchParams(input) {
	return (({ search, next, meta }) => {
		if (input === true) {
			Object.keys(search).forEach((key) => {
				if (meta) (meta.removedAny ||= /* @__PURE__ */ new Set()).add(key);
			});
			return {};
		}
		const result = { ...next(search) };
		if (Array.isArray(input)) input.forEach((key) => {
			delete result[key];
			if (meta) (meta.removedAny ||= /* @__PURE__ */ new Set()).add(key);
		});
		else Object.entries(input).forEach(([key, value]) => {
			if (deepEqual(result[key], value)) {
				delete result[key];
				if (meta) (meta.removed ||= /* @__PURE__ */ new Map()).set(key, value);
			}
		});
		return result;
	});
}
//#endregion
export { retainSearchParams, stripSearchParams };

//# sourceMappingURL=searchMiddleware.js.map