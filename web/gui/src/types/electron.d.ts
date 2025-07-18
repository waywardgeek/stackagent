// TypeScript declarations for Electron API
export interface ElectronAPI {
  // App controls
  getVersion(): Promise<string>;
  quit(): Promise<void>;
  minimize(): Promise<void>;
  maximize(): Promise<void>;
  close(): Promise<void>;
  
  // Menu handlers
  onMenuNewSession(callback: () => void): void;
  onMenuOpenContext(callback: () => void): void;
  onMenuSaveContext(callback: () => void): void;
  onMenuClearAll(callback: () => void): void;
  onMenuToggleTheme(callback: () => void): void;
  onMenuToggleSidebar(callback: () => void): void;
  onMenuSettings(callback: () => void): void;
  
  // Remove listeners
  removeAllListeners(channel: string): void;
  
  // Platform info
  platform: string;
  
  // Check if running in Electron
  isElectron: boolean;
}

export interface ElectronConsole {
  log(message: any): void;
  error(message: any): void;
  warn(message: any): void;
}

declare global {
  interface Window {
    electronAPI?: ElectronAPI;
    electronConsole?: ElectronConsole;
  }
}

export {}; 