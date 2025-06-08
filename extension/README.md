# BookMinder Browser Extension

A cross-platform browser extension that saves bookmarks with page content to the BookMinder API.

## Installation

### Chrome/Chromium/Edge
1. Open browser and go to `chrome://extensions/` (or `edge://extensions/`)
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select the `extension` folder

### Firefox
1. Open Firefox and go to `about:debugging`
2. Click "This Firefox"
3. Click "Load Temporary Add-on"
4. Select any file in the `extension` folder

## Usage

1. Make sure the BookMinder API is running on `http://localhost:9090`
2. Click the extension icon in your browser toolbar
3. Review/edit the auto-filled URL, title, and description
4. Click "Save Bookmark"

## Features

- Auto-extracts page URL, title, and meta description
- Editable fields before saving
- Success/error feedback
- Works with Chrome, Firefox, Safari, and Edge

## Note

You'll need to create icon files (icon16.png, icon48.png, icon128.png) or update the manifest.json to remove icon references.