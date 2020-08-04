// Copyright 2020 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package meta

import "github.com/golangee/reflectplus/internal/tag"

// A ChanDir specified the declared channel direction
type ChanDir string

const (
	SendRecv ChanDir = "SendRecv"
	SendOnly = "SendOnly"
	RecvOnly = "RecvOnly"
)

// An Type is a union tuple of exact one of Basic, Array, Channel, Interface, Map
// Pointer or Struct.
type Type struct {
	Basic     *Basic     `json:",omitempty"`
	Array     *Array     `json:",omitempty"`
	Channel   *Channel   `json:",omitempty"`
	Interface *Interface `json:",omitempty"`
	Map       *Map       `json:",omitempty"`
	Pointer   *Pointer   `json:",omitempty"`
	Slice     *Slice     `json:",omitempty"`
	Struct    *Struct    `json:",omitempty"`
	Named     *Named     `json:",omitempty"`
	Signature *Signature `json:",omitempty"`
}

// Kind returns the first non-nil union value.
func (t Type) Kind() interface{} {
	if t.Basic != nil {
		return t.Basic
	}

	if t.Array != nil {
		return t.Array
	}

	if t.Slice != nil {
		return t.Slice
	}

	if t.Channel != nil {
		return t.Channel
	}

	if t.Interface != nil {
		return t.Interface
	}

	if t.Map != nil {
		return t.Map
	}

	if t.Pointer != nil {
		return t.Pointer
	}

	if t.Struct != nil {
		return t.Struct
	}

	if t.Named != nil {
		return t.Named
	}

	panic("invalid type model")
}

// A Named type is a declared type somewhere in the source. It is not a build-in, however it may be
// also an anonymous type, where the name is just empty.
type Named struct {
	Location    Location
	Doc         string
	Annotations []Annotation `json:",omitempty"`

	// Name is the LHS of the declaration or empty if no such thing
	Name string

	// Underlying is the resolved RHS side of the declaration.
	Underlying DeclId

	// Methods contains the declared methods for this named type (Signature).
	Methods []DeclId `json:",omitempty"`
}

// A Basic type represents a build-in type
type Basic struct {
	Kind BasicKind
}

// Arrays has a length and an according type. Generic declarations (like map[x]y) refer to
// their own type.
type Array struct {
	Len    int64
	DeclId DeclId
}

// A Channel declares the direction and its type.
type Channel struct {
	ChanDir ChanDir
	DeclId  DeclId
}

// An Interface has a set of signatures and embedded types.
type Interface struct {
	// Embeddeds refers only to TypeIds of other embedded Interfaces.
	Embeddeds []DeclId

	// AllMethods refers only to TypeIds of Signatures included by all declared methods, also
	// by embedded ones.
	AllMethods []DeclId
}

// A Map is a generic build-in with two parameters.
type Map struct {
	Key   DeclId
	Value DeclId
}

// A Pointer to a base type (which itself may be again a pointer).
type Pointer struct {
	Base DeclId
}

// Signature represents a declared function or method.
type Signature struct {
	// Receiver is an optional parameter.
	Receiver *Param `json:",omitempty"`

	// Params represent the incoming parameters.
	Params []Param `json:",omitempty"`

	// Results represent the outgoing parameters.
	Results []Param `json:",omitempty"`

	// Variadic indicates if the last parameter is a ...T declaration.
	Variadic bool `json:",omitempty"`
}

// A Param is not a type but declares a tuple of name and type.
type Param struct {
	Pos *Location `json:",omitempty"`

	Doc         string       `json:",omitempty"`
	Annotations []Annotation `json:",omitempty"`

	// The Name of the parameter, if not empty
	Name string

	// The type of the parameter
	DeclId DeclId

	// Tag is the raw string literal. The parsed literal is in Tags
	Tag string `json:",omitempty"`

	// Tags are the parsed form of the Tag literal
	Tags tag.Tags `json:",omitempty"`
}

// A Slice wraps a subset of an array of variable length
type Slice struct {
	// Underlying of the sliced type.
	DeclId DeclId
}

// A Struct contains field definitions. Interestingly the method set does not belong to
// the underlying type, which makes them assignable to each other. However this is also
// true for tags, so it is quite inconsistent.
type Struct struct {
	Fields []Param `json:",omitempty"`
}
