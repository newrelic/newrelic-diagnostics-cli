## Summary

Each `fixtures/<language>` directory serves as the base directory each theoretical application (i.e. `fixtures/node/server.js`). Relative paths to logs and config files (listed below) are the default configurations. No edge cases yet.

`fixtures/<language>/root` indicates `C:\` in Windows (.NET agent), or the root directory in linux for those files stored on the system outside of the application's directory.

## Languages

### Python

 * **Config:** `python/newrelic.ini`
 * **Log:** `python/root/tmp/newrelic-python-agent.log`

### PHP

* **Config (Agent):** `php/root/etc/php5/conf.d/newrelic.ini`
* **Config (Daemon):** `php/root/etc/newrelic/newrelic.cfg`
* **Log (Agent):** `php/root/var/log/newrelic/php_agent.log`
* **Log (Daemon):** `php/root/var/log/newrelic/newrelic-daemon.log`

_Agent config path is variable based on both PHP version and webserver._

### .NET

* **Config:** `dotnet/root/ProgramData/New\ Relic/.NET\ Agent/newrelic.config`
* **Log (Process Profile):** `dotnet/root/ProgramData/New\ Relic/.NET\ Agent/Logs/NewRelic.Profiler.8675.log`
* **Log (Agent):** `dotnet/root/ProgramData/New\ Relic/.NET\ Agent/Logs/newrelic_agent__LM_W3SVC_3_ROOT.log`

#### Corresponding to:

* `C:\ProgramData\New Relic\.NET Agent\newrelic.config`
* `C:\ProgramData\New Relic\.NET Agent\Logs\NewRelic.Profiler.<pid_num>.log`
* `C:\ProgramData\New Relic\.NET Agent\Logs\newrelic_agent__<metabase_path>.log`

#### Java 

* **Config:** `java/newrelic/newrelic.yml`
* **Log:** `java/newrelic/logs/newrelic_agent.log`

#### Node 

* **Config:** `node/newrelic.js`
* **Log:** `node/newrelic_agent.log`

#### Ruby

* **Config:** `ruby/config/newrelic.yml`
* **Log:** `ruby/log/newrelic_agent.log`

#### Go

* **Config:** `go/main.go`