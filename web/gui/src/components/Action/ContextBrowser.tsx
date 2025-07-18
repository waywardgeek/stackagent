import React from 'react';
import { useAppStore } from '@/store';
import { Database, Brain, GitBranch, MessageCircle, User, Bot, Trash2, RefreshCw } from 'lucide-react';
import type { WebSocketEventType } from '@/types';

interface ContextBrowserProps {
  sendMessage: (type: WebSocketEventType, data: any) => void;
}

export const ContextBrowser: React.FC<ContextBrowserProps> = ({ sendMessage }) => {
  const { contextState, protectedMemory, knowledgeBase, messages, clearMessages, connected } = useAppStore();
  
  // Sort messages by timestamp (create a copy first to avoid mutating the store)
  const sortedMessages = [...messages].sort((a, b) => 
    new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  const handleRefreshContext = () => {
    // Request context from backend
    sendMessage('get_context', {});
  };

  return (
    <div className="p-4 space-y-6">
      {/* Conversation Messages */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h3 className="font-medium text-secondary-900 dark:text-secondary-100 flex items-center">
            <MessageCircle className="w-4 h-4 mr-2" />
            Conversation Messages ({messages.length})
          </h3>
          {messages.length > 0 && (
            <button
              onClick={clearMessages}
              className="flex items-center space-x-1 text-xs text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
              title="Clear conversation messages"
            >
              <Trash2 className="w-3 h-3" />
              <span>Clear</span>
            </button>
          )}
        </div>
        <div className="space-y-3 max-h-96 overflow-y-auto">
          {sortedMessages.length === 0 ? (
            <div className="text-center text-secondary-500 dark:text-secondary-400 py-8">
              No messages in this conversation yet
            </div>
          ) : (
            sortedMessages.map((message) => (
              <div key={message.id} className="card-body border-l-4 border-l-primary-400">
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center space-x-2">
                    {message.type === 'user' ? (
                      <User className="w-4 h-4 text-blue-500" />
                    ) : (
                      <Bot className="w-4 h-4 text-green-500" />
                    )}
                    <span className="font-medium text-sm text-secondary-700 dark:text-secondary-300">
                      {message.type === 'user' ? 'User' : 'StackAgent'}
                    </span>
                  </div>
                  <span className="text-xs text-secondary-500 dark:text-secondary-400">
                    {new Date(message.timestamp).toLocaleTimeString()}
                  </span>
                </div>
                <div className="text-sm text-secondary-800 dark:text-secondary-200 mb-2">
                  {message.content}
                </div>
                <div className="flex justify-between items-center text-xs text-secondary-500 dark:text-secondary-400">
                  <span>ID: {message.id}</span>
                  <span>Session: {message.sessionId}</span>
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Context Information */}
      {contextState && (
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-medium text-secondary-900 dark:text-secondary-100 flex items-center">
              <GitBranch className="w-4 h-4 mr-2" />
              Context Information
            </h3>
            <button
              onClick={handleRefreshContext}
              className="flex items-center space-x-1 text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
              title="Refresh context from backend"
            >
              <RefreshCw className="w-3 h-3" />
              <span>Refresh</span>
            </button>
          </div>
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-secondary-600 dark:text-secondary-400">Session ID</span>
              <span className="font-mono text-secondary-900 dark:text-secondary-100 text-xs">
                {contextState.sessionId}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-secondary-600 dark:text-secondary-400">Git Branch</span>
              <span className="font-mono text-secondary-900 dark:text-secondary-100">
                {contextState.gitBranch || 'N/A'}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-secondary-600 dark:text-secondary-400">Memory Entries</span>
              <span className="text-secondary-900 dark:text-secondary-100">
                {contextState.memoryEntries}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-secondary-600 dark:text-secondary-400">Knowledge Entries</span>
              <span className="text-secondary-900 dark:text-secondary-100">
                {contextState.knowledgeEntries}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-secondary-600 dark:text-secondary-400">Command History</span>
              <span className="text-secondary-900 dark:text-secondary-100">
                {contextState.commandHistory}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-secondary-600 dark:text-secondary-400">Last Activity</span>
              <span className="text-secondary-900 dark:text-secondary-100 text-xs">
                {new Date(contextState.lastActivity).toLocaleString()}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-secondary-600 dark:text-secondary-400">Created At</span>
              <span className="text-secondary-900 dark:text-secondary-100 text-xs">
                {new Date(contextState.createdAt).toLocaleString()}
              </span>
            </div>
            {contextState.totalCost !== undefined && (
              <div className="border-t pt-3 mt-3">
                <div className="flex justify-between text-sm">
                  <span className="text-secondary-600 dark:text-secondary-400">Total Cost</span>
                  <span className="text-secondary-900 dark:text-secondary-100 font-mono">
                    ${contextState.totalCost.toFixed(4)}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-secondary-600 dark:text-secondary-400">API Requests</span>
                  <span className="text-secondary-900 dark:text-secondary-100">
                    {contextState.requestCount || 0}
                  </span>
                </div>
                {contextState.cacheStats && (
                  <>
                    <div className="flex justify-between text-sm">
                      <span className="text-secondary-600 dark:text-secondary-400">Cache Hits</span>
                      <span className="text-secondary-900 dark:text-secondary-100">
                        {contextState.cacheStats.cacheHits}
                      </span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-secondary-600 dark:text-secondary-400">Cache Efficiency</span>
                      <span className="text-secondary-900 dark:text-secondary-100">
                        {contextState.cacheStats.cacheEfficiency.toFixed(1)}%
                      </span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-secondary-600 dark:text-secondary-400">Cache Savings</span>
                      <span className="text-secondary-900 dark:text-secondary-100 font-mono text-green-600 dark:text-green-400">
                        ${contextState.cacheStats.totalSavings.toFixed(4)}
                      </span>
                    </div>
                  </>
                )}
              </div>
            )}
          </div>
        </div>
      )}
      
      {/* Protected Memory */}
      <div>
        <h3 className="font-medium text-secondary-900 dark:text-secondary-100 mb-3 flex items-center">
          <Database className="w-4 h-4 mr-2" />
          Protected Memory
        </h3>
        <div className="space-y-2">
          {Object.entries(protectedMemory).map(([key, value]) => (
            <div key={key} className="card-body">
              <div className="flex justify-between items-start">
                <span className="font-medium text-sm text-secondary-700 dark:text-secondary-300">
                  {key}
                </span>
                <span className="text-xs text-secondary-500 dark:text-secondary-400">
                  {value.length} chars
                </span>
              </div>
              <div className="text-sm text-secondary-600 dark:text-secondary-400 mt-1">
                {value.length > 100 ? `${value.substring(0, 100)}...` : value}
              </div>
            </div>
          ))}
          {Object.keys(protectedMemory).length === 0 && (
            <div className="text-center text-secondary-500 dark:text-secondary-400 py-4">
              No protected memory entries
            </div>
          )}
        </div>
      </div>
      
      {/* Knowledge Base */}
      <div>
        <h3 className="font-medium text-secondary-900 dark:text-secondary-100 mb-3 flex items-center">
          <Brain className="w-4 h-4 mr-2" />
          Knowledge Base
        </h3>
        <div className="space-y-2">
          {Object.entries(knowledgeBase).map(([key, value]) => (
            <div key={key} className="card-body">
              <div className="flex justify-between items-start">
                <span className="font-medium text-sm text-secondary-700 dark:text-secondary-300">
                  {key}
                </span>
                <span className="text-xs text-secondary-500 dark:text-secondary-400">
                  {value.length} chars
                </span>
              </div>
              <div className="text-sm text-secondary-600 dark:text-secondary-400 mt-1">
                {value.length > 100 ? `${value.substring(0, 100)}...` : value}
              </div>
            </div>
          ))}
          {Object.keys(knowledgeBase).length === 0 && (
            <div className="text-center text-secondary-500 dark:text-secondary-400 py-4">
              No knowledge entries
            </div>
          )}
        </div>
      </div>

      {/* Debug Information */}
      <div className="mt-6 text-center">
        <div className="flex items-center justify-center space-x-2 mb-2">
          <button
            onClick={handleRefreshContext}
            className="flex items-center space-x-1 text-xs px-3 py-1 bg-primary-100 hover:bg-primary-200 text-primary-700 rounded-md dark:bg-primary-900 dark:hover:bg-primary-800 dark:text-primary-300"
            disabled={!connected}
          >
            <RefreshCw className="w-3 h-3" />
            <span>Refresh Context</span>
          </button>
        </div>
        <div className="text-xs text-secondary-500 dark:text-secondary-400">
          Session: {contextState?.sessionId || 'Not connected'}
        </div>
        <div className="text-xs text-secondary-400 dark:text-secondary-500 mt-2 italic">
          Core principle: Don't be evil
        </div>
      </div>
    </div>
  );
}; 