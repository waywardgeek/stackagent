import React from 'react';
import { useAppStore } from '@/store';
import MessageList from './MessageList';
import { ChatInput } from './ChatInput';
import { FunctionCallList } from './FunctionCallList';
import { TerminalPane } from '../Layout/TerminalPane';
import type { WebSocketEventType } from '@/types';

interface ChatPaneProps {
  sendMessage: (type: WebSocketEventType, data: any) => void;
}

export const ChatPane: React.FC<ChatPaneProps> = ({ sendMessage }) => {
  const { 
    messages, 
    functionCalls, 
    isTerminalVisible,
    currentTerminalOperation,
    allShellOperations,
    hideTerminal 
  } = useAppStore();
  
  return (
    <div className="flex flex-col h-full">
      {/* Chat header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-secondary-200 dark:border-secondary-700">
        <h2 className="text-lg font-semibold text-secondary-900 dark:text-secondary-100">
          Conversation
        </h2>
        <div className="flex items-center space-x-2 text-sm text-secondary-600 dark:text-secondary-400">
          <span>{messages.length} messages</span>
          {functionCalls.length > 0 && (
            <>
              <span>•</span>
              <span>{functionCalls.length} functions</span>
            </>
          )}
        </div>
      </div>
      
      {/* Messages area */}
      <div className="flex-1 flex flex-col min-h-0">
        {messages.length === 0 && functionCalls.length === 0 ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 bg-primary-100 dark:bg-primary-900 rounded-full flex items-center justify-center">
                <div className="w-8 h-8 bg-gradient-to-r from-primary-500 to-primary-600 rounded-lg flex items-center justify-center">
                  <span className="text-white font-bold">S</span>
                </div>
              </div>
              <h3 className="text-lg font-medium text-secondary-900 dark:text-secondary-100 mb-2">
                Welcome to StackAgent
              </h3>
              <p className="text-secondary-600 dark:text-secondary-400 mb-4">
                Your AI coding assistant with persistent memory
              </p>
              <div className="text-sm text-secondary-500 dark:text-secondary-500">
                <p>• Ask questions about your code</p>
                <p>• Run commands and analyze output</p>
                <p>• Context persists across sessions</p>
                <p>• Branch-specific AI memory</p>
              </div>
            </div>
          </div>
        ) : (
          <div className="flex-1 flex flex-col min-h-0">
            <MessageList />
            <FunctionCallList />
          </div>
        )}
      </div>
      
      {/* Chat input */}
      <div className="border-t border-secondary-200 dark:border-secondary-700">
        <ChatInput sendMessage={sendMessage} />
      </div>
      
      {/* Terminal Pane */}
      <TerminalPane
        isVisible={isTerminalVisible}
        onClose={hideTerminal}
        currentOperation={currentTerminalOperation}
        operations={allShellOperations}
      />
    </div>
  );
}; 