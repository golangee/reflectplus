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

import "strconv"

// BasicKind describes the kind of basic type.
type BasicKind int

const (
	Invalid BasicKind = iota // type is invalid

	// predeclared types
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	String
	UnsafePointer

	// aliases
	Byte = Uint8
	Rune = Int32
)

// String returns the universe name of the basic kind
func (b BasicKind) String() string {
	switch b {
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Uintptr:
		return "uintptr"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Complex64:
		return "complex64"
	case Complex128:
		return "complex128"
	case String:
		return "string"
	case UnsafePointer:
		return "unsafe.Pointer"
	default:
		return "unknown-" + strconv.Itoa(int(b))
	}
}

func (b BasicKind) IsFloat() bool {
	switch b {
	case Float32:
		fallthrough
	case Float64:
		return true
	default:
		return false
	}
}

func (b BasicKind) IsString() bool {
	return b == String
}

func (b BasicKind) IsInteger() bool {
	switch b {
	case Int:
		fallthrough
	case Int8:
		fallthrough
	case Int16:
		fallthrough
	case Int32:
		fallthrough
	case Int64:
		fallthrough
	case Uint:
		fallthrough
	case Uint8:
		fallthrough
	case Uint16:
		fallthrough
	case Uint32:
		fallthrough
	case Uint64:
		fallthrough
	case Uintptr:
		return true
	default:
		return false
	}
}

func (b BasicKind) IsUnsigned() bool {
	switch b {
	case Uint:
		fallthrough
	case Uint8:
		fallthrough
	case Uint16:
		fallthrough
	case Uint32:
		fallthrough
	case Uint64:
		fallthrough
	case Uintptr:
		return true
	default:
		return false
	}
}
