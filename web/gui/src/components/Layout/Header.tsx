import React from 'react';
import { useAppStore } from '@/store';
import { 
  Wifi, 
  WifiOff, 
  Moon, 
  Sun, 
  Settings, 
  GitBranch,
  DollarSign,
  Activity
} from 'lucide-react';

export const Header: React.FC = () => {
  const {
    connected,
    theme,
    toggleTheme,
    currentSession,
    contextState,
    isStreaming,
    selectedModel,
    activeCommands,
  } = useAppStore();
  
  return (
    <header className="flex items-center justify-between px-4 py-3 bg-white dark:bg-secondary-800 border-b border-secondary-200 dark:border-secondary-700 shadow-sm">
      {/* Left side - Logo and title */}
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-2">
          <div className="w-8 h-8 bg-gradient-to-r from-primary-500 to-primary-600 rounded-lg flex items-center justify-center">
            <span className="text-white font-bold text-lg">S</span>
          </div>
          <div>
            <h1 className="text-lg font-semibold text-secondary-900 dark:text-secondary-100">
              StackAgent
            </h1>
            <p className="text-xs text-secondary-500 dark:text-secondary-400">
              AI Coding Assistant
            </p>
          </div>
        </div>
        
        {/* Session info */}
        {currentSession && (
          <div className="flex items-center space-x-2 text-sm text-secondary-600 dark:text-secondary-400">
            <div className="w-2 h-2 bg-success-500 rounded-full animate-pulse-fast" />
            <span>Session {currentSession.id.slice(-8)}</span>
          </div>
        )}
      </div>
      
      {/* Center - Context indicators */}
      <div className="flex items-center space-x-4">
        {/* Git branch */}
        {contextState?.gitBranch && (
          <div className="git-branch">
            <GitBranch className="w-4 h-4 mr-1" />
            <span>{contextState.gitBranch}</span>
          </div>
        )}
        
        {/* Active commands */}
        {activeCommands > 0 && (
          <div className="session-info">
            <Activity className="w-4 h-4 mr-1" />
            <span>{activeCommands} running</span>
          </div>
        )}
        
        {/* AI Model */}
        <div className="session-info">
          <span className="text-xs">Model:</span>
          <span className="font-medium ml-1">
            {selectedModel.includes('claude-sonnet-4') ? 'Claude 4' : selectedModel.split('-')[0]}
          </span>
        </div>
        
        {/* Session cost */}
        {currentSession && currentSession.totalCost > 0 && (
          <div className="session-info">
            <DollarSign className="w-4 h-4 mr-1" />
            <span>${currentSession.totalCost.toFixed(4)}</span>
          </div>
        )}
      </div>
      
      {/* Right side - Controls */}
      <div className="flex items-center space-x-3">
        {/* Connection status */}
        <div className="flex items-center space-x-2">
          {connected ? (
            <div className="flex items-center space-x-1 text-success-600 dark:text-success-400">
              <Wifi className="w-4 h-4" />
              <span className="text-sm">Connected</span>
            </div>
          ) : (
            <div className="flex items-center space-x-1 text-error-600 dark:text-error-400">
              <WifiOff className="w-4 h-4" />
              <span className="text-sm">Disconnected</span>
            </div>
          )}
        </div>
        
        {/* Streaming indicator */}
        {isStreaming && (
          <div className="flex items-center space-x-1 text-primary-600 dark:text-primary-400">
            <div className="loading-dots">
              <div className="loading-dot" />
              <div className="loading-dot" />
              <div className="loading-dot" />
            </div>
            <span className="text-sm">AI thinking...</span>
          </div>
        )}
        
        {/* Theme toggle */}
        <button
          onClick={toggleTheme}
          className="p-2 rounded-lg hover:bg-secondary-100 dark:hover:bg-secondary-700 transition-colors"
          aria-label="Toggle theme"
        >
          {theme === 'dark' ? (
            <Sun className="w-5 h-5 text-secondary-600 dark:text-secondary-400" />
          ) : (
            <Moon className="w-5 h-5 text-secondary-600 dark:text-secondary-400" />
          )}
        </button>
        
        {/* Settings */}
        <button
          className="p-2 rounded-lg hover:bg-secondary-100 dark:hover:bg-secondary-700 transition-colors"
          aria-label="Settings"
        >
          <Settings className="w-5 h-5 text-secondary-600 dark:text-secondary-400" />
        </button>
      </div>
    </header>
  );
}; 