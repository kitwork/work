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
	"fmt"
	"time"
)

// Work / work đại diện cho 1 work/workflow node
type Work struct {
	Name    string
	As      string
	Type    Type
	Kind    Kind // short, full, value, list, switch
	Config  map[string]interface{}
	Works   []*Work // chain mặc định
	Success []*Work // nếu OK
	Error   []*Work // nếu lỗi
	Timeout time.Duration
}

// ========================
//  WORK RUNNER
// ========================

func (t *Work) Run(ctx *Context, skipTypes ...string) (err error) {

	fmt.Printf("\n→ Running Work: [%s] %s\n", t.Type, t.Name)

	// 1. chạy work chính
	switch t.Type {
	case TypeScript:
		err = t.Script(ctx)

	case TypeFetch, TypeHTTP, TypeClient, TypeRequest:
		err = t.Request(ctx)

	case TypeCron:
		// return nil

	case TypeLog:
		err = t.Log(ctx)

	case TypeParser:

		err = t.Parse(ctx)

	case TypeForeach:
		fmt.Println("→ foreach chưa implement")

	default:
		err = fmt.Errorf("unknown work type: %s", t.Type)
	}

	// 2. Nếu work chính OK → chạy chuỗi Works
	if err == nil && len(t.Works) > 0 {
		// fmt.Println("→ run works chain...")
		for _, a := range t.Works {
			if err = a.Run(ctx); err != nil {
				break
			}
		}
	}

	// 3. Nếu lỗi → chạy Error branch
	if err != nil && len(t.Error) > 0 {
		fmt.Printf("→ Error in [%s]: %v. Running error branch...\n", t.Name, err)
		for _, e := range t.Error {
			e.Run(ctx)
		}
	}
	// 4. Nếu thành công → chạy Success branch
	if err == nil && len(t.Success) > 0 {
		// fmt.Println("→ Success → chạy success branch")
		for _, s := range t.Success {
			s.Run(ctx)
		}
	}

	return
}
