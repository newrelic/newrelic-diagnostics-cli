$HKLM = 2147483650 #HKEY_LOCAL_MACHINE
$reg = [wmiclass]"\\.\root\default:StdRegprov"
$key = "SYSTEM\CurrentControlSet\Services\W3SVC"
$name = "Environment"
$value = "COR_PROFILER={71DA0A04-7777-4EC6-9643-7D28B46A8A41}","NEWRELIC_INSTALL_PATH=C:\Program Files\New Relic\.NET Agent\"
$reg.SetMultiStringValue($HKLM, $key, $name, $value)
$key = "SYSTEM\CurrentControlSet\Services\WAS"
$name = "Environment"
$value = "COR_ENABLE_PROFILING=1","COR_PROFILER={71DA0A04-7777-4EC6-9643-7D28B46A8A41}"
$reg.SetMultiStringValue($HKLM, $key, $name, $value)
./nrDiag_x64.exe

