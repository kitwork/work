// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"fmt"
)

type Error error

func (t *Work) SetError(ctx *Context) {

	switch t.Kind {
	case KindValue:

		val, err := ctx.evaluator(t.value())
		if err != nil {
			return
		}

		ctx.Error = fmt.Errorf("%s", val)

	default:

	}

	return
}
