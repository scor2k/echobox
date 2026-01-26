#!/usr/bin/env python3
"""
Simple web server for SRE interview
BUG: This file has intentional bugs for candidates to find and fix
"""

import http.server
import socketserver
import json

# BUG 1: Wrong port (should match config.json)
PORT = 8080

# BUG 2: Missing config file loading
# Should load from config.json

class MyHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/plain')
        self.end_headers()
        # BUG 3: Typo in message
        self.wfile.write(b'Hello, SRE Candiate!')  # Missing 'd' in Candidate

def run_server():
    with socketserver.TCPServer(("", PORT), MyHandler) as httpd:
        print(f"Server running on port {PORT}")
        httpd.serve_forever()

if __name__ == "__main__":
    run_server()
