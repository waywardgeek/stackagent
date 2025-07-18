# StackAgent Desktop App

StackAgent is now available as a native desktop application! Get the full power of the revolutionary AI coding assistant with a professional desktop experience.

## 🖥️ Desktop vs Web

| Feature | Desktop App | Web Version |
|---------|-------------|-------------|
| **Experience** | Native window, menus, shortcuts | Browser-based |
| **Installation** | Standalone installer | No installation |
| **System Integration** | System tray, notifications | Limited |
| **Debugging** | Electron DevTools | Browser DevTools |
| **Updates** | Auto-updates (planned) | Refresh page |
| **Offline** | Limited offline features | Requires internet |

## 🚀 Getting Started

### Quick Demo
```bash
# Run the complete desktop app demo
./electron_demo.sh
```

### Manual Setup
```bash
# 1. Install dependencies
cd web/gui
npm install

# 2. Build the frontend
npm run build

# 3. Start the backend server
cd ../..
go run cmd/stackagent-server/main.go &

# 4. Launch desktop app
cd web/gui
npm run electron
```

## 🎛️ Application Menu

The desktop app includes a full application menu with keyboard shortcuts:

### File Menu
- **New Session** (`Ctrl+N`) - Start a fresh session
- **Open Context** (`Ctrl+O`) - Load saved context
- **Exit** (`Ctrl+Q`) - Quit the application

### Edit Menu
- **Undo** (`Ctrl+Z`) - Undo last action
- **Redo** (`Ctrl+Y`) - Redo last action
- **Cut** (`Ctrl+X`) - Cut selection
- **Copy** (`Ctrl+C`) - Copy selection
- **Paste** (`Ctrl+V`) - Paste from clipboard
- **Select All** (`Ctrl+A`) - Select all content

### View Menu
- **Toggle Theme** (`Ctrl+Shift+T`) - Switch dark/light theme
- **Toggle Sidebar** (`Ctrl+B`) - Show/hide sidebar
- **Reload** (`Ctrl+R`) - Reload the app
- **Toggle DevTools** (`F12`) - Open developer tools
- **Zoom In/Out** (`Ctrl+Plus/Minus`) - Adjust zoom level
- **Full Screen** (`F11`) - Toggle full screen

### StackAgent Menu
- **Save Context** (`Ctrl+S`) - Save current context
- **Clear All** (`Ctrl+Shift+C`) - Clear all data
- **Settings** (`Ctrl+,`) - Open settings panel

### Help Menu
- **About StackAgent** - Show app information
- **Learn More** - Open documentation

## 🔧 Development Commands

### Development Mode
```bash
# Start with hot reload (frontend + backend)
npm run electron:dev
```

### Building for Distribution
```bash
# Package app (creates unpacked app)
npm run electron:pack

# Build installer for current platform
npm run electron:dist

# Build for all platforms (Windows, macOS, Linux)
npm run electron:dist-all
```

## 📁 File Structure

```
web/gui/
├── public/
│   ├── electron.js         # Main Electron process
│   ├── preload.js          # Preload script for security
│   └── favicon.ico         # App icon
├── src/
│   ├── hooks/
│   │   └── useElectron.ts  # Electron integration hook
│   └── types/
│       └── electron.d.ts   # TypeScript definitions
├── dist/                   # Built React app
├── dist-electron/          # Packaged Electron app
└── package.json           # Electron configuration
```

## 🔒 Security Features

The desktop app implements modern Electron security practices:

- **Context Isolation** - Renderer process is isolated from Node.js
- **Preload Scripts** - Secure IPC communication
- **No Node Integration** - Renderer cannot access Node.js APIs directly
- **Content Security Policy** - Prevents code injection
- **Secure Defaults** - All external links open in system browser

## 🎨 Platform Integration

### Windows
- **NSIS Installer** - Professional Windows installer
- **Start Menu Integration** - Appears in Windows Start Menu
- **Taskbar Support** - Native taskbar integration

### macOS
- **DMG Installer** - Standard macOS disk image
- **App Store Ready** - Prepared for App Store submission
- **Native Menus** - macOS-style menu bar

### Linux
- **AppImage** - Portable Linux application
- **Desktop Entry** - Integrates with desktop environment
- **System Tray** - Linux system tray support

## 🐛 Debugging

### Desktop App Debugging
```bash
# Start with DevTools open
npm run electron:dev

# In the app, press F12 or Ctrl+Shift+I to open DevTools
```

### Web Version Debugging
```bash
# Start backend server
go run cmd/stackagent-server/main.go

# Open browser to http://localhost:8080
# Use browser DevTools for debugging
```

## 🔄 Updates

### Manual Updates
1. Pull latest code: `git pull`
2. Install dependencies: `npm install`
3. Rebuild: `npm run build`
4. Restart app: `npm run electron`

### Auto-Updates (Planned)
- Automatic update checking
- Background downloads
- Silent installation
- Rollback on failure

## 📊 Performance

The desktop app provides:
- **Faster startup** - No browser overhead
- **Better memory management** - Dedicated process
- **Native performance** - Direct system integration
- **Optimized rendering** - Chromium engine optimizations

## 🔧 Troubleshooting

### Common Issues

**App won't start**
- Check Node.js version: `node --version` (requires 18+)
- Reinstall dependencies: `rm -rf node_modules && npm install`
- Check Electron version: `npm list electron`

**Build fails**
- Clear build cache: `rm -rf dist dist-electron`
- Check TypeScript: `npm run type-check`
- Verify all dependencies: `npm audit`

**WebSocket connection fails**
- Ensure backend is running: `go run cmd/stackagent-server/main.go`
- Check port availability: `lsof -i :8080`
- Verify firewall settings

### Debug Mode
```bash
# Enable debug logging
DEBUG=* npm run electron:dev

# Or specific modules
DEBUG=electron:* npm run electron:dev
```

## 🎯 Next Steps

1. **Test the desktop app** - Run `./electron_demo.sh`
2. **Try all menu options** - Explore File, Edit, View, StackAgent menus
3. **Use keyboard shortcuts** - More efficient than mouse clicks
4. **Compare with web version** - See the differences
5. **Build for distribution** - Create installer for your platform

The desktop app gives you the full StackAgent experience with professional desktop integration! 🚀 