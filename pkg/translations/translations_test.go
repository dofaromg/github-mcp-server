package translations

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTranslationStore(t *testing.T) {
	store := NewTranslationStore()
	assert.NotNil(t, store)
	assert.Equal(t, 0, store.Count())
}

func TestImportExportTranslationKey(t *testing.T) {
	store := NewTranslationStore()

	// Test importing a key
	store.ImportTranslationKey("test_key", "test_value")

	// Test exporting the key
	value, exists := store.ExportTranslationKey("test_key")
	assert.True(t, exists)
	assert.Equal(t, "test_value", value)

	// Test key normalization (uppercase)
	value, exists = store.ExportTranslationKey("TEST_KEY")
	assert.True(t, exists)
	assert.Equal(t, "test_value", value)

	// Test non-existent key
	_, exists = store.ExportTranslationKey("non_existent")
	assert.False(t, exists)
}

func TestImportExportTranslationMap(t *testing.T) {
	store := NewTranslationStore()

	// Import multiple keys
	translations := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	store.ImportTranslationMap(translations)

	// Verify count
	assert.Equal(t, 3, store.Count())

	// Export and verify
	exported := store.ExportTranslationMap()
	assert.Equal(t, 3, len(exported))
	assert.Equal(t, "value1", exported["KEY1"])
	assert.Equal(t, "value2", exported["KEY2"])
	assert.Equal(t, "value3", exported["KEY3"])

	// Verify exported map is a copy (modifying it shouldn't affect store)
	exported["NEW_KEY"] = "new_value"
	assert.Equal(t, 3, store.Count())
}

func TestDeleteTranslationKey(t *testing.T) {
	store := NewTranslationStore()

	store.ImportTranslationKey("key1", "value1")
	store.ImportTranslationKey("key2", "value2")
	assert.Equal(t, 2, store.Count())

	// Delete a key
	store.DeleteTranslationKey("key1")
	assert.Equal(t, 1, store.Count())

	_, exists := store.ExportTranslationKey("key1")
	assert.False(t, exists)

	// Verify other key still exists
	value, exists := store.ExportTranslationKey("key2")
	assert.True(t, exists)
	assert.Equal(t, "value2", value)
}

func TestImportExportFile(t *testing.T) {
	store := NewTranslationStore()

	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "translations.json")

	// Import some keys
	store.ImportTranslationKey("file_key1", "file_value1")
	store.ImportTranslationKey("file_key2", "file_value2")

	// Export to file
	err := store.ExportToFile(testFile)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(testFile)
	require.NoError(t, err)

	// Create a new store and import from file
	newStore := NewTranslationStore()
	err = newStore.ImportFromFile(testFile)
	require.NoError(t, err)

	// Verify imported data
	assert.Equal(t, 2, newStore.Count())
	value, exists := newStore.ExportTranslationKey("file_key1")
	assert.True(t, exists)
	assert.Equal(t, "file_value1", value)
}

func TestImportFromFileErrors(t *testing.T) {
	store := NewTranslationStore()

	// Test non-existent file
	err := store.ImportFromFile("/non/existent/file.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading file")

	// Test invalid JSON file
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("not valid json"), 0o600)
	require.NoError(t, err)

	err = store.ImportFromFile(invalidFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing JSON")
}

func TestConcurrentAccess(t *testing.T) {
	store := NewTranslationStore()

	// Test concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := "concurrent_key"
			store.ImportTranslationKey(key, "value")
			_, _ = store.ExportTranslationKey(key)
		}()
	}
	wg.Wait()

	// Verify no data race occurred
	assert.Equal(t, 1, store.Count())
}

func TestKeyNormalization(t *testing.T) {
	store := NewTranslationStore()

	// Import with lowercase
	store.ImportTranslationKey("lowercase_key", "value1")

	// Export with uppercase should work
	value, exists := store.ExportTranslationKey("LOWERCASE_KEY")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Import with mixed case
	store.ImportTranslationKey("MixedCase_Key", "value2")

	// Export with any case should work
	value, exists = store.ExportTranslationKey("mixedcase_key")
	assert.True(t, exists)
	assert.Equal(t, "value2", value)
}
