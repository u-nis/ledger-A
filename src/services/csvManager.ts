/**
 * CSV Manager - handles CSV and journal file I/O
 * Maintains compatibility with the Go version's data format
 */

import { join } from 'path';
import type { Day, DateRange } from '../types';
import { createDay, isEmpty } from '../types/day';
import { createEntry } from '../types/entry';
import { createDateRange, addDay as addDayToRange } from '../types/dateRange';
import { formatDateISO, parseISODate, addDays } from '../utils';

const DATA_DIR = 'ledger-data';
const CSV_HEADER = 'date,description,cad,idr,screen_time';
const CSV_FILE_NAME = 'data.csv';
const JOURNAL_FILE_NAME = 'entry.md';

export class CSVManager {
  private dataDir: string;

  constructor(dataDir: string = DATA_DIR) {
    this.dataDir = dataDir;
  }

  /**
   * Returns the directory path for a specific date (YYYY/MM/DD)
   */
  getDayDir(date: Date): string {
    const y = date.getFullYear().toString();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    return join(this.dataDir, y, m, d);
  }

  /**
   * Creates the base data directory if it doesn't exist
   */
  async ensureDataDir(): Promise<void> {
    await Bun.write(join(this.dataDir, '.gitkeep'), '');
  }

  /**
   * Creates the day directory if it doesn't exist
   */
  async ensureDayDir(date: Date): Promise<void> {
    const dirPath = this.getDayDir(date);
    // Bun.write creates parent directories automatically
    const keepFile = join(dirPath, '.keep');
    const file = Bun.file(keepFile);
    if (!(await file.exists())) {
      await Bun.write(keepFile, '');
    }
  }

  /**
   * Returns the path for a specific date's CSV file
   */
  getFilePath(date: Date): string {
    return join(this.getDayDir(date), CSV_FILE_NAME);
  }

  /**
   * Returns the path for a specific date's journal file
   */
  getJournalPath(date: Date): string {
    return join(this.getDayDir(date), JOURNAL_FILE_NAME);
  }

  /**
   * Checks if a CSV file exists for a given date
   */
  async fileExists(date: Date): Promise<boolean> {
    const file = Bun.file(this.getFilePath(date));
    return await file.exists();
  }

  /**
   * Checks if a journal file exists for a given date
   */
  async journalExists(date: Date): Promise<boolean> {
    const file = Bun.file(this.getJournalPath(date));
    return await file.exists();
  }

  /**
   * Loads the journal entry for a specific date
   */
  async loadJournal(date: Date): Promise<string> {
    const file = Bun.file(this.getJournalPath(date));
    if (!(await file.exists())) {
      return '';
    }
    return await file.text();
  }

  /**
   * Saves a journal entry for a specific date
   */
  async saveJournal(date: Date, content: string): Promise<void> {
    await this.ensureDayDir(date);
    await Bun.write(this.getJournalPath(date), content);
  }

  /**
   * Deletes the journal file for a specific date
   */
  async deleteJournal(date: Date): Promise<void> {
    const path = this.getJournalPath(date);
    const file = Bun.file(path);
    if (await file.exists()) {
      const { unlink } = await import('fs/promises');
      await unlink(path);
    }
  }

  /**
   * Checks if a date has either CSV or journal data
   */
  async dayHasData(date: Date): Promise<boolean> {
    const [csvExists, journalExists] = await Promise.all([
      this.fileExists(date),
      this.journalExists(date),
    ]);
    return csvExists || journalExists;
  }

  /**
   * Parses a CSV line, handling quoted fields
   */
  private parseCSVLine(line: string): string[] {
    const fields: string[] = [];
    let current = '';
    let inQuotes = false;

    for (let i = 0; i < line.length; i++) {
      const char = line[i];
      if (char === '"') {
        if (inQuotes && line[i + 1] === '"') {
          current += '"';
          i++;
        } else {
          inQuotes = !inQuotes;
        }
      } else if (char === ',' && !inQuotes) {
        fields.push(current);
        current = '';
      } else {
        current += char;
      }
    }
    fields.push(current);
    return fields;
  }

  /**
   * Escapes a field for CSV output
   */
  private escapeCSVField(field: string): string {
    if (field.includes(',') || field.includes('"') || field.includes('\n')) {
      return '"' + field.replace(/"/g, '""') + '"';
    }
    return field;
  }

  /**
   * Loads entries from a CSV file for a specific date
   */
  async loadDay(date: Date): Promise<Day> {
    const day = createDay(date);

    // Load CSV data
    const file = Bun.file(this.getFilePath(date));
    if (await file.exists()) {
      const content = await file.text();
      const lines = content.split('\n').filter(line => line.trim() !== '');

      // Skip header row
      for (let i = 1; i < lines.length; i++) {
        const record = this.parseCSVLine(lines[i]);
        if (record.length < 5) continue; // Skip malformed rows

        const entryDate = parseISODate(record[0]);
        if (!entryDate) continue; // Skip rows with invalid dates

        const cad = parseFloat(record[2]) || 0;
        const idr = parseFloat(record[3]) || 0;
        const screenTime = record[4] || '';

        const entry = createEntry(entryDate, record[1], cad, idr, screenTime);
        day.entries.push(entry);

        // Set screen time from first entry (all entries have same screen time)
        if (day.screenTime === '' && screenTime !== '') {
          day.screenTime = screenTime;
          for (const e of day.entries) {
            e.screenTime = screenTime;
          }
        }
      }
    }

    // Load journal if it exists
    day.journal = await this.loadJournal(date);

    return day;
  }

  /**
   * Saves a day's entries to a CSV file
   */
  async saveDay(day: Day): Promise<void> {
    await this.ensureDayDir(day.date);

    // Only create CSV if there are entries
    if (day.entries.length > 0) {
      const lines: string[] = [CSV_HEADER];

      for (const entry of day.entries) {
        const record = [
          formatDateISO(entry.date),
          this.escapeCSVField(entry.description),
          entry.cad.toFixed(2),
          Math.round(entry.idr).toString(),
          day.screenTime,
        ];
        lines.push(record.join(','));
      }

      await Bun.write(this.getFilePath(day.date), lines.join('\n') + '\n');
    }

    // Save journal if it exists
    if (day.journal !== '') {
      await this.saveJournal(day.date, day.journal);
    }
  }

  /**
   * Deletes the CSV file for a specific date
   */
  async deleteDay(date: Date): Promise<void> {
    const path = this.getFilePath(date);
    const file = Bun.file(path);
    if (await file.exists()) {
      const { unlink } = await import('fs/promises');
      await unlink(path);
    }
  }

  /**
   * Loads all days within a date range
   */
  async loadDateRange(start: Date, end: Date): Promise<DateRange> {
    const dateRange = createDateRange(start, end);

    // Iterate through each day in the range
    let current = new Date(start);
    while (current <= end) {
      if (await this.fileExists(current)) {
        const day = await this.loadDay(current);
        if (!isEmpty(day)) {
          addDayToRange(dateRange, day);
        }
      }
      current = addDays(current, 1);
    }

    return dateRange;
  }

  /**
   * Exports a date range to a new CSV file
   */
  async exportDateRange(dateRange: DateRange, filename: string): Promise<void> {
    await this.ensureDataDir();

    const lines: string[] = [CSV_HEADER];

    for (const day of dateRange.days) {
      for (const entry of day.entries) {
        const record = [
          formatDateISO(entry.date),
          this.escapeCSVField(entry.description),
          entry.cad.toFixed(2),
          Math.round(entry.idr).toString(),
          day.screenTime,
        ];
        lines.push(record.join(','));
      }
    }

    await Bun.write(join(this.dataDir, filename), lines.join('\n') + '\n');
  }

  /**
   * Lists all dates that have data (CSV or journal)
   */
  async listAvailableDates(): Promise<Date[]> {
    const dates: Date[] = [];
    const { readdir } = await import('fs/promises');

    try {
      const years = await readdir(this.dataDir);

      for (const year of years) {
        if (!/^\d{4}$/.test(year)) continue;

        const yearPath = join(this.dataDir, year);
        let months: string[];
        try {
          months = await readdir(yearPath);
        } catch {
          continue;
        }

        for (const month of months) {
          if (!/^\d{2}$/.test(month)) continue;

          const monthPath = join(yearPath, month);
          let days: string[];
          try {
            days = await readdir(monthPath);
          } catch {
            continue;
          }

          for (const day of days) {
            if (!/^\d{2}$/.test(day)) continue;

            // Parse the date from directory structure
            const dateStr = `${year}-${month}-${day}`;
            const date = parseISODate(dateStr);
            if (!date) continue;

            // Check if this day has any data
            if (await this.dayHasData(date)) {
              dates.push(date);
            }
          }
        }
      }
    } catch {
      // Data directory doesn't exist yet
      return [];
    }

    return dates.sort((a, b) => a.getTime() - b.getTime());
  }

  /**
   * Returns the data directory path
   */
  getDataDir(): string {
    return this.dataDir;
  }
}

// Singleton instance
export const csvManager = new CSVManager();
