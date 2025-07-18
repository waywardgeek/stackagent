import React from 'react';
import { useAppStore } from '@/store';
import { User, Bot } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { ShellCommandWidget } from './ShellCommandWidget';
import { FileOperationWidget } from './FileOperationWidget';
import { ShellOperation, FileOperation } from '@/types';

export const MessageList: React.FC = () => {
  const { messages, showTerminal } = useAppStore();
  
  const handleTerminalOpen = (operation: ShellOperation) => {
    showTerminal(operation);
  };

  const handleFileOpen = (operation: FileOperation) => {
    // TODO: Implement file viewer opening
    console.log('Opening file for operation:', operation);
  };
  
  return (
    <div className="scrollable-container p-4 space-y-4">
      {messages.map((message) => (
        <div
          key={message.id}
          className={`flex ${message.type === 'user' ? 'justify-end' : 'justify-start'}`}
        >
          <div
            className={`max-w-[80%] rounded-lg px-4 py-2 ${
              message.type === 'user'
                ? 'chat-message-user'
                : 'chat-message-assistant'
            }`}
          >
            <div className="flex items-center space-x-2 mb-1">
              {message.type === 'user' ? (
                <User className="w-4 h-4" />
              ) : (
                <Bot className="w-4 h-4" />
              )}
              <span className="text-sm font-medium">
                {message.type === 'user' ? 'You' : 'StackAgent'}
              </span>
              <span className="text-xs opacity-75">
                {formatDistanceToNow(new Date(message.timestamp), { addSuffix: true })}
              </span>
            </div>
            <div className="text-sm whitespace-pre-wrap">
              {message.content}
            </div>
            
            {/* Interactive Operation Widgets */}
            {message.operationSummary?.hasOperations && (
              <div className="mt-3">
                {message.operationSummary.shellCommands && message.operationSummary.shellCommands.length > 0 && (
                  <ShellCommandWidget
                    operations={message.operationSummary.shellCommands}
                    onTerminalOpen={handleTerminalOpen}
                  />
                )}
                {message.operationSummary.fileOperations && message.operationSummary.fileOperations.length > 0 && (
                  <FileOperationWidget
                    operations={message.operationSummary.fileOperations}
                    onFileOpen={handleFileOpen}
                  />
                )}
              </div>
            )}
            
            {message.cost && (
              <div className="text-xs opacity-75 mt-1">
                Cost: ${message.cost.toFixed(4)} • {message.tokens} tokens
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}; 