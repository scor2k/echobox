# Available Tools in Interview Environment

This document lists all tools pre-installed in the Echobox interview container for troubleshooting and completing tasks.

## Network Tools

### HTTP/Web
- **curl** - Transfer data from/to URLs
  ```bash
  curl -I https://example.com
  curl -X POST -d "data" https://api.example.com
  ```

- **wget** - Download files
  ```bash
  wget https://example.com/file.txt
  wget -O output.html https://example.com
  ```

### Network Testing
- **nc (netcat)** - Network swiss army knife
  ```bash
  nc -zv localhost 8080        # Port scan
  nc -l 9000                   # Listen on port
  echo "test" | nc host 9000   # Send data
  ```

- **nmap** - Network mapper and port scanner
  ```bash
  nmap localhost               # Scan local ports
  nmap -p 1-1000 localhost     # Scan port range
  nmap -sV localhost           # Service version detection
  ```

- **tcpdump** - Packet capture and analysis
  ```bash
  tcpdump -i any port 8080     # Capture traffic on port
  tcpdump -n -c 100            # Capture 100 packets
  ```

### DNS Tools
- **dig** - DNS lookup utility
  ```bash
  dig example.com
  dig @8.8.8.8 example.com
  dig example.com MX           # Mail records
  ```

- **nslookup** - Query DNS servers
  ```bash
  nslookup example.com
  nslookup example.com 8.8.8.8
  ```

- **host** - DNS lookup
  ```bash
  host example.com
  host -t MX example.com
  ```

### Network Info
- **ip** - Show/manipulate routing, devices
  ```bash
  ip addr                      # Show IP addresses
  ip route                     # Show routing table
  ip link                      # Show network interfaces
  ```

- **ss** - Socket statistics
  ```bash
  ss -tuln                     # Show listening TCP/UDP
  ss -s                        # Summary statistics
  ss -p                        # Show processes
  ```

- **ifconfig** - Configure network interfaces
  ```bash
  ifconfig                     # Show all interfaces
  ifconfig eth0                # Show specific interface
  ```

- **netstat** - Network statistics
  ```bash
  netstat -tuln                # Listening ports
  netstat -an                  # All connections
  ```

- **iftop** - Network bandwidth monitoring
  ```bash
  iftop -i eth0                # Monitor interface
  ```

## System Debugging

### Process Tracing
- **strace** - Trace system calls
  ```bash
  strace ls                    # Trace ls command
  strace -p 1234               # Attach to process
  strace -c ls                 # Count syscalls
  strace -e open,read ls       # Trace specific calls
  ```

- **lsof** - List open files
  ```bash
  lsof -i :8080                # Process using port 8080
  lsof -p 1234                 # Files opened by PID
  lsof /var/log/app.log        # Processes using file
  ```

### Process Management
- **ps** - Process status
  ```bash
  ps aux                       # All processes
  ps -ef                       # Full format
  ps -p 1234                   # Specific PID
  ```

- **top** - Interactive process viewer
  ```bash
  top                          # Real-time view
  top -p 1234                  # Monitor specific PID
  ```

- **htop** - Interactive process viewer (enhanced)
  ```bash
  htop                         # Better than top
  htop -p 1234                 # Monitor PID
  ```

- **iotop** - I/O usage by process
  ```bash
  iotop                        # Disk I/O monitor
  ```

### System Monitoring
- **vmstat** - Virtual memory statistics
  ```bash
  vmstat 1                     # Update every second
  vmstat -s                    # Summary
  ```

- **iostat** - CPU and I/O statistics
  ```bash
  iostat                       # I/O stats
  iostat -x 1                  # Extended, update 1s
  ```

- **free** - Memory usage
  ```bash
  free -h                      # Human readable
  free -m                      # In megabytes
  ```

## Editors

- **vim** - Vi Improved text editor
  ```bash
  vim file.txt
  vim -R file.txt              # Read-only mode
  ```

- **nano** - Easy-to-use text editor
  ```bash
  nano file.txt
  nano -l file.txt             # Show line numbers
  ```

## Text Processing

- **jq** - JSON processor
  ```bash
  echo '{"key":"value"}' | jq '.key'
  cat file.json | jq '.items[]'
  ```

- **grep** - Search text patterns
  ```bash
  grep "error" logfile.txt
  grep -r "pattern" /var/log/
  grep -i "case insensitive" file.txt
  ```

- **sed** - Stream editor
  ```bash
  sed 's/old/new/g' file.txt
  sed -n '10,20p' file.txt     # Print lines 10-20
  ```

- **awk** - Text processing
  ```bash
  awk '{print $1}' file.txt
  awk -F: '{print $1}' /etc/passwd
  ```

## Development Tools

### Python
- **python3** - Python interpreter
  ```bash
  python3 script.py
  python3 -m http.server 8000
  python3 -c "print('hello')"
  ```

- **pip3** - Python package manager
  ```bash
  pip3 install package
  pip3 list
  ```

**Pre-installed Python packages:**
- requests - HTTP library
- pyyaml - YAML parser

### Version Control
- **git** - Version control
  ```bash
  git clone https://...
  git log
  git diff
  ```

### Build Tools
- **make** - Build automation
  ```bash
  make
  make install
  make clean
  ```

## File Management

- **find** - Search for files
  ```bash
  find /var/log -name "*.log"
  find . -type f -mtime -1
  find . -size +10M
  ```

- **tree** - Display directory tree
  ```bash
  tree /var/log
  tree -L 2 /etc
  ```

- **tar** - Archive utility
  ```bash
  tar -czf archive.tar.gz dir/
  tar -xzf archive.tar.gz
  tar -tzf archive.tar.gz      # List contents
  ```

- **zip/unzip** - Compression
  ```bash
  zip archive.zip file1 file2
  unzip archive.zip
  unzip -l archive.zip         # List contents
  ```

- **rsync** - File synchronization
  ```bash
  rsync -av source/ dest/
  ```

## System Information

- **uname** - System information
  ```bash
  uname -a                     # All info
  uname -r                     # Kernel version
  ```

- **df** - Disk space
  ```bash
  df -h                        # Human readable
  df -i                        # Inode usage
  ```

- **du** - Disk usage
  ```bash
  du -sh /var/log              # Directory size
  du -h --max-depth=1          # Subdirectory sizes
  ```

- **uptime** - System uptime
  ```bash
  uptime
  ```

- **who** / **w** - Logged in users
  ```bash
  who
  w
  ```

## Utilities

### Comparison
- **diff** - Compare files
  ```bash
  diff file1.txt file2.txt
  diff -u file1 file2          # Unified format
  ```

### Sorting/Filtering
- **sort** - Sort lines
  ```bash
  sort file.txt
  sort -n -k2 data.txt         # Numeric sort column 2
  ```

- **uniq** - Remove duplicates
  ```bash
  sort file.txt | uniq
  uniq -c file.txt             # Count occurrences
  ```

- **wc** - Word/line count
  ```bash
  wc -l file.txt               # Line count
  wc -w file.txt               # Word count
  ```

### Text Viewing
- **less** - Pager
  ```bash
  less /var/log/messages
  ```

- **head** - View file start
  ```bash
  head -n 20 file.txt
  ```

- **tail** - View file end
  ```bash
  tail -f /var/log/app.log     # Follow updates
  tail -n 100 file.txt
  ```

- **cat** - Concatenate files
  ```bash
  cat file.txt
  cat file1 file2 > combined
  ```

## Complete Tool List

### Networking
✅ curl, wget, nc, nmap, tcpdump
✅ dig, nslookup, host
✅ ip, ss, ifconfig, netstat
✅ iftop, ping, traceroute

### System Debugging
✅ strace, lsof
✅ ps, top, htop, iotop
✅ vmstat, iostat, free

### Development
✅ python3, pip3
✅ git, make
✅ gcc, g++ (if added)

### Editors
✅ vim, nano

### Text Processing
✅ jq, grep, sed, awk
✅ sort, uniq, wc
✅ head, tail, less, cat

### File Management
✅ find, tree
✅ tar, zip, unzip
✅ rsync

### Utilities
✅ bash, bash-completion
✅ coreutils (ls, cp, mv, etc.)
✅ util-linux (many utilities)

## Not Included (Add if Needed)

The following are NOT installed but can be added:
- **Docker** - Not installed in container (don't need inception)
- **kubectl** - Not needed for basic SRE tasks
- **ansible** - Too large, use if needed
- **terraform** - Too large, use if needed
- **gcc/build-essential** - Adds 100MB+, uncomment if needed

## Adding More Tools

If you need additional tools, edit `Dockerfile`:

```dockerfile
RUN apk add --no-cache \
    # ... existing tools ...
    your-package-name \
    && rm -rf /var/cache/apk/*
```

## Size Considerations

Current image size: ~150-200MB (with all tools)

To reduce:
- Remove nmap, tcpdump (large packages)
- Remove python3, pip3 (if not needed for tasks)
- Use minimal toolset

To optimize for specific task types:
- **Debugging tasks**: Keep strace, lsof, netstat
- **Network tasks**: Keep curl, nc, nmap, tcpdump
- **Scripting tasks**: Keep python3, bash-completion
- **Performance tasks**: Keep htop, iotop, vmstat
