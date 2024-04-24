// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.rangefunc

package jsontext

import "iter"

// Tokens returns an iterator over the reference tokens in the JSON pointer,
// starting from the first token until the last token (unless stopped early).
// A token is either a JSON object name or an index to a JSON array element
// encoded as a base-10 integer value.
func (p Pointer) Tokens() iter.Seq[string] {
	return func(yield func(string) bool) {
		for len(p) > 0 {
			if !yield(p.nextToken()) {
				return
			}
		}
	}
}
