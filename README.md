# Mini AWS 🚀

To build the world's best local-first cloud simulator that teaches cloud concepts through practice.

## ✨ Features
- **Compute**: Docker-based instance management (Launch, Stop, List)
- **Storage**: S3-compatible object storage (Upload, Download, Delete)
- **Identity**: API Key authentication

## 🚀 Quick Start
```bash
# 1. Clone & Setup
git clone https://github.com/PoyrazK/Mini_AWS.git
cd Mini_AWS
make run

# 2. Test health
curl localhost:8080/health

# 3. Get an API Key
cloud auth create-demo my-user

# 4. Launch an instance
cloud compute launch --name my-server --image nginx:alpine

# 5. Upload a file
cloud storage upload my-bucket README.md
```

## 🏗️ Architecture
- **Backend**: Go (Clean Architecture)
- **Database**: PostgreSQL (pgx)
- **Infrastructure**: Docker Engine
- **CLI**: Cobra (command-based) + Survey (interactive)

## 📚 Documentation
- [Development Guide](docs/development.md) - Setup on Windows/Mac
- [Database Guide](docs/database.md) - Schema & migrations
- [Storage Guide](docs/guides/storage.md) - Object storage usage

## 📊 KPIs
- Time to Hello World: < 5 min
- API Latency (P95): < 200ms
- CLI Success Rate: > 95%
