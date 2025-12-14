# Work Engine by KitWork

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/doc/go1.25)
[![License: AGPL v3](https://img.shields.io/badge/license-AGPL%203.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/kitwork/work?color=green&label=Release)](https://github.com/kitwork/work/releases)

**Work Engine** is the core automation and workflow execution engine behind 👉 **[KitWork](https://github.com/kitwork/kitwork)**.

It is designed to orchestrate tasks using a **simple YAML-based DSL** or **native Go code**, focusing on **clarity, performance, and extensibility**.

If you like tools such as Zapier or n8n but want **full control in Go**, Work Engine is for you.

---

## 📋 Table of Contents

- [Philosophy](#philosophy)
- [Core Concept: $result](#core-concept-result)
- [Variables: as and alias](#variables-as-and-alias)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Workflow Configuration](#workflow-configuration)
- [Variables & Context](#variables--context)
- [Complete Example](#complete-example)
- [Step-by-Step Execution](#step-by-step-execution)
- [Who Is This For?](#who-is-this-for)
- [License](#license)

---

## Philosophy

Work Engine is built around a few core principles:

| Principle | Description |
|-----------|-------------|
| **Workflow as Data** | Workflows are declarative, readable, versionable, and diff-friendly |
| **Result-driven Execution** | Data flows through the pipeline via `$result`, not via mutable variables |
| **Minimal Abstractions** | No BPMN, no heavy orchestration layers, no magic state |
| **Developer First** | Easy to embed, easy to extend, easy to debug |
| **Pipeline-first Execution** | Explicit data flow with no hidden mutable state |
| **YAML-first, Go-native** | A data transformation pipeline, not just a task runner |

---

## Core Concept: `$result`

### `$result` is the pipeline

`$result` always represents the **output of the most recently executed work step**.  
It is **automatically overwritten** after each step.

```text
work[0] → $result
work[1] → $result (overwrites previous)
work[2] → $result (overwrites previous)
```

**Key Points:**
- You never manually assign `$result`
- Each step reads from `$result` and produces a new `$result`
- Work Engine is **result-driven, not variable-driven**

---

## Variables: `as` and `alias`

Work Engine allows you to **store or reference** the current `$result` using `$`-prefixed variables.

| Keyword | Purpose | Behavior | Example |
|---------|---------|----------|---------|
| `as` | Store a copy | Creates a copy of `$result` in a named variable | `as: $price` |
| `alias` | Create reference | Creates an alias/reference to `$result` | `alias: $list` |

### `as` - Store Copy

```yaml
as: $price
```

**After execution:**
```text
$price  = $result (copy)
$result = unchanged
```

### `alias` - Create Reference

```yaml
alias: $list
```

**After execution:**
```text
$list   = $result (reference)
$result = unchanged
```

> **Note:** Both `as` and `alias` **require the `$` prefix** as they operate on context variables.

---

## Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Lightweight** | High performance with minimal overhead |
| **Pure Go** | No runtime dependencies |
| **Embeddable** | Works as CLI, Server, Worker, or Cron service |

### Workflow Capabilities

| Category | Features | Description |
|----------|----------|-------------|
| **SQL** | `select`, `insert`, `update`, `delete` | Dynamic queries with variables and expressions |
| **HTTP** | `fetch`, `request` | Full support for headers, body, timeout, authentication |
| **Control Flow** | `foreach`, `switch`, `loop`, `wait` | Conditional and iterative execution |
| **Browser** | `chrome`, `screenshot` | Browser automation powered by `chromedp` |
| **System** | `cmd`, `script` | Execute shell commands and scripts |
| **Data Transform** | `target`, `mapping`, `filter` | Transform and manipulate data |

### Observability

- ✅ Structured logging via `slog`
- ✅ Built-in tracing support
- ✅ Step-by-step debugging
- ✅ Workflow execution monitoring

---

## Installation

```bash
go get github.com/kitwork/work
```

**Requirements:**
- Go 1.25 or higher
- Linux, macOS, or Windows

---

## Quick Start

### Minimal Example (Go)

```go
package main

import "github.com/kitwork/work"

func main() {
	err := work.New().
		Secret().   // Load secrets from env/vault
		Schedule(). // Enable cron scheduler
		Router().   // Enable HTTP endpoint routing
		Run()       // Start the engine

	if err != nil {
		panic(err)
	}
}
```

### Work Engine Run Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| **Embedded Engine** | Run workflows programmatically | Integration with existing Go apps |
| **HTTP Server** | Expose workflows as API endpoints | Automation APIs, webhooks |
| **Background Worker** | Execute long-running tasks | Async processing, queues |
| **Scheduled Service** | Cron-like task scheduling | Periodic jobs, maintenance tasks |

---

## Workflow Configuration

A workflow (called a **Work**) is defined as a list of steps under the `work:` key.

### Workflow Structure

```yaml
name: workflow-name
version: 1

work:
  - step1:
      # configuration
  - step2:
      # configuration

success:
  - return:
      # success response

error:
  - return:
      # error response
```

### Execution Flow

Each step follows this pattern:

1. **Read** from the current `$result`
2. **Execute** its logic
3. **Produce** a new `$result` (overwriting the previous one)

---

## Variables & Context

All data access uses `$`-based expressions.

### Built-in Variables

| Variable | Description | Scope |
|----------|-------------|-------|
| `$result` | Output of the last step (pipeline value) | Global |
| `$item` | Current item in `foreach` or `mapping` | Loop scope |
| `$index` | Index of the current item | Loop scope |
| `$body` | HTTP request body | Router context |
| `$queries` | HTTP query parameters | Router context |
| `$headers` | HTTP headers | Router context |
| `$env.NAME` | Environment variable access | Global |
| `$custom` | Custom variables (via `as` or `alias`) | User-defined |

### Variable Access Examples

```yaml
# Access nested data
$result.data.items[0].name

# Environment variables
$env.DATABASE_URL

# HTTP context
$body.user_id
$queries.page
$headers.Authorization

# Loop variables
$item.price
$index
```

---

## Complete Example

### Real-World Workflow: Gold Price API

```yaml
name: gold-price-api
version: 1

work:
  # Step 1: Fetch gold prices from API
  - fetch:
      name: "Lấy giá vàng PNJ"
      url: "https://edge-api.pnj.io/ecom-frontend/v1/get-gold-price?zone=11"
      method: GET
      timeout: 10s
      as: $price

  # Step 2: Focus on the data array
  - target: $price.data

  # Step 3: Transform each item
  - mapping:
      buy:  $item.giaban
      sell: $item.giamua
      code: $item.masp
      name: $item.tensp

  # Step 4: Save transformed result
  - alias: $list

success:
  - return:
      json:
        success: true
        data: $list
        location: $price.chinhanh
        updated_at: $price.updateDate
        source: "https://github.com/kitwork/kitwork"

error:
  - return:
      json:
        success: false
        message: "Failed to fetch or process data"
```

---

## Step-by-Step Execution

### Data Flow Visualization

| Step | Action | `$result` Value | Additional Variables |
|------|--------|----------------|---------------------|
| **1. fetch** | HTTP GET request | Full API response | `$price` = response |
| **2. target** | Focus on data array | `$price.data` (array) | - |
| **3. mapping** | Transform each item | Transformed array | - |
| **4. alias** | Create reference | Unchanged | `$list` = transformed array |
| **5. return** | Build response | Final JSON | - |

### Detailed Explanation

#### Step 1: Fetch Data
```yaml
- fetch:
    url: "https://api.example.com/data"
    as: $price
```
- Makes HTTP request
- Stores response in `$result`
- Creates copy in `$price` variable

#### Step 2: Target Specific Data
```yaml
- target: $price.data
```
- Extracts `data` field from `$price`
- Overwrites `$result` with the array
- `$price` remains unchanged

#### Step 3: Transform Data
```yaml
- mapping:
    buy:  $item.giaban
    sell: $item.giamua
```
- Iterates over each item in `$result`
- Creates new object for each `$item`
- Produces new transformed array in `$result`

#### Step 4: Save Result
```yaml
- alias: $list
```
- Creates `$list` reference to `$result`
- Allows accessing transformed data later
- `$result` remains unchanged

---

## Who Is This For?

| Audience | Why Work Engine? |
|----------|------------------|
| **Backend Engineers** | Build internal tools with Go's type safety and performance |
| **Indie Hackers** | Create automated services without complex infrastructure |
| **DevOps Teams** | Orchestrate workflows with version-controlled YAML |
| **Platform Teams** | Need reliable automation without vendor lock-in |

### Use Cases

- 🔄 API orchestration and data transformation
- 📊 ETL pipelines and data processing
- 🤖 Browser automation and web scraping
- 📧 Notification and alert systems
- 🔗 Webhook handlers and integrations
- ⏰ Scheduled maintenance tasks
- 🔐 Secure data workflows with secret management

---

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

**Copyright © 2025 KitWork**  
**Author:** Huỳnh Nhân Quốc

```
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
```

See the [LICENSE](LICENSE) file for details.