# Kibutsu - Minimal Docker Management UI (Draft)

Kibutsu is a full-stack web application that provides a modern, real-time interface for managing Docker containers, images, and compose projects. Built with Go and SvelteKit, it offers a responsive and intuitive experience for Docker management.

## Features

### Container Management
- Real-time container monitoring and stats
- Start, stop, and restart containers
- Live container logs with terminal emulation
- Container health status and metrics

### Image Management
- List and search Docker images
- Pull new images with progress tracking
- Image history and details
- Clean up unused images

### Docker Compose
- Manage multiple compose projects
- Real-time project status monitoring
- Service scaling and orchestration
- Dependency-aware service management

### System Monitoring
- Real-time resource usage metrics
- WebSocket-based live updates
- System-wide Docker statistics
- Disk usage monitoring

## Technology Stack

### Backend
- Go 1.23+
- Docker Engine API
- WebSocket support
- Middleware chain for security and logging
- Context-aware request handling

### Frontend
- SvelteKit 2.x
- TypeScript
- TailwindCSS
- XTerm.js for terminal support
- Real-time stores with Svelte

## Quick Start

1. Clone the repository:

```bash
git clone https://github.com/yourusername/kibutsu.git
cd kibutsu
```
2. Build the frontend:

```bash
cd frontend
pnpm install
pnpm run build
```
3. Build the backend:

```bash
go build -o kibutsuapi
```
4. Run the application:

```bash
./kibutsuapi
```
5. Access the UI at `http://localhost:8080`

## Development

### Frontend Development

```bash
cd frontend
pnpm install
pnpm dev
```

### Backend Development

```bash
go run main.go
```

### View development

The frontend will be available at `http://localhost:5173` with:
- Hot module replacement
- API proxy to backend
- TypeScript checking
- Tailwind CSS processing

The backend API will be available at `http://localhost:8080` with:
- Auto API version negotiation
- Graceful shutdown
- Request ID tracking
- CORS support
- Structured logging

## API Endpoints

### Container Management
- `GET /api/containers` - List containers
- `POST /api/containers/{id}/start` - Start container
- `POST /api/containers/{id}/stop` - Stop container
- `GET /api/containers/{id}/logs` - Stream container logs
- `GET /api/containers/{id}/stats` - Get container statistics

### Image Management
- `GET /api/images` - List images
- `POST /api/images/pull` - Pull new image
- `DELETE /api/images/{id}` - Remove image
- `GET /api/images/{id}/history` - Get image history

### Compose Operations
- `GET /api/compose/projects` - List compose projects
- `POST /api/compose/projects/{name}/up` - Start project
- `POST /api/compose/projects/{name}/down` - Stop project

### System Information
- `GET /api/system/info` - Get system information
- `GET /api/system/version` - Get Docker version
- `GET /api/system/disk` - Get disk usage

## Configuration

The application uses environment variables for configuration:

```bash
DOCKER_HOST=unix:///var/run/docker.sock # Docker daemon socket
PORT=8080 # Server port
CORS_ORIGIN=http://localhost:5173 # Allowed CORS origin
```

## Architecture

### Frontend Store Management

The frontend uses Svelte's stores for state management. Stores are defined in `src/stores.ts` and used throughout the application.

### Backend Request Handling

The backend handles requests using middleware chains. The main handler is `main.go`, which sets up the middleware and routes.

### WebSocket Integration

The application uses WebSockets for live updates. The WebSocket server is implemented in `websocket.go`, which handles connections and broadcasts updates to connected clients.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Acknowledgments

- Docker Engine API
- SvelteKit team
- XTerm.js contributors
- TailwindCSS community