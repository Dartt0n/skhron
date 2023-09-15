# Skhron
![go workflow](https://github.com/dartt0n/skhron/actions/workflows/go.yml/badge.svg)

Skhron is a simple in-memory storage with active cleaning and rest http api

## Build

```bash
docker build -t skhron-image .
```

## Run

```bash
docker run -e ADDRESS=:9090 -e PERIOD=5 -p 9090:9090 --name skhron skhron-image
```
