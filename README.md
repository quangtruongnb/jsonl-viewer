# JSONL Viewer

A desktop application for viewing and analyzing JSONL (JSON Lines) files with advanced search, filtering, and export capabilities.

## Features

- **File Loading**: Load and parse JSONL files with validation
- **Search & Filter**: Advanced search with Lucene syntax support
- **Field Visibility**: Show/hide specific JSON fields
- **Export**: Export filtered results to JSONL files
- **Cross-platform**: Windows, macOS, and Linux support

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.

## GitHub Actions

This project includes automated build workflows:

### Test Build (`test.yml`)
- Runs on every push and pull request
- Tests builds on Windows, macOS, and Linux
- Ensures the application builds successfully on all platforms

### Release Build (`build.yml`)
- Runs on releases and main branch pushes
- Creates cross-platform executables:
  - **Windows**: `jsonl-viewer.exe` (with embedded WebView2)
  - **macOS**: `jsonl-viewer.app` (universal binary)
  - **Linux**: `jsonl-viewer` (ELF executable)
- Automatically creates release assets when a GitHub release is published

### Usage

1. **For Testing**: Push to any branch to trigger test builds
2. **For Releases**: 
   - Create a GitHub release
   - Assets will be automatically uploaded
   - Download the appropriate file for your platform

### Build Artifacts

- **Windows**: `jsonl-viewer-windows.zip` (contains `jsonl-viewer.exe`)
- **macOS**: `jsonl-viewer-macos.zip` (contains `jsonl-viewer.app`)
- **Linux**: `jsonl-viewer-linux.tar.gz` (contains `jsonl-viewer` executable)
