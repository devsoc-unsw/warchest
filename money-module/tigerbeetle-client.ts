import { createClient } from "tigerbeetle-node";
import { lookup } from "node:dns";
import { promisify } from "node:util";

const lookupAsync = promisify(lookup);
const clusterId = 0n;

/**
 * Resolve address to IP:port — TigerBeetle native client expects IP, not hostname.
 */
async function resolveReplicaAddress(addr: string): Promise<string> {
  const parts = addr.split(":");
  const host = parts[0] ?? "tigerbeetle";
  const port = parts[1] ?? "3000";
  if (/^\d+\.\d+\.\d+\.\d+$/.test(host)) return addr;
  const { address } = await lookupAsync(host, { family: 4 });
  return `${address}:${port}`;
}

/**
 * TigerBeetle client (cluster 0). Resolves hostnames to IP so the native client works in Docker.
 */
let _client: ReturnType<typeof createClient> | null = null;

export async function getTigerbeetle() {
  if (_client) return _client;
  const raw = process.env.TIGERBEETLE_ADDRESS ?? "tigerbeetle:3000";
  const resolved = await resolveReplicaAddress(raw);
  _client = createClient({
    cluster_id: clusterId,
    replica_addresses: [resolved],
  });
  return _client;
}

/** Sync getter for code that runs after init; throws if not inited. */
export function getTigerbeetleSync(): ReturnType<typeof createClient> {
  if (!_client) throw new Error("TigerBeetle not initialized — await getTigerbeetle() first");
  return _client;
}
