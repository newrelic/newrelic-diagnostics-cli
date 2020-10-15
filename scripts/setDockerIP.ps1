$ip = Get-NetIPAddress -InterfaceAlias "vEthernet (HNS Internal NIC)" -AddressFamily IPv4
$docker = $ip.IPAddress+":2375"
$docker
# This returns the docker value which is set to DOCKER_HOST via the bat script command
# for /f %%i in ('powershell .\scripts\setDockerIP.ps1') do set DOCKER_HOST=%%i
