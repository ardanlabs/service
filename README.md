# service

This is a work in process and will attempt to cover best practices around Go based services used in POD architectures and design.

## Docker

```bash
# Build the crud app as a docker image (No local Go installation needed).
docker build -f cmd/crud/dockerfile -t crud-amd64:1 .
```