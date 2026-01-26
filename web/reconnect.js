// Reconnection handler
(function() {
    'use strict';

    // Reconnection state
    let reconnectToken = null;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 3;
    const reconnectDelays = [2000, 5000, 10000]; // Exponential backoff

    // Store token on first successful connection
    window.storeReconnectToken = function(token) {
        reconnectToken = token;
        localStorage.setItem('echobox_reconnect_token', token);
        console.log('Reconnect token stored:', token);
    };

    // Get stored token
    window.getReconnectToken = function() {
        if (reconnectToken) {
            return reconnectToken;
        }
        reconnectToken = localStorage.getItem('echobox_reconnect_token');
        return reconnectToken;
    };

    // Clear reconnect token (on session end)
    window.clearReconnectToken = function() {
        reconnectToken = null;
        localStorage.removeItem('echobox_reconnect_token');
        console.log('Reconnect token cleared');
    };

    // Attempt reconnection
    window.attemptReconnect = function(onSuccess, onFailure) {
        const token = window.getReconnectToken();
        if (!token) {
            console.log('No reconnect token available');
            if (onFailure) onFailure('no_token');
            return;
        }

        if (reconnectAttempts >= maxReconnectAttempts) {
            console.log('Max reconnect attempts reached');
            if (onFailure) onFailure('max_attempts');
            return;
        }

        const attemptNum = reconnectAttempts;
        reconnectAttempts++;

        console.log(`Reconnect attempt ${attemptNum + 1}/${maxReconnectAttempts}...`);

        fetch(`/reconnect?token=${encodeURIComponent(token)}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                return response.json();
            })
            .then(data => {
                console.log('Reconnection approved:', data);
                reconnectAttempts = 0; // Reset on success

                if (onSuccess) {
                    onSuccess(data);
                }
            })
            .catch(error => {
                console.error(`Reconnect attempt ${attemptNum + 1} failed:`, error);

                // Check if we should retry
                if (reconnectAttempts < maxReconnectAttempts) {
                    const delay = reconnectDelays[attemptNum] || 10000;
                    console.log(`Will retry in ${delay/1000}s...`);

                    setTimeout(() => {
                        window.attemptReconnect(onSuccess, onFailure);
                    }, delay);
                } else {
                    console.log('All reconnect attempts failed');
                    if (onFailure) onFailure('all_attempts_failed');
                }
            });
    };

    // Reset reconnection attempts (call on successful connection)
    window.resetReconnectAttempts = function() {
        reconnectAttempts = 0;
    };

})();
