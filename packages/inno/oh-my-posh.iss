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
ChangesEnvironment=yes
SignTool=signtool
SignedUninstaller=yes
CloseApplications=no

[Files]
Source: "bin\oh-my-posh.exe"; DestDir: "{app}\bin"; Flags: sign
Source: "bin\themes\*"; DestDir: "{app}\themes"

[Registry]
Root: "HKA"; Subkey: "{code:GetEnvironmentKey}"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}\bin"; Check: NeedsAddPathHKA(ExpandConstant('{app}\bin'))
Root: "HKA"; Subkey: "{code:GetEnvironmentKey}"; ValueType: string; ValueName: "POSH_THEMES_PATH"; ValueData: {app}\themes; Flags: preservestringtype
Root: "HKA"; Subkey: "{code:GetEnvironmentKey}"; ValueType: string; ValueName: "POSH_INSTALLER"; ValueData: {param:installer|manual}; Flags: preservestringtype

[Code]
function GetEnvironmentKey(Param: string): string;
begin
  if IsAdminInstallMode then
    Result := 'System\CurrentControlSet\Control\Session Manager\Environment'
  else
    Result := 'Environment';
end;

function NeedsAddPathHKA(Param: string): boolean;
var
    OrigPath: string;
begin
    if not RegQueryStringValue(HKA, GetEnvironmentKey(''), 'Path', OrigPath)
    then begin
        Result := True;
        exit;
    end;
    // look for the path with leading and trailing semicolon
    // Pos() returns 0 if not found
    Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;
