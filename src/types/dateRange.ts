import type { Day } from './day';
import type { Entry } from './entry';
import { filterEntries, filteredTotalCAD, filteredTotalIDR } from './day';

/**
 * DateRange represents a range of days
 */
export interface DateRange {
  start: Date;
  end: Date;
  days: Day[];
}

/**
 * Creates a new DateRange
 */
export function createDateRange(start: Date, end: Date): DateRange {
  return {
    start,
    end,
    days: [],
  };
}

/**
 * Adds a day to the range, maintaining chronological order
 */
export function addDay(range: DateRange, day: Day): void {
  range.days.push(day);
  range.days.sort((a, b) => a.date.getTime() - b.date.getTime());
}

/**
 * Returns the sum of CAD for all days in the range
 */
export function totalCAD(range: DateRange): number {
  return range.days.reduce((sum, day) => {
    return sum + day.entries.reduce((daySum, e) => daySum + e.cad, 0);
  }, 0);
}

/**
 * Returns the sum of IDR for all days in the range
 */
export function totalIDR(range: DateRange): number {
  return range.days.reduce((sum, day) => {
    return sum + day.entries.reduce((daySum, e) => daySum + e.idr, 0);
  }, 0);
}

/**
 * Returns all entries from all days, optionally filtered
 */
export function allEntries(range: DateRange, query: string = ''): Entry[] {
  const entries: Entry[] = [];
  for (const day of range.days) {
    entries.push(...filterEntries(day, query));
  }
  return entries;
}

/**
 * Returns the sum of CAD for filtered entries across all days
 */
export function rangeFilteredTotalCAD(range: DateRange, query: string): number {
  return range.days.reduce((sum, day) => sum + filteredTotalCAD(day, query), 0);
}

/**
 * Returns the sum of IDR for filtered entries across all days
 */
export function rangeFilteredTotalIDR(range: DateRange, query: string): number {
  return range.days.reduce((sum, day) => sum + filteredTotalIDR(day, query), 0);
}

/**
 * Returns a human-readable range format
 */
export function formatRangeDisplay(range: DateRange): string {
  const formatDate = (d: Date) => {
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const y = d.getFullYear();
    return `${m}/${day}/${y}`;
  };
  return `${formatDate(range.start)} - ${formatDate(range.end)}`;
}

/**
 * Checks if a day matches the query (for filtering entire days)
 */
export function dayMatchesQuery(day: Day, query: string): boolean {
  if (!query) return true;

  const lowerQuery = query.toLowerCase();

  // Check date formats
  const formatDate = (d: Date, format: string) => {
    const months = [
      'January', 'February', 'March', 'April', 'May', 'June',
      'July', 'August', 'September', 'October', 'November', 'December',
    ];
    const m = d.getMonth();
    const dayNum = d.getDate();
    const y = d.getFullYear();

    switch (format) {
      case 'MM/DD/YYYY':
        return `${String(m + 1).padStart(2, '0')}/${String(dayNum).padStart(2, '0')}/${y}`;
      case 'M/D/YYYY':
        return `${m + 1}/${dayNum}/${y}`;
      case 'MMMM D, YYYY':
        return `${months[m]} ${dayNum}, ${y}`;
      default:
        return '';
    }
  };

  const dateFormats = [
    formatDate(day.date, 'MM/DD/YYYY'),
    formatDate(day.date, 'M/D/YYYY'),
    formatDate(day.date, 'MMMM D, YYYY'),
  ];

  for (const dateStr of dateFormats) {
    if (dateStr.toLowerCase().includes(lowerQuery)) {
      return true;
    }
  }

  // Check screen time
  if (day.screenTime.toLowerCase().includes(lowerQuery)) {
    return true;
  }

  // Check if any entry matches
  for (const entry of day.entries) {
    if (entry.description.toLowerCase().includes(lowerQuery)) {
      return true;
    }
  }

  return false;
}
