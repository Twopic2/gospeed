# Gospeed 
A high-performance encrypted file transfer benchmark written in Go that measures AES encryption/decryption speeds across different file sizes using concurrent operations.

## Overview 
Gospeed is designed to benchmark the performance of AES-256-GCM encryption for file operations. It tests read/write speeds and latency across multiple data sizes, utilizing disk IO. 

*Inspired by gocryptfs*

## Prerequisites 

- Go 1.22 or later (latest version recommended)
- CPU with AES-NI instruction set support _non-aes or softwware based encryption methods will be implemented in the future._ 

## Installation

### Manual Build

```bash
go build -o gospeed .
./gospeed
```

# Expected Output

<img width="685" height="221" alt="Screenshot_20250820_201418" src="https://github.com/user-attachments/assets/683645c8-65ce-46da-b807-be0fe2322c56" />



## Encryption Details
- *Algorithm*: AES-256-GCM (Galois/Counter Mode)
- *Key Size*: 256-bit (32 bytes)
- *Symmetrical key exchange*

