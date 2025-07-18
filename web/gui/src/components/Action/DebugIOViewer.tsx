import React from 'react';
import { useAppStore } from '@/store';
import { Bug, ArrowUp, ArrowDown, Server, AlertCircle, Trash2, Copy } from 'lucide-react';
import type { DebugMessage } from '@/types';

export const DebugIOViewer: React.FC = () => {
  const { debugMessages, clearDebugMessages } = useAppStore();
  
  // Sort messages by timestamp (newest first)
  const sortedMessages = [...debugMessages].sort((a, b) => 
    new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const formatJson = (jsonString: string) => {
    try {
      const parsed = JSON.parse(jsonString);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return jsonString;
    }
  };

  const getIcon = (message: DebugMessage) => {
    if (message.type === 'websocket') {
      return message.direction === 'sent' ? <ArrowUp className="w-4 h-4" /> : <ArrowDown className="w-4 h-4" />;
    }
    if (message.type === 'api') {
      return <Server className="w-4 h-4" />;
    }
    return <AlertCircle className="w-4 h-4" />;
  };

  const getIconColor = (message: DebugMessage) => {
    if (message.type === 'error') return 'text-red-500';
    if (message.direction === 'sent') return 'text-blue-500';
    return 'text-green-500';
  };

  const getTypeLabel = (message: DebugMessage) => {
    if (message.type === 'websocket') {
      return `WebSocket ${message.direction}`;
    }
    if (message.type === 'api') {
      return `API ${message.direction}`;
    }
    return 'Error';
  };

  return (
    <div className="p-4 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="font-medium text-secondary-900 dark:text-secondary-100 flex items-center">
          <Bug className="w-4 h-4 mr-2" />
          JSON I/O Debug ({debugMessages.length})
        </h3>
        {debugMessages.length > 0 && (
          <button
            onClick={clearDebugMessages}
            className="flex items-center space-x-1 text-xs text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
            title="Clear debug messages"
          >
            <Trash2 className="w-3 h-3" />
            <span>Clear</span>
          </button>
        )}
      </div>

      {/* Messages */}
      <div className="space-y-3 max-h-96 overflow-y-auto">
        {sortedMessages.length === 0 ? (
          <div className="text-center text-secondary-500 dark:text-secondary-400 py-8">
            No debug messages yet. Start a conversation to see JSON I/O.
          </div>
        ) : (
          sortedMessages.map((message) => (
            <div key={message.id} className="card-body border-l-4 border-l-secondary-400">
              <div className="flex items-start justify-between mb-2">
                <div className="flex items-center space-x-2">
                  <div className={getIconColor(message)}>
                    {getIcon(message)}
                  </div>
                  <div>
                    <span className="font-medium text-sm text-secondary-700 dark:text-secondary-300">
                      {getTypeLabel(message)}
                    </span>
                    {message.event && (
                      <span className="ml-2 text-xs text-secondary-500 dark:text-secondary-400">
                        {message.event}
                      </span>
                    )}
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-xs text-secondary-500 dark:text-secondary-400">
                    {new Date(message.timestamp).toLocaleTimeString()}
                  </span>
                  <button
                    onClick={() => copyToClipboard(message.rawJson)}
                    className="text-xs text-secondary-500 hover:text-secondary-700 dark:text-secondary-400 dark:hover:text-secondary-200"
                    title="Copy JSON"
                  >
                    <Copy className="w-3 h-3" />
                  </button>
                </div>
              </div>
              
              {/* JSON Content */}
              <div className="bg-secondary-50 dark:bg-secondary-900 p-3 rounded-md">
                <pre className="text-xs text-secondary-800 dark:text-secondary-200 overflow-x-auto whitespace-pre-wrap">
                  {formatJson(message.rawJson)}
                </pre>
              </div>
              
              {/* Summary */}
              <div className="mt-2 text-xs text-secondary-600 dark:text-secondary-400">
                Size: {message.rawJson.length} chars
                {message.data && typeof message.data === 'object' && (
                  <span className="ml-2">
                    Keys: {Object.keys(message.data).join(', ')}
                  </span>
                )}
              </div>
            </div>
          ))
        )}
      </div>

      {/* Instructions */}
      <div className="text-xs text-secondary-500 dark:text-secondary-400 p-3 bg-secondary-50 dark:bg-secondary-800 rounded-md">
        <p className="font-medium mb-1">Debug Information:</p>
        <ul className="space-y-1">
          <li>🔵 Blue arrows: Messages sent to server</li>
          <li>🟢 Green arrows: Messages received from server</li>
          <li>🟡 Server icon: API calls to Claude</li>
          <li>🔴 Alert icon: Errors</li>
        </ul>
      </div>
    </div>
  );
}; 