/**
 * Currency and date formatting utilities
 */

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
];

const SHORT_MONTHS = [
  'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
  'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
];

/**
 * Formats CAD amount with currency symbol
 */
export function formatCAD(amount: number): string {
  if (amount >= 0) {
    return `$${amount.toFixed(2)}`;
  }
  return `-$${(-amount).toFixed(2)}`;
}

/**
 * Formats IDR amount with currency symbol
 */
export function formatIDR(amount: number): string {
  const rounded = Math.round(amount);
  if (rounded >= 0) {
    return `Rp ${rounded.toLocaleString('en-US')}`;
  }
  return `-Rp ${(-rounded).toLocaleString('en-US')}`;
}

/**
 * Formats date as YYYY-MM-DD (for CSV storage)
 */
export function formatDateISO(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

/**
 * Formats date as MM/DD/YYYY (for display)
 */
export function formatDateDisplay(date: Date): string {
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  const y = date.getFullYear();
  return `${m}/${d}/${y}`;
}

/**
 * Formats date as "January 2, 2025" (human-readable)
 */
export function formatDateHuman(date: Date): string {
  return `${MONTHS[date.getMonth()]} ${date.getDate()}, ${date.getFullYear()}`;
}

/**
 * Formats date as "Jan 2" (short format)
 */
export function formatDateShort(date: Date): string {
  return `${SHORT_MONTHS[date.getMonth()]} ${date.getDate()}`;
}

/**
 * Formats date as "Jan 2, 2025 at 3:04 PM"
 */
export function formatDateTimeHuman(date: Date): string {
  const hours = date.getHours();
  const minutes = date.getMinutes();
  const ampm = hours >= 12 ? 'PM' : 'AM';
  const hour12 = hours % 12 || 12;
  const minuteStr = String(minutes).padStart(2, '0');
  return `${SHORT_MONTHS[date.getMonth()]} ${date.getDate()}, ${date.getFullYear()} at ${hour12}:${minuteStr} ${ampm}`;
}

/**
 * Formats a date range for display
 */
export function formatDateRange(start: Date, end: Date): string {
  const startStr = formatDateDisplay(start);
  const endStr = formatDateDisplay(end);

  // Same day
  if (start.getFullYear() === end.getFullYear() &&
      start.getMonth() === end.getMonth() &&
      start.getDate() === end.getDate()) {
    return startStr;
  }

  return `${startStr} - ${endStr}`;
}

/**
 * Truncates a string and adds ellipsis if needed
 */
export function truncate(s: string, maxLen: number): string {
  if (s.length <= maxLen) return s;
  return s.slice(0, maxLen - 3) + '...';
}

/**
 * Pads a string to a fixed width (for table alignment)
 */
export function padRight(s: string, width: number): string {
  if (s.length >= width) return s;
  return s + ' '.repeat(width - s.length);
}

/**
 * Pads a string to a fixed width (right-aligned)
 */
export function padLeft(s: string, width: number): string {
  if (s.length >= width) return s;
  return ' '.repeat(width - s.length) + s;
}

/**
 * Formats screen time duration (e.g., "3h45m")
 */
export function formatScreenTime(input: string): string {
  // Already formatted properly
  if (/^\d+h\d+m$/.test(input) || /^\d+h$/.test(input) || /^\d+m$/.test(input)) {
    return input;
  }
  return input;
}

/**
 * Parses screen time string to minutes
 */
export function parseScreenTime(input: string): number {
  const hourMatch = input.match(/(\d+)h/);
  const minMatch = input.match(/(\d+)m/);

  let minutes = 0;
  if (hourMatch) {
    minutes += parseInt(hourMatch[1], 10) * 60;
  }
  if (minMatch) {
    minutes += parseInt(minMatch[1], 10);
  }

  return minutes;
}
