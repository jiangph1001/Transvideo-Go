# Video Transcoder Management System

A high-performance video transcoding management system built with Go, featuring hardware acceleration support (NVIDIA NVENC, Intel QSV, AMD AMF), SQLite persistence, and a web-based management interface.

## Features

- 🎬 **Multiple Video Codecs**: H.264, HEVC/H.265, AV1
- ⚡ **Hardware Acceleration**: NVIDIA NVENC, Intel QSV, AMD AMF
- 🏫 **Web Management Interface**: Real-time task monitoring
- 📊 **RESTful API**: Programmatic task management
- 📓 **SQLite Persistence**: Task recovery after restart
- 🚀 **Concurrent Processing**: Configurable worker pool
- 👥 **Docker Isolation**: Secure transcoding with dynamic mounting
- 🔄 **Task Recovery**: Automatic restoration of pending tasks

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- NVIDIA/Intel/AMD GPU (optional, for hardware acceleration)

### Method 1: Local Development (Recommended)

```bash
# Check Docker environment
tools/check-docker.sh

# Generate a test video (optional)
make demo-video

# Build and run locally
make build
make run

# Access the web interface
open http://localhost:8080
```

### Method 2: Docker Deployment

```bash
# Configure environment
cp .env.example .env
# Edit .env and configure INPUT_DIR, OUTPUT_DIR

# Run with Docker
make docker-run

# Access the web interface
open http://localhost:8080
```

### Method 3: Quick Test Script

```bash
# Test with your video file
tools/quick-test.sh /path/to/your/video.mp4
```

## Hardware Acceleration

This system supports multiple hardware encoding options for **5-10x faster** transcoding:

| Encoder | Speed | Quality | Requirements |
|---------|-------|---------|--------------|
| **Software (libsvtav1)** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Any CPU |
| **NVIDIA (NVENC)** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | NVIDIA GPU (RTX 10/16/20/30/40 series) |
| **Intel (QSV)** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Intel GPU (Arc/Xe/UHD) |
| **AMD (AMF)** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | AMD GPU (RX 5000/6000/7000 series) |

### Configure Hardware Encoding

Edit `.env` file:

```bash
# Encoder type: software, nvenc, qsv, amf
ENCODER_TYPE=nvenc

# Video codec: h264, hevc, av1
VIDEO_CODEC=hevc

# NVIDIA preset: p1-p7 (p1=fastest, p7=best quality)
# Intel preset: veryfast, faster, fast, medium, slow, slower, veryslow
# AMD preset: speed, balanced, quality
HARDWARE_PRESET=p6

# Quality factor (lower=better quality, larger file)
CRF=28
```

**Example for NVIDIA GPU:**

```bash
ENCODER_TYPE=nvenc
VIDEO_CODEC=hevc  # or h264 for compatibility
HARDWARE_PRESET=p6
CRF=28
MAX_CONCURRENCY=4
INPUT_DIR=/path/to/videos
OUTPUT_DIR=/path/to/encoded
```

**Performance Comparison (10-minute 1080p video):**

| Encoder | Time | CPU/GPU Usage | File Size | Quality |
|---------|------|---------------|-----------|---------|
| Software AV1 | 45 min | 100% CPU | 350MB | Excellent |
| NVENC HEVC | 3 min | 5% CPU, 45% GPU | 500MB | Excellent |
| QSV HEVC | 4 min | 5% CPU, 40% GPU | 480MB | Excellent |

See [HARDWARE_ENCODING.md](HARDWARE_ENCODING.md) for detailed configuration.

## Web Interface

The system provides a clean, responsive web interface at `http://localhost:8080`:

- **Dashboard**: Real-time statistics (pending, running, completed, failed tasks)
- **Task Management**: Add, cancel, and monitor transcoding tasks
- **Auto-refresh**: Updates every 5 seconds
- **Status Indicators**: Color-coded task status (pending/running/completed/failed)

### Adding a Task

1. Enter input filename (e.g., `video.mp4`)
2. Optionally set bitrate (e.g., `1M` for 1 Mbps)
3. Click "Add Task"

The output file will be named `{input_filename}_{codec}.{ext}` (e.g., `video_hevc.mp4`)

## REST API

### Create Task
```bash
POST /api/tasks
Content-Type: application/json

{
  "input_file": "video.mp4",
  "bitrate": "1M",
  "config": {
    "video_codec": "hevc",
    "encoder_type": "nvenc"
  }
}
```

**Response:**
```json
{
  "id": "task_uuid",
  "input_file": "video.mp4",
  "bitrate": "1M",
  "status": "pending",
  "created_at": "2025-11-30T10:00:00Z"
}
```

### List All Tasks
```bash
GET /api/tasks
```

**Response:**
```json
[
  {
    "id": "task_uuid",
    "input_file": "video.mp4",
    "status": "completed",
    "output_file": "video_hevc.mp4",
    "created_at": "2025-11-30T10:00:00Z",
    "started_at": "2025-11-30T10:00:10Z",
    "completed_at": "2025-11-30T10:03:15Z"
  }
]
```

### Get Single Task
```bash
GET /api/tasks/{id}
```

### Cancel Task
```bash
POST /api/tasks/{id}/cancel
```

### Get Statistics
```bash
GET /api/stats
```

**Response:**
```json
{
  "pending": 2,
  "running": 1,
  "completed": 5,
  "failed": 0,
  "total": 8,
  "max_concurrency": 4
}
```

### Update Concurrency
```bash
POST /api/concurrency
Content-Type: application/json

{
  "max_concurrency": 6
}
```

## Configuration

Create a `.env` file based on `.env.example`:

```bash
# Input and output directories (absolute paths)
INPUT_DIR=/path/to/input/videos
OUTPUT_DIR=/path/to/output/videos

# Maximum concurrent transcoding tasks (default: 2)
MAX_CONCURRENCY=4

# Hardware encoding settings
ENCODER_TYPE=nvenc                    # software, nvenc, qsv, amf
VIDEO_CODEC=hevc                      # h264, hevc, av1
HARDWARE_PRESET=p6                    # p1-p7, or balanced/fast/medium
CRF=28                                # Quality factor (20-40)

# Docker settings (optional)
DOCKER_HOST=                          # Leave empty for local Docker

# FFmpeg container settings
FFMPEG_IMAGE=jrottenberg/ffmpeg:ubuntu
AUDIO_CODEC=aac
AUDIO_BITRATE=128k
```

**Recommended Concurrency by CPU Cores:**

| CPU Cores | Software Encoding | Hardware Encoding |
|-----------|-------------------|-------------------|
| 4 cores   | 1-2 tasks         | 2-3 tasks         |
| 8 cores   | 2-3 tasks         | 4-6 tasks         |
| 16 cores  | 4-6 tasks         | 8-12 tasks        |

## Task Persistence (SQLite)

All tasks are automatically saved to SQLite database (`tasks.db`):

- **Automatic Save**: All tasks persisted immediately
- **Restart Recovery**: Pending tasks restored on startup
- **History Retention**: Completed and failed tasks kept for analysis

**Query Database:**
```bash
# View all tasks
sqlite3 tasks.db "SELECT * FROM tasks ORDER BY created_at DESC;"

# View failed tasks
sqlite3 tasks.db "SELECT * FROM tasks WHERE status = 'failed';"

# Backup database
cp tasks.db tasks_backup.db
```

**Recovery Flow:**
```
Application Start
    ↓
Load pending tasks → Re-queue them
    ↓
Load running tasks → Mark as failed (interrupted)
    ↓
Start worker pool
    ↓
Process queued tasks
```

See [DATABASE.md](DATABASE.md) for detailed information.

## Project Structure

```
video-transcoder/
├── main.go                      # Application entry point
├── config/                      # Configuration management
│   └── config.go
├── models/                      # Data models
│   ├── task.go                  # Task structure
│   └── config.go                # Config structure
├── service/                     # Business logic
│   ├── task_service.go          # Task queue & worker pool
│   └── docker_executor.go       # Docker-based transcoding
├── handler/                     # HTTP handlers
│   └── task.go                  # REST API handlers
├── static/                      # Web UI
│   ├── index.html              # Dashboard page
│   └── app.js                  # Frontend logic
├── tools/                       # Utility scripts
│   ├── check-docker.sh         # Docker environment check
│   └── quick-test.sh           # Quick test script
├── .env.example                 # Environment variables template
├── docker-compose.yaml         # Docker Compose configuration
├── Dockerfile                  # Application container
├── Makefile                    # Build & run automation
├── go.mod                       # Go dependencies
└── README.md                    # This file
```

## Makefile Commands

```bash
make build          # Compile the application
make run            # Run locally
make clean          # Clean build artifacts
make test           # Run tests

# Docker commands
make docker-build   # Build Docker image
make docker-run     # Run Docker container
make docker-logs    # View container logs
make docker-stop    # Stop container

# Demo/Testing
make demo-setup     # Create test directories
make demo-video     # Generate test video
make demo           # Complete demo setup

# Development
make fmt            # Format Go code
make lint           # Run linter
make help           # Show all commands
```

## Troubleshooting

### Docker Permission Error
```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Re-login or restart
```

### Port 8080 Already in Use
```bash
# Find process using port
lsof -i :8080
# Or change port in main.go
```

### No Video Output
```bash
# Check input directory
ls -lh test-videos/

# Check Docker logs
make docker-logs

# Verify FFmpeg in container
docker run --rm jrottenberg/ffmpeg:ubuntu ffmpeg -version
```

### Hardware Encoding Not Working

**NVIDIA:**
```bash
# Test GPU access
docker run --rm --gpus all nvidia/cuda:12.2.0-base-ubuntu22.04 nvidia-smi

# Install nvidia-container-toolkit
sudo apt-get install nvidia-container-toolkit
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker
```

**Intel:**
```bash
# Install Intel driver
sudo apt-get install intel-media-va-driver
vainfo  # Should show GPU info
```

See [HARDWARE_ENCODING.md](HARDWARE_ENCODING.md) for detailed troubleshooting.

## Security Considerations

- **No Root**: Runs as non-root user (`transcoder:transcoder`)
- **Dynamic Mounting**: Main container doesn't mount video directories
- **Docker Isolation**: Each transcoding job runs in separate container
- **No Secrets**: Environment variables used instead of hardcoded secrets
- **Socket Access**: Only Docker socket mounted (read-only where possible)

## Performance Tuning

### Optimize for Speed
```bash
VIDEO_CODEC=h264          # Fastest encoding
HARDWARE_PRESET=p1        # Fastest preset (NVENC)
CRF=30                    # Lower quality, smaller file
MAX_CONCURRENCY=8         # More parallel tasks
```

### Optimize for Quality
```bash
VIDEO_CODEC=av1           # Best compression
HARDWARE_PRESET=p7        # Best quality (NVENC)
CRF=20                    # Higher quality, larger file
MAX_CONCURRENCY=2         # Less parallel, more resources per task
```

### Optimize for File Size
```bash
VIDEO_CODEC=av1           # Best compression ratio
HARDWARE_PRESET=veryslow  # Software: slower but smaller
CRF=20                    # Lower CRF = better quality
AUDIO_CODEC=libopus       # Better audio compression
```

## Development

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Make

### Build from Source
```bash
# Clone repository
git clone <repository-url>
cd video-transcoder

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run locally
make run
```

### API Integration Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type CreateTaskRequest struct {
    InputFile string `json:"input_file"`
    Bitrate   string `json:"bitrate"`
    Config    map[string]string `json:"config"`
}

func main() {
    req := CreateTaskRequest{
        InputFile: "video.mp4",
        Bitrate:   "1M",
        Config: map[string]string{
            "video_codec":  "hevc",
            "encoder_type": "nvenc",
        },
    }

    data, _ := json.Marshal(req)
    resp, _ := http.Post("http://localhost:8080/api/tasks", "application/json", bytes.NewBuffer(data))

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Printf("Task created: %v\n", result)
}
```

## License

[Your License Here]

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

- 📚 **Documentation**: See [QUICKSTART.md](QUICKSTART.md) for detailed guide
- 🔧 **Advanced Config**: See [HARDWARE_ENCODING.md](HARDWARE_ENCODING.md) for hardware setup
- 📚 **Database**: See [DATABASE.md](DATABASE.md) for persistence details
- 🔗 **Project Structure**: See [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) for architecture

## Changelog

### v1.0.0 (2025-11-24)
- ✅ Hardware acceleration support (NVENC, QSV, AMF)
- ✅ Multiple video codecs (H.264, HEVC, AV1)
- ✅ Web-based management interface
- ✅ RESTful API
- ✅ SQLite task persistence
- ✅ Docker-based transcoding
- ✅ Concurrent task processing
- ✅ Task recovery on restart
- ✅ Graceful shutdown

---

**Note**: This system is designed for batch video transcoding with hardware acceleration. For live streaming or real-time encoding, consider dedicated streaming solutions.
