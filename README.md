# Mutable
[![Go Report Card](https://goreportcard.com/badge/github.com/askretov/go-mutable)](https://goreportcard.com/report/github.com/askretov/go-mutable)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/1c52c7899e544969b1d83896dbc2b9c4)](https://www.codacy.com/app/askretov/go-mutable?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=askretov/mutable&amp;utm_campaign=Badge_Grade)
[![codecov](https://codecov.io/gh/askretov/mutable/branch/master/graph/badge.svg)](https://codecov.io/gh/askretov/mutable)
[![Build Status](https://travis-ci.org/askretov/go-mutable.svg?branch=master)](https://travis-ci.org/askretov/go-mutable)
[![GoDoc](https://godoc.org/github.com/askretov/mutable?status.svg)](https://godoc.org/github.com/askretov/mutable)
[![Licenses](https://img.shields.io/badge/license-mit-brightgreen.svg)](https://opensource.org/licenses/BSD-3-Clause)

## Introduction
Mutable package provides object changes tracking features and the way to set values to the struct dynamically by a destination field name (including nested structs).\
This package needs Go version 1.9 or later

## Usage
### Installation
```go
go get github.com/askretov/go-mutable
```
### Embedding Mutable
```go
package main

import "github.com/askretov/mutable"

type NestedStruct struct {
	mutable.Mutable
	FieldY string
	FieldZ []int64
}

type MyStruct struct {
	mutable.Mutable
	FieldA string
	FieldB int64        `mutable:"ignored"`
	FieldC NestedStruct `mutable:"deep"`
}

func main() {
    var m = &MyStruct{}
    // Mutable state init
    m.ResetMutableState(m)
}
```
### Tracking changes
```go
// Change values
m.FieldA = "green"
m.FieldC.FieldY = "stone"
// Analyze changes
fmt.Println(m.AnalyzeChanges())
```
*Output:*
```json
{
        "FieldA": {
                "old_value": "",
                "new_value": "green"
        },
        "FieldC": {
                "nested_fields": {
                        "FieldY": {
                                "old_value": "",
                                "new_value": "stone"
                        }
                }
        }
}
```
### Set values dynamically
```go
// Set values
m.SetValue("FieldA", "white")
m.SetValue("FieldC/FieldZ", "[1,2,3]") // You can set typed value or JSON string as well
// Analyze changes
fmt.Println(m.AnalyzeChanges())
```
*Output:*
```json
{
        "FieldA": {
                "old_value": "green",
                "new_value": "white"
        },
        "FieldC": {
                "nested_fields": {
                        "FieldZ": {
                                "old_value": null,
                                "new_value": [1, 2, 3]
                        }
                }
        }
}
```
### Optional settings - Struct tags
Struct field's tag values should be set within **mutable** tag
-   ***ignored*** - specifies ignoring of this field changes tracking
-   ***deep*** - specifies the deep analyze of a field (only for struct kind fields). Instead of regular analysis of a field value itself, every field of nested struct will be analyzed for changes individually.

Example:
```go
type MyStruct struct {
    mutable.Mutable
    FieldA string `mutable:"ignored"`
    FieldB AnotherStructType `mutable:"deep"`
}
```

### Keep in mind
1.  If you use a pointer to struct as field type and want to be able to use **deep** analysis, you have to embed Mutable for such nested field's struct as well.

    Example:
    ```go
    type MyStruct struct {
        mutable.Mutable
        FieldA string
        FieldB *NestedStruct `mutable:"deep"`
    }

    type NestedStruct struct {
        mutable.Mutable // Mutable as well
        FieldY string
        FieldZ string
    }
    ```
2. If you pass a mutable object as an arg to a function, you have to pass it as a pointer to be able to use Mutable features.