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

func (t *Work) Run(ctx *Context) (err error) {

	fmt.Printf("\n→ Running Work: [%s] %s\n", t.Type, t.Name)

	// 1. chạy work chính
	switch t.Type {
	case TypeScript:
		err = t.Script(ctx)

	case TypeFetch, TypeHTTP, TypeClient, TypeRequest:
		err = t.Request(ctx)

	case TypeCron:
		err = t.Cron(ctx)

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
		fmt.Println("→ run works chain...")
		for _, a := range t.Works {
			if err = a.Run(ctx); err != nil {
				break
			}
		}
	}

	// 4. Nếu thành công → chạy Success branch
	if err == nil && len(t.Success) > 0 {
		fmt.Println("→ Success → chạy success branch")
		for _, s := range t.Success {
			s.Run(ctx)
		}
	}

	// 3. Nếu lỗi → chạy Error branch
	if err != nil && len(t.Error) > 0 {
		fmt.Println("→ Error xảy ra → chạy error branch")
		for _, e := range t.Error {
			e.Run(ctx)
		}
	}

	return
}
