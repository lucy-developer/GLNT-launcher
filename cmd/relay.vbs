Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "javaw" & " -jar -Dspring.profiles.active=staging C:\Glnt\GlntSetup\relay\GLNT-Relay-0.0.1-SNAPSHOT.jar", 0, false
Set WshShell = Nothing