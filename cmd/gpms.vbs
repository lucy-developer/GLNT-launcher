Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "javaw" & " -Dspring.profiles.active=staging -jar C:\Glnt\GlntSetup\gpms\gpms-1.0-SNAPSHOT.jar", 0, false
Set WshShell = Nothing