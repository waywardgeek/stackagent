import React, { useState, useEffect, useRef } from 'react';
import { Terminal, X, Maximize2, Minimize2, Copy, Check } from 'lucide-react';
import { ShellOperation } from '@/types';

interface TerminalPaneProps {
  isVisible: boolean;
  onClose: () => void;
  currentOperation?: ShellOperation;
  operations: ShellOperation[];
}

export const TerminalPane: React.FC<TerminalPaneProps> = ({ 
  isVisible, 
  onClose, 
  currentOperation, 
  operations 
}) => {
  const [isMaximized, setIsMaximized] = useState(false);
  const [copied, setCopied] = useState(false);
  const outputRef = useRef<HTMLDivElement>(null);
  
  useEffect(() => {
    if (currentOperation && outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [currentOperation]);

  const handleCopy = async () => {
    if (currentOperation) {
      await navigator.clipboard.writeText(currentOperation.output);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!isVisible) return null;

  const operation = currentOperation || operations[operations.length - 1];
  
  if (!operation) return null;

  return (
    <div className={`
      fixed right-0 top-0 h-full bg-gray-900 text-green-400 font-mono text-sm
      border-l border-gray-700 shadow-xl z-50 transition-all duration-300
      ${isMaximized ? 'w-full' : 'w-96'}
    `}>
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-gray-700 bg-gray-800">
        <div className="flex items-center space-x-2">
          <Terminal className="w-5 h-5 text-green-400" />
          <span className="font-medium text-white">Terminal</span>
        </div>
        <div className="flex items-center space-x-2">
          <button
            onClick={handleCopy}
            className="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors"
            title="Copy output"
          >
            {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
          </button>
          <button
            onClick={() => setIsMaximized(!isMaximized)}
            className="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors"
            title={isMaximized ? "Restore" : "Maximize"}
          >
            {isMaximized ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}
          </button>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors"
            title="Close"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Command Info */}
      <div className="p-3 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center justify-between mb-2">
          <span className="text-gray-300">Command:</span>
          <span className="text-sm text-gray-500">{operation.workingDir}</span>
        </div>
        <div className="bg-gray-900 p-2 rounded">
          <code className="text-green-400">{operation.command}</code>
        </div>
        <div className="flex items-center justify-between mt-2 text-sm">
          <span className="text-gray-400">
            Exit code: <span className={operation.exitCode === 0 ? 'text-green-400' : 'text-red-400'}>
              {operation.exitCode}
            </span>
          </span>
          <span className="text-gray-400">Duration: {operation.duration.toFixed(2)}s</span>
        </div>
      </div>

      {/* Output */}
      <div className="flex-1 overflow-hidden">
        <div 
          ref={outputRef}
          className="h-full overflow-y-auto p-3 bg-gray-900"
          style={{ maxHeight: 'calc(100vh - 200px)' }}
        >
          <pre className="whitespace-pre-wrap text-green-400 text-sm leading-relaxed">
            {operation.output || 'No output'}
          </pre>
        </div>
      </div>

      {/* Footer with operation selector if multiple operations */}
      {operations.length > 1 && (
        <div className="border-t border-gray-700 bg-gray-800 p-2">
          <div className="text-xs text-gray-400 mb-1">Recent Commands:</div>
          <div className="flex space-x-1 overflow-x-auto">
            {operations.slice(-5).map((op) => (
              <button
                key={op.id}
                onClick={() => {/* Handle operation selection */}}
                className={`
                  text-xs px-2 py-1 rounded whitespace-nowrap
                  ${op.id === operation.id 
                    ? 'bg-green-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                  }
                `}
              >
                {op.command.split(' ')[0]}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}; 