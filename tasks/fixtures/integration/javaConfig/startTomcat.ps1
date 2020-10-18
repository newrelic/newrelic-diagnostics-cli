Start-Process -Filepath 'C:/Program Files/Apache Software Foundation/tomcat/apache-tomcat-8.5.12/bin/catalina.bat' -ArgumentList 'start' -NoNewWindow -PassThru 
sleep 30
.\nrdiag.exe 