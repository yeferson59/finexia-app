/**
 * Temporary compatibility re-export.
 *
 * `lib/utils.ts` was a catch-all grab bag. As part of the frontend
 * architecture migration (Phase 1) its helpers moved to themed modules under
 * `$lib/shared/`. This module re-exports them so existing `$lib/utils` imports
 * keep working while callers migrate to the new paths. Prefer importing from
 * `$lib/shared/*` in new code; this file is removed once no imports remain.
 *
 *   cn                                       → $lib/shared/css
 *   formatCalendarDate, todayLocalDateString → $lib/shared/format/date
 *   formatCurrency                           → $lib/shared/format/money
 */
export { cn } from './shared/css';
export { formatCalendarDate, todayLocalDateString } from './shared/format/date';
export { formatCurrency } from './shared/format/money';
