# GitHub Stats API Go

[🇬🇧](README.md) | [🇫🇷](README.fr.md)

![GitHub Stats API Go](https://github.com/sup2ak/github-stats-api-go/actions/workflows/release.yml/badge.svg)
![GitHub CI](https://github.com/sup2ak/github-stats-api-go/actions/workflows/ci.yml/badge.svg)

## Overview

`github-stats-api-go` is a Go-based API that retrieves statistics from GitHub for a specified user. It provides information such as the number of followers, repositories, stars, and organizations associated with the user.

## Features

- Retrieve user statistics from GitHub
- Includes options for fetching stars, followers, following, repositories, and organizations
- Simple HTTP/HTTPS server to handle requests

## Installation

To use this API, you need to have Go installed on your machine. You can install it from the [official Go website](https://golang.org/dl/).

1. Initialize a new Go module:

```bash
go mod init your_project_name
```

2. Install the module using `go get`:

```bash
go get github.com/SUP2Ak/github-stats-api-go
```

3. Create a new Go file (e.g., `main.go`) and import the module:

```go
package main
import (
    "fmt"
    "time"
    githubstats "github.com/sup2ak/github-stats-api-go"
)

func main() {
    // Configuration
    config := githubstats.Config{
        Token: "your_github_token",
        IP:    "0.0.0.0",
        Port:  "8080",
        Scheme: "http", // or "https" null value is "http"
        Path:  "/stats",
        RateLimit: 10,
        CacheDuration: 1 * time.Hour,
        // and more...
    }
    // Create an instance of GStats
    stats := githubstats.GStats{}
    // Connect to the GitHub API
    err := stats.Connect(config)
    if err != nil {
        log.Fatalf("Connection error: %v", err)
    }
}
```

4. Run the program:

```bash
go run main.go
```

## Example Request

To get statistics for the user `sup2ak`, you can make a GET request:

```bash
curl "http://localhost:8080/stats?username=sup2ak"
```

## License

This project is licensed under the [GPL-3.0 License](LICENSE).

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.
