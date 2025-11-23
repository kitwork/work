# Work Engine by KitWork

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/doc/go1.25)
[![License: AGPL v3](https://img.shields.io/badge/license-AGPL%203.0-blue.svg)](LICENSE)

**KitWork Engine** is the core automation engine behind [KitWork](https://github.com/kitwork/kitwork).  
It provides the foundation for running, scheduling, and orchestrating workflows efficiently.

---

## Features

- Lightweight and high-performance workflow engine
- Modular architecture
- Fully compatible with KitWork CLI and Work files
- Open-source under AGPL-3.0

---

## Getting Started

### Install

```bash
go get github.com/kitwork/work
````

### Example Usage

```go
package main

import "github.com/kitwork/work"

func main() {
    // Run tasks from source folder
    err := work.Source("./source/tasks").Run()
    if err != nil {
        panic(err)
    }
}
```

---

## Source Access (AGPL Requirement)

If you run this program on a network server, the **AGPL-3.0** requires that you make the source code of your modified version available to all users interacting with it remotely.
For example, you can provide a "Source" link in your web interface or make it available via a network server.

---

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

```
Copyright (c) 2025 KitWork, Huỳnh Nhân Quốc

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```
