[Setup]
AppId={{B8B2E5D4-7A9E-4E0F-A2C5-8F1D9A7C1234}
AppName=Guardar Partidas
AppVersion=1.0.1
AppPublisher=Between Bytes Software
AppPublisherURL=https://between-bytes-software.com
AppSupportURL=https://between-bytes-software.com
AppUpdatesURL=https://between-bytes-software.com
DefaultDirName={autopf32}\Between Bytes Software\Guardar Partidas
DefaultGroupName=Between Bytes Software
OutputDir=output
OutputBaseFilename=guardar-partidas
SetupIconFile=icono.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
UninstallDisplayIcon={app}\guardar-partidas.exe
ArchitecturesInstallIn64BitMode=x64compatible
DisableProgramGroupPage=yes

[Languages]
Name: "spanish"; MessagesFile: "compiler:Languages\Spanish.isl"

[Tasks]
Name: "desktopicon"; Description: "Crear acceso directo en el escritorio"; GroupDescription: "Accesos directos:"; Flags: unchecked

[Files]
Source: "guardar-partidas.exe"; DestDir: "{app}"; DestName: "guardar-partidas.exe"; Flags: ignoreversion
Source: ".env"; DestDir: "{app}"; Flags: ignoreversion
Source: "icono.ico"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Guardar Partidas"; Filename: "{app}\guardar-partidas.exe"; IconFilename: "{app}\icono.ico"; WorkingDir: "{app}"
Name: "{autodesktop}\Guardar Partidas"; Filename: "{app}\guardar-partidas.exe"; IconFilename: "{app}\icono.ico"; WorkingDir: "{app}"; Tasks: desktopicon

[Run]
Filename: "{app}\guardar-partidas.exe"; Description: "Ejecutar Guardar Partidas"; Flags: nowait postinstall skipifsilent runhidden

[UninstallDelete]
Type: files; Name: "{app}\.env"
Type: files; Name: "{app}\data.json"