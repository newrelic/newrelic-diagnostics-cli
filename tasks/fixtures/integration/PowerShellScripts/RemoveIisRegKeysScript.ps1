Remove-ItemProperty -Path HKLM:\SYSTEM\CurrentControlSet\Services\WAS -Name Environment
Remove-ItemProperty -Path HKLM:\SYSTEM\CurrentControlSet\Services\W3SVC -Name Environment
./nrDiag_x64.exe