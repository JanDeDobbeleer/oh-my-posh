package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// segmentCatalogAliases maps a registered SegmentType to the doc id it is published
// under, for the two cases where the two disagree.
var segmentCatalogAliases = map[SegmentType]string{
	GOLANG:     "golang",
	COPILOTCLI: "copilot-cli",
}

// segmentCatalogEntry is one row of the generated website/plugins/segments/registry.json.
type segmentCatalogEntry struct {
	Type  SegmentType `json:"type"`
	DocID string      `json:"docId"`
}

// buildSegmentCatalog derives the segment catalog census from the live registry in
// config.Segments. EXIT is skipped: it is a legacy alias for STATUS, and both factories
// construct &segments.Status{} (segment_types.go:426 and :483), so it would otherwise
// produce a duplicate entry indistinguishable from STATUS.
func buildSegmentCatalog(t *testing.T) []segmentCatalogEntry {
	t.Helper()

	entries := make([]segmentCatalogEntry, 0, len(Segments))

	for segmentType, factory := range Segments {
		if segmentType == EXIT {
			continue
		}

		_, panicked := callTemplateRecovered(t, segmentType, factory)
		if panicked {
			continue
		}

		docID := string(segmentType)
		if alias, ok := segmentCatalogAliases[segmentType]; ok {
			docID = alias
		}

		entries = append(entries, segmentCatalogEntry{
			Type:  segmentType,
			DocID: docID,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Type < entries[j].Type
	})

	return entries
}

// callTemplateRecovered calls factory().Template(), recovering from a panic so that one
// bad writer fails the test by name instead of taking down the whole suite with an
// unattributed stack trace.
func callTemplateRecovered(t *testing.T, segmentType SegmentType, factory func() SegmentWriter) (tmpl string, panicked bool) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("segment %q panicked calling Template(): %v", segmentType, r)
			panicked = true
		}
	}()

	tmpl = factory().Template()

	return tmpl, panicked
}

// segmentDocIDs walks website/docs/segments/*/ for *.mdx files, skipping overview.mdx,
// and returns the set of doc ids (filename without extension).
func segmentDocIDs(t *testing.T) map[string]bool {
	t.Helper()

	docsDir := filepath.Join("..", "..", "website", "docs", "segments")

	groups, err := os.ReadDir(docsDir)
	require.NoError(t, err)

	docIDs := make(map[string]bool)

	for _, group := range groups {
		if !group.IsDir() {
			continue
		}

		matches, err := filepath.Glob(filepath.Join(docsDir, group.Name(), "*.mdx"))
		require.NoError(t, err)

		for _, match := range matches {
			name := filepath.Base(match)
			if name == "overview.mdx" {
				continue
			}

			docIDs[strings.TrimSuffix(name, ".mdx")] = true
		}
	}

	return docIDs
}

// TestSegmentCatalogMatchesRegistry derives the segment catalog from the live registry and
// asserts it is both internally consistent (every registered type has a doc, every doc maps
// back to a registered type) and matches the committed website/plugins/segments/registry.json.
//
// On mismatch, this test writes the corrected JSON to that path, so the fix for a failing run
// is `go test ./config/... && git add website/plugins/segments/registry.json`.
func TestSegmentCatalogMatchesRegistry(t *testing.T) {
	entries := buildSegmentCatalog(t)

	docIDs := segmentDocIDs(t)

	registeredDocIDs := make(map[string]SegmentType, len(entries))
	for _, entry := range entries {
		registeredDocIDs[entry.DocID] = entry.Type
	}

	var missingDocs []string

	for _, entry := range entries {
		if docIDs[entry.DocID] {
			continue
		}

		missingDocs = append(missingDocs, fmt.Sprintf("%s (docId %q)", entry.Type, entry.DocID))
	}

	assert.Empty(t, missingDocs, "registered segment(s) with no matching doc file: %s", strings.Join(missingDocs, ", "))

	var orphanDocs []string

	for docID := range docIDs {
		if _, ok := registeredDocIDs[docID]; ok {
			continue
		}

		orphanDocs = append(orphanDocs, docID)
	}

	sort.Strings(orphanDocs)
	assert.Empty(t, orphanDocs, "doc file(s) with no matching registered segment: %s", strings.Join(orphanDocs, ", "))

	actual, err := json.MarshalIndent(entries, "", "  ")
	require.NoError(t, err)
	actual = append(actual, '\n')

	registryPath := filepath.Join("..", "..", "website", "plugins", "segments", "registry.json")

	expected, readErr := os.ReadFile(registryPath)
	if readErr == nil && bytes.Equal(expected, actual) {
		return
	}

	require.NoError(t, os.MkdirAll(filepath.Dir(registryPath), 0o750))
	require.NoError(t, os.WriteFile(registryPath, actual, 0o600))

	if readErr != nil {
		t.Fatalf("%s did not exist; wrote the derived catalog (%d entries) — rerun to verify, then git add", registryPath, len(entries))
	}

	t.Fatalf("%s was stale and has been rewritten with the derived catalog (%d entries) — verify the diff, then git add", registryPath, len(entries))
}
