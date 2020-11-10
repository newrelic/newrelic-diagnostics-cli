# Integration Tests

We require that contributors provide test coverage for any new code added to the project. 

We strongly recommend [unit tests](./Unit-Testing.md) over integration tests.

That said, a small number of test cases are best suited for broader end-to-end integration tests.
If you believe you've encountered a case in which integration tests are necessary, please include your reasoning either in a comment in the test or as part of your PR.

-------------------------------------------

The Diagnostics CLI has an integration test structure that allows us to programmatically: 

1. Build Dockerfiles from code 
2. Build a docker image
3. Run that docker image and extract the log output, both from your dev environment OR from within a docker image.

When you run your own integration test, the test Dockerfile structure is built from a public image and includes a freshly built nrdiag binary in /app.

Your task's test coverage should come from [unit tests](unit-testing.md). But if you decide to write an integration test, we'll provide detailed instructions.

### Running integration tests

You can run our current integrations test locally by going into our scripts directory and running 

`./integrationTest.sh` for Linux and Darwin, or `powershell .\integrationTest_windows.ps1` for Windows. Those scripts will invoke our `integration_test.go` test file, which will do the job of finding all  
the `integrationTests.yml` or `integrationTests_windows.yml` files, depending of your operating system.

Each `integrationTests.yml` or `integrationTests_windows.yml` can contain multiple tests. A single test looks like this:

```yml
-  test_name: InvalidXML
    dockerfile_lines: 
     - COPY tasks/base/config/fixtures/validate_testdata2 /app/newrelic.config
    log_entry_expected:  
     - "XML syntax error on line 32: element <log> closed by </configuration>"
     - Failure.*Base/Config/Validate
    log_entry_not_expected:
     - Success.*Dotnet/Config/Agent
```

Our `integration_test.go` test file will parse this syntax and write a dockerfile with the appropriate commands.

Running all these tests in a docker container can take a while. That is why we always suggest to run the unit tests before running the integration test, so a failure in unit tests will cause a faster failure.

### Creating your integration test
Go into the `tasks` directory and find your appropriate category and subcategory. If you do not find there already a `integrationTests.yml` or`integrationTests_windows.yml` file, you can create one. 

To add a new integration test simply create an element in the yml file with the following format:

```yml
- test_name: JavaAppServerJbossNotInstalled
  dockerfile_lines: 
   - "COPY ./fixtures/java/newrelic/newrelic.yml /app"
  docker_cmd: ./nrdiag -filter none
  log_entry_expected:  
   - None.*Java/Appserver/JbossEapCheck
```

- test_name - Name of the test. **THIS MUST HAVE NO SPACES IN IT!!!** This will get translated to ci_<testname> (downcased) in the docker image created. Preferably name your tests in a self-documenting style; meaning, a user should be able to predict what your test will do after solely reading the `test_name`. 

**Recommended:** If you are creating multiple tests for a single task, name them with the same prefix. Tests can be run by specifying a partial regex match, and using the same prefix makes this easier to do. For example, the tests named `JavaPermissionsLogFileSetFromYAML` and `JavaPermissionsWithSysPropSettings` can be run by entering into your command line: `./scripts/integrationTest.sh JavaPermissions` and it will exclude any other test. A `.*` will be appended to the regex match string, so that line would run any tests with a `test_name` that matches the expression `JavaPermissions.*`
- dockerfileLines - strings that will be added to the Dockerfile for this test. This MUST follow the standard [Dockerfile syntax](https://docs.docker.com/engine/reference/builder/), primarily should just be files to copy from elsewhere in the project structure. 
All file references should be relative to the root of the project. 

**Example:** copies a XML file

```yml
 -  test_name: InvalidXML
    dockerfile_lines: 
     - COPY tasks/base/config/fixtures/validate_testdata2 /app/newrelic.config
    docker_cmd: ./nrdiag -v
    log_entry_expected:  
     - "XML syntax error on line 32: element <log> closed by </configuration>"
     - Failure.*Base/Config/Validate
    log_entry_not_expected:
     - Success.*Dotnet/Config/Agent
```

- docker_cmd - adds a custom docker CMD command, wrapped in a shell command to execute the desired actions

**Example:** runs the Diagnostics CLI with a proxy

```yml
 -  test_name: ProxyConnect
    dockerfile_lines: 
     - COPY tasks/base/config/fixtures/validate_testdata2 /app/newrelic.config
    docker_cmd: ./nrdiag -v -p http://127.0.0.1:8888
    log_entry_expected:  
     - "XML syntax error on line 32: element <log> closed by </configuration>"
     - Failure.*Base/Config/Validate
    log_entry_not_expected:
     - Success.*Dotnet/Config/Agent
```

- docker_from - Adds a custom docker FROM command to change the source image the docker container is built from. The integration test suite will automatically add the nrdiag binary and set the work directory to /app when using a custom image source.

**Example:** run the Diagnostics CLI from an ubuntu image

```yml
 -  test_name: ProxyConnect
    docker_from: ubuntu:14.04
    dockerfile_lines: 
     - COPY tasks/base/config/fixtures/validate_testdata2 /app/newrelic.config
    log_entry_expected:  
     - "XML syntax error on line 32: element <log> closed by </configuration>"
     - Failure.*Base/Config/Validate
    log_entry_not_expected:
     - Success.*Dotnet/Config/Agent
```


- log_entry_expected - An array of strings that are regex search strings. The screen output of the integration test will be searched using these strings and a failure to match will cause the test to fail. If you need data from the nrdiag-output.json or a file, it's necessary to add display of this output via the CMD directive.

**Example:** finding two different strings in the log output

```yml
    log_entry_expected:
     - "Please check your config file against a YML linter"
     - "[DEBUG] Done executing tasks"

```

- log_entry_not_expected - An array of strings that are regex search strings. If these are found it will FAIL the test. This allows you to look for unexpected cross over like detecting the Ruby agent is installed when running with a java agent config file.

**Example:** mocking a network response

```yml
 -  test_name: MockSuccesssfulS3Upload
    docker_from: ubuntu:14.04
    dockerfile_lines:
     - RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y install lighttpd openssl ca-certificates
     - COPY tasks/fixtures/integration/mockWebServer/lighttpd.conf /etc/lighttpd/lighttpd.conf
     - RUN mkdir /etc/lighttpd/certs && openssl req -nodes -new -x509  -keyout /etc/lighttpd/certs/lighttpd.pem -out /etc/lighttpd/certs/lighttpd.crt -subj "/C=US/ST=California/L=San Francisco/O=New Relic, Inc./CN=*.newrelic.com"
     - RUN cat /etc/lighttpd/certs/lighttpd.crt >> /etc/lighttpd/certs/lighttpd.pem && cp /etc/lighttpd/certs/lighttpd.crt /usr/local/share/ca-certificates/ && update-ca-certificates --fresh
     - COPY tasks/fixtures/ruby/config/newrelic.yml /app/newrelic.yml
     - 'RUN echo "" >> /var/www/pretends3upload'
     - 'RUN mkdir -p /var/www/api/v1/attachments/ && echo "{\"success\":true}" >> /var/www/api/v1/attachments/upload'
     - 'RUN echo "{\"url\" : \"https://nr-supportlanding.s3.amazonaws.com/pretends3upload\"}" >> /var/www/api/v1/attachments/get_download_url'
    docker_cmd: /etc/init.d/lighttpd start && ./nrdiag -y -a 11111 -v
    hosts_file_additions: 
     - support.newrelic.com:127.0.0.1
     - nr-supportlanding.s3.amazonaws.com:127.0.0.1
    log_entry_expected:
     - 'Reponse: {"success":true}'

```

The key here is the use of a lightweight webserver (this example uses lighttpd) and the `hosts_file_additions` to force network calls to go to the locally running webserver (with some static files serving the responses). The first four lines of the dockerfile_lines section above can likely be copied exactly for most use cases as they will build out the basic webserver (including https with generating a SSL certificate and adding it to the local trust store so connections can be made successfully). Then simply create static files in `/var/www/` with the appropriate path and voila, you've mocked out a network call successfully. 




### Manually building/testing integration test docker images
If you aren't quite sure of what you need to add to your integration test's docker image, you can create/run the local docker images used by an Integration test.

To start out, manually run the integration test to build out a base integrationDockerfile and use as the base for the rest of your test by running:
1. `./integrationTest.sh`
2. Copy/edit the resulting integrationDockerfile to add whatever you want to the integration test
3. Test what the output looks like.

You can build an image from the file: 
1. `docker build --rm -t myTest -f integrationDockerfile .`
2. `docker run --rm myTest`

These can be combined into a single command for convenience with 

`docker build --rm -t myTest -f integrationDockerfile && docker run --rm myTest`. 

#### Running the Diagnostics CLI with options
It's worth noting that while the default integration dockerfile does add `CMD ["./nrdiag"]` to run nrdiag without verbose mode, you can run verbose mode by just adding a `CMD ["./nrdiag", "-v"]` to run nrdiag with verbose or some other series of options and it will be executed instead of the default nrdiag execution because it will appear after the first entry of CMD.

