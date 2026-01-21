/**
 * Entry represents a single ledger entry (transaction)
 */
export interface Entry {
  id: string;          // Unique identifier for undo operations
  date: Date;          // Date of the entry
  description: string; // Description of the transaction
  cad: number;         // Cash flow in CAD
  idr: number;         // Cash flow in IDR
  screenTime: string;  // Screen time for the day (e.g., "3h45m")
}

/**
 * Creates a new entry with a unique ID
 */
export function createEntry(
  date: Date,
  description: string,
  cad: number,
  idr: number,
  screenTime: string = ''
): Entry {
  const descPrefix = description.slice(0, Math.min(8, description.length));
  return {
    id: `${Date.now()}-${descPrefix}`,
    date,
    description,
    cad,
    idr,
    screenTime,
  };
}

/**
 * Creates a deep copy of an entry
 */
export function cloneEntry(entry: Entry): Entry {
  return {
    ...entry,
    date: new Date(entry.date),
  };
}
