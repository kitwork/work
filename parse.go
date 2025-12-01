// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

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

	cfg := Parse{} // mặc định lấy dữ liệu từ ctx.Data

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
