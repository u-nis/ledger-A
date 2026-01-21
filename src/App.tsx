import React, { useState, useEffect, useCallback } from 'react';
import { Menu, MenuSelection } from './components/Menu';
import { Editor, EditorAction } from './components/Editor';
import { DatePicker, DatePickerAction } from './components/DatePicker';
import { DateInput, DateInputAction } from './components/DateInput';
import { RangeView, RangeViewAction } from './components/RangeView';
import { AppState } from './types/state';
import type { Day, DateRange } from './types';
import { createDay } from './types/day';
import { createDateRange } from './types/dateRange';
import { ledgerService } from './services/ledgerService';
import { currencyConverter } from './services/currencyConverter';
import { undoManager } from './services/undoManager';
import { today, isSameDay } from './utils/date';

export function App() {
  const [state, setState] = useState<AppState>(AppState.Menu);
  const [currentDay, setCurrentDay] = useState<Day | null>(null);
  const [currentDate, setCurrentDate] = useState<Date>(today());
  const [dateRange, setDateRange] = useState<DateRange | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Initialize currency converter on mount
  useEffect(() => {
    currencyConverter.refreshRate().catch(() => {
      // Silently fall back to cached/default rate
    });
  }, []);

  const loadDay = useCallback(async (date: Date) => {
    setIsLoading(true);
    try {
      const day = await ledgerService.getDay(date);
      setCurrentDay(day);
      setCurrentDate(date);
    } catch {
      setCurrentDay(createDay(date));
      setCurrentDate(date);
    }
    setIsLoading(false);
  }, []);

  const loadDateRange = useCallback(async (start: Date, end: Date) => {
    setIsLoading(true);
    try {
      const range = await ledgerService.getDateRange(start, end);
      setDateRange(range);
    } catch {
      setDateRange(createDateRange(start, end));
    }
    setIsLoading(false);
  }, []);

  const saveDay = useCallback(async (day: Day) => {
    try {
      await ledgerService.saveDay(day);
    } catch (error) {
      console.error('Failed to save day:', error);
    }
  }, []);

  // Handle Menu selection
  const handleMenuSelect = useCallback(async (selection: MenuSelection) => {
    switch (selection) {
      case MenuSelection.Today:
        await loadDay(today());
        setState(AppState.DayEdit);
        break;
      case MenuSelection.Query:
        setState(AppState.QueryStartDate);
        break;
      case MenuSelection.AddPastDay:
        setState(AppState.DateInput);
        break;
      case MenuSelection.Quit:
        process.exit(0);
        break;
    }
  }, [loadDay]);

  // Handle Editor actions
  const handleEditorAction = useCallback(async (action: EditorAction) => {
    switch (action) {
      case EditorAction.Back:
        if (currentDay) {
          await saveDay(currentDay);
        }
        setState(AppState.Menu);
        break;
      case EditorAction.Saved:
        if (currentDay) {
          await saveDay(currentDay);
        }
        break;
      case EditorAction.Reload:
        // Reload day data (for undo)
        if (currentDate) {
          await loadDay(currentDate);
        }
        break;
    }
  }, [currentDay, currentDate, saveDay, loadDay]);

  // Handle DatePicker actions
  const handleDatePickerAction = useCallback(async (
    action: DatePickerAction,
    startDate?: Date,
    endDate?: Date
  ) => {
    switch (action) {
      case DatePickerAction.Selected:
        if (startDate && endDate) {
          if (isSameDay(startDate, endDate)) {
            // Single day - go to the same editor as "Today"
            await loadDay(startDate);
            setState(AppState.DayEdit);
          } else {
            // Date range - go to range view
            await loadDateRange(startDate, endDate);
            setState(AppState.RangeView);
          }
        }
        break;
      case DatePickerAction.Cancelled:
        setState(AppState.Menu);
        break;
    }
  }, [loadDay, loadDateRange]);

  // Handle DateInput actions
  const handleDateInputAction = useCallback(async (
    action: DateInputAction,
    date?: Date
  ) => {
    switch (action) {
      case DateInputAction.Submitted:
        if (date) {
          await loadDay(date);
          setState(AppState.DayEdit);
        }
        break;
      case DateInputAction.Cancelled:
        setState(AppState.Menu);
        break;
    }
  }, [loadDay]);

  // Handle RangeView actions
  const handleRangeViewAction = useCallback(async (
    action: RangeViewAction,
    selectedDate?: Date
  ) => {
    switch (action) {
      case RangeViewAction.Back:
        setState(AppState.Menu);
        break;
      case RangeViewAction.SelectDay:
        if (selectedDate) {
          await loadDay(selectedDate);
          setState(AppState.DayEdit);
        }
        break;
    }
  }, [loadDay]);

  // Loading state
  if (isLoading) {
    return (
      <box justifyContent="center" alignItems="center" height="100%">
        <text fg="#888888">Loading...</text>
      </box>
    );
  }

  // Render based on current state
  switch (state) {
    case AppState.Menu:
      return <Menu onSelect={handleMenuSelect} />;

    case AppState.DayEdit:
      if (!currentDay) return null;
      return (
        <Editor
          day={currentDay}
          onAction={handleEditorAction}
          onSave={saveDay}
          converter={currencyConverter}
          undo={undoManager}
        />
      );

    case AppState.QueryStartDate:
      return <DatePicker onAction={handleDatePickerAction} />;

    case AppState.DateInput:
      return (
        <DateInput
          title="Enter date to add entries"
          onAction={handleDateInputAction}
        />
      );

    case AppState.RangeView:
      if (!dateRange) return null;
      return <RangeView dateRange={dateRange} onAction={handleRangeViewAction} />;

    default:
      return <Menu onSelect={handleMenuSelect} />;
  }
}
