import React from 'react';
import { FileText } from 'lucide-react';

export const FilePreview: React.FC = () => {
  return (
    <div className="p-4">
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <FileText className="w-16 h-16 text-secondary-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-secondary-900 dark:text-secondary-100 mb-2">
            File Preview
          </h3>
          <p className="text-secondary-600 dark:text-secondary-400">
            File previews coming soon
          </p>
        </div>
      </div>
    </div>
  );
}; 