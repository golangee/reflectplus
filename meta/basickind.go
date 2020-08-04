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

// BasicKind describes the kind of basic type.
type BasicKind string

const (
	Invalid BasicKind = "" // type is invalid

	// predeclared types
	Bool          = "bool"
	Int           = "int"
	Int8          = "int8"
	Int16         = "int16"
	Int32         = "int32"
	Int64         = "int64"
	Uint          = "uint"
	Uint8         = "uint8"
	Uint16        = "uint16"
	Uint32        = "uint32"
	Uint64        = "uint64"
	Uintptr       = "uintptr"
	Float32       = "float32"
	Float64       = "float64"
	Complex64     = "complex64"
	Complex128    = "complex128"
	String        = "string"
	UnsafePointer = "unsafe.Pointer"

	// aliases
	Byte = Uint8
	Rune = Int32
)

// String returns the universe name of the basic kind
func (b BasicKind) String() string {
	return string(b)
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
