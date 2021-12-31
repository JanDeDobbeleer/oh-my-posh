[Setup]
AppName=Oh My Posh
AppVersion=<VERSION>
DefaultDirName={autopf}\oh-my-posh
DefaultGroupName=Oh My Posh
AppPublisher=Jan De Dobbeleer
AppPublisherURL=https://ohmyposh.dev
AppSupportURL=https://github.com/JanDeDobbeleer/oh-my-posh/issues
LicenseFile="bin\COPYING.txt"
OutputBaseFilename=install
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog

[Files]
Source: "bin\oh-my-posh.exe"; DestDir: "{app}\bin"
Source: "bin\themes\*"; DestDir: "{app}\themes"

[Registry]
Root: "HKCU"; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}\bin"; Check: NeedsAddPathHKCU(ExpandConstant('{app}\bin'))
Root: "HKCU"; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}\themes"; Check: NeedsAddPathHKCU(ExpandConstant('{app}\themes'))

[Code]
function NeedsAddPathHKCU(Param: string): boolean;
var
OrigPath: string;
begin
if not RegQueryStringValue(HKEY_CURRENT_USER,
'Environment',
'Path', OrigPath)
then begin
Result := True;
exit;
end;
// look for the path with leading and trailing semicolon
// Pos() returns 0 if not found
Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;
