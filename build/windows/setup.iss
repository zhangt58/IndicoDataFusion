; SEE THE DOCUMENTATION FOR DETAILS ON CREATING .ISS SCRIPT FILES!

#define MyAppName "IndicoDataFusion"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "FRIB, MSU"
#define MyAppURL "https://github.com/zhangt58/IndicoDataFusion"
#define OutputName "IndicoDataFusion"

[Setup]
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={userappdata}\{#MyAppName}
DefaultGroupName=Data Manager
AllowNoIcons=yes
Compression=lzma2
SolidCompression=yes
PrivilegesRequired=lowest
WizardStyle=modern
OutputDir=.\output
OutputBaseFilename={#OutputName}_{#MyAppVersion}
SetupIconFile=.\icon.ico

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "..\bin\indico-data-fusion.exe"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: ".\icon.ico"; DestDir: "{app}"; Flags: ignoreversion

[Run]
// User selected... these files are shown for launch after everything is done
Filename: "{app}\indico-data-fusion.exe"; WorkingDir: "{app}"; Description: "Launch {#MyAppName}"; Flags: postinstall runascurrentuser runhidden skipifsilent;

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\indico-data-fusion.exe"; WorkingDir: "{app}"; Comment: "Launch {#MyAppName}"; IconFilename: "{app}\icon.ico"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"; Comment: "Remove {#MyAppName}"