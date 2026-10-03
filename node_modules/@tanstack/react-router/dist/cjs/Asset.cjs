"use client";
const require_runtime = require("./_virtual/_rolldown/runtime.cjs");
const require_ClientOnly = require("./ClientOnly.cjs");
const require_useRouter = require("./useRouter.cjs");
let react = require("react");
react = require_runtime.__toESM(react, 1);
let react_jsx_runtime = require("react/jsx-runtime");
let _tanstack_router_core_isServer = require("@tanstack/router-core/isServer");
//#region src/Asset.tsx
var INLINE_CSS_HYDRATION_ATTR = "data-tsr-inline-css";
var noopScriptHandler = () => {};
function setScriptAttrs(script, attrs) {
	if (!attrs) return;
	for (const [key, value] of Object.entries(attrs)) if (key !== "suppressHydrationWarning" && value !== void 0 && value !== false) script.setAttribute(key, typeof value === "boolean" ? "" : String(value));
}
function Asset(asset) {
	const { attrs, children, nonce, preventScriptHoist } = asset;
	switch (asset.tag) {
		case "title": return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("title", {
			...attrs,
			suppressHydrationWarning: true,
			children
		});
		case "meta": return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("meta", {
			...attrs,
			suppressHydrationWarning: true
		});
		case "link": return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("link", {
			...attrs,
			precedence: attrs?.precedence ?? (attrs?.rel === "stylesheet" ? "default" : void 0),
			nonce,
			suppressHydrationWarning: true
		});
		case "style":
			if (asset.inlineCss && (process.env.TSS_INLINE_CSS_ENABLED === "true" || process.env.TSS_INLINE_CSS_ENABLED === void 0 && _tanstack_router_core_isServer.isServer)) return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(InlineCssStyle, {
				attrs,
				nonce,
				children
			});
			return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("style", {
				...attrs,
				dangerouslySetInnerHTML: { __html: children },
				nonce
			});
		case "script": return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(Script, {
			attrs,
			preventScriptHoist,
			children
		});
		default: return null;
	}
}
function InlineCssStyle({ attrs, children, nonce }) {
	const isInlineCssPlaceholder = children === void 0;
	const [hydratedInlineCss] = react.useState(() => {
		if (!isInlineCssPlaceholder || typeof document === "undefined") return;
		return document.querySelector(`style[${INLINE_CSS_HYDRATION_ATTR}]`)?.textContent ?? void 0;
	});
	const html = isInlineCssPlaceholder ? hydratedInlineCss ?? "" : children ?? "";
	return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("style", {
		...attrs,
		[INLINE_CSS_HYDRATION_ATTR]: "",
		dangerouslySetInnerHTML: { __html: html },
		nonce,
		suppressHydrationWarning: true
	});
}
function Script({ attrs, children, preventScriptHoist }) {
	const router = require_useRouter.useRouter();
	const hydrated = require_ClientOnly.useHydrated();
	const dataScript = typeof attrs?.type === "string" && attrs.type !== "" && attrs.type !== "text/javascript" && attrs.type !== "module";
	if (process.env.NODE_ENV !== "production" && attrs?.src && typeof children === "string" && children.trim().length) console.warn("[TanStack Router] <Script> received both `src` and `children`. The `children` content will be ignored. Remove `children` or remove `src`.");
	react.useEffect(() => {
		if (dataScript) return;
		if (attrs?.src) {
			const normSrc = (() => {
				try {
					const base = document.baseURI || window.location.href;
					return new URL(attrs.src, base).href;
				} catch {
					return attrs.src;
				}
			})();
			for (const el of document.querySelectorAll("script[src]")) if (el.src === normSrc) return;
			const script = document.createElement("script");
			setScriptAttrs(script, attrs);
			document.head.appendChild(script);
			return () => script.remove();
		}
		if (typeof children === "string") {
			const typeAttr = typeof attrs?.type === "string" ? attrs.type : "text/javascript";
			const nonceAttr = typeof attrs?.nonce === "string" ? attrs.nonce : void 0;
			for (const el of document.querySelectorAll("script:not([src])")) {
				if (!(el instanceof HTMLScriptElement)) continue;
				const sType = el.getAttribute("type") ?? "text/javascript";
				const sNonce = el.getAttribute("nonce") ?? void 0;
				if (el.textContent === children && sType === typeAttr && sNonce === nonceAttr) return;
			}
			const script = document.createElement("script");
			script.textContent = children;
			setScriptAttrs(script, attrs);
			document.head.appendChild(script);
			return () => script.remove();
		}
	}, [
		attrs,
		children,
		dataScript
	]);
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) {
		if (attrs?.src) {
			if (!preventScriptHoist) return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("script", {
				...attrs,
				suppressHydrationWarning: true
			});
			return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("script", {
				...attrs,
				onLoad: noopScriptHandler,
				suppressHydrationWarning: true
			});
		}
		if (typeof children === "string") return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("script", {
			...attrs,
			dangerouslySetInnerHTML: { __html: children },
			suppressHydrationWarning: true
		});
		return null;
	}
	if (dataScript && typeof children === "string") return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("script", {
		...attrs,
		suppressHydrationWarning: true,
		dangerouslySetInnerHTML: { __html: children }
	});
	if (!hydrated) {
		if (attrs?.src) return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("script", {
			...attrs,
			suppressHydrationWarning: true
		});
		if (typeof children === "string") return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("script", {
			...attrs,
			dangerouslySetInnerHTML: { __html: children },
			suppressHydrationWarning: true
		});
	}
	return null;
}
//#endregion
exports.Asset = Asset;

//# sourceMappingURL=Asset.cjs.map