/**
 * Temporary compatibility re-export.
 *
 * The API client moved to `$lib/api/client` as part of the frontend
 * architecture migration (Phase 1). This module keeps the old import path
 * (`$lib/server/api`) working for the ~24 loaders/actions that still import
 * `authedFetch`/`authedFetchSafe` directly, so Phase 1 stays a mechanical
 * move without touching every route.
 *
 * Phase 2 migrates those callers to the typed domain modules under `$lib/api`
 * and deletes this file. Do not add new imports of `$lib/server/api`.
 */
export { authedFetch, authedFetchSafe } from '$lib/api/client';
