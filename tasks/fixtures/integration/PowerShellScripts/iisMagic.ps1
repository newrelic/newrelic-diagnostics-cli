<#
Import-Module WebAdministration
New-Item 'IIS:\Sites\Default Web Site\DemoApp' -physicalPath 'C:\app\App' -type Application
$Acl = Get-Acl 'C:\app'
$Ar = New-Object  system.security.accesscontrol.filesystemaccessrule("Everyone","FullControl","Allow")
$Acl.SetAccessRule($Ar)
Set-Acl 'C:\app' $Acl
iisreset
$webclient = New-Object Net.WebClient
$webclient.DownloadString("http://localhost:80/DemoApp");
#>

Import-Module WebAdministration
$iisAppPoolName = "TestPool"
$iisAppPoolDotNetVersion = "v4.0"
$iisAppName = "TestApp"
$directoryPath = "C:\app\App"

#navigate to the app pools root
cd IIS:\AppPools\

#check if the app pool exists
if (!(Test-Path $iisAppPoolName -pathType container))
{
    #create the app pool
    $appPool = New-Item $iisAppPoolName
    $appPool | Set-ItemProperty -Name "managedRuntimeVersion" -Value $iisAppPoolDotNetVersion
}

#navigate to the sites root
cd IIS:\Sites\

#check if the site exists
if (Test-Path $iisAppName -pathType container)
{
    return
}

#create the site
$iisApp = New-Item $iisAppName -bindings @{protocol="http";bindingInformation=":8080:"} -physicalPath $directoryPath
$iisApp | Set-ItemProperty -Name "applicationPool" -Value $iisAppPoolName


$Acl = Get-Acl 'C:\app\App'
$Ar = New-Object  system.security.accesscontrol.filesystemaccessrule('IIS_IUSRS','FullControl','ContainerInherit', 'InheritOnly','Allow')
$Ar2 = New-Object  system.security.accesscontrol.filesystemaccessrule('IIS_IUSRS','FullControl','ObjectInherit', 'InheritOnly','Allow')
$Acl.AddAccessRule($Ar)
$Acl.AddAccessRule($Ar2)
Set-Acl 'C:\app\App' $Acl
iisreset
$webclient = New-Object Net.WebClient
$webclient.DownloadString("http://localhost:8080/");
./nrDiag_x64.exe