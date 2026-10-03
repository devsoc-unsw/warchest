//#region src/config.ts
const createRouterConfig = (options) => {
	return {
		serializationAdapters: options.serializationAdapters,
		defaultSsr: options.defaultSsr
	};
};
//#endregion
exports.createRouterConfig = createRouterConfig;

//# sourceMappingURL=config.cjs.map