# Security Documentation

## Overview

Echobox is designed to provide a secure, isolated environment for conducting technical interviews while maintaining assessment integrity through anti-cheat measures.

## Threat Model

### Assets to Protect
1. **Interview Integrity**: Prevent cheating (paste, external help)
2. **System Resources**: Prevent resource exhaustion (DoS)
3. **Host System**: Prevent container escape or host compromise
4. **Recorded Data**: Ensure tamper-proof session recordings

### Threat Actors
1. **Malicious Candidate**: Attempts to cheat or game the system
2. **Curious Candidate**: Explores beyond permitted scope
3. **External Attacker**: Attempts to compromise host (unlikely, but considered)

### Attack Vectors
1. Copy-paste from external sources
2. Resource exhaustion (CPU, memory, disk, processes)
3. Network abuse (if enabled)
4. Container escape attempts
5. Recording tampering

## Security Measures Implemented

### 1. Anti-Cheat System

**Client-Side Prevention:**
- ✅ Paste events blocked (Ctrl+V, right-click, drag-drop)
- ✅ Context menu disabled
- ✅ Clipboard API blocked
- ✅ Focus tracking (window/tab switching detected)

**Server-Side Detection:**
- ✅ Input rate limiting (30 chars/sec threshold)
- ✅ Burst detection (30 chars in 100ms)
- ✅ Typing pattern analysis (WPM, anomalies)
- ✅ Multi-level event logging (info/warning/critical)

**Post-Session Analysis:**
- ✅ Automated verdict generation (CLEAN, SUSPICIOUS, etc.)
- ✅ Pattern anomaly detection
- ✅ Recommendation system
- ✅ Typing statistics (WPM, std dev)

**Limitations:**
- ⚠️ Client-side checks can be bypassed (browser dev tools)
- ⚠️ Sophisticated automation might evade detection
- ⚠️ No video proof (only keystroke logs)
- ✅ Multiple layers make cheating difficult and detectable

### 2. Container Isolation

**Implemented:**
- ✅ Non-root user execution (UID 1000: candidate)
- ✅ Capability dropping (CAP_DROP=ALL + selective add)
- ✅ Security option: no-new-privileges
- ✅ Resource limits (CPU, memory, pids)
- ✅ Single-use containers (no state persistence)
- ✅ Minimal attack surface (Alpine Linux)

**Not Implemented (By Design):**
- ❌ Read-only filesystem - Candidates need write access for tasks
- ❌ Mandatory network isolation - Some tasks may require internet

**Mitigation:**
- Writable filesystem is necessary for interview tasks
- Resource limits prevent abuse
- Non-root prevents privilege escalation
- Network isolation is optional (enable if tasks allow)

### 3. Resource Protection

**Memory Limits:**
```bash
--memory="512m"              # Hard limit
--memory-reservation="256m"   # Soft limit
```
Prevents memory exhaustion attacks.

**CPU Limits:**
```bash
--cpus="0.5"
```
Limits CPU-intensive processes to 50% of one core.

**Process Limits:**
```bash
--pids-limit=100
--ulimit nproc=512:1024
```
Prevents fork bombs and process exhaustion.

**Disk I/O:**
- Containers are single-use (limited lifetime)
- Session recordings have bounded size
- No long-term storage in containers

### 4. Recording Integrity

**Tamper Protection:**
- ✅ SHA-256 hashing of all recorded files
- ✅ Hashes stored in metadata.json
- ✅ Verification available via analyze.sh
- ✅ Progressive saves (10s flush interval)

**Data Integrity:**
- ✅ Buffered I/O prevents data loss
- ✅ Graceful shutdown ensures all data saved
- ✅ Multiple recording layers (keystrokes, terminal, websocket, events)

**Limitations:**
- ⚠️ Recordings stored on host filesystem (ensure proper permissions)
- ⚠️ No encryption at rest (add if storing sensitive data)

### 5. Network Security

**Options:**

**Full Isolation (Maximum Security):**
```bash
docker run --network none echobox:latest
```
- No external network access
- Localhost only
- Best for self-contained tasks

**Bridge Network (Default):**
```bash
docker run echobox:latest
```
- Allows external access
- Needed for tasks requiring package installation
- Required for tasks involving API calls

**Recommendation:**
- Use `--network none` when possible
- Enable network only if tasks explicitly require it
- Monitor egress traffic if network is enabled

### 6. Access Control

**Container Access:**
- ✅ Non-root user (UID 1000)
- ✅ No sudo/privilege escalation
- ✅ Limited capabilities

**Web Interface:**
- ⚠️ No authentication (add basic auth if needed)
- ✅ Security headers (CSP, X-Frame-Options, etc.)
- ✅ Single-use URLs (one candidate per container)

**Recommendation:**
- Deploy behind reverse proxy with auth
- Use firewall rules to restrict access
- Consider VPN for sensitive interviews

## Known Limitations

### 1. Client-Side Bypass
**Risk**: Determined attacker can modify JavaScript
**Mitigation**: Server-side detection catches anomalies
**Impact**: Medium (detectable in post-analysis)

### 2. Writable Filesystem
**Risk**: Candidate can create files, run scripts
**Mitigation**: Non-root user, resource limits, single-use container
**Impact**: Low (by design, required for tasks)

### 3. No Video Proctoring
**Risk**: Cannot verify identity or see screen
**Mitigation**: Keystroke analysis, timing patterns
**Impact**: Medium (consider for high-stakes interviews)

### 4. Network Access (If Enabled)
**Risk**: External resource access, data exfiltration
**Mitigation**: Optional network isolation, egress monitoring
**Impact**: Medium (disable network if not needed)

### 5. No Real-Time Monitoring
**Risk**: Cheating only detected post-session
**Mitigation**: Client-side blocks common methods
**Impact**: Low (most cheating attempts logged)

## Security Best Practices

### Deployment

1. **Use Docker:**
   ```bash
   # Always run in containers
   docker run --security-opt=no-new-privileges:true \
              --cap-drop=ALL \
              echobox:latest
   ```

2. **Resource Limits:**
   ```bash
   # Always set limits
   --memory="512m" --cpus="0.5" --pids-limit=100
   ```

3. **Network Isolation:**
   ```bash
   # Default to no network
   --network none

   # Enable only if needed
   --network bridge
   ```

4. **Single-Use Containers:**
   ```bash
   # Auto-remove after exit
   docker run --rm echobox:latest
   ```

5. **Monitoring:**
   ```bash
   # Watch container logs
   docker logs -f echobox-candidate1

   # Monitor resources
   docker stats echobox-candidate1
   ```

### Session Review

1. **Always review recordings:**
   ```bash
   ./scripts/analyze.sh sessions/candidate_*/
   ./scripts/replay.sh sessions/candidate_*/
   ```

2. **Check analysis verdict:**
   ```bash
   jq '.verdict, .confidence_score, .flags' sessions/*/analysis.json
   ```

3. **Review anti-cheat events:**
   ```bash
   cat sessions/*/events.log
   ```

4. **Verify file integrity:**
   ```bash
   cd sessions/candidate_*/
   jq -r '.file_hashes | to_entries[] | "\(.value)  \(.key)"' metadata.json | shasum -a 256 -c
   ```

### Production Hardening

**Additional Measures:**

1. **Reverse Proxy:**
   ```nginx
   # nginx config
   location / {
       proxy_pass http://localhost:8080;
       proxy_set_header X-Real-IP $remote_addr;
       auth_basic "Interview Session";
       auth_basic_user_file /etc/nginx/.htpasswd;
   }
   ```

2. **Firewall Rules:**
   ```bash
   # Allow only specific IPs
   iptables -A INPUT -p tcp --dport 8080 -s <candidate_ip> -j ACCEPT
   iptables -A INPUT -p tcp --dport 8080 -j DROP
   ```

3. **TLS/HTTPS:**
   ```bash
   # Use reverse proxy with Let's Encrypt
   # Or configure Go server with TLS
   ```

4. **Audit Logging:**
   ```bash
   # Log all container events
   docker run --log-driver=syslog echobox:latest
   ```

## Compliance Considerations

### Data Privacy (GDPR, etc.)

- **Personal Data**: Candidate name, IP address, session recordings
- **Retention**: Define retention policy (e.g., 90 days)
- **Access**: Limit who can view recordings
- **Deletion**: Implement data deletion process

**Recommendations:**
- Obtain consent before recording
- Anonymize recordings if possible
- Secure session storage (encrypt at rest)
- Implement data retention policy

### Assessment Integrity

- **Anti-Cheat**: Multiple detection layers
- **Tamper-Proof**: SHA-256 hashing
- **Audit Trail**: Complete session recording
- **Fair Use**: Clearly communicate restrictions

## Incident Response

### If Cheating Suspected

1. Review analysis.json verdict
2. Check events.log for paste attempts
3. Replay session with scriptreplay
4. Examine typing patterns (WPM anomalies)
5. Compare solutions with other candidates
6. Document findings

### If Container Compromised

1. Immediately stop container
2. Preserve session recordings
3. Analyze logs for attack vectors
4. Review security settings
5. Update security measures
6. Report if necessary

### If Data Breach

1. Identify scope (which sessions affected)
2. Secure remaining data
3. Notify affected candidates (if required)
4. Implement additional encryption
5. Review access logs

## Security Roadmap

### Implemented (Current)
- ✅ Multi-layer anti-cheat
- ✅ Container isolation
- ✅ Resource limits
- ✅ Recording integrity
- ✅ Non-root execution

### Future Enhancements

**High Priority:**
- [ ] Add basic authentication to web interface
- [ ] Implement TLS/HTTPS support
- [ ] Add rate limiting to HTTP endpoints
- [ ] Implement session encryption at rest

**Medium Priority:**
- [ ] Real-time anti-cheat alerts
- [ ] Browser fingerprinting enhancements
- [ ] AppArmor/SELinux profile
- [ ] Network egress monitoring

**Low Priority:**
- [ ] Video proctoring integration
- [ ] Multi-factor authentication
- [ ] Advanced ML-based cheat detection
- [ ] Blockchain-based recording verification

## Responsible Disclosure

If you discover a security vulnerability:

1. **Do not** publicly disclose without notification
2. Email: security@yourcompany.com (replace with actual)
3. Include: Description, steps to reproduce, impact
4. Allow 90 days for remediation before disclosure

## Security Audit Checklist

- [ ] All Docker security options enabled
- [ ] Resource limits configured
- [ ] Non-root user verified
- [ ] Network isolation reviewed
- [ ] Anti-cheat detection tested
- [ ] Recording integrity verified
- [ ] Access controls implemented
- [ ] Monitoring and logging enabled
- [ ] Documentation reviewed
- [ ] Incident response plan ready

## Conclusion

Echobox implements defense-in-depth with multiple security layers:

1. **Prevention**: Client-side anti-cheat blocks common methods
2. **Detection**: Server-side monitoring catches anomalies
3. **Analysis**: Post-session verdict identifies suspicious patterns
4. **Isolation**: Docker containers limit blast radius
5. **Integrity**: Cryptographic hashing ensures tamper-evidence

While no system is 100% secure, Echobox provides strong protection for technical interview assessments.

**For production use**: Review this document, implement recommended hardening, and adapt to your specific security requirements.
