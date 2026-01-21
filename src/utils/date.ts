/**
 * Date parsing utilities
 * Matches the Go version's flexible date parsing
 */

export interface ParseResult {
  startDate: Date;
  endDate: Date;
  isRange: boolean;
  isMonth: boolean;
  valid: boolean;
  error: string;
}

const monthNames: Record<string, number> = {
  jan: 0, january: 0,
  feb: 1, february: 1,
  mar: 2, march: 2,
  apr: 3, april: 3,
  may: 4,
  jun: 5, june: 5,
  jul: 6, july: 6,
  aug: 7, august: 7,
  sep: 8, september: 8,
  oct: 9, october: 9,
  nov: 10, november: 10,
  dec: 11, december: 11,
};

/**
 * Attempts to parse a month name (full or abbreviated)
 */
function parseMonth(s: string): { month: number; ok: boolean } {
  s = s.toLowerCase().trim();
  if (s in monthNames) {
    return { month: monthNames[s], ok: true };
  }
  // Try prefix matching for partial names
  for (const [name, m] of Object.entries(monthNames)) {
    if (s.length >= 3 && name.startsWith(s)) {
      return { month: m, ok: true };
    }
  }
  return { month: 0, ok: false };
}

/**
 * Creates a Date object at midnight local time
 */
function createDate(year: number, month: number, day: number): Date {
  return new Date(year, month, day, 0, 0, 0, 0);
}

/**
 * Gets the last day of a month
 */
function lastDayOfMonth(year: number, month: number): Date {
  // Month+1, day 0 gives last day of previous month
  return new Date(year, month + 1, 0, 0, 0, 0, 0);
}

/**
 * Parses flexible date input and returns a ParseResult
 * Supported formats:
 * - Single dates: "jan 5 2025", "jan,5,2025", "01/05/2025", "1/5/2025"
 * - Date ranges: "jan 5 - jan 20", "jan 5 to jan 20", "1/5 - 1/20"
 * - Whole months: "jan", "january", "jan 2025"
 * - Month ranges: "jan - feb", "jan - mar 2025"
 */
export function parseDateInput(input: string, refYear: number): ParseResult {
  input = input.trim();
  if (!input) {
    return { startDate: new Date(), endDate: new Date(), isRange: false, isMonth: false, valid: false, error: 'empty input' };
  }

  // Check if it's a range (contains " - " or " to ")
  let parts: string[] = [];
  if (input.includes(' - ')) {
    parts = input.split(' - ').map(s => s.trim());
  } else if (input.toLowerCase().includes(' to ')) {
    parts = input.toLowerCase().split(' to ').map(s => s.trim());
  } else if (input.includes('-') && !input.includes('/')) {
    // Handle "jan-feb" without spaces
    parts = input.split('-').map(s => s.trim());
  }

  if (parts.length === 2) {
    // Parse as range
    const start = parseSingleDate(parts[0], refYear);
    const end = parseSingleDate(parts[1], refYear);

    if (!start.valid) {
      return { ...start, error: 'invalid start date: ' + start.error };
    }
    if (!end.valid) {
      return { ...end, error: 'invalid end date: ' + end.error };
    }

    let startDate = start.startDate;
    let endDate = end.isMonth ? end.endDate : end.startDate;

    // Swap if end is before start
    if (endDate < startDate) {
      [startDate, endDate] = [endDate, startDate];
    }

    return {
      startDate,
      endDate,
      isRange: true,
      isMonth: false,
      valid: true,
      error: '',
    };
  }

  // Parse as single date or month
  return parseSingleDate(input, refYear);
}

/**
 * Parses a single date or month
 */
function parseSingleDate(input: string, refYear: number): ParseResult {
  input = input.trim();

  // Try numeric formats first: MM/DD/YYYY, M/D/YYYY, MM/DD, M/D
  const numericResult = parseNumericDate(input, refYear);
  if (numericResult.valid) {
    return numericResult;
  }

  // Try month name formats: "jan 5 2025", "jan 5", "jan,5,2025", "jan 2025", "jan"
  const textResult = parseTextDate(input, refYear);
  if (textResult.valid) {
    return textResult;
  }

  return { startDate: new Date(), endDate: new Date(), isRange: false, isMonth: false, valid: false, error: 'unrecognized format' };
}

/**
 * Parses numeric date formats
 */
function parseNumericDate(input: string, refYear: number): ParseResult {
  const invalid: ParseResult = { startDate: new Date(), endDate: new Date(), isRange: false, isMonth: false, valid: false, error: '' };

  // Pattern: MM/DD/YYYY or M/D/YYYY
  const fullPattern = /^(\d{1,2})\/(\d{1,2})\/(\d{4})$/;
  let matches = input.match(fullPattern);
  if (matches) {
    const month = parseInt(matches[1], 10);
    const day = parseInt(matches[2], 10);
    const year = parseInt(matches[3], 10);

    if (month < 1 || month > 12) {
      return { ...invalid, error: 'invalid month' };
    }

    const date = createDate(year, month - 1, day);
    return {
      startDate: date,
      endDate: date,
      isRange: false,
      isMonth: false,
      valid: true,
      error: '',
    };
  }

  // Pattern: MM/DD or M/D (use refYear)
  const shortPattern = /^(\d{1,2})\/(\d{1,2})$/;
  matches = input.match(shortPattern);
  if (matches) {
    const month = parseInt(matches[1], 10);
    const day = parseInt(matches[2], 10);

    if (month < 1 || month > 12) {
      return { ...invalid, error: 'invalid month' };
    }

    const date = createDate(refYear, month - 1, day);
    return {
      startDate: date,
      endDate: date,
      isRange: false,
      isMonth: false,
      valid: true,
      error: '',
    };
  }

  return invalid;
}

/**
 * Parses text-based date formats with month names
 */
function parseTextDate(input: string, refYear: number): ParseResult {
  const invalid: ParseResult = { startDate: new Date(), endDate: new Date(), isRange: false, isMonth: false, valid: false, error: '' };

  // Try slash format with month name first: "Jan/02/2025" or "Jan/02"
  if (input.includes('/')) {
    const parts = input.split('/');
    if (parts.length >= 2) {
      const { month, ok } = parseMonth(parts[0]);
      if (ok) {
        const day = parseInt(parts[1], 10);
        if (!isNaN(day) && day >= 1 && day <= 31) {
          let year = refYear;
          if (parts.length >= 3) {
            const y = parseInt(parts[2], 10);
            if (!isNaN(y) && y >= 1900 && y <= 2100) {
              year = y;
            }
          }
          const date = createDate(year, month, day);
          return {
            startDate: date,
            endDate: date,
            isRange: false,
            isMonth: false,
            valid: true,
            error: '',
          };
        }
      }
    }
  }

  // Normalize separators (replace commas with spaces)
  input = input.replace(/,/g, ' ').toLowerCase();

  // Split into tokens
  const tokens = input.split(/\s+/).filter(t => t.length > 0);
  if (tokens.length === 0) {
    return { ...invalid, error: 'empty input' };
  }

  // First token should be a month
  const { month, ok } = parseMonth(tokens[0]);
  if (!ok) {
    return { ...invalid, error: 'expected month name' };
  }

  // Just month name: "jan" or "january"
  if (tokens.length === 1) {
    const start = createDate(refYear, month, 1);
    const end = lastDayOfMonth(refYear, month);
    return {
      startDate: start,
      endDate: end,
      isRange: true,
      isMonth: true,
      valid: true,
      error: '',
    };
  }

  // Check if second token is a year: "jan 2025"
  if (tokens.length === 2) {
    const year = parseInt(tokens[1], 10);
    if (!isNaN(year) && year >= 1900 && year <= 2100) {
      const start = createDate(year, month, 1);
      const end = lastDayOfMonth(year, month);
      return {
        startDate: start,
        endDate: end,
        isRange: true,
        isMonth: true,
        valid: true,
        error: '',
      };
    }

    // Second token is a day: "jan 5"
    const day = parseInt(tokens[1], 10);
    if (!isNaN(day) && day >= 1 && day <= 31) {
      const date = createDate(refYear, month, day);
      return {
        startDate: date,
        endDate: date,
        isRange: false,
        isMonth: false,
        valid: true,
        error: '',
      };
    }
  }

  // Month, day, year: "jan 5 2025"
  if (tokens.length >= 3) {
    const day = parseInt(tokens[1], 10);
    const year = parseInt(tokens[2], 10);

    if (!isNaN(day) && !isNaN(year) && day >= 1 && day <= 31 && year >= 1900 && year <= 2100) {
      const date = createDate(year, month, day);
      return {
        startDate: date,
        endDate: date,
        isRange: false,
        isMonth: false,
        valid: true,
        error: '',
      };
    }
  }

  return { ...invalid, error: 'unrecognized format' };
}

/**
 * Parses an ISO date string (YYYY-MM-DD) to a Date
 */
export function parseISODate(dateStr: string): Date | null {
  const match = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!match) return null;

  const year = parseInt(match[1], 10);
  const month = parseInt(match[2], 10) - 1;
  const day = parseInt(match[3], 10);

  return createDate(year, month, day);
}

/**
 * Calculates the number of days between two dates (inclusive)
 */
export function daysBetween(start: Date, end: Date): number {
  const startMs = new Date(start.getFullYear(), start.getMonth(), start.getDate()).getTime();
  const endMs = new Date(end.getFullYear(), end.getMonth(), end.getDate()).getTime();

  const diffMs = Math.abs(endMs - startMs);
  return Math.floor(diffMs / (24 * 60 * 60 * 1000)) + 1;
}

/**
 * Gets the start of day (midnight)
 */
export function startOfDay(date: Date): Date {
  return createDate(date.getFullYear(), date.getMonth(), date.getDate());
}

/**
 * Checks if two dates are the same day
 */
export function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() &&
         a.getMonth() === b.getMonth() &&
         a.getDate() === b.getDate();
}

/**
 * Adds days to a date
 */
export function addDays(date: Date, days: number): Date {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
}

/**
 * Gets today's date at midnight
 */
export function today(): Date {
  const now = new Date();
  return createDate(now.getFullYear(), now.getMonth(), now.getDate());
}
