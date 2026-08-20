# Orbit-LB 🪐

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![HTMX](https://img.shields.io/badge/HTMX-1.9.11-336699?style=for-the-badge&logo=htmx)](https://htmx.org/)
[![Templ](https://img.shields.io/badge/Templ-HTML_in_Go-FFD700?style=for-the-badge&logo=go)](https://templ.guide/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

**Orbit-LB** is an incredibly sleek, real-time visualizer for a custom Layer 7 Reverse Proxy and Load Balancer built in Go. It simulates, manages, and visualizes complex distributed systems infrastructure directly in the browser using HTMX and Templ.

> **Kernel Mode Activated:** Built by Gummybearansh to engineer uncompromising backend infrastructure.

## 🚀 The Tech Real: How it Works

Under the hood, Orbit-LB is not just a UI trick—it's a functioning proxy simulation that demonstrates core distributed systems engineering concepts:

- **Goroutine-backed Ephemeral Nodes:** When you horizontally scale ("Add Server"), the backend issues a command to spawn an isolated server process (Goroutine) on a discrete port on the host machine. 
- **L7 Proxy Interception:** The central proxy node terminates incoming HTTP/TCP connections and multiplexes the payload traffic across the active cluster.
- **Service Discovery & Routing:** The proxy maintains an active registry of healthy backend instances. It actively routes payloads using a deterministic Round-Robin algorithm.
- **Resource Optimization (TTL):** Nodes are fully ephemeral. To optimize resources, any server node that receives 0 traffic for 20 continuous seconds will automatically trigger a graceful self-termination and deregister from the proxy.

## ⚡ Tech Stack

- **Backend System:** Pure `Go`. Handles the HTTP routing, the proxy logic, concurrent state management (mutex-locked registries), and ephemeral node lifecycles.
- **Frontend Interactivity:** `HTMX`. Achieves real-time server-side state syncing and DOM updates without writing massive React states. The UI polls backend health routes seamlessly.
- **Templating Engine:** `a-h/templ`. Type-safe HTML generation directly in Go, ensuring the UI perfectly reflects the strict backend data structures.
- **Styling & Animation:** `TailwindCSS` with custom `requestAnimationFrame` JavaScript loops for the intricate orbital physics, hover-pausing, and particle traffic routing.

## 🛠️ Quick Start

Want to boot up the cluster on your local machine? 

```bash
# Clone the repository
git clone https://github.com/gummybearansh/orbit-lb.git
cd orbit-lb

# Generate the Templ files (requires templ CLI)
templ generate

# Boot the Proxy
go run cmd/dashboard/main.go
```
Visit `http://localhost:3000` to access the Kernel Mode control panel.

## 🎮 Interacting with the Cluster

Once booted, the app provides a highly polished interactive tour. You can also use native keyboard shortcuts for maximum efficiency:
- **`A`**: Issue a syscall to spawn a new server node (Horizontal Scaling).
- **`Space` (Tap)**: Dispatch a single HTTP payload.
- **`Space` (Hold)**: Sustain an automated barrage of payloads based on your configured RPS rate. Hover over any node to freeze the cluster and inspect the multiplexed traffic!
