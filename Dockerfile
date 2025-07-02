FROM golang:1.24.1

WORKDIR /major

# Install build dependencies (ffmpeg is required by your Go app)
RUN apt-get update && apt-get install -y \
    build-essential \
    libopenblas-dev \
    git \
    ffmpeg \
 && apt-get clean && rm -rf /var/lib/apt/lists/*

# Copy your Go app files into the container
COPY . .

# Expose the app port
EXPOSE 8081

# Run the server using go run
CMD ["go", "run", "main.go"]
