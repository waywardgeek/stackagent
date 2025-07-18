const { contextBridge, ipcRenderer } = require('electron');

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld('electronAPI', {
  // App controls
  getVersion: () => ipcRenderer.invoke('app-version'),
  quit: () => ipcRenderer.invoke('app-quit'),
  minimize: () => ipcRenderer.invoke('app-minimize'),
  maximize: () => ipcRenderer.invoke('app-maximize'),
  close: () => ipcRenderer.invoke('app-close'),
  
  // Menu handlers
  onMenuNewSession: (callback) => ipcRenderer.on('menu-new-session', callback),
  onMenuOpenContext: (callback) => ipcRenderer.on('menu-open-context', callback),
  onMenuSaveContext: (callback) => ipcRenderer.on('menu-save-context', callback),
  onMenuClearAll: (callback) => ipcRenderer.on('menu-clear-all', callback),
  onMenuToggleTheme: (callback) => ipcRenderer.on('menu-toggle-theme', callback),
  onMenuToggleSidebar: (callback) => ipcRenderer.on('menu-toggle-sidebar', callback),
  onMenuSettings: (callback) => ipcRenderer.on('menu-settings', callback),
  
  // Remove listeners
  removeAllListeners: (channel) => ipcRenderer.removeAllListeners(channel),
  
  // Platform info
  platform: process.platform,
  
  // Check if running in Electron
  isElectron: true
});

// Expose console for debugging
contextBridge.exposeInMainWorld('electronConsole', {
  log: (message) => console.log(message),
  error: (message) => console.error(message),
  warn: (message) => console.warn(message)
}); 