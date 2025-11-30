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
				return errors.New("host not exits auto SSL")
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
