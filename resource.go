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
	"time"
)

// ResourceType describes type of the remote resource: directory or file.
type ResourceType int

const (
	// Unknown resource type. Indicates ad error.
	ResourceTypeUnknown ResourceType = iota
	// Directory retource type.
	ResourceTypeDir
	// File retource type.
	ResourceTypeFile
)

// UnmarshalJSON unmarshales server's JSON response to ResourceType.
func (rt *ResourceType) UnmarshalJSON(data []byte) error {
	s := string(data)

	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return fmt.Errorf("invalid type JSON value %s", s)
	}

	switch s[1 : len(s)-1] {
	case "dir":
		*rt = ResourceTypeDir
	case "file":
		*rt = ResourceTypeFile
	default:
		*rt = ResourceTypeUnknown
	}

	return nil
}

// Resource describes remote resource.
type Resource struct {
	// Type describes type of the resource: directory or file.
	Type ResourceType `json:"type"`
	// Name of the directory or file.
	Name string `json:"name"`
	// Created specifies time of the resource creation.
	Created time.Time `json:"created"`
	// Modified specifies time of the resource modification.
	Modified time.Time `json:"modified"`
	// Path is a full path to the resource.
	Path string `json:"path"`
	// MD5 Hash of the file. Empty for directories.
	Hash string `json:"md5"`
	// Size of the file. Zero for directories.
	Size int `json:"size"`
	// SubList is a directory childrens list.
	SubList *ResourceList `json:"_embedded"`
}

// ResourceList is a server's list of resources response model.
type ResourceList struct {
	// Items contains resources list itself.
	Items []*Resource `json:"items"`
	// Limit is a size of the requested resources page.
	Limit int `json:"limit"`
	// Offset is a page offset.
	Offset int `json:"offset"`
}
