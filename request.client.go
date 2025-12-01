// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"fmt"
	"net/http"
)

func (r *Request) Client(ctx *Context) error {
	fmt.Printf("→ [http] %s %s\n", r.Method, r.URL)

	req, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return err
	}

	for k, v := range r.Header {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("→ Response status:", resp.Status)
	return nil
}
