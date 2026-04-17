/**
 * Input validation utilities for candle-engine.
 *
 * Rules that are "obviously wrong" on the client side are enforced here so
 * invalid data is never sent to the backend.
 */

/**
 * The earliest date accepted as a valid range boundary.
 *
 * Must match the backend constant `MinDate` in
 * `backend/pkg/webserver/api/validation/validation.go`.
 */
export const MIN_DATE = new Date('1900-01-01T00:00:00.000Z');

export interface DateRangeValidation {
  fromError: string | null;
  toError: string | null;
}

/**
 * Parse a datetime-local string (YYYY-MM-DDTHH:mm) into a Date.
 * Returns null when the value is empty or unparseable.
 */
function parseDatetimeLocal(value: string): Date | null {
  if (!value) return null;
  const d = new Date(value);
  return isNaN(d.getTime()) ? null : d;
}

/**
 * Validate a date-range pair represented as datetime-local strings.
 *
 * Rules:
 *  - Each non-empty date must be on or after MIN_DATE (1900-01-01).
 *  - When both dates are present, `from` must not be later than `to`.
 *
 * @param from - datetime-local string for the range start (may be empty)
 * @param to   - datetime-local string for the range end (may be empty)
 * @returns    An object with `fromError` and `toError` (null when valid).
 */
export function validateDateRange(from: string, to: string): DateRangeValidation {
  const fromDate = parseDatetimeLocal(from);
  const toDate = parseDatetimeLocal(to);

  let fromError: string | null = null;
  let toError: string | null = null;

  if (from && fromDate === null) {
    fromError = 'Invalid date';
  } else if (fromDate !== null && fromDate < MIN_DATE) {
    fromError = 'Date must be on or after 1900-01-01';
  }

  if (to && toDate === null) {
    toError = 'Invalid date';
  } else if (toDate !== null && toDate < MIN_DATE) {
    toError = 'Date must be on or after 1900-01-01';
  }

  if (fromDate !== null && toDate !== null && fromDate > toDate) {
    fromError = fromError ?? '"From" must not be later than "To"';
    toError = toError ?? '"To" must not be earlier than "From"';
  }

  return { fromError, toError };
}

/** Returns true when the given DateRangeValidation has no errors. */
export function isDateRangeValid(validation: DateRangeValidation): boolean {
  return validation.fromError === null && validation.toError === null;
}
