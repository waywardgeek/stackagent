import { useHotkeys } from 'react-hotkeys-hook';
import { useAppStore } from '@/store';

export const useKeyboardShortcuts = () => {
  const {
    toggleTheme,
    toggleSidebar,
    selectFunctionCall,
    functionCalls,
    clearMessages,
    clearFunctionCalls,
    clearCommandExecutions,
    addNotification,
  } = useAppStore();
  
  // Theme toggle
  useHotkeys('ctrl+shift+t', () => {
    toggleTheme();
    addNotification({
      type: 'info',
      title: 'Theme Toggled',
      message: 'Switched color theme',
      timestamp: new Date(),
      duration: 2000,
    });
  }, {
    preventDefault: true,
    description: 'Toggle dark/light theme',
  });
  
  // Sidebar toggle
  useHotkeys('ctrl+b', () => {
    toggleSidebar();
  }, {
    preventDefault: true,
    description: 'Toggle sidebar',
  });
  
  // Command palette (future implementation)
  useHotkeys('ctrl+k', () => {
    // TODO: Implement command palette
    addNotification({
      type: 'info',
      title: 'Command Palette',
      message: 'Command palette coming soon!',
      timestamp: new Date(),
      duration: 3000,
    });
  }, {
    preventDefault: true,
    description: 'Open command palette',
  });
  
  // Navigate function calls
  useHotkeys('ctrl+j', () => {
    const currentIndex = functionCalls.findIndex(fc => fc.selected);
    const nextIndex = currentIndex < functionCalls.length - 1 ? currentIndex + 1 : 0;
    
    if (functionCalls[nextIndex]) {
      selectFunctionCall(functionCalls[nextIndex].id);
    }
  }, {
    preventDefault: true,
    description: 'Navigate to next function call',
  });
  
  useHotkeys('ctrl+shift+j', () => {
    const currentIndex = functionCalls.findIndex(fc => fc.selected);
    const prevIndex = currentIndex > 0 ? currentIndex - 1 : functionCalls.length - 1;
    
    if (functionCalls[prevIndex]) {
      selectFunctionCall(functionCalls[prevIndex].id);
    }
  }, {
    preventDefault: true,
    description: 'Navigate to previous function call',
  });
  
  // Clear actions
  useHotkeys('ctrl+shift+c', () => {
    clearMessages();
    clearFunctionCalls();
    clearCommandExecutions();
    addNotification({
      type: 'success',
      title: 'Chat Cleared',
      message: 'All messages and function calls cleared',
      timestamp: new Date(),
      duration: 3000,
    });
  }, {
    preventDefault: true,
    description: 'Clear all messages and function calls',
  });
  
  // Focus input (future implementation)
  useHotkeys('/', () => {
    const inputElement = document.querySelector('[data-chat-input]') as HTMLInputElement;
    if (inputElement) {
      inputElement.focus();
    }
  }, {
    preventDefault: true,
    description: 'Focus chat input',
  });
  
  // Copy selected function call output
  useHotkeys('ctrl+c', () => {
    const selectedFC = functionCalls.find(fc => fc.selected);
    if (selectedFC && selectedFC.result) {
      navigator.clipboard.writeText(JSON.stringify(selectedFC.result, null, 2));
      addNotification({
        type: 'success',
        title: 'Copied',
        message: 'Function call result copied to clipboard',
        timestamp: new Date(),
        duration: 2000,
      });
    }
  }, {
    preventDefault: false, // Allow default copy behavior when no function call is selected
    description: 'Copy selected function call result',
  });
  
  // Refresh/reload
  useHotkeys('ctrl+r', () => {
    window.location.reload();
  }, {
    preventDefault: true,
    description: 'Reload application',
  });
  
  // Help dialog (future implementation)
  useHotkeys('ctrl+/', () => {
    // TODO: Implement help dialog
    addNotification({
      type: 'info',
      title: 'Keyboard Shortcuts',
      message: 'Help dialog coming soon!',
      timestamp: new Date(),
      duration: 3000,
    });
  }, {
    preventDefault: true,
    description: 'Show keyboard shortcuts help',
  });
  
  // Quick actions with numbers
  useHotkeys('ctrl+1', () => {
    // TODO: Quick action 1
    addNotification({
      type: 'info',
      title: 'Quick Action 1',
      message: 'Quick actions coming soon!',
      timestamp: new Date(),
      duration: 2000,
    });
  }, {
    preventDefault: true,
    description: 'Quick action 1',
  });
  
  useHotkeys('ctrl+2', () => {
    // TODO: Quick action 2
    addNotification({
      type: 'info',
      title: 'Quick Action 2',
      message: 'Quick actions coming soon!',
      timestamp: new Date(),
      duration: 2000,
    });
  }, {
    preventDefault: true,
    description: 'Quick action 2',
  });
  
  useHotkeys('ctrl+3', () => {
    // TODO: Quick action 3
    addNotification({
      type: 'info',
      title: 'Quick Action 3',
      message: 'Quick actions coming soon!',
      timestamp: new Date(),
      duration: 2000,
    });
  }, {
    preventDefault: true,
    description: 'Quick action 3',
  });
  
  // Search functionality (future implementation)
  useHotkeys('ctrl+f', () => {
    // TODO: Implement search
    addNotification({
      type: 'info',
      title: 'Search',
      message: 'Search functionality coming soon!',
      timestamp: new Date(),
      duration: 3000,
    });
  }, {
    preventDefault: true,
    description: 'Search in chat',
  });
  
  // Export session (future implementation)
  useHotkeys('ctrl+e', () => {
    // TODO: Implement export
    addNotification({
      type: 'info',
      title: 'Export',
      message: 'Export functionality coming soon!',
      timestamp: new Date(),
      duration: 3000,
    });
  }, {
    preventDefault: true,
    description: 'Export current session',
  });
  
  // Return available shortcuts for help display
  return {
    shortcuts: [
      { key: 'Ctrl+Shift+T', description: 'Toggle dark/light theme' },
      { key: 'Ctrl+B', description: 'Toggle sidebar' },
      { key: 'Ctrl+K', description: 'Open command palette' },
      { key: 'Ctrl+J', description: 'Navigate to next function call' },
      { key: 'Ctrl+Shift+J', description: 'Navigate to previous function call' },
      { key: 'Ctrl+Shift+C', description: 'Clear all messages and function calls' },
      { key: '/', description: 'Focus chat input' },
      { key: 'Ctrl+C', description: 'Copy selected function call result' },
      { key: 'Ctrl+R', description: 'Reload application' },
      { key: 'Ctrl+/', description: 'Show keyboard shortcuts help' },
      { key: 'Ctrl+1-3', description: 'Quick actions' },
      { key: 'Ctrl+F', description: 'Search in chat' },
      { key: 'Ctrl+E', description: 'Export current session' },
    ],
  };
}; 