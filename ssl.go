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

func SSL() *tls.Config { //Tự Động tạo HTTPS -  Auto create SSL

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy, // policy In Folder
		Cache:      autocert.DirCache("./certs"),
	}

	return &tls.Config{
		GetCertificate: m.GetCertificate,
		NextProtos:     []string{"http/1.1", "acme-tls/1"},
		// MinVersion: tls.VersionTLS12,
	}
}

func hostPolicy(ctx context.Context, host string) error {
	src := NewSource("ssl")
	if !src.HasFile(host) {
		return errors.New("host not exits ssl")
	}
	return nil // Nil là sẽ đăng kí SSL cho tên miền này
}
