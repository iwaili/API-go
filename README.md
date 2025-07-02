his repository provides a lightweight transcription API built in Go, integrated with Whisper.cpp, a fast and efficient C++ implementation of OpenAI’s Whisper model. The API accepts audio uploads, segments them into 5-second chunks using FFmpeg, and streams transcription results in real time via newline-delimited JSON.

The server is designed for concurrency, with audio chunks processed in parallel using goroutines. API access is secured using Bloom filter–based key validation. The system is optimized for performance on Apple Silicon (including M-series chips) and is containerized for easy deployment using Docker.

This project showcases an end-to-end, on-device speech-to-text pipeline with a focus on performance, modularity, and low-latency operation, without relying on external cloud services.
