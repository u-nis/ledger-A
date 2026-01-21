/**
 * Ledger Service - Core CRUD operations for days/entries
 */

import type { Day, Entry, DateRange } from '../types';
import { createDay, addEntry, removeEntry, getEntry, updateEntry, setScreenTime } from '../types/day';
import { createEntry } from '../types/entry';
import { CSVManager, csvManager } from './csvManager';

export class LedgerService {
  private csv: CSVManager;
  private dayCache: Map<string, Day> = new Map();

  constructor(csvManager?: CSVManager) {
    this.csv = csvManager || new CSVManager();
  }

  /**
   * Gets the cache key for a date
   */
  private getCacheKey(date: Date): string {
    return `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
  }

  /**
   * Gets or loads a day from cache/disk
   */
  async getDay(date: Date): Promise<Day> {
    const key = this.getCacheKey(date);

    // Check cache first
    const cached = this.dayCache.get(key);
    if (cached) {
      return cached;
    }

    // Load from disk
    const day = await this.csv.loadDay(date);
    this.dayCache.set(key, day);
    return day;
  }

  /**
   * Saves a day to disk and updates cache
   */
  async saveDay(day: Day): Promise<void> {
    const key = this.getCacheKey(day.date);
    this.dayCache.set(key, day);
    await this.csv.saveDay(day);
  }

  /**
   * Creates a new day if it doesn't exist
   */
  async createDay(date: Date): Promise<Day> {
    const key = this.getCacheKey(date);

    // Check if already exists
    const existing = this.dayCache.get(key);
    if (existing) {
      return existing;
    }

    // Check disk
    if (await this.csv.dayHasData(date)) {
      return await this.getDay(date);
    }

    // Create new day
    const day = createDay(date);
    this.dayCache.set(key, day);
    return day;
  }

  /**
   * Adds an entry to a day
   */
  async addEntryToDay(
    date: Date,
    description: string,
    cad: number,
    idr: number,
    screenTime: string = ''
  ): Promise<Entry> {
    const day = await this.getDay(date);
    const entry = createEntry(date, description, cad, idr, screenTime || day.screenTime);
    addEntry(day, entry);
    await this.saveDay(day);
    return entry;
  }

  /**
   * Removes an entry from a day
   */
  async removeEntryFromDay(date: Date, entryId: string): Promise<Entry | undefined> {
    const day = await this.getDay(date);
    const removed = removeEntry(day, entryId);
    if (removed) {
      await this.saveDay(day);
    }
    return removed;
  }

  /**
   * Updates an entry in a day
   */
  async updateEntryInDay(date: Date, entry: Entry): Promise<boolean> {
    const day = await this.getDay(date);
    const success = updateEntry(day, entry);
    if (success) {
      await this.saveDay(day);
    }
    return success;
  }

  /**
   * Gets an entry by ID from a day
   */
  async getEntryFromDay(date: Date, entryId: string): Promise<Entry | undefined> {
    const day = await this.getDay(date);
    return getEntry(day, entryId);
  }

  /**
   * Sets screen time for a day
   */
  async setDayScreenTime(date: Date, screenTime: string): Promise<void> {
    const day = await this.getDay(date);
    setScreenTime(day, screenTime);
    await this.saveDay(day);
  }

  /**
   * Sets journal for a day
   */
  async setDayJournal(date: Date, journal: string): Promise<void> {
    const day = await this.getDay(date);
    day.journal = journal;
    await this.saveDay(day);
  }

  /**
   * Loads a date range
   */
  async getDateRange(start: Date, end: Date): Promise<DateRange> {
    return await this.csv.loadDateRange(start, end);
  }

  /**
   * Lists all available dates with data
   */
  async listAvailableDates(): Promise<Date[]> {
    return await this.csv.listAvailableDates();
  }

  /**
   * Checks if a date has data
   */
  async dateHasData(date: Date): Promise<boolean> {
    return await this.csv.dayHasData(date);
  }

  /**
   * Clears the day cache
   */
  clearCache(): void {
    this.dayCache.clear();
  }

  /**
   * Invalidates cache for a specific date
   */
  invalidateCache(date: Date): void {
    const key = this.getCacheKey(date);
    this.dayCache.delete(key);
  }

  /**
   * Gets the CSV manager for direct access if needed
   */
  getCSVManager(): CSVManager {
    return this.csv;
  }
}

// Singleton instance
export const ledgerService = new LedgerService(csvManager);
