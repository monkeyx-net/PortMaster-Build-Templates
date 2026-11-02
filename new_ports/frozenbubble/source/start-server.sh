#!/bin/bash
# Frozen Bubble Server Startup Script

echo "========================================"
echo "  Frozen Bubble Network Server"
echo "========================================"
echo ""

# Navigate to server directory
cd "$(dirname "$0")/build/server"

if [ ! -f "./fb-server" ]; then
    echo "ERROR: Server binary not found!"
    echo "Please run 'cmake --build build' first"
    exit 1
fi

# Parse command line arguments
PORT=1511
DEBUG_MODE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -d|--debug)
            DEBUG_MODE="-d"
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -p, --port PORT    Set server port (default: 1511)"
            echo "  -d, --debug        Run in debug mode (logs to console)"
            echo "  -h, --help         Show this help message"
            echo ""
            echo "The server always enables both:"
            echo "  - TCP on PORT for LAN and internet (Net Game) connections"
            echo "  - UDP on PORT for LAN broadcast discovery"
            echo ""
            echo "Example:"
            echo "  $0 -p 1234 -d      # Run on port 1234 with debug output"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h for help"
            exit 1
            ;;
    esac
done

echo "Starting server on port $PORT..."
echo "  TCP: native clients (desktop, Android)"
echo "  WebSocket: browser clients (WASM / itch.io) — same port, auto-detected"
echo "  UDP: LAN broadcast discovery"
echo "Press Ctrl+C to stop the server"
echo ""
echo "Server logs:"
echo "----------------------------------------"

# -q: don't register with frozen-bubble.org master server (offline since ~2012)
# -l: enable UDP broadcast so clients can auto-discover on LAN
# -z: no language requirement
./fb-server -p $PORT -q -l -z $DEBUG_MODE
