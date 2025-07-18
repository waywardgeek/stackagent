import React from 'react';
import { useAppStore, selectSelectedFunctionCall } from '@/store';
import { FunctionCallDetails } from './FunctionCallDetails';
import { ContextBrowser } from './ContextBrowser';
import { CommandOutput } from './CommandOutput';
import { FilePreview } from './FilePreview';
import type { ActionView } from '@/types';
import { 
  Eye, 
  Terminal, 
  FileText, 
  Database, 
  Info,
  Zap
} from 'lucide-react';

export const ActionPane: React.FC = () => {
  const selectedFunctionCall = useAppStore(selectSelectedFunctionCall);
  const { functionCalls, commandExecutions } = useAppStore();
  
  // Determine what to show in the action pane
  const getActiveView = (): ActionView => {
    if (selectedFunctionCall) {
      return 'function-call';
    }
    
    // Show most recent function call if none selected
    if (functionCalls.length > 0) {
      return 'function-call';
    }
    
    // Show most recent command execution
    if (commandExecutions.length > 0) {
      return 'command-output';
    }
    
    // Default to context browser
    return 'context';
  };
  
  const activeView = getActiveView();
  
  return (
    <div className="flex flex-col h-full">
      {/* Action pane header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-secondary-200 dark:border-secondary-700">
        <h2 className="text-lg font-semibold text-secondary-900 dark:text-secondary-100">
          {activeView === 'function-call' && 'Function Details'}
          {activeView === 'command-output' && 'Command Output'}
          {activeView === 'context' && 'Context Browser'}
          {activeView === 'file-preview' && 'File Preview'}
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
          <ContextBrowser />
        )}
        
        {activeView === 'file-preview' && (
          <FilePreview />
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
      
      {/* Quick actions */}
      <div className="border-t border-secondary-200 dark:border-secondary-700 p-3">
        <div className="flex items-center space-x-2">
          <button className="btn-ghost text-xs">
            <Terminal className="w-4 h-4 mr-1" />
            Commands
          </button>
          <button className="btn-ghost text-xs">
            <FileText className="w-4 h-4 mr-1" />
            Files
          </button>
          <button className="btn-ghost text-xs">
            <Database className="w-4 h-4 mr-1" />
            Context
          </button>
          <button className="btn-ghost text-xs">
            <Info className="w-4 h-4 mr-1" />
            Info
          </button>
        </div>
      </div>
    </div>
  );
}; 