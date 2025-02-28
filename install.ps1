# Проверяем и создаем директорию для бинарных файлов
$binPath = Join-Path $env:USERPROFILE "bin"
if (-not (Test-Path $binPath)) {
    Write-Host "Creating bin directory at $binPath"
    New-Item -ItemType Directory -Force -Path $binPath
}

# Компилируем приложение
Write-Host "Building freelancy..."
go build -o freelancy.exe

# Копируем исполняемый файл
Write-Host "Installing freelancy to $binPath"
Copy-Item -Force -Path "freelancy.exe" -Destination (Join-Path $binPath "freelancy.exe")

# Добавляем путь в PATH если его там нет
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$binPath*") {
    Write-Host "Adding bin directory to PATH"
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$binPath", "User")
    $env:Path = "$env:Path;$binPath"
}

Write-Host "`nInstallation complete!"
Write-Host "You can now run 'freelancy' from any terminal."
Write-Host "Note: You may need to restart your terminal for PATH changes to take effect." 