import { pgTable, serial, text, timestamp } from "drizzle-orm/pg-core";

/**
 * Shared schema — used by both app and money-module.
 * Add your tables here.
 */
export const example = pgTable("example", {
  id: serial("id").primaryKey(),
  name: text("name").notNull(),
  createdAt: timestamp("created_at", { withTimezone: true }).defaultNow().notNull(),
});
