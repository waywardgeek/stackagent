import React, { useEffect, useRef } from 'react';
import { useAppStore } from '@/store';
import { DualPaneLayout } from '@/components/Layout/DualPaneLayout';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import { useElectron } from '@/hooks/useElectron';

const App: React.FC = () => {
  const { 
    theme, 
    connected, 
    setConnected, 
    addNotification 
  } = useAppStore();
  

  
  // Initialize WebSocket connection
  const { connect, disconnect, isConnected, sendMessage } = useWebSocket();
  const connectRef = useRef(connect);
  const disconnectRef = useRef(disconnect);
  const sendMessageRef = useRef(sendMessage);
  
  // Update refs when functions change
  useEffect(() => {
    connectRef.current = connect;
    disconnectRef.current = disconnect;
    sendMessageRef.current = sendMessage;
  }, [connect, disconnect, sendMessage]);
  
  // Set up keyboard shortcuts
  useKeyboardShortcuts();
  
  // Initialize Electron integration
  const { isElectron, platform, version } = useElectron();
  
  // Apply theme to document
  useEffect(() => {
    const root = document.documentElement;
    if (theme === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }, [theme]);
  
  // Initialize WebSocket connection on mount (only once)
  useEffect(() => {
    const timer = setTimeout(() => {
      connectRef.current();
    }, 100); // Small delay to ensure everything is ready
    
    return () => {
      clearTimeout(timer);
      // Don't disconnect on cleanup - let the WebSocket hook handle it
    };
  }, []);
  
  // Update store with connection status
  useEffect(() => {
    setConnected(isConnected);
  }, [isConnected, setConnected]);
  
  // Show connection status notifications
  useEffect(() => {
    if (connected) {
      addNotification({
        type: 'success',
        title: 'Connected',
        message: 'Successfully connected to StackAgent backend',
        timestamp: new Date(),
      });
    } else {
      addNotification({
        type: 'error',
        title: 'Disconnected',
        message: 'Lost connection to StackAgent backend',
        timestamp: new Date(),
      });
    }
  }, [connected, addNotification]);
  
  // Show Electron app info on startup
  useEffect(() => {
    if (isElectron) {
      addNotification({
        type: 'info',
        title: 'Desktop App',
        message: `Running StackAgent ${version} on ${platform}`,
        timestamp: new Date(),
      });
    }
  }, [isElectron, version, platform, addNotification]);

  // Handle window resize events
  useEffect(() => {
    const handleResize = () => {
      // Force a layout recalculation
      const root = document.getElementById('root');
      if (root) {
        root.style.height = '100vh';
        root.style.display = 'block';
        // Trigger reflow
        root.offsetHeight;
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return (
    <div className="App h-full">
      <DualPaneLayout sendMessage={sendMessageRef.current} />
    </div>
  );
};

export default App; 