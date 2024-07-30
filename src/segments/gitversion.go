package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type gitVersion struct {
	Major                           int    `json:"Major"`
	Minor                           int    `json:"Minor"`
	Patch                           int    `json:"Patch"`
	PreReleaseTag                   string `json:"PreReleaseTag"`
	PreReleaseTagWithDash           string `json:"PreReleaseTagWithDash"`
	PreReleaseLabel                 string `json:"PreReleaseLabel"`
	PreReleaseLabelWithDash         string `json:"PreReleaseLabelWithDash"`
	PreReleaseNumber                int    `json:"PreReleaseNumber"`
	WeightedPreReleaseNumber        int    `json:"WeightedPreReleaseNumber"`
	BuildMetaData                   int    `json:"BuildMetaData"`
	BuildMetaDataPadded             string `json:"BuildMetaDataPadded"`
	FullBuildMetaData               string `json:"FullBuildMetaData"`
	MajorMinorPatch                 string `json:"MajorMinorPatch"`
	SemVer                          string `json:"SemVer"`
	LegacySemVer                    string `json:"LegacySemVer"`
	LegacySemVerPadded              string `json:"LegacySemVerPadded"`
	AssemblySemVer                  string `json:"AssemblySemVer"`
	AssemblySemFileVer              string `json:"AssemblySemFileVer"`
	FullSemVer                      string `json:"FullSemVer"`
	InformationalVersion            string `json:"InformationalVersion"`
	BranchName                      string `json:"BranchName"`
	EscapedBranchName               string `json:"EscapedBranchName"`
	Sha                             string `json:"Sha"`
	ShortSha                        string `json:"ShortSha"`
	NuGetVersionV2                  string `json:"NuGetVersionV2"`
	NuGetVersion                    string `json:"NuGetVersion"`
	NuGetPreReleaseTagV2            string `json:"NuGetPreReleaseTagV2"`
	NuGetPreReleaseTag              string `json:"NuGetPreReleaseTag"`
	VersionSourceSha                string `json:"VersionSourceSha"`
	CommitsSinceVersionSource       int    `json:"CommitsSinceVersionSource"`
	CommitsSinceVersionSourcePadded string `json:"CommitsSinceVersionSourcePadded"`
	UncommittedChanges              int    `json:"UncommittedChanges"`
	CommitDate                      string `json:"CommitDate"`
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
