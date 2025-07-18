import React from 'react';
import { useAppStore } from '@/store';
import { Database, Brain, GitBranch } from 'lucide-react';

export const ContextBrowser: React.FC = () => {
  const { contextState, protectedMemory, knowledgeBase } = useAppStore();
  
  return (
    <div className="p-4 space-y-6">
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
      
      {contextState && (
        <div>
          <h3 className="font-medium text-secondary-900 dark:text-secondary-100 mb-3 flex items-center">
            <GitBranch className="w-4 h-4 mr-2" />
            Context Information
          </h3>
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-secondary-600 dark:text-secondary-400">Session ID</span>
              <span className="font-mono text-secondary-900 dark:text-secondary-100">
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
          </div>
        </div>
      )}
    </div>
  );
}; 