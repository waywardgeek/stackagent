import React from 'react';
import { useAppStore } from '@/store';
import { Zap, Clock, CheckCircle, XCircle } from 'lucide-react';

export const FunctionCallList: React.FC = () => {
  const { functionCalls, selectFunctionCall } = useAppStore();
  
  return (
    <div className="space-y-2 p-4">
      {functionCalls.map((fc) => (
        <div
          key={fc.id}
          onClick={() => selectFunctionCall(fc.id)}
          className={`
            function-call cursor-pointer transition-all
            ${fc.status === 'completed' ? 'function-call-completed' : ''}
            ${fc.status === 'failed' ? 'function-call-failed' : ''}
            ${fc.status === 'running' ? 'function-call-running' : ''}
            ${fc.selected ? 'ring-2 ring-primary-500' : ''}
          `}
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Zap className="w-4 h-4" />
              <span className="font-medium">{fc.name}</span>
              {fc.status === 'running' && <div className="loading-spinner" />}
              {fc.status === 'completed' && <CheckCircle className="w-4 h-4 text-success-600" />}
              {fc.status === 'failed' && <XCircle className="w-4 h-4 text-error-600" />}
            </div>
            <div className="flex items-center space-x-2 text-sm text-secondary-600 dark:text-secondary-400">
              <Clock className="w-4 h-4" />
              <span>{fc.duration || 'Running...'}</span>
            </div>
          </div>
          <div className="text-sm text-secondary-600 dark:text-secondary-400 mt-1">
            {Object.keys(fc.arguments).length > 0 && (
              <span>Arguments: {Object.keys(fc.arguments).join(', ')}</span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}; 