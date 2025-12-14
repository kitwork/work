// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

func (t *Work) Alias(ctx *Context) error {
	var alias string

	switch t.Kind {
	case KindValue:
		if t.As != "" {
			alias = t.As
		} else {
			alias = t.value()
		}
	default:

	}
	return ctx.alias(alias)
}

func (t *Work) Target(ctx *Context) error {
	var target string

	switch t.Kind {
	case KindValue:
		target = t.value()
	case KindList:

	default:
	}

	return ctx.target(target)
}
