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
package cluster

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// KeyStore is a store for public keys.
type KeyStore struct {
	basePath string
}

// NewKeyStore creates a new KeyStore
func NewKeyStore(basePath string) *KeyStore {
	return &KeyStore{
		basePath: basePath,
	}
}

// Init initializes the key store, creating the store directory if needed.
func (s *KeyStore) Init() error {
	return os.MkdirAll(s.basePath, 0760)
}

func fileExists(path string) bool {
	// XXX: There's a subtle bug: if stat fails for another reason that the file
	// not existing, we return the file exists.
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (s *KeyStore) keyPath(name string) string {
	return filepath.Join(s.basePath, name)
}

func (s *KeyStore) keyExists(name string) bool {
	return fileExists(s.keyPath(name))
}

// Store adds the key to the store.
func (s *KeyStore) Store(name, key string) error {
	if s.keyExists(name) {
		return errors.Errorf("key store: store: key '%s' already exists", name)
	}

	if err := os.WriteFile(s.keyPath(name), []byte(key), 0644); err != nil {
		return errors.Wrap(err, "key store: write")
	}

	return nil
}

// Get retrieves a key from the store.
func (s *KeyStore) Get(name string) ([]byte, error) {
	if !s.keyExists(name) {
		return nil, errors.Errorf("key store: get: unknown key '%s'", name)
	}
	return os.ReadFile(s.keyPath(name))
}

// Remove removes a key from the store.
func (s *KeyStore) Remove(name string) error {
	if !s.keyExists(name) {
		return errors.Errorf("key store: remove: unknown key '%s'", name)
	}
	if err := os.Remove(s.keyPath(name)); err != nil {
		return errors.Wrap(err, "key store: remove")
	}
	return nil
}
