<#
.SYNOPSIS
    安装 HAP 应用到 HarmonyOS 设备

.DESCRIPTION
    编译生成 HAP 包并安装到连接的手机上。
    HAP 文件路径: entry/build/default/outputs/default/entry-default-signed.hap

.PARAMETER ProjectPath
    项目根目录路径。默认为当前目录。

.PARAMETER SkipBuild
    跳过编译步骤，直接安装已有的 HAP。

.PARAMETER Uninstall
    安装前先卸载旧版本。

.PARAMETER BundleName
    应用包名。默认为 com.example.template。

.EXAMPLE
    .\install_hap.ps1
    .\install_hap.ps1 -ProjectPath "C:\Projects\MyLib"
    .\install_hap.ps1 -SkipBuild
    .\install_hap.ps1 -Uninstall
#>

param(
    [string]$ProjectPath = ".",
    [switch]$SkipBuild,
    [switch]$Uninstall,
    [string]$BundleName = "com.example.template"
)

$ErrorActionPreference = "Stop"

function Write-ColorText {
    param([string]$Text, [string]$Color = "White", [switch]$NoNewline)
    if ($NoNewline) {
        Write-Host $Text -ForegroundColor $Color -NoNewline
    } else {
        Write-Host $Text -ForegroundColor $Color
    }
}

# Switch to project directory
Push-Location $ProjectPath

try {
    $hapPath = "entry\build\default\outputs\default\entry-default-signed.hap"

    Write-Host ""
    Write-ColorText "========================================" "Cyan"
    Write-ColorText "  HarmonyOS HAP Installation Tool" "Cyan"
    Write-ColorText "========================================" "Cyan"
    Write-Host ""

    # Step 1: Check hdc
    Write-ColorText "[Step 1] Checking hdc tool..." "White" -NoNewline
    $hdc = Get-Command hdc -ErrorAction SilentlyContinue
    if (-not $hdc) {
        Write-ColorText " NOT FOUND" "Red"
        Write-ColorText "         Please add hdc to PATH" "Yellow"
        exit 1
    }
    Write-ColorText " OK" "Green"

    # Step 2: Check device connection
    Write-ColorText "[Step 2] Checking device connection..." "White" -NoNewline
    try {
        $targets = & hdc list targets 2>&1
        $targetList = ($targets | Where-Object { $_ -and $_ -ne "[Empty]" -and $_ -notmatch "^\s*$" })
        if (-not $targetList -or $targetList.Count -eq 0) {
            Write-ColorText " NO DEVICE" "Red"
            Write-ColorText "         Please connect device and enable USB debugging" "Yellow"
            exit 1
        }
        Write-ColorText " OK: $($targetList[0])" "Green"
    } catch {
        Write-ColorText " ERROR" "Red"
        Write-ColorText "         Failed to check device: $($_.Exception.Message)" "Yellow"
        exit 1
    }

    # Step 3: Build (optional)
    if (-not $SkipBuild) {
        Write-ColorText "[Step 3] Building HAP..." "Cyan"

        Write-ColorText "  Cleaning old build artifacts..." "Gray"
        try {
            & hvigorw clean 2>&1 | Out-Null
        } catch {
            Write-ColorText "  Warning: Clean failed, continuing..." "Yellow"
        }

        Write-ColorText "  Building (hvigorw assembleHap)..." "Gray"
        $buildOutput = & hvigorw assembleHap 2>&1
        $buildResult = $LASTEXITCODE

        if ($buildResult -ne 0) {
            Write-ColorText "  BUILD FAILED" "Red"
            Write-Host ""
            Write-ColorText "--- Build Error Output ---" "Yellow"
            $buildOutput | Select-Object -Last 30 | ForEach-Object { Write-Host $_ }
            Write-ColorText "--- End of Build Error ---" "Yellow"
            Write-Host ""
            Write-ColorText "Please fix the errors and retry" "Yellow"
            exit 1
        }
        Write-ColorText "  Build successful" "Green"
    } else {
        Write-ColorText "[Step 3] Skipped build (-SkipBuild)" "Gray"
    }

    # Step 4: Check HAP file
    Write-ColorText "[Step 4] Checking HAP file..." "White" -NoNewline
    if (-not (Test-Path $hapPath)) {
        Write-ColorText " NOT FOUND" "Red"
        Write-ColorText "         HAP file does not exist: $hapPath" "Yellow"
        Write-ColorText "         Please run 'hvigorw assembleHap' first" "Yellow"
        exit 1
    }
    $hapSize = (Get-Item $hapPath).Length
    $hapSizeMB = [math]::Round($hapSize / 1MB, 2)
    Write-ColorText " OK ($hapSizeMB MB)" "Green"

    # Step 5: Uninstall old version (optional)
    if ($Uninstall) {
        Write-ColorText "[Step 5] Uninstalling old version ($BundleName)..." "White" -NoNewline
        try {
            & hdc uninstall $BundleName 2>&1 | Out-Null
            Write-ColorText " OK" "Green"
        } catch {
            Write-ColorText " Failed (may not exist)" "Yellow"
        }
    }

    # Step 6: Install
    $stepNum = if ($Uninstall) { "6" } else { "5" }
    Write-ColorText "[Step $stepNum] Installing HAP to device..." "White" -NoNewline
    try {
        $installOutput = & hdc install $hapPath 2>&1
        $installStr = $installOutput -join " "

        if ($installStr -match "success") {
            Write-ColorText " SUCCESS" "Green"
        } else {
            Write-ColorText " FAILED" "Red"
            Write-Host ""
            Write-ColorText "--- Installation Error ---" "Yellow"
            Write-Host $installStr
            Write-ColorText "--- End of Error ---" "Yellow"

            if ($installStr -match "already exist") {
                Write-Host ""
                Write-ColorText "Hint: Old version exists, use -Uninstall to remove it" "Yellow"
                Write-ColorText "  .\install_hap.ps1 -Uninstall" "Cyan"
            }
            if ($installStr -match "signature") {
                Write-Host ""
                Write-ColorText "Hint: Signature verification failed" "Yellow"
                Write-ColorText "      Ensure bundleName is 'com.example.template'" "Yellow"
            }
            exit 1
        }
    } catch {
        Write-ColorText " ERROR" "Red"
        Write-ColorText "         $($_.Exception.Message)" "Yellow"
        exit 1
    }

    Write-Host ""
    Write-ColorText "========================================" "Cyan"
    Write-ColorText "  INSTALLATION COMPLETE" "Green"
    Write-ColorText "  Bundle: $BundleName" "Gray"
    Write-ColorText "========================================" "Cyan"
    Write-Host ""

} catch {
    Write-ColorText "Unexpected error: $($_.Exception.Message)" "Red"
    exit 1
} finally {
    Pop-Location
}
