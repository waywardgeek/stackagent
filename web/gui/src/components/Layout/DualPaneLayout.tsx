import React, { useState, useRef, useCallback } from 'react';
import { useAppStore } from '@/store';
import { ChatPane } from '@/components/Chat/ChatPane';
import { ActionPane } from '@/components/Action/ActionPane';
import { Header } from '@/components/Layout/Header';
import { StatusBar } from '@/components/Layout/StatusBar';
import type { WebSocketEventType } from '@/types';

interface DualPaneLayoutProps {
  sendMessage: (type: WebSocketEventType, data: any) => void;
}

export const DualPaneLayout: React.FC<DualPaneLayoutProps> = ({ sendMessage }) => {
  const { leftPaneWidth, setLeftPaneWidth } = useAppStore();
  const [isDragging, setIsDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);
  
  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging || !containerRef.current) return;
    
    const container = containerRef.current;
    const rect = container.getBoundingClientRect();
    const newLeftWidth = ((e.clientX - rect.left) / rect.width) * 100;
    
    // Limit the pane width between 20% and 80%
    const clampedWidth = Math.max(20, Math.min(80, newLeftWidth));
    setLeftPaneWidth(clampedWidth);
  }, [isDragging, setLeftPaneWidth]);
  
  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);
  
  // Add global mouse event listeners when dragging
  React.useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
      
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
      };
    }
  }, [isDragging, handleMouseMove, handleMouseUp]);
  
  return (
    <div className="dual-pane-layout bg-secondary-50 dark:bg-secondary-900">
      {/* Header */}
      <Header />
      
      {/* Main content area with dual panes */}
      <div 
        ref={containerRef}
        className="dual-pane-content relative"
      >
        {/* Left Pane - Chat */}
        <div 
          className="pane bg-white dark:bg-secondary-800 border-r border-secondary-200 dark:border-secondary-700"
          style={{ width: `${leftPaneWidth}%` }}
        >
          <ChatPane sendMessage={sendMessage} />
        </div>
        
        {/* Resizable Divider */}
        <div
          className={`
            w-1 bg-secondary-200 dark:bg-secondary-700 
            hover:bg-secondary-300 dark:hover:bg-secondary-600 
            cursor-col-resize select-none transition-colors
            ${isDragging ? 'bg-primary-500 dark:bg-primary-400' : ''}
          `}
          onMouseDown={handleMouseDown}
          role="separator"
          aria-orientation="vertical"
          aria-label="Resize panes"
        >
          {/* Visual indicator for the divider */}
          <div className="h-full w-full relative">
            <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
              <div className="w-0.5 h-8 bg-secondary-400 dark:bg-secondary-500 rounded-full" />
            </div>
          </div>
        </div>
        
        {/* Right Pane - Action Details */}
        <div 
          className="pane bg-white dark:bg-secondary-800"
          style={{ width: `${100 - leftPaneWidth}%` }}
        >
          <ActionPane sendMessage={sendMessage} />
        </div>
      </div>
      
      {/* Status Bar */}
      <StatusBar />
    </div>
  );
}; 