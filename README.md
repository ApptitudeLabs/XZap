# Xclean

> Clean up your Xcode workspace, simulators, runtimes, and more — with style.

![xclean logo](./assets/logo.png)
---

## Features

- 🧹 **Clean** DerivedData, Archives, ModuleCache, SwiftPM Cache
- 📱 **List simulators** with storage usage
- 📈 **Highlight critical simulators** (above 3GB or your threshold)
- 🧹 **List and remove runtimes**
- 🭹 **Interactive cleanup** of large simulators
- 🎝 **Dry-run**, **Force Clean**, **Summary-only** modes
- 📋 **Save reports** to file with `--output`
- 💻 **Built for macOS** – supports Intel & Apple Silicon
- 🎨 **Beautiful CLI** with emojis, colors, and spinners

---

## Installation

```bash
git clone https://github.com/ApptitudeLabs/xclean.git
cd xclean
make build-mac
./bin/xclean_darwin_arm64 list sims
```

---

## CLI Usage Examples

```bash
# Clean Xcode caches
xclean clean                              # Clean DerivedData, Archives, ModuleCache, SwiftPM
xclean clean --dry-run                    # Preview what would be deleted

# List and manage simulators
xclean list sims                          # List simulators with storage usage
xclean list sims --threshold 2            # Only show simulators larger than 2GB
xclean list sims --summary-only           # Only print total space and counts
xclean list sims --output report.txt      # Save full report to file
xclean list sims --clean                  # Interactively delete large sims (>3GB)
xclean list sims --clean --dry-run        # Simulate what would be cleaned
xclean list sims --clean --force-clean    # Delete without confirmation
xclean cleansims                          # Delete all sims over 2GB (with confirmation)
xclean cleansims --dry-run                # Preview what would be deleted
xclean cleansims --force                  # Delete without confirmation

# Manage runtimes
xclean list runtimes                      # List installed Xcode runtimes
xclean remove runtime "iOS 17.5"          # Remove a specific runtime
xclean remove runtime "iOS 17.5" --force  # Force remove runtime without asking
```

---

## Screenshot

> Professional terminal output with emoji indicators, colors, and spinner animations.

<!-- ![screenshot](./assets/screenshot.png) -->

---

## Roadmap

- 📦 Homebrew tap install (`brew install xclean`)
- 🖥️ Export JSON or Markdown reports

---

## License

MIT © [Apptitude Labs](https://github.com/ApptitudeLabs)

