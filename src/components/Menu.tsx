import React from 'react';
import { useKeyboard } from '@opentui/react';
import { Box as StyledBox } from './shared/Box';
import { StatusBar } from './shared/StatusBar';
import { formatDateDisplay } from '../utils/format';

export enum MenuSelection {
  None = 'None',
  Today = 'Today',
  Query = 'Query',
  AddPastDay = 'AddPastDay',
  Quit = 'Quit',
}

interface MenuItem {
  key: string;
  label: string;
  description: string;
  selection: MenuSelection;
}

interface MenuProps {
  onSelect: (selection: MenuSelection) => void;
}

const LOGO = `
 ╦  ╔═╗╔╦╗╔═╗╔═╗╦═╗   ╔═╗
 ║  ║╣  ║║║ ╦║╣ ╠╦╝───╠═╣
 ╩═╝╚═╝═╩╝╚═╝╚═╝╩╚═   ╩ ╩`;

export function Menu({ onSelect }: MenuProps) {
  const [selected, setSelected] = React.useState(0);

  const today = formatDateDisplay(new Date());

  const items: MenuItem[] = [
    { key: '1', label: `Today (${today})`, description: 'View and edit today\'s entries', selection: MenuSelection.Today },
    { key: '2', label: 'Query', description: 'View a single day or date range', selection: MenuSelection.Query },
    { key: '3', label: 'Add Entry for Past Day', description: 'Add entries for a day you missed', selection: MenuSelection.AddPastDay },
  ];

  useKeyboard((event) => {
    const key = event.name;
    if (key === 'up') {
      setSelected(prev => Math.max(0, prev - 1));
    } else if (key === 'down') {
      setSelected(prev => Math.min(items.length - 1, prev + 1));
    } else if (key === 'return' || key === 'space') {
      onSelect(items[selected].selection);
    } else if (key === '1') {
      onSelect(MenuSelection.Today);
    } else if (key === '2') {
      onSelect(MenuSelection.Query);
    } else if (key === '3') {
      onSelect(MenuSelection.AddPastDay);
    } else if (key === 'q') {
      onSelect(MenuSelection.Quit);
    }
  });

  const helpItems = [
    { key: '↑/↓', description: 'navigate' },
    { key: 'Enter', description: 'select' },
    { key: 'q', description: 'quit' },
  ];

  return (
    <StyledBox
      title="LEDGER-A"
      footer={<StatusBar helpItems={helpItems} />}
    >
      <box flexDirection="column" alignItems="center" gap={1}>
        {/* Logo */}
        <text fg="#FFFFFF">
          <b>{LOGO}</b>
        </text>

        {/* Subtitle */}
        <text fg="#888888">
          <i>Daily Finance Tracker</i>
        </text>

        {/* Spacer */}
        <box height={2} />

        {/* Menu items */}
        <box flexDirection="column" gap={1}>
          {items.map((item, index) => {
            const isSelected = index === selected;
            const cursor = isSelected ? '► ' : '  ';

            return (
              <box key={item.key} flexDirection="row" gap={1}>
                <text fg="#FFFFFF">
                  <b>{cursor}</b>
                </text>
                <text fg="#FFFFFF">
                  <b>[{item.key}]</b>
                </text>
                <box width={30}>
                  {isSelected ? (
                    <text fg="#FFFFFF" bg="#444444">
                      <b>{item.label}</b>
                    </text>
                  ) : (
                    <text fg="#CCCCCC">
                      {item.label}
                    </text>
                  )}
                </box>
                <text fg="#888888">
                  <i>{item.description}</i>
                </text>
              </box>
            );
          })}
        </box>
      </box>
    </StyledBox>
  );
}
