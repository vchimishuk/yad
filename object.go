// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// Yad is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Yad is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Yad. If not, see <http://www.gnu.org/licenses/>.

package yad

import (
	"fmt"
	"net/url"
	"strings"
)

// Status of operation type.
type Status int

const (
	StatusFailure = iota
	StatusInProgress
	StatusSuccess
)

// Error is a Yandex.Disk server error response object.
type Error struct {
	StatusCode  int
	Message     string `json:"message"`
	Description string `json:"description"`
	Err         string `json:"error"`
}

// Error returns string representation of Error object.
func (err *Error) Error() string {
	return fmt.Sprintf("%d: %s", err.StatusCode, err.Description)
}

// Link is a Yandex.Disk Link response object.
type Link struct {
	Href      string `json:"href"`
	Method    string `json:"method"`
	Templated bool   `json:"templated"`
}

// IsOperation returns true if current Link represents operation object,
// which can be checked for its status. Otherwise Link represents remote
// object, t.g. newly copied file.
func (l *Link) IsOperation() bool {
	return strings.Contains(l.Href, "/disk/operations")
}

// Operation returns operation ID.
// Returns empty string if Link doesn't represent operation.
func (l *Link) Operation() string {
	if l.IsOperation() {
		u, err := url.Parse(l.Href)
		if err == nil {
			parts := strings.Split(u.RequestURI(), "/")

			if len(parts) > 0 {
				return parts[len(parts)-1]
			} else {
				return ""
			}
		}
	}

	return ""
}

// Stats information about Yandex.Disk account.
type Stats struct {
	// Total disk space in bytes.
	Total int64 `json:"total_space"`
	// Used space in bytes.
	Used int64 `json:"used_space"`
	// Trash occupied space in bytes.
	Trash int64 `json:"trash_size"`
	// System folders list.
	SystemFolders map[string]string `json:"system_folders"`
}
