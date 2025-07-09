# WBFY - Web Browser For Your CLI

A web-based terminal emulator that allows you to run command-line programs in your browser. Perfect for educational environments to teach Linux commands and test DSA code interactively.

## Features

- Run any command-line program in a browser terminal
- Interactive input/output with proper terminal emulation
- Automatic window resizing
- Cross-platform support
- Robust WebSocket communication
- Auto-reconnection on connection loss

## Usage

```bash
./wbfy [command] [args...]
```

For example:
```bash
# Run a Python interpreter
./wbfy python3

# Run a specific script
./wbfy python3 path/to/script.py

# Run a shell
./wbfy bash
```

## Integration with Education Platforms

WBFY is designed to be easily integrated with educational platforms:

1. **Container Integration**: Each student can get their own isolated environment
2. **DSA Practice**: Students can write and test algorithms in real-time
3. **Linux Commands Learning**: Practice Linux commands in a safe, web-based environment

## How It Works

1. A PTY (pseudo-terminal) is created on the server side
2. The command is executed in this PTY
3. Input/output is streamed via WebSockets between the browser and server
4. xterm.js provides the terminal UI in the browser

## Development

### Building from Source

```bash
# Get dependencies
go get -u ./...

# Build the binary
go build -o wbfy

# Run
./wbfy [command] [args...]
```

### Customizing the Frontend

The web interface is located in the `web/` directory:
- `index.html` - Main HTML file
- `css/` - Stylesheets (if you want to add custom styles)
- `js/` - JavaScript files (if you want to extend functionality)

## Troubleshooting

### Input Not Working

If terminal input isn't working:
1. Ensure the terminal has focus (click on it)
2. Check browser console for WebSocket errors
3. Verify you're running a command that accepts input

### Connection Issues

If you see connection errors:
1. Make sure the server is running
2. Check if port 8080 is already in use
3. Try a different browser

## License

MIT
