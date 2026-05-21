//go:build linux && !darwin

package segments

// PowerShell script that queries the Spotify session from the Windows
// System Media Transport Controls (SMTC). Output is a single line of
// "<status>|<title>|<artist>|<album>|<trackNumber>" where <status> is the
// lowercased PlaybackStatus enum value (playing/paused/stopped/closed/
// opened/changing), or "closed||||0" when no Spotify SMTC session exists.
//
// Used from WSL: the Linux binary cannot call WinRT directly, so it shells
// out to the Windows interop powershell.exe to read the host's SMTC list.
// On native Windows the binary instead calls combase.dll directly via
// runtime.QuerySpotifySMTC — see runtime/smtc_windows.go.
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
