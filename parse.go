// KitWork - Work Engine Core
// Copyright (C) 2025 Huỳnh Nhân Quốc

// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License version 3 (AGPL-3.0).

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the AGPL-3.0 License along with this program.
// If not, see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"errors"
	"fmt"
)

type Parse struct {
	Value interface{} `work:"value,ignore"`
	To    string      `work:"to,required,fallback"`
}

func (t *Work) Parse(ctx *Context) error {
	if t.Type != TypeParser {
		return errors.New("type is not parse")
	}

	cfg := Parse{Value: ctx.Result} // mặc định lấy dữ liệu từ ctx.Data

	if err := t.classify(ctx, &cfg); err != nil {
		return err
	}
	fmt.Print(cfg)
	return cfg.Handle(ctx)
}

func (p *Parse) Handle(ctx *Context) error {
	if p.Value == nil {
		return errors.New("parse: no input data in ctx.Result")
	}

	switch p.To {
	case "json":

	case "string":

	case "int":

	default:

	}

	return nil
}
