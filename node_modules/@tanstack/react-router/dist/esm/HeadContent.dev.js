"use client";
import { useHydrated } from "./ClientOnly.js";
import { useRouter } from "./useRouter.js";
import { Asset } from "./Asset.js";
import { useTags } from "./headContentUtils.js";
import { DEV_STYLES_ATTR } from "@tanstack/router-core";
import * as React$1 from "react";
import { createElement } from "react";
import { Fragment, jsx } from "react/jsx-runtime";
//#region src/HeadContent.dev.tsx
/**
* Render route-managed head tags (title, meta, links, styles, head scripts).
* Place inside the document head of your app shell.
*
* Development version: filters out dev styles link after hydration and
* includes a fallback cleanup effect for hydration mismatch cases.
*
* @link https://tanstack.com/router/latest/docs/framework/react/guide/document-head-management
*/
function HeadContent(props) {
	const tags = useTags(props.assetCrossOrigin);
	const nonce = useRouter().options.ssr?.nonce;
	const hydrated = useHydrated();
	React$1.useEffect(() => {
		if (hydrated) document.querySelectorAll(`link[${DEV_STYLES_ATTR}]`).forEach((el) => el.remove());
	}, [hydrated]);
	return /* @__PURE__ */ jsx(Fragment, { children: (hydrated ? tags.filter((tag) => tag.tag !== "link" || tag.attrs?.[DEV_STYLES_ATTR] !== true) : tags).map((tag) => /* @__PURE__ */ createElement(Asset, {
		...tag,
		key: `tsr-meta-${JSON.stringify(tag)}`,
		nonce
	})) });
}
//#endregion
export { HeadContent };

//# sourceMappingURL=HeadContent.dev.js.map