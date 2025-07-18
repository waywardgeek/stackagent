import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Terminal, 
  FileText, 
  Settings, 
  Clock, 
  Loader2,
  Activity,
  Zap
} from 'lucide-react';
import { useAppStore, selectActiveFunctionCalls, selectActiveShellCommands, selectActiveFileOperations, selectOperationMetrics } from '@/store';
import type { FunctionCall, ShellOperation, FileOperation } from '@/types';

interface LiveOperationProps {
  operation: FunctionCall | ShellOperation | FileOperation;
  type: 'function' | 'shell' | 'file';
}

const LiveOperation: React.FC<LiveOperationProps> = ({ operation, type }) => {
  const [elapsedTime, setElapsedTime] = useState(0);

  useEffect(() => {
    const startTime = 'startTime' in operation ? operation.startTime : operation.timestamp;
    const timer = setInterval(() => {
      setElapsedTime(Date.now() - startTime.getTime());
    }, 100);

    return () => clearInterval(timer);
  }, [operation]);

  const formatTime = (ms: number) => {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`;
    }
    return `${seconds}s`;
  };

  const getIcon = () => {
    switch (type) {
      case 'function':
        return <Settings className="w-4 h-4" />;
      case 'shell':
        return <Terminal className="w-4 h-4" />;
      case 'file':
        return <FileText className="w-4 h-4" />;
      default:
        return <Activity className="w-4 h-4" />;
    }
  };

  const getTitle = () => {
    switch (type) {
      case 'function':
        return (operation as FunctionCall).name;
      case 'shell':
        return (operation as ShellOperation).command;
      case 'file':
        const fileOp = operation as FileOperation;
        return `${fileOp.type}: ${fileOp.filePath}`;
      default:
        return 'Unknown operation';
    }
  };

  const getOutput = () => {
    if ('streamingOutput' in operation && operation.streamingOutput) {
      return operation.streamingOutput;
    }
    if ('output' in operation && operation.output) {
      return operation.output;
    }
    if ('streamingContent' in operation && operation.streamingContent) {
      return operation.streamingContent;
    }
    if ('content' in operation && operation.content) {
      return operation.content;
    }
    return '';
  };

  const progress = operation.progress || 0;
  const hasOutput = getOutput().length > 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 mb-3 border border-gray-200 dark:border-gray-600"
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <div className="p-1 bg-blue-100 dark:bg-blue-900 rounded">
            {getIcon()}
          </div>
          <span className="font-medium text-gray-900 dark:text-gray-100 truncate">
            {getTitle()}
          </span>
          <div className="flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400">
            <Clock className="w-3 h-3" />
            {formatTime(elapsedTime)}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {progress > 0 && (
            <div className="text-xs text-gray-500 dark:text-gray-400">
              {Math.round(progress * 100)}%
            </div>
          )}
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
          >
            <Loader2 className="w-4 h-4 text-blue-500" />
          </motion.div>
        </div>
      </div>

      {progress > 0 && (
        <div className="mb-3">
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <motion.div
              className="bg-blue-500 h-2 rounded-full"
              initial={{ width: 0 }}
              animate={{ width: `${progress * 100}%` }}
              transition={{ duration: 0.3 }}
            />
          </div>
        </div>
      )}

      {hasOutput && (
        <div className="mt-3">
          <div className="bg-gray-900 dark:bg-gray-950 rounded p-3 overflow-auto max-h-32">
            <pre className="text-xs text-green-400 font-mono whitespace-pre-wrap">
              {getOutput()}
            </pre>
          </div>
        </div>
      )}
    </motion.div>
  );
};

interface LiveOperationsDisplayProps {
  className?: string;
}

const LiveOperationsDisplay: React.FC<LiveOperationsDisplayProps> = ({ className = '' }) => {
  const activeFunctionCalls = useAppStore(selectActiveFunctionCalls);
  const activeShellCommands = useAppStore(selectActiveShellCommands);
  const activeFileOperations = useAppStore(selectActiveFileOperations);
  const operationMetrics = useAppStore(selectOperationMetrics);

  const totalActiveOperations = activeFunctionCalls.length + activeShellCommands.length + activeFileOperations.length;

  if (totalActiveOperations === 0) {
    return null;
  }

  return (
    <div className={`${className}`}>
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="bg-white dark:bg-gray-900 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-4 mb-4"
      >
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
              <Activity className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                Live Operations
              </h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                {totalActiveOperations} operation{totalActiveOperations !== 1 ? 's' : ''} in progress
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
            <div className="flex items-center gap-1">
              <Zap className="w-4 h-4" />
              {operationMetrics.operationsPerMinute}/min
            </div>
            <div className="flex items-center gap-1">
              <Clock className="w-4 h-4" />
              {operationMetrics.averageResponseTime > 0 ? 
                `${operationMetrics.averageResponseTime.toFixed(1)}ms avg` : 
                'N/A'
              }
            </div>
          </div>
        </div>

        <div className="space-y-3">
          <AnimatePresence mode="popLayout">
            {activeFunctionCalls.map((operation) => (
              <LiveOperation
                key={operation.id}
                operation={operation}
                type="function"
              />
            ))}
            {activeShellCommands.map((operation) => (
              <LiveOperation
                key={operation.id}
                operation={operation}
                type="shell"
              />
            ))}
            {activeFileOperations.map((operation) => (
              <LiveOperation
                key={operation.id}
                operation={operation}
                type="file"
              />
            ))}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  );
};

export default LiveOperationsDisplay; 