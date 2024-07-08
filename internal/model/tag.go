// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"database/sql/driver"
	"slices"
)

type Tag struct {
	Val string
}

func (t Tag) Equal(intag Tag) bool {
	if t.Val == intag.Val {
		return true
	}
	return false
}

var RandomizedMacAddressTag = Tag{Val: "RandomizedMACAddress"}

func Add(tag Tag, tags []Tag) []Tag {
	for _, x := range tags {
		if x.Equal(tag) {
			return tags
		}
	}
	return append(tags, tag)
}

func Remove(tag Tag, tags []Tag) []Tag {
	rmidx := -1
	for i, x := range tags {
		if x.Equal(tag) {
			rmidx = i
		}
	}
	if rmidx == -1 {
		return tags
	}
	return slices.Delete(tags, rmidx, rmidx+1)
}

func (t Tag) Value() (driver.Value, error) {
	return t.Val, nil
}

func (t *Tag) Scan(src interface{}) error {
	// TODO: fixme
	// switch src := src.(type) {
	// case string:
	// 	t.Val = src
	// }
	return nil
}
