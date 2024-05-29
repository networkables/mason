// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tags

import (
	"slices"
)

type Tag struct {
	Value string
}

func (t Tag) Equal(intag Tag) bool {
	if t.Value == intag.Value {
		return true
	}
	return false
}

var RandomizedMacAddressTag = Tag{Value: "RandomizedMACAddress"}

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
