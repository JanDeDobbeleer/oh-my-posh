# Typed Segment Configuration - Implementation Guide

## Current Architecture

### How Segments Work Today

1. **Config Loading**: Configs are loaded via `config.Load()` which:
   - Calls `parseConfigFile()` -> `readConfig()`
   - Uses standard JSON/YAML/TOML unmarshalers
   - Unmarshals into `Config` struct containing `[]Block`
   - Each Block contains `[]*Segment`

2. **Segment Structure** (`src/config/segment.go`):
   ```go
   type Segment struct {
       Type       SegmentType
       Properties properties.Map  // Generic property bag
       // ... other fields
       writer     SegmentWriter
   }
   ```

3. **Segment Initialization** (`src/config/segment_types.go`):
   ```go
   func (segment *Segment) MapSegmentWithWriter(env runtime.Environment) error {
       writer := Segments[segment.Type]()  // Factory function
       wrapper := &properties.Wrapper{Properties: segment.Properties}
       writer.Init(wrapper, env)  // Pass properties
       segment.writer = writer
       return nil
   }
   ```

4. **SegmentWriter Interface**:
   ```go
   type SegmentWriter interface {
       Enabled() bool
       Template() string
       SetText(text string)
       SetIndex(index int)
       Text() string
       Init(props properties.Properties, env runtime.Environment)
       CacheKey() (string, bool)
   }
   ```

5. **Segment Implementation** (e.g., `src/segments/status.go`):
   ```go
   type Status struct {
       Base
       String  string
       Meaning string
       Error   bool
   }
   
   func (s *Status) Init(props properties.Properties, env runtime.Environment) {
       s.Base.Init(props, env)
   }
   
   func (s *Status) Enabled() bool {
       // Uses s.props.GetString(StatusTemplate, "{{ .Code }}")
       // Uses s.props.GetBool(AlwaysEnabled, false)
   }
   ```

## Target Architecture

### How Segments Should Work

1. **Typed Segment Structs**: Each segment is a self-contained struct with typed fields:
   ```go
   type Status struct {
       // Embedded base for common functionality
       SegmentBase
       
       // Typed configuration fields with defaults
       StatusTemplate  string `json:"status_template" default:"{{ .Code }}"`
       StatusSeparator string `json:"status_separator" default:"|"`
       AlwaysEnabled   bool   `json:"always_enabled"`
       
       // Runtime state (not serialized)
       String  string `json:"-"`
       Meaning string `json:"-"`
       Error   bool   `json:"-"`
   }
   ```

2. **SegmentBase**: Common fields extracted to shared struct:
   ```go
   type SegmentBase struct {
       Type       SegmentType
       Template   string
       Foreground color.Ansi
       // ... all common fields
       
       // Runtime fields
       env     runtime.Environment `json:"-"`
       text    string             `json:"-"`
       index   int                `json:"-"`
   }
   ```

3. **Simplified Interface**: Remove properties parameter:
   ```go
   type SegmentWriter interface {
       Enabled() bool
       Template() string
       SetText(text string)
       SetIndex(index int)
       Text() string
       Init(env runtime.Environment)  // No props!
       CacheKey() (string, bool)
   }
   ```

4. **Polymorphic Unmarshaling**: Custom unmarshal logic:
   ```go
   func (b *Block) UnmarshalJSON(data []byte) error {
       // Parse segments as raw JSON
       // Peek at "type" field
       // Create concrete segment based on type
       // Unmarshal into concrete type
       // Call ApplyDefaults()
       // Append to segments list
   }
   ```

## Step-by-Step Migration Path

### Phase 1: Infrastructure (COMPLETED ✅)
- [x] Create `ApplyDefaults()` helper
- [x] Add comprehensive tests
- [x] Document migration plan

### Phase 2: Create SegmentBase

**File**: `src/config/segment_base.go`

```go
package config

import (
    "time"
    
    "github.com/jandedobbeleer/oh-my-posh/src/color"
    "github.com/jandedobbeleer/oh-my-posh/src/runtime"
    "github.com/jandedobbeleer/oh-my-posh/src/template"
)

// SegmentBase contains common configuration fields for all segments
type SegmentBase struct {
    // Configuration fields
    Type                   SegmentType    `json:"type" toml:"type" yaml:"type"`
    Alias                  string         `json:"alias,omitempty" toml:"alias,omitempty" yaml:"alias,omitempty"`
    Template               string         `json:"template,omitempty" toml:"template,omitempty" yaml:"template,omitempty"`
    Templates              template.List  `json:"templates,omitempty" toml:"templates,omitempty" yaml:"templates,omitempty"`
    TemplatesLogic         template.Logic `json:"templates_logic,omitempty" toml:"templates_logic,omitempty" yaml:"templates_logic,omitempty"`
    Foreground             color.Ansi     `json:"foreground,omitempty" toml:"foreground,omitempty" yaml:"foreground,omitempty"`
    ForegroundTemplates    template.List  `json:"foreground_templates,omitempty" toml:"foreground_templates,omitempty" yaml:"foreground_templates,omitempty"`
    Background             color.Ansi     `json:"background,omitempty" toml:"background,omitempty" yaml:"background,omitempty"`
    BackgroundTemplates    template.List  `json:"background_templates,omitempty" toml:"background_templates,omitempty" yaml:"background_templates,omitempty"`
    PowerlineSymbol        string         `json:"powerline_symbol,omitempty" toml:"powerline_symbol,omitempty" yaml:"powerline_symbol,omitempty"`
    LeadingPowerlineSymbol string         `json:"leading_powerline_symbol,omitempty" toml:"leading_powerline_symbol,omitempty" yaml:"leading_powerline_symbol,omitempty"`
    LeadingDiamond         string         `json:"leading_diamond,omitempty" toml:"leading_diamond,omitempty" yaml:"leading_diamond,omitempty"`
    TrailingDiamond        string         `json:"trailing_diamond,omitempty" toml:"trailing_diamond,omitempty" yaml:"trailing_diamond,omitempty"`
    Interactive            bool           `json:"interactive,omitempty" toml:"interactive,omitempty" yaml:"interactive,omitempty"`
    Timeout                time.Duration  `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty"`
    MaxWidth               int            `json:"max_width,omitempty" toml:"max_width,omitempty" yaml:"max_width,omitempty"`
    MinWidth               int            `json:"min_width,omitempty" toml:"min_width,omitempty" yaml:"min_width,omitempty"`
    Index                  int            `json:"index,omitempty" toml:"index,omitempty" yaml:"index,omitempty"`
    Force                  bool           `json:"force,omitempty" toml:"force,omitempty" yaml:"force,omitempty"`
    Newline                bool           `json:"newline,omitempty" toml:"newline,omitempty" yaml:"newline,omitempty"`
    ExcludeFolders         []string       `json:"exclude_folders,omitempty" toml:"exclude_folders,omitempty" yaml:"exclude_folders,omitempty"`
    IncludeFolders         []string       `json:"include_folders,omitempty" toml:"include_folders,omitempty" yaml:"include_folders,omitempty"`
    Tips                   []string       `json:"tips,omitempty" toml:"tips,omitempty" yaml:"tips,omitempty"`
    Cache                  *Cache         `json:"cache,omitempty" toml:"cache,omitempty" yaml:"cache,omitempty"`
    
    // Runtime-only fields (not serialized)
    env        runtime.Environment `json:"-" toml:"-" yaml:"-"`
    text       string              `json:"-" toml:"-" yaml:"-"`
    index      int                 `json:"-" toml:"-" yaml:"-"`
    styleCache SegmentStyle        `json:"-" toml:"-" yaml:"-"`
    name       string              `json:"-" toml:"-" yaml:"-"`
    Needs      []string            `json:"-" toml:"-" yaml:"-"`
    Duration   time.Duration       `json:"-" toml:"-" yaml:"-"`
    NameLength int                 `json:"-" toml:"-" yaml:"-"`
    Enabled    bool                `json:"-" toml:"-" yaml:"-"`
    restored   bool                `json:"-" toml:"-" yaml:"-"`
}

// Name returns the segment name (alias or type)
func (s *SegmentBase) Name() string {
    if s.name != "" {
        return s.name
    }
    if s.Alias != "" {
        s.name = s.Alias
        return s.name
    }
    // Convert SegmentType to title case string
    s.name = string(s.Type)
    return s.name
}

// SetText sets the rendered text
func (s *SegmentBase) SetText(text string) {
    s.text = text
}

// Text returns the rendered text
func (s *SegmentBase) Text() string {
    return s.text
}

// SetIndex sets the segment index
func (s *SegmentBase) SetIndex(index int) {
    s.index = index
}

// GetEnv returns the runtime environment
func (s *SegmentBase) GetEnv() runtime.Environment {
    return s.env
}

// SetEnv sets the runtime environment
func (s *SegmentBase) SetEnv(env runtime.Environment) {
    s.env = env
}
```

### Phase 3: Update SegmentWriter Interface

**File**: `src/config/segment_types.go`

Change line 22 from:
```go
Init(props properties.Properties, env runtime.Environment)
```

To:
```go
Init(env runtime.Environment)
```

This is a BREAKING CHANGE that affects all segments.

### Phase 4: Implement Polymorphic Unmarshaling

**Option A**: Custom UnmarshalJSON on Block

Add to `src/config/block.go`:

```go
func (b *Block) UnmarshalJSON(data []byte) error {
    // Use type alias to avoid recursion
    type Alias Block
    aux := &struct {
        RawSegments []json.RawMessage `json:"segments"`
        *Alias
    }{
        Alias: (*Alias)(b),
    }
    
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    
    // Clear segments before repopulating
    b.Segments = nil
    
    for _, rawSeg := range aux.RawSegments {
        segment, err := unmarshalSegment(rawSeg)
        if err != nil {
            return err
        }
        b.Segments = append(b.Segments, segment)
    }
    
    return nil
}

func unmarshalSegment(data []byte) (*Segment, error) {
    // Peek at type field
    var typeCheck struct {
        Type SegmentType `json:"type"`
    }
    if err := json.Unmarshal(data, &typeCheck); err != nil {
        return nil, err
    }
    
    // Create concrete writer based on type
    writer := createTypedSegment(typeCheck.Type)
    if writer == nil {
        // Fall back to old property-based system for non-migrated segments
        var seg Segment
        if err := json.Unmarshal(data, &seg); err != nil {
            return nil, err
        }
        return &seg, nil
    }
    
    // Unmarshal into concrete type
    if err := json.Unmarshal(data, writer); err != nil {
        return nil, err
    }
    
    // Apply defaults
    if err := ApplyDefaults(writer); err != nil {
        return nil, err
    }
    
    // Wrap in Segment struct for backwards compatibility
    seg := &Segment{
        Type:   typeCheck.Type,
        writer: writer,
    }
    
    return seg, nil
}

func createTypedSegment(t SegmentType) SegmentWriter {
    switch t {
    case STATUS:
        return &segments.Status{}
    case PATH:
        return &segments.Path{}
    case GIT:
        return &segments.Git{}
    // Add other migrated segments here
    default:
        return nil  // Falls back to property-based
    }
}
```

**Option B**: Keep dual system

Modify `MapSegmentWithWriter` to detect new vs old segments:

```go
func (segment *Segment) MapSegmentWithWriter(env runtime.Environment) error {
    segment.env = env
    
    // If writer already set (from polymorphic unmarshal), just init it
    if segment.writer != nil {
        segment.writer.Init(env)
        return nil
    }
    
    // Old property-based system
    if segment.Properties == nil {
        segment.Properties = make(properties.Map)
    }
    
    f, ok := Segments[segment.Type]
    if !ok {
        return errors.New("unable to map writer")
    }
    
    writer := f()
    
    // Check if writer supports new interface
    if newWriter, ok := writer.(interface{ Init(runtime.Environment) }); ok {
        newWriter.Init(env)
    } else {
        // Old interface
        wrapper := &properties.Wrapper{Properties: segment.Properties}
        writer.Init(wrapper, env)
    }
    
    segment.writer = writer
    return nil
}
```

### Phase 5: Migrate Status Segment (Example)

**File**: `src/segments/status.go`

```go
package segments

import (
    "strconv"
    "strings"
    
    "github.com/jandedobbeleer/oh-my-posh/src/config"
    "github.com/jandedobbeleer/oh-my-posh/src/runtime"
    "github.com/jandedobbeleer/oh-my-posh/src/template"
    "github.com/jandedobbeleer/oh-my-posh/src/text"
)

type Status struct {
    config.SegmentBase
    
    // Configuration
    StatusTemplate  string `json:"status_template" toml:"status_template" yaml:"status_template" default:"{{ .Code }}"`
    StatusSeparator string `json:"status_separator" toml:"status_separator" yaml:"status_separator" default:"|"`
    AlwaysEnabled   bool   `json:"always_enabled" toml:"always_enabled" yaml:"always_enabled"`
    
    // Runtime state
    String  string `json:"-"`
    Meaning string `json:"-"`
    Error   bool   `json:"-"`
    Code    int    `json:"-"`  // Exposed for template
}

func (s *Status) Template() string {
    if s.SegmentBase.Template != "" {
        return s.SegmentBase.Template
    }
    return " {{ .String }} "
}

func (s *Status) Init(env runtime.Environment) {
    s.SetEnv(env)
}

func (s *Status) Enabled() bool {
    status, pipeStatus := s.GetEnv().StatusCodes()
    s.Code = status
    s.String = s.formatStatus(status, pipeStatus)
    s.Meaning = template.GetReasonFromStatus(status)
    
    if s.AlwaysEnabled {
        return true
    }
    
    return s.Error
}

func (s *Status) CacheKey() (string, bool) {
    return "", false
}

func (s *Status) formatStatus(status int, pipeStatus string) string {
    if status != 0 {
        s.Error = true
    }
    
    if pipeStatus == "" {
        s.Code = status
        if txt, err := template.Render(s.StatusTemplate, s); err == nil {
            return txt
        }
        return strconv.Itoa(status)
    }
    
    builder := text.NewBuilder()
    
    var context struct {
        Code int
    }
    
    splitted := strings.Split(pipeStatus, " ")
    for i, codeStr := range splitted {
        write := func(txt string) {
            if i > 0 {
                builder.WriteString(s.StatusSeparator)
            }
            builder.WriteString(txt)
        }
        
        code, err := strconv.Atoi(codeStr)
        if err != nil {
            write(codeStr)
            continue
        }
        
        if code != 0 {
            s.Error = true
        }
        
        context.Code = code
        
        txt, err := template.Render(s.StatusTemplate, context)
        if err != nil {
            write(codeStr)
            continue
        }
        
        write(txt)
    }
    
    return builder.String()
}
```

### Phase 6: Update Tests

**File**: `src/segments/status_test.go` (create or update)

```go
func TestStatusEnabled(t *testing.T) {
    env := new(mock.Environment)
    env.On("StatusCodes").Return(0, "")
    
    // Old way:
    // s := &Status{}
    // s.Init(properties.Map{AlwaysEnabled: true}, env)
    
    // New way:
    s := &Status{AlwaysEnabled: true}
    s.Init(env)
    
    assert.True(t, s.Enabled())
}
```

### Phase 7: Update Base Segment

**File**: `src/segments/base.go`

```go
package segments

import (
    "github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

// Base provides common functionality for legacy segments
// New segments should embed config.SegmentBase instead
type Base struct {
    env runtime.Environment
    Segment *Segment
}

type Segment struct {
    Text  string
    Index int
}

func (b *Base) Text() string {
    return b.Segment.Text
}

func (b *Base) SetText(text string) {
    b.Segment.Text = text
}

func (b *Base) SetIndex(index int) {
    b.Segment.Index = index
}

func (b *Base) Init(env runtime.Environment) {
    b.Segment = &Segment{}
    b.env = env
}

func (b *Base) CacheKey() (string, bool) {
    return "", false
}
```

## Testing Strategy

1. **Unit Tests**: Test each segment's Enabled() and Template() logic
2. **Integration Tests**: Test full config unmarshaling
3. **Migration Tests**: Ensure old configs still work (if maintaining compatibility)

## Rollout Plan

1. Implement infrastructure (DONE)
2. Implement ONE segment end-to-end (Status) to validate approach
3. Test thoroughly
4. Migrate Path segment
5. Migrate Git segment
6. Update documentation
7. Gradually migrate remaining segments
8. Remove old properties system once all migrated

## Risks

- **Breaking Change**: All configs need updating if we don't maintain compatibility
- **Large Scope**: 90+ segments to eventually migrate
- **Testing Burden**: Each segment needs thorough testing

## Benefits

- **Type Safety**: Catch configuration errors at compile time
- **Better IDE Support**: Autocomplete for config fields
- **Clearer Defaults**: Expressed in struct tags
- **Easier Maintenance**: No more property string constants
