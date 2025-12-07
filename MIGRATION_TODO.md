# Typed Segment Configuration Migration - TODO

## Overview
This document tracks the migration from property-based segment configuration to typed struct-based configuration.

## Completed
- [x] Created `src/config/defaults.go` with `ApplyDefaults()` helper function
- [x] Added comprehensive tests for `ApplyDefaults()` in `src/config/defaults_test.go`
- [x] All tests passing for defaults functionality

## Next Steps

### 1. Create SegmentBase Structure (HIGH PRIORITY)
Create `src/config/segment_base.go` with a `SegmentBase` struct that contains common fields:
- Type, Alias, Template, Templates, TemplatesLogic
- Foreground, Background, PowerlineSymbol, LeadingPowerlineSymbol
- LeadingDiamond, TrailingDiamond
- Interactive, Timeout, MaxWidth, MinWidth
- Index, Force, Newline
- ExcludeFolders, IncludeFolders, Tips
- Cache
- Runtime-only fields: env, writer, styleCache, name, Needs, Duration, NameLength, Enabled, restored

### 2. Modify SegmentWriter Interface (HIGH PRIORITY)
In `src/config/segment_types.go`:
- Change `Init(props properties.Properties, env runtime.Environment)` 
- To: `Init(env runtime.Environment)`

This is a breaking change that affects all segments.

### 3. Update Base Segment (HIGH PRIORITY)
In `src/segments/base.go`:
- Remove `props properties.Properties` field
- Update `Init()` signature to match new interface
- Keep `env` and `Segment` fields

### 4. Migrate Status Segment (FIRST IMPLEMENTATION)
Start with Status as it's the simplest:

File: `src/segments/status.go`
```go
type Status struct {
    Base
    
    // Config fields
    StatusTemplate  string `json:"status_template" toml:"status_template" yaml:"status_template" default:"{{ .Code }}"`
    StatusSeparator string `json:"status_separator" toml:"status_separator" yaml:"status_separator" default:"|"`
    AlwaysEnabled   bool   `json:"always_enabled" toml:"always_enabled" yaml:"always_enabled"`
    
    // Runtime fields (not serialized)
    String  string `json:"-"`
    Meaning string `json:"-"`
    Error   bool   `json:"-"`
}

func (s *Status) Init(env runtime.Environment) {
    s.Base.Init(env)
    // ApplyDefaults will be called by the unmarshaler
}

func (s *Status) Enabled() bool {
    status, pipeStatus := s.env.StatusCodes()
    s.String = s.formatStatus(status, pipeStatus)
    s.Meaning = template.GetReasonFromStatus(status)
    
    if s.AlwaysEnabled {
        return true
    }
    return s.Error
}

func (s *Status) formatStatus(status int, pipeStatus string) string {
    statusTemplate := s.StatusTemplate
    // ... rest of logic using s.StatusTemplate and s.StatusSeparator directly
}
```

### 5. Implement Polymorphic Unmarshaling (CRITICAL)
In `src/config/block.go` or `src/config/segment.go`:

Add custom `UnmarshalJSON` for Block that:
1. Parses segments as `[]json.RawMessage`
2. For each segment, peek at the `type` field
3. Create the appropriate concrete segment struct based on type
4. Unmarshal into that struct
5. Call `ApplyDefaults(segment)`
6. Append to Block.Segments as SegmentWriter

Example:
```go
func (b *Block) UnmarshalJSON(data []byte) error {
    type Alias Block
    aux := &struct {
        Segments []json.RawMessage `json:"segments"`
        *Alias
    }{
        Alias: (*Alias)(b),
    }
    
    if err := json.Unmarshal(data, aux); err != nil {
        return err
    }
    
    for _, rawSeg := range aux.Segments {
        // Peek at type
        var typeCheck struct {
            Type SegmentType `json:"type"`
        }
        json.Unmarshal(rawSeg, &typeCheck)
        
        // Create concrete segment
        segment := createSegment(typeCheck.Type)
        if segment == nil {
            return fmt.Errorf("unknown segment type: %s", typeCheck.Type)
        }
        
        // Unmarshal into concrete type
        if err := json.Unmarshal(rawSeg, segment); err != nil {
            return err
        }
        
        // Apply defaults
        if err := ApplyDefaults(segment); err != nil {
            return err
        }
        
        b.Segments = append(b.Segments, segment)
    }
    
    return nil
}

func createSegment(t SegmentType) SegmentWriter {
    switch t {
    case STATUS:
        return &segments.Status{}
    case PATH:
        return &segments.Path{}
    case GIT:
        return &segments.Git{}
    // Add other migrated segments here
    default:
        return nil
    }
}
```

### 6. Migrate Path Segment
Similar to Status but with more fields. Key properties to migrate:
- Style (various path styles)
- FolderSeparatorIcon/Template
- HomeIcon, FolderIcon
- MappedLocations
- MaxDepth, MaxWidth
- And many more - see path.go for full list

### 7. Migrate Git Segment
Most complex of the three. Properties to migrate:
- FetchStatus, FetchPushStatus, FetchWorktreeCount, etc.
- Various icons (BranchIcon, BranchAheadIcon, etc.)
- IgnoreStatus, DisableWithJJ
- See git.go for full list (100+ lines of properties)

### 8. Update Tests
For each migrated segment:
- Update tests that use `properties.Map{}` to construct segment configs
- Instead, construct segment structs directly or use JSON unmarshaling
- Example:
  ```go
  // Old:
  s := &Status{}
  s.Init(properties.Map{StatusTemplate: "custom"}, env)
  
  // New:
  s := &Status{StatusTemplate: "custom"}
  s.Init(env)
  ```

### 9. Handle Non-Migrated Segments
Option A (Coexistence): Keep dual system
- Add a check in MapSegmentWithWriter to detect if segment supports new Init signature
- Use reflection or type assertion to determine which Init to call
- Allows gradual migration

Option B (As specified in problem statement):
- Comment out all non-migrated segment registrations in segment_types.go
- Add clear TODO comments
- This breaks non-migrated segments temporarily

### 10. Documentation
Update docs for migrated segments:
- website/docs/segments/git.md
- website/docs/segments/path.md
- website/docs/segments/status.md
- Add migration guide in website/docs/configuration/migration.mdx

## Testing Strategy
1. Test ApplyDefaults thoroughly (DONE ✓)
2. Test each migrated segment's Enabled() and Template() in isolation
3. Test JSON/TOML/YAML unmarshaling for each segment
4. Test full config loading with migrated segments
5. Integration tests with actual prompts

## Risks and Considerations
- This is a BREAKING CHANGE - configs will need updates
- Non-migrated segments will be temporarily broken if we comment them out
- Need to update all documentation
- Users will need migration instructions

## Future Work
After git/path/status are proven working:
- Migrate remaining segments one by one
- Remove old properties.Properties system entirely
- Simplify SegmentWriter interface further
