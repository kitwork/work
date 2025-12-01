// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

type Kind string

const (
	KindFull      Kind = "full"
	KindShort     Kind = "short"
	KindValue     Kind = "value"
	KindList      Kind = "list"
	KindPrimitive Kind = "primitive"
	KindUnknown   Kind = "unknown"
)
