import React from 'react';
import { FunctionCall } from '@/types';
import { formatDistanceToNow } from 'date-fns';

interface FunctionCallDetailsProps {
  functionCall?: FunctionCall;
}

export const FunctionCallDetails: React.FC<FunctionCallDetailsProps> = ({ functionCall }) => {
  if (!functionCall) {
    return (
      <div className="p-4 text-center text-secondary-600 dark:text-secondary-400">
        No function call selected
      </div>
    );
  }
  
  return (
    <div className="p-4 space-y-4">
      <div>
        <h3 className="font-medium text-secondary-900 dark:text-secondary-100 mb-2">
          {functionCall.name}
        </h3>
        <div className={`status-${functionCall.status}`}>
          {functionCall.status}
        </div>
      </div>
      
      <div>
        <h4 className="font-medium text-secondary-900 dark:text-secondary-100 mb-2">
          Arguments
        </h4>
        <div className="code-block">
          <pre>{JSON.stringify(functionCall.arguments, null, 2)}</pre>
        </div>
      </div>
      
      {functionCall.result && (
        <div>
          <h4 className="font-medium text-secondary-900 dark:text-secondary-100 mb-2">
            Result
          </h4>
          <div className="code-block">
            <pre>{JSON.stringify(functionCall.result, null, 2)}</pre>
          </div>
        </div>
      )}
      
      {functionCall.error && (
        <div>
          <h4 className="font-medium text-error-600 dark:text-error-400 mb-2">
            Error
          </h4>
          <div className="code-block bg-error-50 dark:bg-error-900/20 border-error-200 dark:border-error-800">
            <pre className="text-error-700 dark:text-error-300">{functionCall.error}</pre>
          </div>
        </div>
      )}
      
      <div className="text-sm text-secondary-600 dark:text-secondary-400">
        <p>Started: {formatDistanceToNow(new Date(functionCall.startTime), { addSuffix: true })}</p>
        {functionCall.endTime && (
          <p>Ended: {formatDistanceToNow(new Date(functionCall.endTime), { addSuffix: true })}</p>
        )}
        {functionCall.duration && (
          <p>Duration: {functionCall.duration}ms</p>
        )}
      </div>
    </div>
  );
}; 