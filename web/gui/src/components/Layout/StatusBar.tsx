import React from 'react';
import { useAppStore } from '@/store';
import { 
  Database, 
  MessageSquare, 
  Command, 
  FileText, 
  Brain, 
  Clock,
  HardDrive,
  Zap
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

export const StatusBar: React.FC = () => {
  const {
    contextState,
    messages,
    functionCalls,
    activeCommands,
    protectedMemory,
    workspaceState,
    knowledgeBase,
    reconnectAttempts,
    connected,
  } = useAppStore();
  
  return (
    <footer className="flex items-center justify-between px-4 py-2 bg-secondary-100 dark:bg-secondary-800 border-t border-secondary-200 dark:border-secondary-700 text-sm">
      {/* Left side - Context statistics */}
      <div className="flex items-center space-x-6">
        {/* Memory entries */}
        <div className="flex items-center space-x-1 text-secondary-600 dark:text-secondary-400">
          <Database className="w-4 h-4" />
          <span>{Object.keys(protectedMemory).length} memory</span>
        </div>
        
        {/* Knowledge entries */}
        <div className="flex items-center space-x-1 text-secondary-600 dark:text-secondary-400">
          <Brain className="w-4 h-4" />
          <span>{Object.keys(knowledgeBase).length} knowledge</span>
        </div>
        
        {/* Messages */}
        <div className="flex items-center space-x-1 text-secondary-600 dark:text-secondary-400">
          <MessageSquare className="w-4 h-4" />
          <span>{messages.length} messages</span>
        </div>
        
        {/* Function calls */}
        <div className="flex items-center space-x-1 text-secondary-600 dark:text-secondary-400">
          <Zap className="w-4 h-4" />
          <span>{functionCalls.length} functions</span>
        </div>
        
        {/* Active files */}
        {workspaceState && workspaceState.activeFiles.length > 0 && (
          <div className="flex items-center space-x-1 text-secondary-600 dark:text-secondary-400">
            <FileText className="w-4 h-4" />
            <span>{workspaceState.activeFiles.length} files</span>
          </div>
        )}
        
        {/* Command history */}
        {contextState && contextState.commandHistory > 0 && (
          <div className="flex items-center space-x-1 text-secondary-600 dark:text-secondary-400">
            <Command className="w-4 h-4" />
            <span>{contextState.commandHistory} commands</span>
          </div>
        )}
      </div>
      
      {/* Center - Working directory */}
      {workspaceState?.workingDir && (
        <div className="flex items-center space-x-1 text-secondary-600 dark:text-secondary-400">
          <HardDrive className="w-4 h-4" />
          <span className="font-mono text-xs">
            {workspaceState.workingDir.length > 50 
              ? `...${workspaceState.workingDir.slice(-50)}`
              : workspaceState.workingDir
            }
          </span>
        </div>
      )}
      
      {/* Right side - Status indicators */}
      <div className="flex items-center space-x-4">
        {/* Last activity */}
        {contextState?.lastActivity && (
          <div className="flex items-center space-x-1 text-secondary-600 dark:text-secondary-400">
            <Clock className="w-4 h-4" />
            <span>
              {formatDistanceToNow(new Date(contextState.lastActivity), { addSuffix: true })}
            </span>
          </div>
        )}
        
        {/* Connection status */}
        <div className="flex items-center space-x-1">
          {connected ? (
            <div className="flex items-center space-x-1 text-success-600 dark:text-success-400">
              <div className="w-2 h-2 bg-success-500 rounded-full" />
              <span>Online</span>
            </div>
          ) : (
            <div className="flex items-center space-x-1 text-error-600 dark:text-error-400">
              <div className="w-2 h-2 bg-error-500 rounded-full" />
              <span>
                {reconnectAttempts > 0 ? `Reconnecting (${reconnectAttempts})` : 'Offline'}
              </span>
            </div>
          )}
        </div>
        
        {/* Active commands indicator */}
        {activeCommands > 0 && (
          <div className="flex items-center space-x-1 text-primary-600 dark:text-primary-400">
            <div className="loading-spinner" />
            <span>{activeCommands} running</span>
          </div>
        )}
        
        {/* Session ID */}
        {contextState?.sessionId && (
          <div className="text-xs text-secondary-500 dark:text-secondary-400 font-mono">
            Session: {contextState.sessionId.slice(-8)}
          </div>
        )}
      </div>
    </footer>
  );
}; 