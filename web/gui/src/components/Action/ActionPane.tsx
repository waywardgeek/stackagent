import React, { useState } from 'react';
import { useAppStore, selectSelectedFunctionCall } from '@/store';
import { FunctionCallDetails } from './FunctionCallDetails';
import { ContextBrowser } from './ContextBrowser';
import { CommandOutput } from './CommandOutput';
import { FilePreview } from './FilePreview';
import { DebugIOViewer } from './DebugIOViewer';
import type { ActionView, WebSocketEventType } from '@/types';
import { 
  Eye, 
  Terminal, 
  FileText, 
  Database, 
  Zap,
  Bug
} from 'lucide-react';

interface ActionPaneProps {
  sendMessage: (type: WebSocketEventType, data: any) => void;
}

export const ActionPane: React.FC<ActionPaneProps> = ({ sendMessage }) => {
  const selectedFunctionCall = useAppStore(selectSelectedFunctionCall);
  const { functionCalls, commandExecutions } = useAppStore();
  const [activeView, setActiveView] = useState<ActionView>('context');
  
  return (
    <div className="flex flex-col h-full">
      {/* Action pane header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-secondary-200 dark:border-secondary-700">
        <h2 className="text-lg font-semibold text-secondary-900 dark:text-secondary-100">
          {activeView === 'function-call' && 'Function Details'}
          {activeView === 'command-output' && 'Command Output'}
          {activeView === 'context' && 'Context Browser'}
          {activeView === 'file-preview' && 'File Preview'}
          {activeView === 'debug-io' && 'JSON I/O Debug'}
        </h2>
        
        {/* View indicators */}
        <div className="flex items-center space-x-2">
          {selectedFunctionCall && (
            <div className="flex items-center space-x-1 text-sm text-secondary-600 dark:text-secondary-400">
              <Zap className="w-4 h-4" />
              <span>{selectedFunctionCall.name}</span>
            </div>
          )}
        </div>
      </div>
      
      {/* Content area */}
      <div className="flex-1 overflow-hidden">
        {activeView === 'function-call' && (
          <FunctionCallDetails functionCall={selectedFunctionCall || functionCalls[functionCalls.length - 1]} />
        )}
        
        {activeView === 'command-output' && (
          <CommandOutput execution={commandExecutions[commandExecutions.length - 1]} />
        )}
        
        {activeView === 'context' && (
          <ContextBrowser sendMessage={sendMessage} />
        )}
        
        {activeView === 'file-preview' && (
          <FilePreview />
        )}
        
        {activeView === 'debug-io' && (
          <DebugIOViewer />
        )}
        
        {/* Default empty state */}
        {!selectedFunctionCall && functionCalls.length === 0 && commandExecutions.length === 0 && (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-4 bg-secondary-100 dark:bg-secondary-800 rounded-full flex items-center justify-center">
                <Eye className="w-8 h-8 text-secondary-400" />
              </div>
              <h3 className="text-lg font-medium text-secondary-900 dark:text-secondary-100 mb-2">
                No Action Selected
              </h3>
              <p className="text-secondary-600 dark:text-secondary-400 mb-4">
                Function call details will appear here
              </p>
              <div className="text-sm text-secondary-500 dark:text-secondary-500">
                <p>• Click on function calls to see details</p>
                <p>• View command outputs and results</p>
                <p>• Browse context and memory</p>
                <p>• Preview modified files</p>
              </div>
            </div>
          </div>
        )}
      </div>
      
      {/* Tab Navigation */}
      <div className="border-t border-secondary-200 dark:border-secondary-700 p-3">
        <div className="flex items-center space-x-2">
          <button 
            onClick={() => setActiveView('context')}
            className={`btn-ghost text-xs ${activeView === 'context' ? 'bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-300' : ''}`}
          >
            <Database className="w-4 h-4 mr-1" />
            Context
          </button>
          <button 
            onClick={() => setActiveView('debug-io')}
            className={`btn-ghost text-xs ${activeView === 'debug-io' ? 'bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-300' : ''}`}
          >
            <Bug className="w-4 h-4 mr-1" />
            Debug I/O
          </button>
          {functionCalls.length > 0 && (
            <button 
              onClick={() => setActiveView('function-call')}
              className={`btn-ghost text-xs ${activeView === 'function-call' ? 'bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-300' : ''}`}
            >
              <Zap className="w-4 h-4 mr-1" />
              Functions
            </button>
          )}
          {commandExecutions.length > 0 && (
            <button 
              onClick={() => setActiveView('command-output')}
              className={`btn-ghost text-xs ${activeView === 'command-output' ? 'bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-300' : ''}`}
            >
              <Terminal className="w-4 h-4 mr-1" />
              Commands
            </button>
          )}
          <button 
            onClick={() => setActiveView('file-preview')}
            className={`btn-ghost text-xs ${activeView === 'file-preview' ? 'bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-300' : ''}`}
          >
            <FileText className="w-4 h-4 mr-1" />
            Files
          </button>
        </div>
      </div>
    </div>
  );
}; 