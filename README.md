# 🚀 WebServer: High-Performance Router

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](LICENSE)

A lightweight, blazing-fast HTTP multiplexer for Go, built on a custom **Trie (Prefix Tree)** data structure. Designed for developers who want the power of a modern framework without the "bloat."

---

## ✨ Features

* **⚡ Trie-Based Routing:** O(n) lookup time where *n* is the number of path segments.
* **🎭 Dynamic Parameters:** Support for `:id` style segments (e.g., `/users/:id/profile`).
* **🌐 Catch-All Wildcards:** Use `*` to capture entire sub-paths—perfect for static file servers.
* **📂 Static Asset Excellence:** Built-in `ServeFileSystem` with automatic directory listing.
* **🧹 Zero Dependencies:** Only uses the Go standard library.

---

## 🛠️ Quick Start

### 1. Installation
```bash
go get [github.com/zuadi/webServer](https://github.com/zuadi/webServer)
````
### 2. Basic Usage (RESTful Example)
```bash
package main

import (
    "net/http"
    "webServer/router"
)

func main() {
    ws := router.NewRouter()

    // GET Request
    ws.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Welcome to the Trie! 🌳"))
    })

    // POST Request (REST API)
    ws.Post("/api/data", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusCreated)
        w.Write([]byte(`{"status": "success"}`))
    })

    // Serve Static Files with Wildcards
    // This correctly strips the prefix and handles directory listings!
    ws.ServeFileSystem("/assets/*", "./static")

    http.ListenAndServe(":4040", ws)
}
```

---

## 📂 Route Matching Priority

GopherTrie uses a specific priority system to ensure the most logical route wins:

1. **Exact Match:** /users/profile always beats a parameter.

2. **Parameter Match:** /users/:id captures single segments.

3. **Wildcard Match:** /static/* captures everything remaining if no other match is found.

Note: Directory listings are supported on /folder/* paths by correctly handling the trailing slash logic within the Trie.

