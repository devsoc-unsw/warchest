export type AssetCrossOrigin = 'anonymous' | 'use-credentials';
export type ScriptFormat = 'module' | 'iife';
export declare const DEV_STYLES_ATTR = "data-tanstack-router-dev-styles";
export type AssetCrossOriginConfig = AssetCrossOrigin | Partial<Record<'script' | 'stylesheet', AssetCrossOrigin>>;
export type ManifestAssetLink = string | {
    href: string;
    crossOrigin?: AssetCrossOrigin;
};
export declare function getAssetCrossOrigin(assetCrossOrigin: AssetCrossOriginConfig | undefined, kind: 'script' | 'stylesheet'): AssetCrossOrigin | undefined;
export declare function getManifestScriptFormat(manifest: {
    scriptFormat?: ScriptFormat;
} | undefined): ScriptFormat;
export declare function getScriptPreloadAttrs(manifest: {
    scriptFormat?: ScriptFormat;
} | undefined, link: ManifestAssetLink, assetCrossOrigin?: AssetCrossOriginConfig): {
    rel: 'modulepreload' | 'preload';
    as?: 'script';
    href: string;
    crossOrigin?: AssetCrossOrigin;
};
export declare function resolveManifestAssetLink(link: ManifestAssetLink): {
    href: string;
    crossOrigin?: AssetCrossOrigin;
};
export type Manifest = {
    scriptFormat?: ScriptFormat;
    inlineStyle?: ManifestInlineCss;
    routes: Record<string, ManifestRoute>;
};
export type ServerManifest = {
    scriptFormat?: ScriptFormat;
    inlineCss?: ServerManifestInlineCss;
    routes: Record<string, ServerManifestRoute>;
};
export type ServerManifestInlineCss = {
    styles: Record<string, string>;
    templates?: Record<string, InlineCssTemplate>;
};
export type InlineCssTemplate = {
    strings: Array<string>;
    urls: Array<string>;
};
export type ManifestRoute = {
    filePath?: string;
    preloads?: Array<ManifestAssetLink>;
    scripts?: Array<ManifestScript>;
    css?: Array<ManifestCssLink>;
};
export type ServerManifestRoute = ManifestRoute;
export type ManifestRouteAssets = Pick<ManifestRoute, 'preloads' | 'scripts' | 'css'>;
export type RouterManagedTitleTag = {
    tag: 'title';
    attrs?: Record<string, any>;
    children: string;
};
export type RouterManagedMetaTag = {
    tag: 'meta';
    attrs?: Record<string, any>;
    children?: never;
};
export type RouterManagedLinkTag = {
    tag: 'link';
    attrs?: Record<string, any>;
    children?: never;
};
export type RouterManagedScriptTag = {
    tag: 'script';
    attrs?: Record<string, any>;
    children?: string;
};
export type ManifestScript = Omit<RouterManagedScriptTag, 'tag'>;
export type RouterManagedStyleTag = {
    tag: 'style';
    attrs?: Record<string, any>;
    children?: string;
    inlineCss?: true;
};
export type RouterManagedTag = RouterManagedTitleTag | RouterManagedMetaTag | RouterManagedLinkTag | RouterManagedScriptTag | RouterManagedStyleTag;
export declare function appendUniqueUserTags(target: Array<RouterManagedTag>, tags: Array<RouterManagedTag>): void;
export type ManifestCssLink = string | {
    href: string;
    crossOrigin?: AssetCrossOrigin;
    [DEV_STYLES_ATTR]?: true;
};
export type ManifestInlineCss = {
    attrs?: Record<string, any>;
    children?: string;
};
export type RouterManagedInlineCssTag = RouterManagedStyleTag & {
    inlineCss: true;
};
export declare function getStylesheetHref(asset: ManifestCssLink): string;
export declare function resolveManifestCssLink(link: ManifestCssLink): {
    href: string;
    crossOrigin?: AssetCrossOrigin;
    "data-tanstack-router-dev-styles"?: true;
};
export declare function createInlineCssStyleAsset(css: string): ManifestInlineCss;
export declare function createInlineCssPlaceholderAsset(): ManifestInlineCss;
