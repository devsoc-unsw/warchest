const require_runtime = require("./_virtual/_rolldown/runtime.cjs");
const require_useRouter = require("./useRouter.cjs");
let _tanstack_router_core = require("@tanstack/router-core");
let react = require("react");
react = require_runtime.__toESM(react, 1);
let _tanstack_react_store = require("@tanstack/react-store");
let _tanstack_router_core_isServer = require("@tanstack/router-core/isServer");
//#region src/headContentUtils.tsx
function buildTagsFromMatches(router, nonce, matches, assetCrossOrigin) {
	const routeMeta = matches.map((match) => match.meta).filter((meta) => meta !== void 0);
	const resultMeta = [];
	const metaByAttribute = {};
	let title;
	for (let i = routeMeta.length - 1; i >= 0; i--) {
		const metas = routeMeta[i];
		for (let j = metas.length - 1; j >= 0; j--) {
			const m = metas[j];
			if (!m) continue;
			if (m.title) {
				if (!title) title = {
					tag: "title",
					children: m.title
				};
			} else if ("script:ld+json" in m) try {
				const json = JSON.stringify(m["script:ld+json"]);
				resultMeta.push({
					tag: "script",
					attrs: { type: "application/ld+json" },
					children: (0, _tanstack_router_core.escapeHtml)(json)
				});
			} catch {}
			else {
				const attribute = m.name ?? m.property;
				if (attribute) if (metaByAttribute[attribute]) continue;
				else metaByAttribute[attribute] = true;
				resultMeta.push({
					tag: "meta",
					attrs: {
						...m,
						nonce
					}
				});
			}
		}
	}
	if (title) resultMeta.push(title);
	if (nonce) resultMeta.push({
		tag: "meta",
		attrs: {
			property: "csp-nonce",
			content: nonce
		}
	});
	resultMeta.reverse();
	const constructedLinks = matches.flatMap((match) => match.links ?? []).filter((link) => link !== void 0).map((link) => ({
		tag: "link",
		attrs: {
			...link,
			nonce
		}
	}));
	const manifest = router.ssr?.manifest;
	const manifestCssTags = [];
	if (manifest) {
		matches.forEach((match) => {
			(manifest.routes[match.routeId]?.css)?.forEach((link) => {
				const resolvedLink = (0, _tanstack_router_core.resolveManifestCssLink)(link);
				manifestCssTags.push({
					tag: "link",
					attrs: {
						rel: "stylesheet",
						...resolvedLink,
						crossOrigin: (0, _tanstack_router_core.getAssetCrossOrigin)(assetCrossOrigin, "stylesheet") ?? resolvedLink.crossOrigin,
						suppressHydrationWarning: true,
						nonce
					}
				});
			});
		});
		if (manifest.inlineStyle) manifestCssTags.push({
			tag: "style",
			attrs: {
				...manifest.inlineStyle.attrs,
				nonce
			},
			children: manifest.inlineStyle.children,
			inlineCss: true
		});
	}
	const preloadLinks = [];
	if (manifest) matches.forEach((match) => {
		manifest.routes[match.routeId]?.preloads?.forEach((preload) => {
			preloadLinks.push({
				tag: "link",
				attrs: {
					...(0, _tanstack_router_core.getScriptPreloadAttrs)(manifest, preload, assetCrossOrigin),
					nonce
				}
			});
		});
	});
	const styles = matches.flatMap((match) => match.styles ?? []).filter((style) => style !== void 0).map(({ children, ...attrs }) => ({
		tag: "style",
		attrs: {
			...attrs,
			nonce
		},
		children
	}));
	const headScripts = matches.flatMap((match) => match.headScripts ?? []).filter((script) => script !== void 0).map(({ children, ...script }) => ({
		tag: "script",
		attrs: {
			...script,
			nonce
		},
		children
	}));
	const tags = [];
	(0, _tanstack_router_core.appendUniqueUserTags)(tags, resultMeta);
	tags.push(...preloadLinks);
	(0, _tanstack_router_core.appendUniqueUserTags)(tags, constructedLinks);
	tags.push(...manifestCssTags);
	(0, _tanstack_router_core.appendUniqueUserTags)(tags, styles);
	(0, _tanstack_router_core.appendUniqueUserTags)(tags, headScripts);
	return tags;
}
/**
* Build the list of head/link/meta/script tags to render for active matches.
* Used internally by `HeadContent`.
*/
var useTags = (assetCrossOrigin) => {
	const router = require_useRouter.useRouter();
	const nonce = router.options.ssr?.nonce;
	if (_tanstack_router_core_isServer.isServer ?? router.isServer) return buildTagsFromMatches(router, nonce, router.stores.matches.get(), assetCrossOrigin);
	const routeMeta = (0, _tanstack_react_store.useStore)(router.stores.matches, (matches) => {
		return matches.map((match) => match.meta).filter((meta) => meta !== void 0);
	}, _tanstack_router_core.deepEqual);
	const meta = react.useMemo(() => {
		const resultMeta = [];
		const metaByAttribute = {};
		let title;
		for (let i = routeMeta.length - 1; i >= 0; i--) {
			const metas = routeMeta[i];
			for (let j = metas.length - 1; j >= 0; j--) {
				const m = metas[j];
				if (!m) continue;
				if (m.title) {
					if (!title) title = {
						tag: "title",
						children: m.title
					};
				} else if ("script:ld+json" in m) try {
					const json = JSON.stringify(m["script:ld+json"]);
					resultMeta.push({
						tag: "script",
						attrs: { type: "application/ld+json" },
						children: (0, _tanstack_router_core.escapeHtml)(json)
					});
				} catch {}
				else {
					const attribute = m.name ?? m.property;
					if (attribute) if (metaByAttribute[attribute]) continue;
					else metaByAttribute[attribute] = true;
					resultMeta.push({
						tag: "meta",
						attrs: {
							...m,
							nonce
						}
					});
				}
			}
		}
		if (title) resultMeta.push(title);
		if (nonce) resultMeta.push({
			tag: "meta",
			attrs: {
				property: "csp-nonce",
				content: nonce
			}
		});
		resultMeta.reverse();
		return resultMeta;
	}, [routeMeta, nonce]);
	const links = (0, _tanstack_react_store.useStore)(router.stores.matches, (matches) => {
		return matches.flatMap((match) => match.links ?? []).filter((link) => link !== void 0).map((link) => ({
			tag: "link",
			attrs: {
				...link,
				nonce
			}
		}));
	}, _tanstack_router_core.deepEqual);
	const manifestCssTags = (0, _tanstack_react_store.useStore)(router.stores.matches, (matches) => {
		const manifest = router.ssr?.manifest;
		const tags = [];
		if (!manifest) return tags;
		matches.forEach((match) => {
			manifest.routes[match.routeId]?.css?.forEach((link) => {
				const resolvedLink = (0, _tanstack_router_core.resolveManifestCssLink)(link);
				tags.push({
					tag: "link",
					attrs: {
						rel: "stylesheet",
						...resolvedLink,
						crossOrigin: (0, _tanstack_router_core.getAssetCrossOrigin)(assetCrossOrigin, "stylesheet") ?? resolvedLink.crossOrigin,
						suppressHydrationWarning: true,
						nonce
					}
				});
			});
		});
		if (manifest.inlineStyle) tags.push({
			tag: "style",
			attrs: {
				...manifest.inlineStyle.attrs,
				nonce
			},
			children: manifest.inlineStyle.children,
			inlineCss: true
		});
		return tags;
	}, _tanstack_router_core.deepEqual);
	const preloadLinks = (0, _tanstack_react_store.useStore)(router.stores.matches, (matches) => {
		const preloadLinks = [];
		const manifest = router.ssr?.manifest;
		if (!manifest) return preloadLinks;
		matches.forEach((match) => {
			manifest.routes[match.routeId]?.preloads?.forEach((preload) => {
				preloadLinks.push({
					tag: "link",
					attrs: {
						...(0, _tanstack_router_core.getScriptPreloadAttrs)(manifest, preload, assetCrossOrigin),
						nonce
					}
				});
			});
		});
		return preloadLinks;
	}, _tanstack_router_core.deepEqual);
	const styles = (0, _tanstack_react_store.useStore)(router.stores.matches, (matches) => {
		return matches.flatMap((match) => match.styles ?? []).filter((style) => style !== void 0).map(({ children, ...attrs }) => ({
			tag: "style",
			attrs: {
				...attrs,
				nonce
			},
			children
		}));
	}, _tanstack_router_core.deepEqual);
	const headScripts = (0, _tanstack_react_store.useStore)(router.stores.matches, (matches) => {
		return matches.flatMap((match) => match.headScripts ?? []).filter((script) => script !== void 0).map(({ children, ...script }) => ({
			tag: "script",
			attrs: {
				...script,
				nonce
			},
			children
		}));
	}, _tanstack_router_core.deepEqual);
	const tags = [];
	(0, _tanstack_router_core.appendUniqueUserTags)(tags, meta);
	tags.push(...preloadLinks);
	(0, _tanstack_router_core.appendUniqueUserTags)(tags, links);
	tags.push(...manifestCssTags);
	(0, _tanstack_router_core.appendUniqueUserTags)(tags, styles);
	(0, _tanstack_router_core.appendUniqueUserTags)(tags, headScripts);
	return tags;
};
//#endregion
exports.useTags = useTags;

//# sourceMappingURL=headContentUtils.cjs.map