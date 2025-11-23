package work

import (
	"errors"
	"fmt"
	"time"
)

// Schedule định nghĩa một job với nhiều style khác nhau
type Cron struct {
	// Cron expression kiểu chuẩn Unix
	Cron string `work:"cron,omitempty"`

	// Daily: chạy hàng ngày lúc HH:MM
	Daily string `work:"daily,omitempty"`

	// Every: chạy theo duration như "10s", "5m", "2h"
	Every string `work:"every,omitempty"`

	// Weekly: chạy hằng tuần vào ngày cụ thể và giờ HH:MM
	Weekly string `work:"weekly,omitempty"` // ví dụ: "Monday 15:00"

	// Monthly: chạy hàng tháng vào ngày X và giờ HH:MM
	Monthly string `work:"monthly,omitempty"` // ví dụ: "5 10:00" → ngày 5 hàng tháng lúc 10:00

	// Hourly: chạy mỗi giờ, có thể chỉ định phút
	Hourly string `work:"hourly,omitempty"` // ví dụ: "15" → phút thứ 15 mỗi giờ

	// Tag: dùng để gán tag cho job
	Tag string `work:"tag,omitempty"`

	// Limit: giới hạn số lần chạy
	Limit int `work:"limit,omitempty"`

	// StartAt / EndAt: thời gian bắt đầu và kết thúc
	StartAt *time.Time `work:"start_at,omitempty"`
	EndAt   *time.Time `work:"end_at,omitempty"`
}

func (t *Work) Cron(ctx *Context) error {
	if t.Type != TypeCron {
		return errors.New("type is not cron")
	}

	fmt.Println("→ [cron] chạy lịch ...")
	return nil
}
