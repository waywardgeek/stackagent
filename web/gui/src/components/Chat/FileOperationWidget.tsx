import React, { useState } from 'react';
import { 
  File, 
  FileText, 
  Edit3, 
  Search, 
  FolderOpen, 
  ChevronDown, 
  ChevronRight,
  Eye,
  Diff
} from 'lucide-react';
import { FileOperation } from '@/types';

interface FileOperationWidgetProps {
  operations: FileOperation[];
  onFileOpen?: (operation: FileOperation) => void;
}

export const FileOperationWidget: React.FC<FileOperationWidgetProps> = ({ 
  operations, 
  onFileOpen 
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  
  const totalOperations = operations.length;
  const uniqueFiles = [...new Set(operations.map(op => op.filePath))].length;
  const operationTypes = [...new Set(operations.map(op => op.type))];

  const handleClick = () => {
    if (operations.length === 1) {
      onFileOpen?.(operations[0]);
    } else {
      setIsExpanded(!isExpanded);
    }
  };

  const getOperationIcon = (type: string) => {
    switch (type) {
      case 'read':
        return <FileText className="w-4 h-4 text-blue-500" />;
      case 'write':
        return <File className="w-4 h-4 text-green-500" />;
      case 'edit':
        return <Edit3 className="w-4 h-4 text-orange-500" />;
      case 'search':
        return <Search className="w-4 h-4 text-purple-500" />;
      case 'list':
        return <FolderOpen className="w-4 h-4 text-gray-500" />;
      default:
        return <File className="w-4 h-4 text-gray-500" />;
    }
  };

  const getOperationSummary = () => {
    if (totalOperations === 1) {
      const op = operations[0];
      return `${op.type} ${op.filePath}`;
    }
    
    if (uniqueFiles === 1) {
      return `${totalOperations} operations on ${operations[0].filePath}`;
    }
    
    return `${totalOperations} operations on ${uniqueFiles} files`;
  };

  const formatFileSize = (size?: number) => {
    if (!size) return '';
    if (size < 1024) return `${size}B`;
    if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)}KB`;
    return `${(size / 1024 / 1024).toFixed(1)}MB`;
  };

  const getActionIcon = (operation: FileOperation) => {
    if (operation.type === 'edit' && operation.changes) {
      return <Diff className="w-4 h-4" />;
    }
    return <Eye className="w-4 h-4" />;
  };

  return (
    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 my-2">
      <div 
        className="p-3 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        onClick={handleClick}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            {getOperationIcon(operations[0].type)}
            <span className="font-medium text-gray-900 dark:text-gray-100">
              File Operations
            </span>
            {totalOperations > 1 && (
              isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />
            )}
          </div>
          <div className="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
            <span>{operationTypes.join(', ')}</span>
            {uniqueFiles > 1 && <span>• {uniqueFiles} files</span>}
          </div>
        </div>
        
        {totalOperations === 1 && (
          <div className="mt-2 text-sm">
            <code className="bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded text-gray-800 dark:text-gray-200">
              {getOperationSummary()}
            </code>
            {operations[0].size && (
              <span className="ml-2 text-gray-500 dark:text-gray-400">
                ({formatFileSize(operations[0].size)})
              </span>
            )}
          </div>
        )}
      </div>

      {isExpanded && totalOperations > 1 && (
        <div className="border-t border-gray-200 dark:border-gray-700">
          {operations.map((op) => (
            <div 
              key={op.id}
              className="p-3 border-b border-gray-200 dark:border-gray-700 last:border-b-0 hover:bg-gray-50 dark:hover:bg-gray-750 cursor-pointer"
              onClick={(e) => {
                e.stopPropagation();
                onFileOpen?.(op);
              }}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  {getOperationIcon(op.type)}
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {op.type}
                  </span>
                  <code className="text-sm bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded text-gray-800 dark:text-gray-200">
                    {op.filePath}
                  </code>
                </div>
                <div className="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
                  {op.size && <span>{formatFileSize(op.size)}</span>}
                  {getActionIcon(op)}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}; 