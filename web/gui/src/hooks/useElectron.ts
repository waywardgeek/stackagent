import { useEffect, useState } from 'react';
import { useAppStore } from '@/store';
import type { ElectronAPI } from '@/types/electron';

export const useElectron = () => {
  const [isElectron, setIsElectron] = useState(false);
  const [electronAPI, setElectronAPI] = useState<ElectronAPI | null>(null);
  const [platform, setPlatform] = useState<string>('');
  const [version, setVersion] = useState<string>('');
  
  const { 
    toggleTheme, 
    toggleSidebar, 
    clearMessages, 
    clearFunctionCalls,
    clearCommandExecutions,
    clearNotifications,
    addNotification 
  } = useAppStore();

  useEffect(() => {
    // Check if running in Electron
    if (window.electronAPI) {
      setIsElectron(true);
      setElectronAPI(window.electronAPI);
      setPlatform(window.electronAPI.platform);
      
      // Get app version
      window.electronAPI.getVersion().then(setVersion);
      
      // Set up menu event listeners
      const handleNewSession = () => {
        // Clear all data for new session
        clearMessages();
        clearFunctionCalls();
        clearCommandExecutions();
        clearNotifications();
        
        addNotification({
          type: 'info',
          title: 'New Session',
          message: 'Started a new session',
          timestamp: new Date(),
        });
      };
      
      const handleOpenContext = () => {
        addNotification({
          type: 'info',
          title: 'Open Context',
          message: 'Context loading feature coming soon',
          timestamp: new Date(),
        });
      };
      
      const handleSaveContext = () => {
        addNotification({
          type: 'success',
          title: 'Save Context',
          message: 'Context saved successfully',
          timestamp: new Date(),
        });
      };
      
      const handleClearAll = () => {
        clearMessages();
        clearFunctionCalls();
        clearCommandExecutions();
        clearNotifications();
        
        addNotification({
          type: 'info',
          title: 'Cleared All',
          message: 'All data has been cleared',
          timestamp: new Date(),
        });
      };
      
      const handleToggleTheme = () => {
        toggleTheme();
      };
      
      const handleToggleSidebar = () => {
        toggleSidebar();
      };
      
      const handleSettings = () => {
        addNotification({
          type: 'info',
          title: 'Settings',
          message: 'Settings panel coming soon',
          timestamp: new Date(),
        });
      };
      
      // Register menu event listeners
      window.electronAPI.onMenuNewSession(handleNewSession);
      window.electronAPI.onMenuOpenContext(handleOpenContext);
      window.electronAPI.onMenuSaveContext(handleSaveContext);
      window.electronAPI.onMenuClearAll(handleClearAll);
      window.electronAPI.onMenuToggleTheme(handleToggleTheme);
      window.electronAPI.onMenuToggleSidebar(handleToggleSidebar);
      window.electronAPI.onMenuSettings(handleSettings);
      
      // Cleanup listeners on unmount
      return () => {
        window.electronAPI?.removeAllListeners('menu-new-session');
        window.electronAPI?.removeAllListeners('menu-open-context');
        window.electronAPI?.removeAllListeners('menu-save-context');
        window.electronAPI?.removeAllListeners('menu-clear-all');
        window.electronAPI?.removeAllListeners('menu-toggle-theme');
        window.electronAPI?.removeAllListeners('menu-toggle-sidebar');
        window.electronAPI?.removeAllListeners('menu-settings');
      };
    }
  }, [
    toggleTheme,
    toggleSidebar,
    clearMessages,
    clearFunctionCalls,
    clearCommandExecutions,
    clearNotifications,
    addNotification
  ]);

  // App control functions
  const quit = async () => {
    if (electronAPI) {
      await electronAPI.quit();
    }
  };

  const minimize = async () => {
    if (electronAPI) {
      await electronAPI.minimize();
    }
  };

  const maximize = async () => {
    if (electronAPI) {
      await electronAPI.maximize();
    }
  };

  const close = async () => {
    if (electronAPI) {
      await electronAPI.close();
    }
  };

  return {
    isElectron,
    platform,
    version,
    quit,
    minimize,
    maximize,
    close,
  };
}; 