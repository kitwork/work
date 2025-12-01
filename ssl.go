// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"context"
	"crypto/tls"
	"errors"

	"golang.org/x/crypto/acme/autocert"
)

func SSL(source ...string) *tls.Config { //Tự Động tạo HTTPS -  Auto create SSL
	src := NewSource("ssl", source...)

	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		HostPolicy: func(ctx context.Context, host string) error {
			if !src.HasFile(host) {
				return errors.New("host not exits ssl")
			}
			return nil // Nil là sẽ đăng kí SSL cho tên miền này
		}, // policy In Folder
		Cache: autocert.DirCache("./certs"),
	}

	return &tls.Config{
		GetCertificate: m.GetCertificate,
		NextProtos:     []string{"http/1.1", "acme-tls/1"},
		// MinVersion: tls.VersionTLS12,
	}
}
