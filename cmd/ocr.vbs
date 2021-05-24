Set WshShell = CreateObject("WScript.Shell")
WshShell.Run chr(34) & "C:\Glnt\GlntSetup\GlntProxySvr.exe" & Chr(34), 0
Set WshShell = Nothing