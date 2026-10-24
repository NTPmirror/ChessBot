# ChessBot

[English](README.md) | [Русский](README_RU.md)

A high-performance chess automation engine written in Go. It hooks into an active Android screen (via direct root access on Termux or through ADB on a desktop host), reads the board layout using template matching, evaluates optimal moves in real time, and executes input taps automatically.

Created by **ytika**.

---

## Key Features

- **Multi-Platform Support**: Runs on Windows, Linux, and directly on Android devices via Termux (requires root with KernelSU or Magisk).
- **Fast Multithreaded Search**:
  - Alpha-Beta search with parallel root search distributed across CPU cores (`GOMAXPROCS`).
  - Ray-based attack detection (`IsSquareAttacked`) bypassing heavy legal-move generation in check validation.
  - Quiescence search with MVV-LVA move ordering to prevent horizon-effect blunders.
  - Killer move heuristics for accelerated beta-cutoffs.
  - 16-shard concurrent transposition table protected by read-write mutexes.
- **Live Web Dashboard**:
  - Optional WebSocket-driven local browser interface displaying board state, highlight paths, and move logs.

---

## Project Structure

```
.
├── Code/
│   ├── assets/       # Embedded platform binaries and assets
│   ├── board/        # Board stabilization, turn inference, match state
│   ├── config/       # JSONC config loader with comment stripping, versioning
│   ├── engine/       # Move generation, rules, evaluation, parallel minimax
│   ├── recognizer/   # Frame capture (screencap), ADB/root tap, template vision
│   ├── terminal/     # Clean ANSI console renderer
│   ├── web/          # Embedded WebSocket server & live interactive chessboard
│   └── main.go       # Core orchestration loop
├── Figures/          # Reference piece templates and board samples
├── act.go            # Cross-compilation and build management utility
├── config.jsonc      # Runtime configuration file
└── main.go           # Application entrypoint
```

---

## Getting Started

### Prerequisites

- **Go 1.22+** installed on your build machine.
- An Android device running Duolingo Chess:
  - **Option A (Standalone on phone)**: Termux with root privileges (KernelSU / Magisk).
  - **Option B (Desktop via ADB)**: USB Debugging enabled on your device connected to PC.

---

### Building

Use `act.go` to build standalone releases for your target platform:

```bash
# Build for all platforms (Windows, Linux, Android)
go run act.go build all

# Build specifically for Android (ARM64 ELF)
go run act.go build android

# Build for Windows (AMD64 EXE with bundled ADB)
go run act.go build windows

# Build for Linux (AMD64 ELF with bundled Linux ADB)
go run act.go build linux
```

Binaries are emitted into the `Builds/<version>/<Platform>/` directory.

---

### Running on Mobile (Android / Termux)

> # ⚠️ WARNING: ROOT ACCESS REQUIRED (KernelSU / Magisk / APatch)!
> **The bot cannot operate standalone on Android without Root access, as it relies on direct privileged commands for screen capture (`screencap`) and touch synthesis (`input tap`).**

1. Download the pre-built **Android** archive from the release directory (`Builds/<version>/Android/`).
2. Extract the archive into any folder in **Termux** (e.g. `~/Duolingo`).
3. Execute the `start.sh` script:
   ```bash
   ./start.sh
   ```
   The script automatically sets executable permissions and invokes the binary via root (`su`).

---

### Running from Desktop (Windows / Linux)

Running from a PC does **not** require root access on the phone, but USB Debugging must be enabled.

#### 1. Enabling USB Debugging on your phone:
1. Open **Settings** → **About phone** (or *Software information*).
2. Tap **"Build number"** rapidly 7 times until you see the message *"You are now a developer!"*.
3. Go back to main settings → **System** (or *Additional settings*) → **Developer options**.
4. Enable **"USB debugging"**.
5. *(For Xiaomi / HyperOS / MIUI devices)*: also enable **"USB debugging (Security settings)"** to allow touch input emulation.
6. Connect your phone to your PC via a USB cable. When prompted on the phone display, check **"Always allow from this computer"** and tap **Allow**.

#### 2. Launching:
1. Download and extract the pre-built release for your OS (**Windows** or **Linux**). Platform-appropriate ADB binaries are already bundled.
2. Launch the bot:
   ```bash
   # Windows
   .\ChessBot.exe

   # Linux
   ./chessbot
   ```
3. To open the real-time live web chessboard, append the `-WEB` flag:
   ```bash
   .\ChessBot.exe -WEB
   ```
   Navigate to `http://localhost:5873` in your browser.

---

## Configuration (`config.jsonc`)

```jsonc
{
  // Polling delay between screen captures in milliseconds
  "interval_ms": 10,

  // Local web dashboard port
  "web_port": 5873,

  // Start web server automatically
  "enable_web": false,

  // Vertical board offset ratio for Oscar bot mode
  "y_ratio_oscar": 0.355,

  // Vertical board offset ratio for online PvP mode
  "y_ratio_people": 0.336,

  // Directory containing reference piece silhouettes
  "figures_dir": "./Figures",

  // Matching tolerance for noisy or highlighted cells
  "max_error_threshold": 500.0,

  // Consecutive identical detections required to validate a state change
  "stability_threshold": 2,

  // Automatically execute taps on the screen
  "auto_play": true,

  // Automatically tap XP claim buttons on match completion
  "auto_claim_xp": true,

  // Automatically start new matches from the lobby
  "auto_start_game": true,

  // Opponent mode: "People" (online) or "Oscar" (built-in bot)
  "game_type": "People"
}
```
