/**
 * Money module: TigerBeetle + shared Drizzle DB.
 */

import { getTigerbeetle } from "./tigerbeetle-client.ts";
import { createDb } from "./db/index.ts";

export { createDb } from "./db/index.ts";
export { getTigerbeetle, getTigerbeetleSync } from "./tigerbeetle-client.ts";

const port = Number(process.env.PORT ?? 6767);

async function main() {
  const db = createDb();
  await getTigerbeetle();

  // this starts a http server
  const server = Bun.serve({
    port,
    fetch(req) {
      const url = new URL(req.url);
      if (url.pathname === "/health") {
        return new Response(JSON.stringify({ ok: true }), {
          headers: { "Content-Type": "application/json" },
        });
      }
      return new Response("Not Found", { status: 404 });
    },
  });

  console.log(`money-module listening on http://localhost:${server.port}`);
}

main();
