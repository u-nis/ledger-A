import type { Entry } from './entry';
import { cloneEntry } from './entry';

/**
 * Day represents all entries for a single day
 */
export interface Day {
  date: Date;
  entries: Entry[];
  screenTime: string;
  journal: string; // Markdown journal entry for the day
}

/**
 * Creates a new Day instance
 */
export function createDay(date: Date): Day {
  return {
    date,
    entries: [],
    screenTime: '',
    journal: '',
  };
}

/**
 * Checks if the day has a journal entry
 */
export function hasJournal(day: Day): boolean {
  return day.journal !== '';
}

/**
 * Adds an entry to the day
 */
export function addEntry(day: Day, entry: Entry): void {
  entry.screenTime = day.screenTime;
  day.entries.push(entry);
}

/**
 * Removes an entry by ID, returns the removed entry or undefined
 */
export function removeEntry(day: Day, id: string): Entry | undefined {
  const index = day.entries.findIndex(e => e.id === id);
  if (index === -1) return undefined;
  const [removed] = day.entries.splice(index, 1);
  return removed;
}

/**
 * Gets an entry by ID
 */
export function getEntry(day: Day, id: string): Entry | undefined {
  return day.entries.find(e => e.id === id);
}

/**
 * Updates an existing entry
 */
export function updateEntry(day: Day, entry: Entry): boolean {
  const index = day.entries.findIndex(e => e.id === entry.id);
  if (index === -1) return false;
  day.entries[index] = entry;
  return true;
}

/**
 * Sets the screen time for all entries in the day
 */
export function setScreenTime(day: Day, screenTime: string): void {
  day.screenTime = screenTime;
  for (const entry of day.entries) {
    entry.screenTime = screenTime;
  }
}

/**
 * Returns the sum of all CAD amounts
 */
export function totalCAD(day: Day): number {
  return day.entries.reduce((sum, e) => sum + e.cad, 0);
}

/**
 * Returns the sum of all IDR amounts
 */
export function totalIDR(day: Day): number {
  return day.entries.reduce((sum, e) => sum + e.idr, 0);
}

/**
 * Checks if an entry matches the search query (vim-style)
 */
export function entryMatchesQuery(entry: Entry, query: string): boolean {
  if (!query) return true;

  const lowerQuery = query.toLowerCase();

  // Search in description
  if (entry.description.toLowerCase().includes(lowerQuery)) {
    return true;
  }

  // Search in date (multiple formats)
  const dateFormats = [
    formatDate(entry.date, 'MM/DD/YYYY'),
    formatDate(entry.date, 'M/D/YYYY'),
    formatDate(entry.date, 'MMMM D, YYYY'),
    formatDate(entry.date, 'MMM D'),
  ];
  for (const dateStr of dateFormats) {
    if (dateStr.toLowerCase().includes(lowerQuery)) {
      return true;
    }
  }

  // Search in CAD amount
  if (entry.cad.toFixed(2).includes(lowerQuery)) {
    return true;
  }

  // Search in IDR amount
  if (Math.round(entry.idr).toString().includes(lowerQuery)) {
    return true;
  }

  return false;
}

/**
 * Filters entries matching the search query
 */
export function filterEntries(day: Day, query: string): Entry[] {
  if (!query) return day.entries;
  return day.entries.filter(e => entryMatchesQuery(e, query));
}

/**
 * Returns the sum of CAD for filtered entries
 */
export function filteredTotalCAD(day: Day, query: string): number {
  return filterEntries(day, query).reduce((sum, e) => sum + e.cad, 0);
}

/**
 * Returns the sum of IDR for filtered entries
 */
export function filteredTotalIDR(day: Day, query: string): number {
  return filterEntries(day, query).reduce((sum, e) => sum + e.idr, 0);
}

/**
 * Returns true if the day has no entries and no journal
 */
export function isEmpty(day: Day): boolean {
  return day.entries.length === 0 && day.journal === '';
}

/**
 * Creates a deep copy of a day
 */
export function cloneDay(day: Day): Day {
  return {
    date: new Date(day.date),
    entries: day.entries.map(cloneEntry),
    screenTime: day.screenTime,
    journal: day.journal,
  };
}

// Helper function for date formatting (will be replaced by utils/format.ts)
function formatDate(date: Date, format: string): string {
  const months = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December',
  ];
  const shortMonths = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

  const m = date.getMonth();
  const d = date.getDate();
  const y = date.getFullYear();

  switch (format) {
    case 'MM/DD/YYYY':
      return `${String(m + 1).padStart(2, '0')}/${String(d).padStart(2, '0')}/${y}`;
    case 'M/D/YYYY':
      return `${m + 1}/${d}/${y}`;
    case 'MMMM D, YYYY':
      return `${months[m]} ${d}, ${y}`;
    case 'MMM D':
      return `${shortMonths[m]} ${d}`;
    default:
      return date.toISOString();
  }
}
