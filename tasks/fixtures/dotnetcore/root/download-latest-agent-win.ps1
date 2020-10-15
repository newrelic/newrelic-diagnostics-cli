# this script downloads the latest dotnetcore agent for windows x64
# saves it as ./newrelic.zip
add-type @"
using System.Net;
using System.Security.Cryptography.X509Certificates;
public class TrustAllCertsPolicy : ICertificatePolicy {
    public bool CheckValidationResult(
        ServicePoint srvPoint, X509Certificate certificate,
        WebRequest request, int certificateProblem) {
        return true;
    }
}
"@
$allProtocols = [System.Net.SecurityProtocolType]'Ssl3,Tls,Tls11,Tls12'
[System.Net.ServicePointManager]::SecurityProtocol = $allProtocols
[System.Net.ServicePointManager]::CertificatePolicy = New-Object TrustAllCertsPolicy

$output = "newrelic.zip"
$baseUrl = "https://download.newrelic.com"
$searchUrl = "https://nr-downloads-main.s3.amazonaws.com/?delimiter=/&prefix=dot_net_agent/core_20/current/"
[xml]$xml = Invoke-WebRequest $searchUrl -UseBasicParsing
$fileName = $xml.ListBucketResult.Contents.Key | Where-Object{$_ -match "dot_net_agent/core_20/current/newrelic-netcore20-agent-win_[0-9]+[.][0-9]+[.][0-9]+[.][0-9]+_x64[.]zip"}
$fileUrl = "$baseUrl/$fileName"
Invoke-WebRequest -Uri $fileUrl -OutFile "$output" -UseBasicParsing