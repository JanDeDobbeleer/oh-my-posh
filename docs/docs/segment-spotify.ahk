DetectHiddenWindows, On
WinGet, winInfo, List, ahk_exe Spotify.exe
indexer := 3
thisID := winInfo%indexer%
WinGetTitle, playing, ahk_id %thisID%
DetectHiddenWindows, Off
FileAppend, %playing%, *
