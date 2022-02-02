package engine

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/segments"
	"strings"
)

const (
	colorBackground = properties.Property("color_background")

	prefix  = properties.Property("prefix")
	postfix = properties.Property("postfix")
)

func (cfg *Config) Migrate(env environment.Environment) {
	for _, block := range cfg.Blocks {
		for _, segment := range block.Segments {
			segment.migrate(env, cfg.Version)
		}
	}
	for _, segment := range cfg.Tooltips {
		segment.migrate(env, cfg.Version)
	}
	cfg.updated = true
	cfg.Version = configVersion
}

func (segment *Segment) migrate(env environment.Environment, version int) {
	if version < 1 {
		segment.migrationOne(env)
	}
}

func (segment *Segment) migrationOne(env environment.Environment) {
	if err := segment.mapSegmentWithWriter(env); err != nil {
		return
	}
	// General properties that need replacement
	segment.migratePropertyKey("display_version", properties.FetchVersion)
	delete(segment.Properties, "enable_hyperlink")
	switch segment.Type { // nolint:exhaustive
	case TEXT:
		segment.migratePropertyValue("text", properties.SegmentTemplate)
	case GIT:
		hasTemplate := segment.hasProperty(properties.SegmentTemplate)
		segment.migratePropertyKey("display_status", segments.FetchStatus)
		segment.migratePropertyKey("display_stash_count", segments.FetchStashCount)
		segment.migratePropertyKey("display_worktree_count", segments.FetchWorktreeCount)
		segment.migratePropertyKey("display_upstream_icon", segments.FetchUpstreamIcon)
		segment.migrateTemplate()
		segment.migrateIconOverride("local_working_icon", " \uF044 ")
		segment.migrateIconOverride("local_staged_icon", " \uF046 ")
		segment.migrateIconOverride("stash_count_icon", " \uF692 ")
		segment.migrateIconOverride("worktree_count_icon", " \uf1bb ")
		segment.migrateIconOverride("status_separator_icon", " |")
		if segment.Properties.GetBool(properties.Property("status_colors_enabled"), false) {
			segment.migrateColorOverride("local_changes_color", "{{ if or (.Working.Changed) (.Staging.Changed) }}%s{{ end }}")
			segment.migrateColorOverride("ahead_and_behind_color", "{{ if and (gt .Ahead 0) (gt .Behind 0) }}%s{{ end }}")
			segment.migrateColorOverride("behind_color", "{{ if gt .Ahead 0 }}%s{{ end }}")
			segment.migrateColorOverride("ahead_color", "{{ if gt .Behind 0 }}%s{{ end }}")
		}
		if !hasTemplate {
			segment.migrateInlineColorOverride("working_color", "{{ .Working.String }}")
			segment.migrateInlineColorOverride("staging_color", "{{ .Staging.String }}")
		}
		// legacy properties
		delete(segment.Properties, "display_branch_status")
		delete(segment.Properties, "display_status_detail")
		delete(segment.Properties, "status_colors_enabled")
	case BATTERY:
		segment.migrateTemplate()
		segment.migrateColorOverride("charged_color", `{{ if eq "Full" .State.String }}%s{{ end }}`)
		segment.migrateColorOverride("charging_color", `{{ if eq "Charging" .State.String }}%s{{ end }}`)
		segment.migrateColorOverride("discharging_color", `{{ if eq "Discharging" .State.String }}%s{{ end }}`)
		stateList := []string{`"Discharging"`}
		if segment.Properties.GetBool(properties.Property("display_charging"), true) {
			stateList = append(stateList, `"Charging"`)
		}
		if segment.Properties.GetBool(properties.Property("display_charged"), true) {
			stateList = append(stateList, `"Full"`)
		}
		if len(stateList) < 3 {
			enabledTemplate := "{{ $stateList := list %s }}{{ if has .State.String $stateList }}{{ .Icon }}{{ .Percentage }}{{ end }}"
			template := segment.Properties.GetString(properties.SegmentTemplate, segment.writer.Template())
			template = strings.ReplaceAll(template, "{{ .Icon }}{{ .Percentage }}", fmt.Sprintf(enabledTemplate, strings.Join(stateList, " ")))
			segment.Properties[properties.SegmentTemplate] = template
		}
		// legacy properties
		delete(segment.Properties, "display_charging")
		delete(segment.Properties, "display_charged")
		delete(segment.Properties, "battery_icon")
	case PYTHON:
		segment.migrateTemplate()
		segment.migratePropertyKey("display_virtual_env", segments.FetchVirtualEnv)
	case SESSION:
		hasTemplate := segment.hasProperty(properties.SegmentTemplate)
		segment.migrateTemplate()
		segment.migrateIconOverride("ssh_icon", "\uf817 ")
		template := segment.Properties.GetString(properties.SegmentTemplate, segment.writer.Template())
		if !segment.Properties.GetBool(properties.Property("display_host"), true) {
			template = strings.ReplaceAll(template, "@{{ .HostName }}", "")
		}
		if !segment.Properties.GetBool(properties.Property("display_user"), true) {
			template = strings.ReplaceAll(template, "@", "")
			template = strings.ReplaceAll(template, "{{ .UserName }}", "")
		}
		segment.Properties[properties.SegmentTemplate] = template
		segment.migrateIconOverride("user_info_separator", "@")
		if !hasTemplate {
			segment.migrateInlineColorOverride("user_color", "{{ .UserName }}")
			segment.migrateInlineColorOverride("host_color", "{{ .HostName }}")
		}
	case NODE:
		segment.migrateTemplate()
		segment.migratePropertyKey("display_package_manager", segments.FetchPackageManager)
		enableVersionMismatch := "enable_version_mismatch"
		if segment.Properties.GetBool(properties.Property(enableVersionMismatch), false) {
			delete(segment.Properties, properties.Property(enableVersionMismatch))
			segment.migrateColorOverride("version_mismatch_color", "{{ if .Mismatch }}%s{{ end }}")
		}
	case EXIT:
		segment.migrateTemplate()
		template := segment.Properties.GetString(properties.SegmentTemplate, segment.writer.Template())
		displayExitCode := properties.Property("display_exit_code")
		if !segment.Properties.GetBool(displayExitCode, true) {
			delete(segment.Properties, displayExitCode)
			template = strings.ReplaceAll(template, " {{ .Meaning }}", "")
		}
		alwaysNumeric := properties.Property("always_numeric")
		if segment.Properties.GetBool(alwaysNumeric, false) {
			delete(segment.Properties, alwaysNumeric)
			template = strings.ReplaceAll(template, ".Meaning", ".Code")
		}
		segment.Properties[properties.SegmentTemplate] = template
		segment.migrateIconOverride("success_icon", "\uf42e")
		segment.migrateIconOverride("error_icon", "\uf00d")
		segment.migrateColorOverride("error_color", "{{ if gt .Code 0 }}%s{{ end }}")
	default:
		segment.migrateTemplate()
	}
	delete(segment.Properties, colorBackground)
}

func (segment *Segment) hasProperty(property properties.Property) bool {
	for key := range segment.Properties {
		if key == property {
			return true
		}
	}
	return false
}

func (segment *Segment) migratePropertyValue(property properties.Property, value interface{}) {
	if !segment.hasProperty(property) {
		return
	}
	segment.Properties[property] = value
}

func (segment *Segment) migratePropertyKey(oldProperty, newProperty properties.Property) {
	if !segment.hasProperty(oldProperty) {
		return
	}
	value := segment.Properties[oldProperty]
	delete(segment.Properties, oldProperty)
	segment.Properties[newProperty] = value
}

func (segment *Segment) migrateTemplate() {
	defer segment.migratePreAndPostFix()
	if segment.hasProperty(properties.SegmentTemplate) {
		// existing template, ensure to add default pre/postfix values
		if !segment.hasProperty(prefix) {
			segment.Properties[prefix] = " "
		}
		if !segment.hasProperty(postfix) {
			segment.Properties[postfix] = " "
		}
		return
	}
	segment.Properties[properties.SegmentTemplate] = segment.writer.Template()
}

func (segment *Segment) migrateIconOverride(property properties.Property, overrideValue string) {
	if !segment.hasProperty(property) {
		return
	}
	template := segment.Properties.GetString(properties.SegmentTemplate, segment.writer.Template())
	if strings.Contains(template, overrideValue) {
		template = strings.ReplaceAll(template, overrideValue, segment.Properties.GetString(property, ""))
	}
	segment.Properties[properties.SegmentTemplate] = template
	delete(segment.Properties, property)
}

func (segment *Segment) migrateColorOverride(property properties.Property, template string) {
	if !segment.hasProperty(property) {
		return
	}
	color := segment.Properties.GetColor(property, "")
	delete(segment.Properties, property)
	if len(color) == 0 {
		return
	}
	background := segment.Properties.GetBool(colorBackground, false)
	colorTemplate := fmt.Sprintf(template, color)
	if background {
		segment.BackgroundTemplates = append(segment.BackgroundTemplates, colorTemplate)
		return
	}
	segment.ForegroundTemplates = append(segment.ForegroundTemplates, colorTemplate)
}

func (segment *Segment) migrateInlineColorOverride(property properties.Property, old string) {
	if !segment.hasProperty(property) {
		return
	}
	color := segment.Properties.GetColor(property, "")
	delete(segment.Properties, property)
	if len(color) == 0 {
		return
	}
	colorTemplate := fmt.Sprintf("<%s>%s</>", color, old)
	template := segment.Properties.GetString(properties.SegmentTemplate, segment.writer.Template())
	template = strings.ReplaceAll(template, old, colorTemplate)
	segment.Properties[properties.SegmentTemplate] = template
}

func (segment *Segment) migratePreAndPostFix() {
	template := segment.Properties.GetString(properties.SegmentTemplate, segment.writer.Template())
	defaultValue := " "
	if segment.hasProperty(prefix) {
		prefix := segment.Properties.GetString(prefix, defaultValue)
		template = strings.TrimPrefix(template, defaultValue)
		template = prefix + template
		delete(segment.Properties, "prefix")
	}
	if segment.hasProperty(postfix) {
		postfix := segment.Properties.GetString(postfix, defaultValue)
		template = strings.TrimSuffix(template, defaultValue)
		template += postfix
		delete(segment.Properties, "postfix")
	}
	segment.Properties[properties.SegmentTemplate] = template
}
