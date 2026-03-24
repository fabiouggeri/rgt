@echo off
set JAVA_HOME=%~dp0\java
set PATH=%~dp0\java\bin;%PATH%
start javaw -cp ./rgt-admin-client.jar org.rgt.gui.MainWindow