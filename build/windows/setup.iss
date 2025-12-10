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
DefaultGroupName=Indico Data Fusion
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
Source: "..\bin\idf.exe"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: ".\icon.ico"; DestDir: "{app}"; Flags: ignoreversion

[Run]
// User selected... these files are shown for launch after everything is done
Filename: "{app}\idf.exe"; WorkingDir: "{app}"; Description: "Launch {#MyAppName}"; Flags: postinstall runascurrentuser skipifsilent;

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\idf.exe"; WorkingDir: "{app}"; Comment: "Launch {#MyAppName}"; IconFilename: "{app}\icon.ico"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"; Comment: "Remove {#MyAppName}"
