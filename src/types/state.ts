import type { Day } from './day';
import type { DateRange } from './dateRange';

/**
 * Application states - matches the Go version's state machine
 */
export enum AppState {
  Menu = 'Menu',
  DayEdit = 'DayEdit',
  RangeView = 'RangeView',
  DateInput = 'DateInput',
  QueryStartDate = 'QueryStartDate',
  QueryEndDate = 'QueryEndDate',
}

/**
 * Editor modes within DayEdit state
 */
export enum EditorMode {
  Normal = 'Normal',
  Search = 'Search',
  InlineEdit = 'InlineEdit',
  ScreenTime = 'ScreenTime',
  Journal = 'Journal',
}

/**
 * Column indices for inline editing
 */
export enum EditColumn {
  Description = 0,
  IDR = 1,
  CAD = 2,
}

/**
 * Menu option indices
 */
export enum MenuOption {
  Today = 0,
  Query = 1,
  GoToDate = 2,
}

/**
 * Notification type for status messages
 */
export interface Notification {
  message: string;
  isError: boolean;
  timestamp: number;
}

/**
 * Main application state
 */
export interface AppContext {
  // Current state
  state: AppState;
  previousState: AppState | null;

  // Menu state
  menuSelection: MenuOption;

  // Current day being viewed/edited
  currentDay: Day | null;
  selectedEntryIndex: number;

  // Date range for query view
  dateRange: DateRange | null;
  queryStartDate: Date | null;

  // Editor state
  editorMode: EditorMode;
  editColumn: EditColumn;
  editValue: string;
  searchQuery: string;
  deleteConfirmPending: boolean;

  // Date input state
  dateInputValue: string;
  dateInputTarget: 'goto' | 'queryStart' | 'queryEnd';

  // Notification
  notification: Notification | null;

  // Loading state
  isLoading: boolean;
}

/**
 * Creates the initial application context
 */
export function createInitialContext(): AppContext {
  return {
    state: AppState.Menu,
    previousState: null,
    menuSelection: MenuOption.Today,
    currentDay: null,
    selectedEntryIndex: 0,
    dateRange: null,
    queryStartDate: null,
    editorMode: EditorMode.Normal,
    editColumn: EditColumn.Description,
    editValue: '',
    searchQuery: '',
    deleteConfirmPending: false,
    dateInputValue: '',
    dateInputTarget: 'goto',
    notification: null,
    isLoading: false,
  };
}
