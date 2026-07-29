package cli

import (
	"encoding/json"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func recordedEnvelope(t *testing.T, enabled bool, data any) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(data)
	require.NoError(t, err)

	envelope, err := json.Marshal(config.RecordedSegment{Data: raw, Enabled: enabled})
	require.NoError(t, err)

	return envelope
}

func TestSanitizeDataDocument_ScrubsIdentityFields(t *testing.T) {
	cfg := &config.Config{
		Blocks: []*config.Block{
			{
				Segments: []*config.Segment{
					{Type: config.GIT},
					{Type: config.AWS},
					{Type: config.AZ},
					{Type: config.PATH},
					{Type: config.SYSTEMINFO},
					{Type: config.BATTERY},
					{Type: config.TEXT, Alias: "greeting"},
					{Type: config.AWS, Alias: "aws-disabled"},
				},
			},
		},
	}

	env := map[string]json.RawMessage{
		"UserName":    mustJSON(t, "jande"),
		"HostName":    mustJSON(t, "surface-pro"),
		"PWD":         mustJSON(t, "/home/jande/dev/oh-my-posh"),
		"PSWD":        mustJSON(t, "/c/Users/Jande/dev/oh-my-posh"),
		"Folder":      mustJSON(t, "oh-my-posh"),
		"AbsolutePWD": mustJSON(t, `C:\Users\Jande\dev\oh-my-posh`),
		"OS":          mustJSON(t, "windows"),
	}

	segments := map[string]json.RawMessage{
		"git": recordedEnvelope(t, true, map[string]any{
			"Segment": map[string]any{"Text": " main "},
			"User":    map[string]any{"Name": "Jan De Dobbeleer", "Email": "jan@example.com"},
			"Dir":     `C:\Users\Jande\dev\oh-my-posh`,
			// A real git remote embedding the GitHub username must survive
			// untouched: this is exactly what a naive string replace would
			// corrupt.
			"UpstreamURL": "https://github.com/JanDeDobbeleer/oh-my-posh",
		}),
		"aws": recordedEnvelope(t, true, map[string]any{
			"Segment":     map[string]any{"Text": " prod "},
			"AccessKeyID": "AKIAABCDEFGHIJKLMNOP",
		}),
		// Shape taken from an actual unsanitized recording of the az segment
		// (AzureSubscription, segments/az.go), not from re-reading the struct
		// definition: that is exactly how the id/tenantId/homeTenantId/user
		// keys below were previously gotten wrong (scrubbed as "ID"/"TenantID"/
		// "HomeTenantID"/"User", which never matched anything and leaked a
		// real email into 8 committed fixtures via user.name).
		"az": recordedEnvelope(t, true, map[string]any{
			"Segment":           map[string]any{"Text": " sub "},
			"Origin":            "CLI",
			"user":              map[string]any{"name": "jan.example@gmail.com", "type": "user"},
			"id":                "11111111-2222-3333-4444-555555555555",
			"name":              "MVP",
			"state":             "Enabled",
			"tenantId":          "66666666-7777-8888-9999-aaaaaaaaaaaa",
			"tenantDisplayName": "",
			"environmentName":   "AzureCloud",
			"homeTenantId":      "bbbbbbbb-cccc-dddd-eeee-ffffffffffff",
			"managedByTenants":  []any{},
			"isDefault":         true,
		}),
		"path": recordedEnvelope(t, true, map[string]any{
			"Segment": map[string]any{"Text": " oh-my-posh "},
			"Path":    `C:\Users\Jande\dev\oh-my-posh`,
			"Folders": []any{
				map[string]any{"Name": "Jande", "Path": `C:\Users\Jande`},
				map[string]any{"Name": "dev", "Path": `C:\Users\Jande\dev`},
			},
		}),
		"sysinfo": recordedEnvelope(t, true, map[string]any{
			"Segment":                 map[string]any{"Text": " 42 "},
			"PhysicalTotalMemory":     float64(17179869184),
			"PhysicalAvailableMemory": float64(1234567),
			"PhysicalPercentUsed":     93.7,
			"Load1":                   3.14,
			// Disks (runtime.SystemInfo) is a map keyed by drive/device name
			// whose entries carry their own serialNumber/label - real hardware
			// identifiers on a machine where the OS reports them, found while
			// auditing every field against an actual recording rather than
			// trusting the struct definition.
			"Disks": map[string]any{
				"C:": map[string]any{"serialNumber": "WD-REAL1234567", "label": "Windows"},
			},
		}),
		"battery": recordedEnvelope(t, true, map[string]any{
			"Segment":    map[string]any{"Text": " 88 "},
			"Percentage": float64(12),
			"State":      float64(4),
			"Icon":       "",
		}),
		"greeting": recordedEnvelope(t, true, map[string]any{
			"Segment": map[string]any{"Text": " jande@surface-pro "},
		}),
		// A recorded-but-disabled segment still carries its last-known writer
		// state (config/segment.go's restoreData unmarshals it before checking
		// Enabled), so it must be scrubbed identically to an enabled one - and
		// its Enabled flag must survive the round-trip unchanged.
		"aws-disabled": recordedEnvelope(t, false, map[string]any{
			"Segment":     map[string]any{"Text": " stale "},
			"AccessKeyID": "AKIAZZZZZZZZZZZZZZZZ",
		}),
	}

	doc := map[string]any{
		"version":  config.DataVersion,
		"env":      env,
		"segments": segments,
	}

	raw, err := json.MarshalIndent(doc, "", "  ")
	require.NoError(t, err)

	sanitized, err := sanitizeDataDocument(raw, cfg)
	require.NoError(t, err)

	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(sanitized, &root))

	var gotEnv map[string]string
	require.NoError(t, json.Unmarshal(root["env"], &gotEnv))

	assert.Equal(t, "alice", gotEnv["UserName"])
	assert.Equal(t, "contoso-devbox", gotEnv["HostName"])
	assert.Equal(t, "~/dev/oh-my-posh", gotEnv["PWD"])
	assert.Equal(t, "~/dev/oh-my-posh", gotEnv["PSWD"])
	assert.Equal(t, "~/dev/oh-my-posh", gotEnv["AbsolutePWD"])
	assert.Equal(t, "windows", gotEnv["OS"], "non-identity keys must survive untouched")

	var gotSegments map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(root["segments"], &gotSegments))

	var git map[string]any
	require.True(t, unwrapRecorded(t, gotSegments["git"], &git))
	assert.Equal(t, "", git["Segment"].(map[string]any)["Text"])
	user := git["User"].(map[string]any)
	assert.Equal(t, "Alice Example", user["Name"])
	assert.Equal(t, "alice@contoso.com", user["Email"])
	assert.Equal(t, "~/dev/oh-my-posh", git["Dir"])
	assert.Equal(t, "https://github.com/JanDeDobbeleer/oh-my-posh", git["UpstreamURL"],
		"a real git remote URL is project data, not personal identity, and must not be corrupted")

	var aws map[string]any
	require.True(t, unwrapRecorded(t, gotSegments["aws"], &aws))
	assert.Equal(t, "", aws["AccessKeyID"])

	var az map[string]any
	require.True(t, unwrapRecorded(t, gotSegments["az"], &az))
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", az["id"])
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", az["tenantId"])
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", az["homeTenantId"])
	assert.Equal(t, "alice@contoso.com", az["user"].(map[string]any)["name"])
	assert.Equal(t, "MVP", az["name"], "the subscription display name is not personal identity and must survive untouched")

	var path map[string]any
	require.True(t, unwrapRecorded(t, gotSegments["path"], &path))
	assert.Equal(t, "~/dev/oh-my-posh", path["Path"])
	for _, entry := range path["Folders"].([]any) {
		folder := entry.(map[string]any)
		assert.Equal(t, "~/dev/oh-my-posh", folder["Path"])
	}

	var sysinfo map[string]any
	require.True(t, unwrapRecorded(t, gotSegments["sysinfo"], &sysinfo))
	assert.InDelta(t, 34359738368, sysinfo["PhysicalTotalMemory"], 0)
	assert.InDelta(t, 12884901888, sysinfo["PhysicalAvailableMemory"], 0)
	assert.InDelta(t, 62.5, sysinfo["PhysicalPercentUsed"], 0.001)
	assert.InDelta(t, 0.52, sysinfo["Load1"], 0.001)
	assert.Empty(t, sysinfo["Disks"], "per-disk serial numbers must be dropped, not merely overwritten field by field")

	var battery map[string]any
	require.True(t, unwrapRecorded(t, gotSegments["battery"], &battery))
	assert.InDelta(t, 70, battery["Percentage"], 0)
	assert.InDelta(t, 3, battery["State"], 0)

	var greeting map[string]any
	require.True(t, unwrapRecorded(t, gotSegments["greeting"], &greeting))
	assert.Equal(t, "", greeting["Segment"].(map[string]any)["Text"])

	var awsDisabled map[string]any
	enabled := unwrapRecorded(t, gotSegments["aws-disabled"], &awsDisabled)
	assert.False(t, enabled, "the disabled flag itself must survive sanitization untouched")
	assert.Equal(t, "", awsDisabled["AccessKeyID"], "a disabled segment's stale writer state must still be scrubbed")
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(v)
	require.NoError(t, err)

	return raw
}
