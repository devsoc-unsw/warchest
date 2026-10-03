"use client";
import { useLayoutEffect, usePrevious } from "./utils.js";
import { useRouter } from "./useRouter.js";
import { getLocationChangeInfo, trimPathRight } from "@tanstack/router-core";
import * as React$1 from "react";
import { batch, useStore } from "@tanstack/react-store";
//#region src/Transitioner.tsx
function Transitioner() {
	const router = useRouter();
	const mountLoadForRouter = React$1.useRef({
		router,
		mounted: false
	});
	const [isTransitioning, setIsTransitioning] = React$1.useState(false);
	const isLoading = useStore(router.stores.isLoading, (value) => value);
	const hasPending = useStore(router.stores.hasPending, (value) => value);
	const previousIsLoading = usePrevious(isLoading);
	const isAnyPending = isLoading || isTransitioning || hasPending;
	const previousIsAnyPending = usePrevious(isAnyPending);
	const isPagePending = isLoading || hasPending;
	const previousIsPagePending = usePrevious(isPagePending);
	router.startTransition = (fn) => {
		setIsTransitioning(true);
		React$1.startTransition(() => {
			fn();
			setIsTransitioning(false);
		});
	};
	React$1.useEffect(() => {
		const unsub = router.history.subscribe(router.load);
		const nextLocation = router.buildLocation({
			to: router.latestLocation.pathname,
			search: true,
			params: true,
			hash: true,
			state: true,
			_includeValidateSearch: true
		});
		if (trimPathRight(router.latestLocation.publicHref) !== trimPathRight(nextLocation.publicHref)) router.commitLocation({
			...nextLocation,
			replace: true
		});
		return () => {
			unsub();
		};
	}, [router, router.history]);
	useLayoutEffect(() => {
		if (typeof window !== "undefined" && router.ssr || mountLoadForRouter.current.router === router && mountLoadForRouter.current.mounted) return;
		mountLoadForRouter.current = {
			router,
			mounted: true
		};
		const tryLoad = async () => {
			try {
				await router.load();
			} catch (err) {
				console.error(err);
			}
		};
		tryLoad();
	}, [router]);
	useLayoutEffect(() => {
		if (previousIsLoading && !isLoading) router.emit({
			type: "onLoad",
			...getLocationChangeInfo(router.stores.location.get(), router.stores.resolvedLocation.get())
		});
	}, [
		previousIsLoading,
		router,
		isLoading
	]);
	useLayoutEffect(() => {
		if (previousIsPagePending && !isPagePending) router.emit({
			type: "onBeforeRouteMount",
			...getLocationChangeInfo(router.stores.location.get(), router.stores.resolvedLocation.get())
		});
	}, [
		isPagePending,
		previousIsPagePending,
		router
	]);
	useLayoutEffect(() => {
		if (previousIsAnyPending && !isAnyPending) {
			const changeInfo = getLocationChangeInfo(router.stores.location.get(), router.stores.resolvedLocation.get());
			router.emit({
				type: "onResolved",
				...changeInfo
			});
			batch(() => {
				router.stores.status.set("idle");
				router.stores.resolvedLocation.set(router.stores.location.get());
			});
		}
	}, [
		isAnyPending,
		previousIsAnyPending,
		router
	]);
	return null;
}
//#endregion
export { Transitioner };

//# sourceMappingURL=Transitioner.js.map