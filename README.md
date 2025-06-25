# Golem 🗿

[![Build Status](https://img.shields.io/github/actions/workflow/status/Advik-B/Golem/go.yml?branch=main)](https://github.com/Advik-B/Golem/actions)

**Golem is a from-scratch implementation of a high-performance Minecraft server written entirely in Go.** Inspired by the feature-rich and highly configurable Purpur server, Golem aims to leverage the power of Go's concurrency and simplicity to create a modern, maintainable, and extremely fast server platform for Minecraft.

A core feature of Golem is its first-class support for plugins written in **JavaScript** via the [goja](https://github.com/dop251/goja) engine, providing a safe, isolated, and easy-to-use API for server customization.

---

### Core Philosophy

The Golem project is guided by a few key principles:

1.  **Performance through Concurrency:** Leverage Go's native goroutines and channels to handle player connections, world processing, and other tasks concurrently, aiming for superior performance and scalability.
2.  **Simplicity and Maintainability:** By building from the ground up in Go, we avoid the complexities of a heavily patched Java codebase. The goal is a server that is easier to understand, maintain, and contribute to.
3.  **Extreme Configurability:** Like its inspiration, Purpur, Golem will offer deep and granular control over gameplay mechanics, mob behavior, and server settings through a simple YAML configuration.
4.  **Modern Extensibility:** Provide a powerful and safe plugin API using JavaScript. This lowers the barrier to entry for developers and allows for rapid development of server customizations without needing to compile Go code.

### ⚠️ Current Status: Highly Experimental

Golem is in the very early stages of development. The project is currently focused on building the core infrastructure. **It is not yet playable.**

**Progress Checklist:**
-   [x] TCP Server & Connection Handling
-   [x] Server List Ping (Status & Ping Protocol)
-   [ ] **(In Progress)** Login Protocol & Initial World Entry
-   [ ] World State Management (Chunks, Blocks, Entities)
-   [ ] Main Game Loop & Entity Ticking
-   [ ] Player Movement & Basic Physics
-   [ ] World Persistence (Anvil Format I/O)
-   [ ] Purpur Feature Implementation
-   [ ] Goja Plugin System API

### Planned Features

-   **Go-Native Engine:** A completely fresh implementation with no Java dependencies, designed for performance.
-   **Purpur-Inspired Gameplay:**
    -   Hundreds of configuration options to tailor the exact Minecraft experience you want.
    -   Unique gameplay mechanics like ridable mobs, advanced enchantments, and custom block behaviors.
-   **JavaScript Plugin API (via Goja):**
    -   **Isolated Runtimes:** Each plugin runs in its own JS environment, preventing one plugin from crashing the entire server.
    -   **Clean and Safe API:** A well-defined Go API exposed to JavaScript, providing powerful tools without exposing raw server internals.
    -   **Event-Driven Architecture:** Register JS callbacks for game events like player joins, block breaks, and more.
    -   **Simple Command System:** Easily register custom server commands directly from JavaScript.

### Technology Stack

-   **Language:** [Go](https://go.dev/)
-   _(planned)_ **Scripting Engine:** [goja](https://github.com/dop251/goja) (ECMAScript 5.1+ implementation)

### Getting Started

To get the project running, you'll need Go installed on your system (version 1.21 or newer is recommended).

```bash
# 1. Clone the repository
git clone https://github.com/Advik-B/Golem.git
cd Golem

# 2. Tidy dependencies
go mod tidy

# 3. Run the server
go run ./cmd/golem
```

The server will start listening on `0.0.0.0:25565`. You can now ping the server from your Minecraft client.

### Contributing

We are actively seeking contributors! Whether you're a Go expert, have experience with the Minecraft protocol, or just want to help with documentation, there's a place for you here.

1.  **Find Something to Work On:** Check out our **[ROADMAP.md](ROADMAP.md)** and the open issues on our [GitHub Issues](https://github.com/Advik-B/Golem/issues) page.
2.  **Discuss:** It's always a good idea to comment on an issue or start a discussion before you begin a major implementation.
3.  **Code:** Follow standard Go idioms and best practices. Ensure your code is formatted with `gofmt`.
4.  **Pull Request:** Submit a PR with a clear description of the changes you've made and why.

### License

Golem is licensed under the **Apache-2.0 license**. See the [`LICENSE.txt`](./LICENSE.txt) file for more details.
