import { drizzle } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import * as schema from "./schema";

export { schema };
export type Schema = typeof schema;

/**
 * Create a Drizzle client connected to Postgres.
 * Uses DATABASE_URL from env if no url is provided.
 */
export function createDb(url?: string) {
  const connectionString = url ?? process.env.DATABASE_URL;
  if (!connectionString) {
    throw new Error("DATABASE_URL (or url argument) is required");
  }
  const client = postgres(connectionString, { max: 10 });
  return drizzle(client, { schema });
}
