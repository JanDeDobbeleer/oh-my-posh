package cli

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
)

const (
	// sanitizedPath replaces every absolute path a fixture could otherwise
	// leak (env PWD/PSWD/AbsolutePWD, git.Dir, path.Path/Location/Folders[].Path)
	// with the same synthetic value website/segment_data.json uses, so a
	// sanitized fixture and the website tell the same story.
	sanitizedPath = "~/dev/oh-my-posh"

	// sanitizedGUID replaces az subscription/tenant IDs, which are otherwise
	// real GUIDs identifying the recording machine's Azure account.
	sanitizedGUID = "00000000-0000-0000-0000-000000000000"
)

// sanitizeEnv overwrites the identity-carrying keys of a recorded env section in
// place, matching the synthetic identity website/segment_data.json already uses
// so a sanitized fixture and the website tell the same story. Structural, not a
// string replace: replacing a username string wholesale would also rewrite it
// inside unrelated recorded values (e.g. a git remote URL that happens to embed
// the same string) while still missing case-sensitive OS-specific spellings
// (C:\Users\Name, %USERNAME%). Overwriting known keys by name has neither
// problem.
func sanitizeEnv(env map[string]json.RawMessage) {
	set := func(key, value string) {
		if _, ok := env[key]; !ok {
			return
		}

		raw, err := json.Marshal(value)
		if err != nil {
			return
		}

		env[key] = raw
	}

	set("UserName", "alice")
	set("HostName", "contoso-devbox")
	set("PWD", sanitizedPath)
	set("PSWD", sanitizedPath)
	set("Folder", "oh-my-posh")
	set("AbsolutePWD", sanitizedPath)
}

// sanitizeSegmentText clears Segment.Text - the fully rendered string every
// writer embeds via segments.Base (e.g. " jande@surface-pro " for a session
// segment) - on any writer shaped generic map. Safe to zero unconditionally:
// Render recomputes it from the segment's template on every prompt, so a
// replayed fixture never reads the stale value back.
func sanitizeSegmentText(data map[string]any) {
	segment, ok := data["Segment"].(map[string]any)
	if !ok {
		return
	}

	if _, ok := segment["Text"]; ok {
		segment["Text"] = ""
	}
}

// sanitizeSegmentData scrubs the identity-carrying fields specific to a few
// segment writers. Keyed on the segment's configured type rather than
// pattern-matching field names in the decoded map, since alias collisions
// (config.Segment.DataKey) mean the map key in the recorded document is not
// reliably the segment type.
func sanitizeSegmentData(segType config.SegmentType, data map[string]any) {
	sanitizeSegmentText(data)

	switch segType {
	case config.GIT:
		if user, ok := data["User"].(map[string]any); ok {
			if _, ok := user["Name"]; ok {
				user["Name"] = "Alice Example"
			}

			if _, ok := user["Email"]; ok {
				user["Email"] = "alice@contoso.com"
			}
		}

		if _, ok := data["Dir"]; ok {
			data["Dir"] = sanitizedPath
		}
	case config.AWS:
		if _, ok := data["AccessKeyID"]; ok {
			data["AccessKeyID"] = ""
		}
	case config.AZ:
		// AzureSubscription (segments/az.go), embedded anonymously in Az,
		// carries explicit lowercase/camelCase json tags (json:"id",
		// json:"tenantId", json:"homeTenantId", json:"user") instead of the Go
		// field names - unlike every other segment scrubbed here. Verified
		// against an actual unsanitized recording, not just the struct
		// definition, after the Go-field-name assumption below turned out
		// wrong for exactly this segment and leaked a real email address into
		// 8 committed fixtures (az.User.Name -> the serialized "user.name").
		if _, ok := data["id"]; ok {
			data["id"] = sanitizedGUID
		}

		if _, ok := data["tenantId"]; ok {
			data["tenantId"] = sanitizedGUID
		}

		if _, ok := data["homeTenantId"]; ok {
			data["homeTenantId"] = sanitizedGUID
		}

		if user, ok := data["user"].(map[string]any); ok {
			if _, ok := user["name"]; ok {
				user["name"] = "alice@contoso.com"
			}
		}
	case config.PATH:
		if _, ok := data["Path"]; ok {
			data["Path"] = sanitizedPath
		}

		// Location mirrors env.Flags().AbsolutePWD (segments/path.go) - the same
		// absolute-path leak as env.AbsolutePWD itself, just duplicated onto the
		// writer.
		if _, ok := data["Location"]; ok {
			data["Location"] = sanitizedPath
		}

		if folders, ok := data["Folders"].([]any); ok {
			for _, entry := range folders {
				folder, ok := entry.(map[string]any)
				if !ok {
					continue
				}

				if _, ok := folder["Path"]; ok {
					folder["Path"] = sanitizedPath
				}
			}
		}
	case config.SYSTEMINFO:
		// Matches website/segment_data.json's synthetic figures so a sanitized
		// fixture and the website tell the same story, and so goldens don't
		// encode whatever memory/load happened to be true on the recording
		// machine.
		for key, value := range map[string]any{
			"PhysicalTotalMemory":     int64(34359738368),
			"PhysicalAvailableMemory": int64(12884901888),
			"PhysicalFreeMemory":      int64(12884901888),
			"PhysicalPercentUsed":     62.5,
			"SwapTotalMemory":         0,
			"SwapFreeMemory":          0,
			"SwapPercentUsed":         0.0,
			"Load1":                   0.52,
			"Load5":                   0.48,
			"Load15":                  0.42,
		} {
			if _, ok := data[key]; ok {
				data[key] = value
			}
		}

		// Disks (runtime.SystemInfo, embedded via SystemInfo.SystemInfo) is a
		// map keyed by drive/device name, whose gopsutil-defined per-disk
		// entries carry their own serialNumber/label fields - real hardware
		// identifiers on a machine where the OS reports them (empty on this
		// one, but not guaranteed empty on the next). No bundled theme
		// template references .Disks (grepped), so it is not needed for
		// rendering; dropped rather than partially scrubbed.
		if _, ok := data["Disks"]; ok {
			data["Disks"] = map[string]any{}
		}
	case config.BATTERY:
		// battery.State 3 is Charging (runtime/battery/battery.go); Icon is left
		// blank like the website fixture rather than recomputed, since it is a
		// consequence of the (now overwritten) State plus per-theme option
		// glyphs, not identity.
		if _, ok := data["Percentage"]; ok {
			data["Percentage"] = 70
		}

		if _, ok := data["State"]; ok {
			data["State"] = 3
		}

		if _, ok := data["Icon"]; ok {
			data["Icon"] = ""
		}
	default:
		// Every other segment writer's recorded fields are either not
		// identity-carrying (version numbers, icons, boolean flags, ...) or
		// covered by the sanitizeSegmentText clear above.
	}
}

// sanitizeDataDocument re-parses a data document built by buildDataDocument and
// scrubs identity from it in place, returning the re-marshaled bytes. cfg
// supplies each segment's configured type, keyed by DataKey, so sanitization
// survives an alias.
func sanitizeDataDocument(doc []byte, cfg *config.Config) ([]byte, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(doc, &root); err != nil {
		return nil, err
	}

	typeByKey := make(map[string]config.SegmentType)

	for _, block := range cfg.Blocks {
		for _, segment := range block.Segments {
			typeByKey[segment.DataKey()] = segment.Type
		}
	}

	if envRaw, ok := root["env"]; ok {
		var env map[string]json.RawMessage
		if err := json.Unmarshal(envRaw, &env); err != nil {
			return nil, err
		}

		sanitizeEnv(env)

		raw, err := json.Marshal(env)
		if err != nil {
			return nil, err
		}

		root["env"] = raw
	}

	if segmentsRaw, ok := root["segments"]; ok {
		var segments map[string]json.RawMessage
		if err := json.Unmarshal(segmentsRaw, &segments); err != nil {
			return nil, err
		}

		for key, raw := range segments {
			var envelope config.RecordedSegment
			if err := json.Unmarshal(raw, &envelope); err != nil {
				return nil, err
			}

			var data map[string]any
			if err := json.Unmarshal(envelope.Data, &data); err != nil {
				return nil, err
			}

			sanitizeSegmentData(typeByKey[key], data)

			sanitizedData, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}

			envelope.Data = sanitizedData

			sanitizedRaw, err := json.Marshal(envelope)
			if err != nil {
				return nil, err
			}

			segments[key] = sanitizedRaw
		}

		raw, err := json.Marshal(segments)
		if err != nil {
			return nil, err
		}

		root["segments"] = raw
	}

	return json.MarshalIndent(root, "", "  ")
}
