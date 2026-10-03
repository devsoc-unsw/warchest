import { AnyRoute } from './route.js';
import { RouterState } from './router.js';
import { FullSearchSchema } from './routeInfo.js';
import { ParsedLocation } from './location.js';
import { AnyRedirect } from './redirect.js';
import { AnyRouteMatch } from './Matches.js';
export interface RouterReadableStore<TValue> {
    get: () => TValue;
}
export interface RouterWritableStore<TValue> extends RouterReadableStore<TValue> {
    set: ((updater: (prev: TValue) => TValue) => void) & ((value: TValue) => void);
}
export type RouterBatchFn = (fn: () => void) => void;
export type MutableStoreFactory = <TValue>(initialValue: TValue) => RouterWritableStore<TValue>;
export type ReadonlyStoreFactory = <TValue>(read: () => TValue) => RouterReadableStore<TValue>;
export type GetStoreConfig = (opts: {
    isServer?: boolean;
}) => StoreConfig;
export type StoreConfig = {
    createMutableStore: MutableStoreFactory;
    createReadonlyStore: ReadonlyStoreFactory;
    batch: RouterBatchFn;
    init?: (stores: RouterStores<AnyRoute>) => void;
};
type MatchStore = RouterWritableStore<AnyRouteMatch> & {
    routeId?: string;
};
type ReadableStore<TValue> = RouterReadableStore<TValue>;
/** SSR non-reactive createMutableStore */
export declare function createNonReactiveMutableStore<TValue>(initialValue: TValue): RouterWritableStore<TValue>;
/** SSR non-reactive createReadonlyStore */
export declare function createNonReactiveReadonlyStore<TValue>(read: () => TValue): RouterReadableStore<TValue>;
export interface RouterStores<in out TRouteTree extends AnyRoute> {
    status: RouterWritableStore<RouterState<TRouteTree>['status']>;
    loadedAt: RouterWritableStore<number>;
    isLoading: RouterWritableStore<boolean>;
    isTransitioning: RouterWritableStore<boolean>;
    location: RouterWritableStore<ParsedLocation<FullSearchSchema<TRouteTree>>>;
    resolvedLocation: RouterWritableStore<ParsedLocation<FullSearchSchema<TRouteTree>> | undefined>;
    statusCode: RouterWritableStore<number>;
    redirect: RouterWritableStore<AnyRedirect | undefined>;
    matchesId: RouterWritableStore<Array<string>>;
    pendingIds: RouterWritableStore<Array<string>>;
    matches: ReadableStore<Array<AnyRouteMatch>>;
    pendingMatches: ReadableStore<Array<AnyRouteMatch>>;
    cachedMatches: ReadableStore<Array<AnyRouteMatch>>;
    firstId: ReadableStore<string | undefined>;
    hasPending: ReadableStore<boolean>;
    matchRouteDeps: ReadableStore<{
        locationHref: string;
        resolvedLocationHref: string | undefined;
        status: RouterState<TRouteTree>['status'];
    }>;
    __store: RouterReadableStore<RouterState<TRouteTree>>;
    matchStores: Map<string, MatchStore>;
    pendingMatchStores: Map<string, MatchStore>;
    cachedMatchStores: Map<string, MatchStore>;
    /**
     * Get a computed store that resolves a routeId to its current match state.
     * Returns the same cached store instance for repeated calls with the same key.
     * The computed depends on matchesId + the individual match store, so
     * subscribers are only notified when the resolved match state changes.
     */
    getRouteMatchStore: (routeId: string) => RouterReadableStore<AnyRouteMatch | undefined>;
    setMatches: (nextMatches: Array<AnyRouteMatch>) => void;
    setPending: (nextMatches: Array<AnyRouteMatch>) => void;
    setCached: (nextMatches: Array<AnyRouteMatch>) => void;
}
export declare function createRouterStores<TRouteTree extends AnyRoute>(initialState: RouterState<TRouteTree>, config: StoreConfig): RouterStores<TRouteTree>;
export {};
