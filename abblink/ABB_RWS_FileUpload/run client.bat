@echo off
echo Changing to drive D:
d:
if errorlevel 1 (
    echo Failed to change to drive D:
    pause
    exit /b 1
)

echo Changing directory to \csharp\python
cd \csharp\python
if errorlevel 1 (
    echo Failed to change directory
    pause
    exit /b 1
)

echo Running Python script...
python file-upload-client.py 127.0.0.1 12345 d:\ftp\abb4.txt
if errorlevel 1 (
    echo Failed to run Python script
    pause
    exit /b 1
)

pause

