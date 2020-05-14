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

package parser

import (
	"reflect"
	"testing"
)

func TestParseAnnotation(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    Annotation
		wantErr bool
	}{

		{"valid-0", `@a.b.c("Text":"hello", "Num":5, "Float":3.4, "Enabled":false)//hello`,
			Annotation{
				Name:   "a.b.c",
				Values: map[string]interface{}{"Text": "hello", "Num": float64(5), "Float": 3.4, "Enabled": false},
			}, false,
		},

		{"valid-1", `@a.b.c({"Text":"hello", "Num":5, "Float":3.4, "Enabled":false}) // ignored braces in comment ) "`,
			Annotation{
				Name:   "a.b.c",
				Values: map[string]interface{}{"Text": "hello", "Num": float64(5), "Float": 3.4, "Enabled": false},
			}, false,
		},

		{"valid-2", `@a("hello")// this is a shortcut for {"value":"hello"}`,
			Annotation{
				Name:   "a",
				Values: map[string]interface{}{"value": "hello"},
			}, false,
		},
		{"valid-3", `@a()// this is a shortcut for {}`,
			Annotation{
				Name:   "a",
				Values: map[string]interface{}{},
			}, false,
		},
		{"valid-4", `@a// this is a shortcut for {}`,
			Annotation{
				Name:   "a",
				Values: map[string]interface{}{},
			}, false,
		},
		{"valid-5", `@a("a":"b")`,
			Annotation{
				Name:   "a",
				Values: map[string]interface{}{"a": "b"},
			}, false,
		},
		{"valid-6", `@a(5)`,
			Annotation{
				Name:   "a",
				Values: map[string]interface{}{"value": float64(5)},
			}, false,
		},
		/* DeepEqual fails
		{"valid-7", `@a("anyKey":"anyValue","num":5,"bool":true,"nested":{"care":{"of":["your", "head"]}})`,
			Annotation{
				Name:   "a",
				Values: map[string]interface{}{"anyKey": "anyValue", "num": float64(5), "bool": true, "nested": map[string]interface{}{"care": map[string]interface{}{"of": []string{"your", "head"}}}},
			}, false,
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAnnotation(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAnnotation() error = %v, wantErr %v => %+v", err, tt.wantErr, got)
				return
			}
			if !reflect.DeepEqual(got.Values, tt.want.Values) {
				t.Errorf("expected %v but got %v", tt.want.Values, got.Values)
			}
		})
	}
}

func TestMultiline(t *testing.T) {
	textBlock := `
stuff
 @Repo
 @Repo()
 @Repo({}) // comments allowed, outer {} can be omitted 
 @Repo({"value":5})
 @Repo(5) // implicitly wrapped into {"value": 5}
 @Repo("text") // implicitly wrapped into {"value": "text"}
 @Repo("value":"te:xt") // this is fine 
 @Repo("values":["can","be","multiple"])
 @Repo("anyKey":"anyValue","num":5,"bool":true,"nested":{"care":{"of":["your", "head"]}})
 @Repo(   """
    this is 
    a multiline string 
    or json literal.
    However line breaks and additional start/ending spaces are discarded and replaced by 
    single spaces.
 """
)
otherstuff

   @ee.sql.Schema("""
   "dialect":"mysql", "version":1, "group":"some_name", "value":
   "CREATE TABLE IF NOT EXISTS 'some_table_name'
   (
    	'group'              VARCHAR(255) NOT NULL,
		'version'            BIGINT       NOT NULL,
	'script'             VARCHAR(255) NOT NULL,
   	'type'               VARCHAR(12)  NOT NULL,
	'checksum'           CHAR(64)     NOT NULL,
	'applied_at'         TIMESTAMP    NOT NULL,
	'execution_duration' BIGINT       NOT NULL,
	'status'             VARCHAR(12)  NOT NULL,
	'log'                TEXT         NOT NULL,
    	PRIMARY KEY ('group', 'version')
	 )"
   """)

`

	annotations, err := ParseAnnotations(textBlock)
	if err != nil {
		t.Fatal(err)
	}

	if len(annotations) != 11 {
		t.Fatal(annotations)
	}

	if annotations[9].Values["value"] != "this is a multiline string or json literal. However line breaks and additional start/ending spaces are discarded and replaced by single spaces." {
		t.Fatal(annotations[9])
	}
}
