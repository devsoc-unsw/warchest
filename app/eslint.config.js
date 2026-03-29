import { tanstackConfig } from '@tanstack/eslint-config'

export default [
    ...tanstackConfig,
  {
    // Ignore configuration files and specific JS files
    // so the TypeScript parser doesn't try to read them.
    ignores: [
      "eslint.config.js",
      "prettier.config.js",
      "test/traffic_sim.js" 
    ]
  }
]
