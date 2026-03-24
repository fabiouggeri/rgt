SET RELEASE=-Drelease

IF /I "%1"=="debug" SET RELEASE=

call mvn clean install -Ddistribution %RELEASE%
IF NOT "%ERRORLEVEL%"=="0" GOTO ERRO

GOTO FIM

:ERRO
echo "Error building RGT"

:FIM
