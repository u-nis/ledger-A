import React from 'react';

interface HelpItem {
  key: string;
  description: string;
}

interface StatusBarProps {
  rateInfo?: string;
  modeInfo?: string;
  notification?: string;
  helpItems: HelpItem[];
}

/**
 * A ribbon-style footer/status bar matching the Go version's RenderRibbonFooter
 */
export function StatusBar({ rateInfo, modeInfo, notification, helpItems }: StatusBarProps) {
  return (
    <box flexDirection="row" width="100%" gap={0}>
      {/* Mode ribbon (leftmost, auto width) */}
      {modeInfo && (
        <box backgroundColor="#CCCCCC" paddingLeft={2} paddingRight={2}>
          <text fg="#000000">
            <b>{modeInfo}</b>
          </text>
        </box>
      )}

      {/* Rate ribbon (left side) */}
      {rateInfo && (
        <box backgroundColor="#444444" paddingLeft={2} paddingRight={2}>
          <text fg="#FFFFFF">
            <b>{rateInfo}</b>
          </text>
        </box>
      )}

      {/* Help/controls ribbon */}
      <box
        backgroundColor="#222222"
        paddingLeft={2}
        paddingRight={2}
        flexDirection="row"
      >
        {helpItems.map((item, index) => (
          <React.Fragment key={index}>
            <text fg="#FFFFFF">{item.key}</text>
            <text fg="#888888">{' ' + item.description + '  '}</text>
          </React.Fragment>
        ))}
      </box>

      {/* Notification ribbon (right side, takes remaining space) */}
      <box backgroundColor="#333333" paddingLeft={2} paddingRight={2} flexGrow={1} justifyContent="flex-end">
        <text fg="#FFFFFF">{notification || ''}</text>
      </box>
    </box>
  );
}

interface NotificationProps {
  message: string;
  isError?: boolean;
}

/**
 * A notification component for displaying messages
 */
export function Notification({ message, isError = false }: NotificationProps) {
  if (!message) return null;

  return (
    <box backgroundColor="#222222" paddingLeft={1} paddingRight={1}>
      <text fg={isError ? '#CCCCCC' : '#FFFFFF'}>
        {message}
      </text>
    </box>
  );
}
