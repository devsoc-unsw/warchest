import { AnyRouter } from './router.cjs';
import { ParsedLocation } from './location.cjs';
export type ScrollRestorationEntry = {
    scrollX: number;
    scrollY: number;
};
export type ScrollRestorationOptions = {
    getKey?: (location: ParsedLocation) => string;
    scrollBehavior?: ScrollToOptions['behavior'];
};
export declare const storageKey = "tsr-scroll-restoration-v1_3";
/**
 * The default `getKey` function for `useScrollRestoration`.
 * It returns the `key` from the location state or the `href` of the location.
 *
 * The `location.href` is used as a fallback to support the use case where the location state is not available like the initial render.
 */
export declare const defaultGetScrollRestorationKey: (location: ParsedLocation) => string;
export declare function getElementScrollRestorationEntry(router: AnyRouter, options: ({
    id: string;
    getElement?: () => Window | Element | undefined | null;
} | {
    id?: string;
    getElement: () => Window | Element | undefined | null;
}) & {
    getKey?: (location: ParsedLocation) => string;
}): ScrollRestorationEntry | undefined;
export declare function setupScrollRestoration(router: AnyRouter, force?: boolean): void;
