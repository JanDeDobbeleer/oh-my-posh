package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type gitVersion struct {
	NuGetVersionV2                  string `json:"NuGetVersionV2"`
	FullSemVer                      string `json:"FullSemVer"`
	CommitDate                      string `json:"CommitDate"`
	AssemblySemVer                  string `json:"AssemblySemVer"`
	PreReleaseTagWithDash           string `json:"PreReleaseTagWithDash"`
	PreReleaseLabel                 string `json:"PreReleaseLabel"`
	PreReleaseLabelWithDash         string `json:"PreReleaseLabelWithDash"`
	AssemblySemFileVer              string `json:"AssemblySemFileVer"`
	CommitsSinceVersionSourcePadded string `json:"CommitsSinceVersionSourcePadded"`
	VersionSourceSha                string `json:"VersionSourceSha"`
	BuildMetaDataPadded             string `json:"BuildMetaDataPadded"`
	FullBuildMetaData               string `json:"FullBuildMetaData"`
	MajorMinorPatch                 string `json:"MajorMinorPatch"`
	NuGetVersion                    string `json:"NuGetVersion"`
	LegacySemVer                    string `json:"LegacySemVer"`
	LegacySemVerPadded              string `json:"LegacySemVerPadded"`
	PreReleaseTag                   string `json:"PreReleaseTag"`
	NuGetPreReleaseTag              string `json:"NuGetPreReleaseTag"`
	SemVer                          string `json:"SemVer"`
	InformationalVersion            string `json:"InformationalVersion"`
	BranchName                      string `json:"BranchName"`
	EscapedBranchName               string `json:"EscapedBranchName"`
	Sha                             string `json:"Sha"`
	ShortSha                        string `json:"ShortSha"`
	NuGetPreReleaseTagV2            string `json:"NuGetPreReleaseTagV2"`
	BuildMetaData                   int    `json:"BuildMetaData"`
	Major                           int    `json:"Major"`
	PreReleaseNumber                int    `json:"PreReleaseNumber"`
	Minor                           int    `json:"Minor"`
	CommitsSinceVersionSource       int    `json:"CommitsSinceVersionSource"`
	WeightedPreReleaseNumber        int    `json:"WeightedPreReleaseNumber"`
	UncommittedChanges              int    `json:"UncommittedChanges"`
	Patch                           int    `json:"Patch"`
}

type GitVersion struct {
	props properties.Properties
	env   runtime.Environment

	gitVersion
}

func (n *GitVersion) Template() string {
	return " {{ .MajorMinorPatch }} "
}

func (n *GitVersion) Enabled() bool {
	gitversion := "gitversion"
	if !n.env.HasCommand(gitversion) {
		return false
	}

	response, err := n.env.RunCommand(gitversion, "-output", "json")
	if err != nil {
		return false
	}

	n.gitVersion = gitVersion{}
	err = json.Unmarshal([]byte(response), &n.gitVersion)

	return err == nil
}

func (n *GitVersion) Init(props properties.Properties, env runtime.Environment) {
	n.props = props
	n.env = env
}
