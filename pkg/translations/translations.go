package translations

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type TranslationHelperFunc func(key string, defaultValue string) string

func NullTranslationHelper(_ string, defaultValue string) string {
	return defaultValue
}

// TranslationStore provides thread-safe granular access to translation keys.
// This supports particle (granular) translation operations for importing and
// exporting individual translation entries.
type TranslationStore struct {
	mu   sync.RWMutex
	keys map[string]string
}

// NewTranslationStore creates a new TranslationStore with an empty key map.
func NewTranslationStore() *TranslationStore {
	return &TranslationStore{
		keys: make(map[string]string),
	}
}

// ImportTranslationKey imports a single translation key-value pair into the store.
// The key is normalized to uppercase for consistency.
func (ts *TranslationStore) ImportTranslationKey(key, value string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.keys[strings.ToUpper(key)] = value
}

// ExportTranslationKey exports (retrieves) a single translation value by key.
// Returns the value and a boolean indicating if the key exists.
// The key is normalized to uppercase for lookup.
func (ts *TranslationStore) ExportTranslationKey(key string) (string, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	value, exists := ts.keys[strings.ToUpper(key)]
	return value, exists
}

// ImportTranslationMap imports multiple translation key-value pairs from a map.
// All keys are normalized to uppercase for consistency.
func (ts *TranslationStore) ImportTranslationMap(translations map[string]string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for k, v := range translations {
		ts.keys[strings.ToUpper(k)] = v
	}
}

// ExportTranslationMap exports all translation key-value pairs as a map.
// Returns a copy of the internal map to prevent external modification.
func (ts *TranslationStore) ExportTranslationMap() map[string]string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	result := make(map[string]string, len(ts.keys))
	for k, v := range ts.keys {
		result[k] = v
	}
	return result
}

// ImportFromFile imports translations from a JSON file at the specified path.
// Returns an error if the file cannot be read or parsed.
func (ts *TranslationStore) ImportFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	ts.ImportTranslationMap(translations)
	return nil
}

// ExportToFile exports all translations to a JSON file at the specified path.
// Returns an error if the file cannot be created or written.
func (ts *TranslationStore) ExportToFile(path string) error {
	ts.mu.RLock()
	jsonData, err := json.MarshalIndent(ts.keys, "", "  ")
	ts.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("error marshaling map to JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0o600); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// DeleteTranslationKey removes a single translation key from the store.
// The key is normalized to uppercase before deletion.
func (ts *TranslationStore) DeleteTranslationKey(key string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.keys, strings.ToUpper(key))
}

// Count returns the number of translation keys in the store.
func (ts *TranslationStore) Count() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.keys)
}

func TranslationHelper() (TranslationHelperFunc, func()) {
	var translationKeyMap = map[string]string{}
	v := viper.New()

	// Load from JSON file
	v.SetConfigName("github-mcp-server-config")
	v.SetConfigType("json")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		// ignore error if file not found as it is not required
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Could not read JSON config: %v", err)
		}
	}

	// create a function that takes both a key, and a default value and returns either the default value or an override value
	return func(key string, defaultValue string) string {
			key = strings.ToUpper(key)
			if value, exists := translationKeyMap[key]; exists {
				return value
			}
			// check if the env var exists
			if value, exists := os.LookupEnv("GITHUB_MCP_" + key); exists {
				// TODO I could not get Viper to play ball reading the env var
				translationKeyMap[key] = value
				return value
			}

			v.SetDefault(key, defaultValue)
			translationKeyMap[key] = v.GetString(key)
			return translationKeyMap[key]
		}, func() {
			// dump the translationKeyMap to a json file
			if err := DumpTranslationKeyMap(translationKeyMap); err != nil {
				log.Fatalf("Could not dump translation key map: %v", err)
			}
		}
}

// DumpTranslationKeyMap writes the translation map to a json file called github-mcp-server-config.json
func DumpTranslationKeyMap(translationKeyMap map[string]string) error {
	file, err := os.Create("github-mcp-server-config.json")
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer func() { _ = file.Close() }()

	// marshal the map to json
	jsonData, err := json.MarshalIndent(translationKeyMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling map to JSON: %v", err)
	}

	// write the json data to the file
	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}
