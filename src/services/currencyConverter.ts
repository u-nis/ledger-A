/**
 * Currency Converter - API client with caching for CAD/IDR conversion
 */

import { join } from 'path';
import { readFileSync } from 'fs';
import { formatDateTimeHuman } from '../utils/format';

const CACHE_FILE_NAME = '.rate_cache.json';
const DEFAULT_CAD_TO_IDR = 11800.0;
const API_BASE_URL = 'https://api.frankfurter.app';

interface RateCache {
  cad_to_idr: number;
  last_updated: string; // ISO date string
}

export class CurrencyConverter {
  private cacheDir: string;
  private cache: RateCache;
  private offline: boolean = false;
  private lastError: Error | null = null;

  constructor(cacheDir: string = 'ledger-data') {
    this.cacheDir = cacheDir;
    this.cache = {
      cad_to_idr: DEFAULT_CAD_TO_IDR,
      last_updated: '',
    };
    this.loadCacheSync();
  }

  /**
   * Gets the full path to the cache file
   */
  private getCachePath(): string {
    return join(this.cacheDir, CACHE_FILE_NAME);
  }

  /**
   * Loads the cached rate from disk (sync for constructor)
   */
  private loadCacheSync(): void {
    try {
      // Use synchronous file read for constructor
      const text = readFileSync(this.getCachePath(), 'utf-8');
      const data = JSON.parse(text) as RateCache;
      this.cache = data;
    } catch {
      // No cache exists, use default
      this.cache = {
        cad_to_idr: DEFAULT_CAD_TO_IDR,
        last_updated: '',
      };
    }
  }

  /**
   * Loads the cached rate from disk
   */
  async loadCache(): Promise<void> {
    try {
      const file = Bun.file(this.getCachePath());
      if (await file.exists()) {
        const text = await file.text();
        const data = JSON.parse(text) as RateCache;
        this.cache = data;
      }
    } catch {
      // No cache exists, use default
    }
  }

  /**
   * Saves the current rate to disk
   */
  async saveCache(): Promise<void> {
    await Bun.write(this.getCachePath(), JSON.stringify(this.cache, null, 2));
  }

  /**
   * Fetches the latest CAD to IDR rate from the API
   */
  async fetchCADToIDR(): Promise<number> {
    try {
      const response = await fetch(`${API_BASE_URL}/latest?from=CAD&to=IDR`);
      if (!response.ok) {
        throw new Error(`API returned ${response.status}`);
      }

      const data = await response.json() as { rates: { IDR: number } };
      return data.rates.IDR;
    } catch (error) {
      throw error;
    }
  }

  /**
   * Refreshes the exchange rate from the API
   */
  async refreshRate(): Promise<void> {
    try {
      const rate = await this.fetchCADToIDR();
      this.cache = {
        cad_to_idr: rate,
        last_updated: new Date().toISOString(),
      };
      this.offline = false;
      this.lastError = null;
      await this.saveCache();
    } catch (error) {
      this.offline = true;
      this.lastError = error instanceof Error ? error : new Error(String(error));
      throw error;
    }
  }

  /**
   * Gets the current CAD to IDR rate
   */
  getCADToIDRRate(): number {
    return this.cache.cad_to_idr;
  }

  /**
   * Gets the current IDR to CAD rate
   */
  getIDRToCADRate(): number {
    if (this.cache.cad_to_idr === 0) {
      return 0;
    }
    return 1.0 / this.cache.cad_to_idr;
  }

  /**
   * Converts CAD to IDR
   */
  cadToIDR(cad: number): number {
    return cad * this.cache.cad_to_idr;
  }

  /**
   * Converts IDR to CAD
   */
  idrToCAD(idr: number): number {
    if (this.cache.cad_to_idr === 0) {
      return 0;
    }
    return idr / this.cache.cad_to_idr;
  }

  /**
   * Returns true if the last API call failed
   */
  isOffline(): boolean {
    return this.offline;
  }

  /**
   * Returns the last error from the API
   */
  getLastError(): Error | null {
    return this.lastError;
  }

  /**
   * Returns when the rate was last updated
   */
  getLastUpdated(): Date | null {
    if (!this.cache.last_updated) {
      return null;
    }
    return new Date(this.cache.last_updated);
  }

  /**
   * Returns a human-readable last updated string
   */
  getLastUpdatedString(): string {
    if (!this.cache.last_updated) {
      return 'never (using default rate)';
    }
    return formatDateTimeHuman(new Date(this.cache.last_updated));
  }

  /**
   * Returns a status message about the current rate
   */
  getStatusMessage(): string {
    if (this.offline) {
      if (!this.cache.last_updated) {
        return `Offline - using default rate (1 CAD = ${Math.round(this.cache.cad_to_idr)} IDR)`;
      }
      const lastDate = new Date(this.cache.last_updated);
      const shortMonths = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      return `Offline - using cached rate from ${shortMonths[lastDate.getMonth()]} ${lastDate.getDate()}`;
    }
    const lastDate = new Date(this.cache.last_updated);
    const shortMonths = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    return `Rate: 1 CAD = ${Math.round(this.cache.cad_to_idr)} IDR (updated ${shortMonths[lastDate.getMonth()]} ${lastDate.getDate()})`;
  }

  /**
   * Returns a formatted string of the current rate
   */
  formatRate(): string {
    return `1 CAD = ${Math.round(this.cache.cad_to_idr)} IDR`;
  }
}

// Singleton instance
export const currencyConverter = new CurrencyConverter();
