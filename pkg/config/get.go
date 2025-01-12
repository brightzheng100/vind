/*
Copyright © 2019-2023 footloose developers
Copyright © 2024-2025 Bright Zheng <bright.zheng@outlook.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func pathSplit(r rune) bool {
	return r == '.' || r == '[' || r == ']' || r == '"'
}

// GetValueFromConfig returns specific value from object given a string path
func GetValueFromConfig(stringPath string, object interface{}) (interface{}, error) {
	keyPath := strings.FieldsFunc(stringPath, pathSplit)
	v := reflect.ValueOf(object)
	for _, key := range keyPath {
		keyUpper := strings.Title(key)
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			v = v.FieldByName(keyUpper)
			if !v.IsValid() {
				return nil, fmt.Errorf("%v key does not exist", keyUpper)
			}
		} else if v.Kind() == reflect.Slice {
			index, errConv := strconv.Atoi(keyUpper)
			if errConv != nil {
				return nil, fmt.Errorf("%v is not an index", key)
			}
			v = v.Index(index)
		} else {
			return nil, fmt.Errorf("%v is neither a slice or a struct", v)
		}
	}
	return v.Interface(), nil
}
