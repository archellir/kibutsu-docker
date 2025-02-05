# Use Alpine as base
FROM alpine:3.19

# Install required build dependencies
RUN apk add --no-cache \
    nodejs \
    npm \
    go \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy frontend files first
COPY frontend/ frontend/
RUN cd frontend && \
    npm ci && \
    npm run build && \
    cd .. && \
    rm -rf frontend/node_modules

# Copy Go files
COPY . .

# Build Go binary with embedded frontend
RUN mkdir -p backend/static && \
    cp -r frontend/build/* backend/static/ && \
    go build -o kibutsuapi

# Clean up build dependencies
RUN apk del nodejs npm go git && \
    rm -rf /root/.npm /root/.cache /var/cache/apk/*

# Expose port
EXPOSE 8080

# Create volume for Docker socket
VOLUME /var/run/docker.sock

# Run application
CMD ["./kibutsuapi"]