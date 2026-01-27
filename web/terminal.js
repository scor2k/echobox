// Terminal and WebSocket management
(function() {
    'use strict';

    // State
    let ws = null;
    let term = null;
    let fitAddon = null;
    let connected = false;
    let sessionEnded = false;

    // DOM elements
    const terminalEl = document.getElementById('terminal');
    const statusDot = document.getElementById('status-dot');
    const statusText = document.getElementById('status-text');
    const finishBtn = document.getElementById('finish-btn');
    const modal = document.getElementById('confirm-modal');
    const modalCancel = document.getElementById('modal-cancel');
    const modalConfirm = document.getElementById('modal-confirm');

    // Initialize terminal
    function initTerminal() {
        term = new Terminal({
            cursorBlink: true,
            fontSize: 14,
            fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
            theme: {
                background: '#1e1e1e',
                foreground: '#d4d4d4',
                cursor: '#d4d4d4',
                black: '#000000',
                red: '#cd3131',
                green: '#0dbc79',
                yellow: '#e5e510',
                blue: '#2472c8',
                magenta: '#bc3fbc',
                cyan: '#11a8cd',
                white: '#e5e5e5',
                brightBlack: '#666666',
                brightRed: '#f14c4c',
                brightGreen: '#23d18b',
                brightYellow: '#f5f543',
                brightBlue: '#3b8eea',
                brightMagenta: '#d670d6',
                brightCyan: '#29b8db',
                brightWhite: '#e5e5e5'
            },
            allowProposedApi: true
        });

        // Fit addon for responsive terminal
        fitAddon = new FitAddon.FitAddon();
        term.loadAddon(fitAddon);

        term.open(terminalEl);
        fitAddon.fit();

        // Handle resize
        window.addEventListener('resize', () => {
            fitAddon.fit();
            sendResize();
        });

        // Prevent default paste behavior - multiple layers
        term.attachCustomKeyEventHandler((event) => {
            // Block Ctrl+V and Cmd+V
            if ((event.ctrlKey || event.metaKey) && event.key === 'v') {
                event.preventDefault();
                event.stopPropagation();
                showNotification('Paste is disabled for this assessment');
                logAntiCheatEvent('paste_attempt', { source: 'keyboard_shortcut' });
                return false;
            }
            return true;
        });

        // Prevent right-click paste
        terminalEl.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            e.stopPropagation();
            showNotification('Context menu is disabled');
            return false;
        }, true); // Use capture phase

        // Prevent paste events - multiple listeners for coverage
        const blockPaste = (e) => {
            e.preventDefault();
            e.stopPropagation();
            e.stopImmediatePropagation();
            showNotification('Paste is disabled for this assessment');
            logAntiCheatEvent('paste_attempt', { source: 'paste_event' });
            return false;
        };

        // Add to terminal element
        terminalEl.addEventListener('paste', blockPaste, true); // Capture phase
        terminalEl.addEventListener('paste', blockPaste, false); // Bubble phase

        // Add to document level (catches all paste attempts)
        document.addEventListener('paste', blockPaste, true);

        // Override clipboard API
        if (navigator.clipboard) {
            const originalReadText = navigator.clipboard.readText;
            navigator.clipboard.readText = function() {
                showNotification('Clipboard access is disabled');
                logAntiCheatEvent('paste_attempt', { source: 'clipboard_api' });
                return Promise.reject(new Error('Clipboard access disabled'));
            };
        }

        // Block drag and drop (another paste method)
        const blockDragDrop = (e) => {
            e.preventDefault();
            e.stopPropagation();
            showNotification('Drag and drop is disabled');
            logAntiCheatEvent('paste_attempt', { source: 'drag_drop' });
            return false;
        };

        terminalEl.addEventListener('drop', blockDragDrop, true);
        terminalEl.addEventListener('dragover', blockDragDrop, true);
        document.addEventListener('drop', blockDragDrop, true);

        // Detect rapid input (potential paste)
        let inputBuffer = [];
        let lastInputTime = Date.now();

        term.onData((data) => {
            const now = Date.now();
            const timeDiff = now - lastInputTime;

            // Reset buffer if more than 200ms since last input
            if (timeDiff > 200) {
                inputBuffer = [];
            }

            inputBuffer.push(data);
            lastInputTime = now;

            // BLOCK rapid input that suggests paste (>20 chars in single event)
            if (data.length > 20) {
                showNotification('Paste blocked - large input detected');
                logAntiCheatEvent('paste_blocked', {
                    chars: data.length,
                    source: 'large_input_block'
                });
                // Don't send to server - BLOCKED!
                return;
            }

            // Check for rapid input (>30 chars in <100ms suggests paste)
            if (inputBuffer.length > 30 && timeDiff < 100) {
                showNotification('Rapid input detected - please type manually');
                logAntiCheatEvent('rapid_input', {
                    chars: inputBuffer.length,
                    time_ms: timeDiff
                });
                // Don't send to server - BLOCKED!
                return;
            }

            // Send to server (only if passed all checks)
            if (connected && ws && ws.readyState === WebSocket.OPEN) {
                ws.send(data);
            }
        });

        // Window focus tracking
        let windowFocused = true;
        window.addEventListener('focus', () => {
            if (!windowFocused) {
                windowFocused = true;
                logAntiCheatEvent('window_focus', { gained: true });
            }
        });

        window.addEventListener('blur', () => {
            if (windowFocused) {
                windowFocused = false;
                logAntiCheatEvent('window_focus', { gained: false });
            }
        });

        // Visibility API for tab switching
        document.addEventListener('visibilitychange', () => {
            logAntiCheatEvent('tab_visibility', {
                hidden: document.hidden
            });
        });
    }

    // Initialize WebSocket
    function initWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = protocol + '//' + window.location.host + '/ws';

        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            connected = true;
            updateStatus('connected', 'Connected');
            finishBtn.disabled = false;

            // Send initial resize
            sendResize();
        };

        ws.onclose = (event) => {
            connected = false;
            finishBtn.disabled = true;

            console.log('WebSocket closed:', event.code, event.reason);

            // Don't reconnect if session ended
            if (sessionEnded) {
                console.log('Session ended, not reconnecting');
                updateStatus('disconnected', 'Session Ended');
                return;
            }

            updateStatus('disconnected', 'Reconnecting...');

            // Check if we have a reconnect token
            if (typeof window.attemptReconnect === 'function') {
                window.attemptReconnect(
                    // On success
                    (data) => {
                        console.log('Reconnection successful, restoring terminal...');
                        // Restore terminal buffer if provided
                        if (data.terminal && data.terminal.buffer) {
                            term.write(data.terminal.buffer);
                        }
                        // Reconnect WebSocket
                        initWebSocket();
                    },
                    // On failure
                    (reason) => {
                        console.log('Reconnection failed:', reason);
                        // Try fresh connection after delay
                        setTimeout(() => {
                            if (!connected) {
                                console.log('Attempting fresh connection...');
                                initWebSocket();
                            }
                        }, 3000);
                    }
                );
            } else {
                // Fallback: simple reconnect
                setTimeout(() => {
                    if (!connected) {
                        console.log('Attempting to reconnect...');
                        initWebSocket();
                    }
                }, 2000);
            }
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            updateStatus('disconnected', 'Connection Error');
        };

        ws.onmessage = (event) => {
            if (event.data instanceof Blob) {
                // Handle binary data
                const reader = new FileReader();
                reader.onload = () => {
                    term.write(new Uint8Array(reader.result));
                };
                reader.readAsArrayBuffer(event.data);
            } else {
                // Check if it's a JSON message
                try {
                    const msg = JSON.parse(event.data);
                    if (msg.type === 'session_ended') {
                        console.log('Session ended:', msg.data);
                        sessionEnded = true;

                        // Show session ended message
                        term.write('\r\n\r\n\x1b[1;33m=== Session Ended ===\x1b[0m\r\n');
                        term.write('The shell has exited.\r\n');
                        term.write('Your session has been saved.\r\n');
                        term.write('You may close this window.\r\n');

                        updateStatus('disconnected', 'Session Ended');
                        finishBtn.disabled = true;

                        // Clear reconnect token
                        if (typeof window.clearReconnectToken === 'function') {
                            window.clearReconnectToken();
                        }

                        return;
                    }
                } catch (e) {
                    // Not JSON, treat as terminal data
                }

                // Handle text data
                term.write(event.data);
            }
        };
    }

    // Send resize message to server
    function sendResize() {
        if (!connected || !ws || ws.readyState !== WebSocket.OPEN) {
            return;
        }

        const msg = {
            type: 'resize',
            data: {
                cols: term.cols,
                rows: term.rows
            }
        };

        ws.send(JSON.stringify(msg));
    }

    // Update connection status
    function updateStatus(state, text) {
        statusDot.className = 'status-dot ' + state;
        statusText.textContent = text;
    }

    // Show notification
    function showNotification(message, duration = 3000) {
        const notification = document.createElement('div');
        notification.className = 'notification';
        notification.textContent = message;
        document.body.appendChild(notification);

        setTimeout(() => {
            notification.remove();
        }, duration);
    }

    // Log anti-cheat event
    function logAntiCheatEvent(eventType, data) {
        if (!connected || !ws || ws.readyState !== WebSocket.OPEN) {
            return;
        }

        const msg = {
            type: 'anticheat',
            data: {
                event: eventType,
                timestamp: Date.now(),
                ...data
            }
        };

        ws.send(JSON.stringify(msg));
    }

    // Finish session
    function finishSession() {
        if (ws && ws.readyState === WebSocket.OPEN) {
            const msg = {
                type: 'finish',
                data: { timestamp: Date.now() }
            };
            ws.send(JSON.stringify(msg));
        }

        // Close connection
        if (ws) {
            ws.close();
        }

        // Show completion message
        term.write('\r\n\r\n\x1b[1;32m=== Session Ended ===\x1b[0m\r\n');
        term.write('Your interview session has been saved.\r\n');
        term.write('You may close this window.\r\n');

        finishBtn.disabled = true;
    }

    // Modal handlers
    finishBtn.addEventListener('click', () => {
        modal.classList.add('show');
    });

    modalCancel.addEventListener('click', () => {
        modal.classList.remove('show');
    });

    modalConfirm.addEventListener('click', () => {
        modal.classList.remove('show');
        finishSession();
    });

    // Prevent closing modal by clicking outside
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            // Optionally allow closing by clicking backdrop
            // modal.classList.remove('show');
        }
    });

    // Initialize everything when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            initTerminal();
            initWebSocket();
        });
    } else {
        initTerminal();
        initWebSocket();
    }

})();
