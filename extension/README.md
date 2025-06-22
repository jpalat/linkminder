# BookMinder Browser Extension

A cross-browser extension that saves bookmarks with page content to the BookMinder API. Works on Chrome, Firefox, Edge, and Safari.

## 🚀 Quick Install (Pre-built Packages)

### Option 1: Download & Install
1. **Build packages**: Run `./build.sh` in the extension directory
2. **Install for your browser**:
   - **Chrome/Edge**: Load `build/bookminder-chrome-v1.0.zip` 
   - **Firefox**: Load `build/bookminder-firefox-v1.0.zip`
   - **Safari**: Load `build/bookminder-safari-v1.0.zip`

### Option 2: Development Install
Load the extension directory directly for development:

#### Chrome/Chromium/Edge
1. Open `chrome://extensions/` (or `edge://extensions/`)
2. Enable "Developer mode" 
3. Click "Load unpacked"
4. Select the `extension` folder

#### Firefox
1. Open `about:debugging`
2. Click "This Firefox" 
3. Click "Load Temporary Add-on"
4. Select `manifest_v2.json` in the `extension` folder

#### Safari
1. Install Xcode from the Mac App Store
2. Open Terminal and run:
   ```bash
   xcrun safari-web-extension-converter build/safari/
   ```
3. Follow the prompts to create an Xcode project
4. Open the generated Xcode project
5. Build and run the project (⌘+R)
6. Enable the extension in Safari Preferences → Extensions

## 🛠️ Building Packages

### Prerequisites
- `zip` command (install with `apt install zip` on Linux)
- **For Safari**: Xcode (macOS only) for Safari Web Extension Converter

### Build Commands
```bash
# Build all browser packages
./build.sh

# Or use NPM scripts
npm run build

# Clean build directory
npm run clean
```

### Generated Packages
- `bookminder-chrome-v1.0.zip` - Chrome/Edge (Manifest V3)
- `bookminder-firefox-v1.0.zip` - Firefox (Manifest V2)
- `bookminder-safari-v1.0.zip` - Safari (Manifest V3, requires Xcode conversion)

## ⚙️ Configuration

### First Setup
1. **Install extension** using instructions above
2. **Configure API URL**:
   - **Chrome/Edge**: Right-click extension icon → "Options"
   - **Firefox**: `about:addons` → BookMinder → Preferences
   - **Safari**: Safari menu → Preferences → Extensions → BookMinder → Preferences
3. **Enter API URL**: `http://localhost:9090` (or your server URL)
4. **Click "Save Settings"**

### API Server
Make sure the BookMinder API is running:
```bash
cd /path/to/linkminder
go run main.go
```

## 📖 Usage

1. **Navigate** to any webpage
2. **Click** the BookMinder extension icon
3. **Review** auto-filled title, URL, description
4. **Select action**:
   - **Read Later** - Save for later review
   - **Working** - Add to a project (specify topic)
   - **Share** - Mark for sharing (specify recipient)
5. **Click** "Save Bookmark" or "Save & Close Tab"

## ✨ Features

- **Cross-browser compatibility** (Chrome, Firefox, Edge, Safari)
- **Configurable API endpoint** via options page
- **Auto-extracts** page content, title, and meta description
- **Smart content detection** using semantic HTML selectors
- **Action-based organization** (read-later, working, share)
- **Topic/project management** with autocomplete
- **Tab management** (save & close option)
- **Error handling** with user feedback
- **Restricted page detection** (extension pages, about: URLs)

## 🔧 Development

### File Structure
```
extension/
├── manifest.json          # Chrome/Edge/Safari (V3)
├── manifest_v2.json       # Firefox (V2)
├── background.js          # Service worker/background script
├── content.js            # Page content extraction
├── popup.html/js         # Extension popup interface
├── options.html/js       # Settings page
├── build.sh             # Cross-browser build script
├── package.json         # NPM scripts
└── icons/               # Extension icons
```

### API Integration
The extension communicates with the BookMinder API:
- `POST /bookmark` - Save bookmark data
- `GET /topics` - Fetch available topics

### Browser Differences
- **Chrome/Edge/Safari**: Uses Manifest V3 with service workers
- **Firefox**: Uses Manifest V2 with background scripts
- **URL restrictions**: Handles browser-specific extension URLs

## 🐛 Troubleshooting

### Common Issues
1. **"API URL not configured"** - Open extension options and set API URL
2. **"Failed to save bookmark"** - Check API server is running
3. **Content not extracted** - Page may use dynamic content loading
4. **Extension not loading** - Verify manifest version matches browser

### Debug Mode
- **Chrome/Edge**: `chrome://extensions/` → Details → Inspect views
- **Firefox**: `about:debugging` → Inspect
- **Safari**: Develop menu → Web Extension Background Content (enable Develop menu first)
- Check browser console for error messages

### Safari-Specific Notes
- Requires macOS and Xcode for development
- Extension must be code-signed for distribution
- Use Safari Web Extension Converter for manifest conversion
- Enable Develop menu: Safari → Preferences → Advanced → Show Develop menu