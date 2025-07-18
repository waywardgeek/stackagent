import React from 'react';
import { CommandExecution } from '@/types';

interface CommandOutputProps {
  execution?: CommandExecution;
}

export const CommandOutput: React.FC<CommandOutputProps> = ({ execution }) => {
  if (!execution) {
    return (
      <div className="p-4 text-center text-secondary-600 dark:text-secondary-400">
        No command execution selected
      </div>
    );
  }
  
  return (
    <div className="p-4 space-y-4">
      <div>
        <h3 className="font-medium text-secondary-900 dark:text-secondary-100 mb-2">
          Command: {execution.command}
        </h3>
        <div className={`status-${execution.status}`}>
          {execution.status}
        </div>
      </div>
      
      {execution.output && (
        <div>
          <h4 className="font-medium text-secondary-900 dark:text-secondary-100 mb-2">
            Output
          </h4>
          <div className="code-block">
            <pre className="whitespace-pre-wrap text-sm">{execution.output}</pre>
          </div>
        </div>
      )}
      
      {execution.error && (
        <div>
          <h4 className="font-medium text-error-600 dark:text-error-400 mb-2">
            Error
          </h4>
          <div className="code-block bg-error-50 dark:bg-error-900/20 border-error-200 dark:border-error-800">
            <pre className="text-error-700 dark:text-error-300 whitespace-pre-wrap text-sm">
              {execution.error}
            </pre>
          </div>
        </div>
      )}
      
      <div className="text-sm text-secondary-600 dark:text-secondary-400">
        <p>Handle ID: {execution.handleId}</p>
        <p>Exit Code: {execution.exitCode ?? 'N/A'}</p>
        {execution.lineCount && <p>Lines: {execution.lineCount}</p>}
        {execution.duration && <p>Duration: {execution.duration}ms</p>}
      </div>
    </div>
  );
}; 