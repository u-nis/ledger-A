/**
 * Undo Manager - 100-action undo stack for ledger operations
 */

import type { Entry } from '../types';
import { cloneEntry } from '../types/entry';
import { truncate } from '../utils/format';
import { LedgerService, ledgerService } from './ledgerService';

export enum ActionType {
  AddEntry = 'AddEntry',
  DeleteEntry = 'DeleteEntry',
  EditEntry = 'EditEntry',
  SetScreenTime = 'SetScreenTime',
  SetJournal = 'SetJournal',
}

export interface UndoAction {
  type: ActionType;
  date: Date;
  entry?: Entry;           // For entry operations
  oldEntry?: Entry;        // For edit operations (previous state)
  screenTime?: string;     // For screen time operations
  oldScreenTime?: string;  // Previous screen time
  journal?: string;        // For journal operations
  oldJournal?: string;     // Previous journal
  description: string;     // Human-readable description for notification
}

const MAX_STACK_SIZE = 100;

export class UndoStack {
  private actions: UndoAction[] = [];

  /**
   * Adds an action to the undo stack
   */
  push(action: UndoAction): void {
    this.actions.push(action);

    // Trim if exceeds max size
    if (this.actions.length > MAX_STACK_SIZE) {
      this.actions.shift();
    }
  }

  /**
   * Removes and returns the last action from the stack
   */
  pop(): UndoAction | undefined {
    return this.actions.pop();
  }

  /**
   * Returns the last action without removing it
   */
  peek(): UndoAction | undefined {
    return this.actions[this.actions.length - 1];
  }

  /**
   * Returns true if the stack is empty
   */
  isEmpty(): boolean {
    return this.actions.length === 0;
  }

  /**
   * Returns the number of actions in the stack
   */
  size(): number {
    return this.actions.length;
  }

  /**
   * Removes all actions from the stack
   */
  clear(): void {
    this.actions = [];
  }

  /**
   * Records an add entry action
   */
  pushAddEntry(date: Date, entry: Entry): void {
    this.push({
      type: ActionType.AddEntry,
      date,
      entry: cloneEntry(entry),
      description: `Added '${truncate(entry.description, 20)}'`,
    });
  }

  /**
   * Records a delete entry action
   */
  pushDeleteEntry(date: Date, entry: Entry): void {
    this.push({
      type: ActionType.DeleteEntry,
      date,
      entry: cloneEntry(entry),
      description: `Deleted '${truncate(entry.description, 20)}'`,
    });
  }

  /**
   * Records an edit entry action
   */
  pushEditEntry(date: Date, oldEntry: Entry, newEntry: Entry): void {
    this.push({
      type: ActionType.EditEntry,
      date,
      entry: cloneEntry(newEntry),
      oldEntry: cloneEntry(oldEntry),
      description: `Edited '${truncate(oldEntry.description, 20)}'`,
    });
  }

  /**
   * Records a screen time change action
   */
  pushSetScreenTime(date: Date, oldScreenTime: string, newScreenTime: string): void {
    this.push({
      type: ActionType.SetScreenTime,
      date,
      screenTime: newScreenTime,
      oldScreenTime,
      description: `Changed screen time to '${newScreenTime}'`,
    });
  }

  /**
   * Records a journal change action
   */
  pushSetJournal(date: Date, oldJournal: string, newJournal: string): void {
    this.push({
      type: ActionType.SetJournal,
      date,
      journal: newJournal,
      oldJournal,
      description: `Updated journal`,
    });
  }
}

export class UndoManager {
  private stack: UndoStack;
  private service: LedgerService;

  constructor(service?: LedgerService) {
    this.stack = new UndoStack();
    this.service = service || ledgerService;
  }

  /**
   * Gets the undo stack
   */
  getStack(): UndoStack {
    return this.stack;
  }

  /**
   * Performs the undo operation and returns a description of what was undone
   */
  async undo(): Promise<string | null> {
    const action = this.stack.pop();
    if (!action) {
      return null;
    }

    const day = await this.service.getDay(action.date);

    switch (action.type) {
      case ActionType.AddEntry:
        // Undo add = remove the entry
        if (action.entry) {
          const index = day.entries.findIndex(e => e.id === action.entry!.id);
          if (index !== -1) {
            day.entries.splice(index, 1);
          }
          await this.service.saveDay(day);
          return `Undo: Removed '${truncate(action.entry.description, 20)}'`;
        }
        break;

      case ActionType.DeleteEntry:
        // Undo delete = restore the entry
        if (action.entry) {
          day.entries.push(action.entry);
          await this.service.saveDay(day);
          return `Undo: Restored '${truncate(action.entry.description, 20)}'`;
        }
        break;

      case ActionType.EditEntry:
        // Undo edit = restore old entry state
        if (action.oldEntry) {
          const index = day.entries.findIndex(e => e.id === action.oldEntry!.id);
          if (index !== -1) {
            day.entries[index] = action.oldEntry;
          }
          await this.service.saveDay(day);
          return `Undo: Reverted '${truncate(action.oldEntry.description, 20)}'`;
        }
        break;

      case ActionType.SetScreenTime:
        // Undo screen time = restore old screen time
        if (action.oldScreenTime !== undefined) {
          day.screenTime = action.oldScreenTime;
          for (const entry of day.entries) {
            entry.screenTime = action.oldScreenTime;
          }
          await this.service.saveDay(day);
          return `Undo: Restored screen time to '${action.oldScreenTime}'`;
        }
        break;

      case ActionType.SetJournal:
        // Undo journal = restore old journal
        if (action.oldJournal !== undefined) {
          day.journal = action.oldJournal;
          await this.service.saveDay(day);
          return `Undo: Restored journal`;
        }
        break;
    }

    return null;
  }

  /**
   * Returns true if there are actions to undo
   */
  canUndo(): boolean {
    return !this.stack.isEmpty();
  }

  /**
   * Records an add entry action for undo
   */
  recordAddEntry(date: Date, entry: Entry): void {
    this.stack.pushAddEntry(date, entry);
  }

  /**
   * Records a delete entry action for undo
   */
  recordDeleteEntry(date: Date, entry: Entry): void {
    this.stack.pushDeleteEntry(date, entry);
  }

  /**
   * Records an edit entry action for undo
   */
  recordEditEntry(date: Date, oldEntry: Entry, newEntry: Entry): void {
    this.stack.pushEditEntry(date, oldEntry, newEntry);
  }

  /**
   * Records a screen time change for undo
   */
  recordSetScreenTime(date: Date, oldScreenTime: string, newScreenTime: string): void {
    this.stack.pushSetScreenTime(date, oldScreenTime, newScreenTime);
  }

  /**
   * Records a journal change for undo
   */
  recordSetJournal(date: Date, oldJournal: string, newJournal: string): void {
    this.stack.pushSetJournal(date, oldJournal, newJournal);
  }

  /**
   * Clears the undo stack
   */
  clear(): void {
    this.stack.clear();
  }
}

// Singleton instance
export const undoManager = new UndoManager(ledgerService);
