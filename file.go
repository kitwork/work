// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"os"
)

type File struct {
	Content []byte
}

func Save(path string, content []byte) error {
	return os.WriteFile(path, content, 0o644)
}

func Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}
