# Gospeed 
A high-performance encrypted file transfer benchmark written in Go that measures AES encryption/decryption speeds across different file sizes using concurrent operations.

## Overview 
Gospeed is designed to benchmark the performance of AES-256-GCM encryption for file operations. It tests read/write speeds and latency across multiple data sizes, utilizing disk IO. 

*Inspired by gocryptfs*

## Prerequisites 

- Go 1.23 or later (latest version recommended)
- CPU with AES-NI instruction set support _non-aes or softwware based encryption methods will be implemented in the future._ 

## Installation

### Manual Build

```bash
go build -o gospeed .
./gospeed
```

# Expected Output

<img width="625" height="226" alt="Screenshot 2025-09-04 at 6 23 20 PM" src="https://github.com/user-attachments/assets/81ff6c31-ee41-488b-913d-1ac3c4ce6fb8" />


## Encryption Details
- *Algorithm*: AES-256-GCM (Galois/Counter Mode)
- *Key Size*: 256-bit (32 bytes)
- *Symmetrical key exchange*

