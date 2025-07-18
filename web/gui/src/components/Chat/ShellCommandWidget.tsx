import React, { useState } from 'react';
import { Terminal, ChevronDown, ChevronRight, Clock, CheckCircle, XCircle } from 'lucide-react';
import { ShellOperation } from '@/types';

interface ShellCommandWidgetProps {
  operations: ShellOperation[];
  onTerminalOpen?: (operation: ShellOperation) => void;
}

export const ShellCommandWidget: React.FC<ShellCommandWidgetProps> = ({ 
  operations, 
  onTerminalOpen 
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  
  const totalCommands = operations.length;
  const failedCommands = operations.filter(op => op.exitCode !== 0).length;
  const totalDuration = operations.reduce((sum, op) => sum + op.duration, 0);

  const handleClick = () => {
    if (operations.length === 1) {
      onTerminalOpen?.(operations[0]);
    } else {
      setIsExpanded(!isExpanded);
    }
  };

  const getStatusIcon = (exitCode: number) => {
    return exitCode === 0 ? (
      <CheckCircle className="w-4 h-4 text-green-500" />
    ) : (
      <XCircle className="w-4 h-4 text-red-500" />
    );
  };

  return (
    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 my-2">
      <div 
        className="p-3 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        onClick={handleClick}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Terminal className="w-5 h-5 text-blue-500" />
            <span className="font-medium text-gray-900 dark:text-gray-100">
              {totalCommands === 1 ? 'Shell Command' : `${totalCommands} Shell Commands`}
            </span>
            {totalCommands > 1 && (
              isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />
            )}
          </div>
          <div className="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
            <Clock className="w-4 h-4" />
            <span>{totalDuration.toFixed(2)}s</span>
            {failedCommands > 0 && (
              <span className="text-red-500">({failedCommands} failed)</span>
            )}
          </div>
        </div>
        
        {totalCommands === 1 && (
          <div className="mt-2 text-sm">
            <code className="bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded text-gray-800 dark:text-gray-200">
              {operations[0].command}
            </code>
          </div>
        )}
      </div>

      {isExpanded && totalCommands > 1 && (
        <div className="border-t border-gray-200 dark:border-gray-700">
          {operations.map((op) => (
            <div 
              key={op.id}
              className="p-3 border-b border-gray-200 dark:border-gray-700 last:border-b-0 hover:bg-gray-50 dark:hover:bg-gray-750 cursor-pointer"
              onClick={(e) => {
                e.stopPropagation();
                onTerminalOpen?.(op);
              }}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  {getStatusIcon(op.exitCode)}
                  <code className="text-sm bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded text-gray-800 dark:text-gray-200">
                    {op.command}
                  </code>
                </div>
                <div className="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
                  <span>{op.duration.toFixed(2)}s</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}; 