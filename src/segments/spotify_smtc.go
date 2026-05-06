//go:build windows || (linux && !darwin)

package segments

import "strings"

// PowerShell script that queries the Spotify session from the Windows
// System Media Transport Controls (SMTC). Output is a single line of
// "<status>|<title>|<artist>|<album>|<trackNumber>" where <status> is the
// lowercased PlaybackStatus enum value (playing/paused/stopped/closed/
// opened/changing), or "closed||||0" when no Spotify SMTC session exists.
//
// Works for both the native Win32 Spotify client and the Microsoft Store
// version (SourceAppUserModelId starts with "SpotifyAB.SpotifyMusic_").
// On WSL, the same script is invoked via the Windows interop powershell.exe
// so the Linux side reads the same SMTC sessions as the Windows host.
const spotifySMTCScript = `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$null = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager,Windows.Media.Control,ContentType=WindowsRuntime]
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$asTaskGeneric = [System.WindowsRuntimeSystemExtensions].GetMethods() |
    Where-Object { $_.Name -eq 'AsTask' -and $_.IsGenericMethodDefinition -and $_.GetGenericArguments().Length -eq 1 -and $_.GetParameters().Length -eq 1 } |
    Select-Object -First 1
function Await($t, $rt) {
    $netTask = $asTaskGeneric.MakeGenericMethod($rt).Invoke($null, @($t))
    $netTask.Wait(-1) | Out-Null
    $netTask.Result
}
$mgrType = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager,Windows.Media.Control,ContentType=WindowsRuntime]
$mgr = Await ($mgrType::RequestAsync()) ([Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager])
$session = $mgr.GetSessions() | Where-Object { $_.SourceAppUserModelId -match 'Spotify' } | Select-Object -First 1
if (-not $session) { Write-Output 'closed||||0'; return }
$status = $session.GetPlaybackInfo().PlaybackStatus.ToString().ToLower()
$props = Await ($session.TryGetMediaPropertiesAsync()) ([Windows.Media.Control.GlobalSystemMediaTransportControlsSessionMediaProperties])
'{0}|{1}|{2}|{3}|{4}' -f $status, $props.Title, $props.Artist, $props.AlbumTitle, $props.TrackNumber`

// querySMTC invokes the SMTC PowerShell script and stores the result in s.
// Returns true when the segment should be displayed.
func (s *Spotify) querySMTC() bool {
	output, err := s.env.RunCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", spotifySMTCScript)
	if err != nil {
		return false
	}
	return s.parseSMTCOutput(output)
}

func (s *Spotify) parseSMTCOutput(output string) bool {
	output = strings.TrimSpace(output)
	if output == "" {
		return false
	}

	parts := strings.SplitN(output, "|", 5)
	if len(parts) != 5 {
		return false
	}

	switch parts[0] {
	case playing:
		s.Status = playing
	case paused:
		s.Status = paused
	default:
		// stopped, closed, opened, changing — segment hidden, matching macOS/Linux behavior.
		s.Status = stopped
		return false
	}

	s.Track = parts[1]
	s.Artist = parts[2]

	// Spotify ads expose the campaign as a regular "Music" SMTC entry but with no
	// AlbumTitle and TrackNumber=0 — neither is true for real tracks. Use the pair
	// to flag ads (single condition would false-positive on Spotify Singles etc.).
	if s.Status == playing && parts[3] == "" && parts[4] == "0" {
		s.Status = ad
	}

	s.resolveIcon()
	return true
}
