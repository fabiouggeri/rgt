@echo off
SET BUILD_RELEASE=-Drelease 
SET GOAL=

IF /I "%1"=="install" SET GOAL=install
IF /I "%2"=="install" SET GOAL=install
IF /I "%1"=="deploy" SET GOAL=deploy
IF /I "%2"=="deploy" SET GOAL=deploy

IF /I "%1"=="dev" SET BUILD_RELEASE=
IF /I "%2"=="dev" SET BUILD_RELEASE=
IF /I "%1"=="dev" SET BUILD_RELEASE=
IF /I "%2"=="dev" SET BUILD_RELEASE=

IF /I "%GOAL%"=="" SET GOAL=install

echo "====================================="
echo "| Cleaning...                       |"
echo "====================================="
call mvn clean

echo %GOAL%...
echo "====================================="
echo "| Building xHarbour (BCC) 32bits... |"
echo "====================================="
call mvn %GOAL% -Ddistribution %BUILD_RELEASE% -Dcc=bcc -Dclp=xharbour -Dopt=speed
IF NOT "%ERRORLEVEL%"=="0" GOTO ERRO

echo "====================================="
echo "| Building Harbour (MSC) 32bits...  |"
echo "====================================="
call mvn %GOAL% -Ddistribution %BUILD_RELEASE% -Dcc=msc -Darch=32 -Dclp=harbour -Dopt=speed
IF NOT "%ERRORLEVEL%"=="0" GOTO ERRO

echo "====================================="
echo "| Building Harbour (Mingw) 32bits...|" 
echo "====================================="
call mvn %GOAL% -Ddistribution %BUILD_RELEASE% -Dcc=mingw -Darch=32 -Dclp=harbour -Dopt=speed
IF NOT "%ERRORLEVEL%"=="0" GOTO ERRO

rem biblioteca ming64 e compativel com clang
rem echo "====================================="
rem echo "| Building Harbour (CLang) 32bits...|" 
rem echo "====================================="
rem call mvn %GOAL% -Ddistribution %BUILD_RELEASE% -Dcc=clang -Darch=32 -Dclp=harbour
rem IF NOT "%ERRORLEVEL%"=="0" GOTO ERRO

echo "====================================="
echo "| Building Harbour (MSC) 64bits...  |"
echo "====================================="
call mvn %GOAL% -Ddistribution %BUILD_RELEASE% -Dcc=msc -Dclp=harbour -Dopt=speed
IF NOT "%ERRORLEVEL%"=="0" GOTO ERRO

echo "====================================="
echo "| Building Harbour (Mingw) 64bits...|"
echo "====================================="
call mvn %GOAL% -Ddistribution %BUILD_RELEASE% -Dcc=mingw -Dclp=harbour -Dopt=speed
IF NOT "%ERRORLEVEL%"=="0" GOTO ERRO

rem biblioteca ming64 e compativel com clang
rem echo "====================================="
rem echo "| Building Harbour (CLang) 64bits...|"
rem echo "====================================="
rem call mvn %GOAL% -Ddistribution %BUILD_RELEASE% -Dcc=clang -Dclp=harbour
rem IF NOT "%ERRORLEVEL%"=="0" GOTO ERRO

GOTO FIM

:ERRO
echo "Error building RGT"

:FIM
SET BUILD_RELEASE=
SET GOAL=
@echo on
