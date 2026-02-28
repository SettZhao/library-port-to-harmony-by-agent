<#
.SYNOPSIS
    运行 HarmonyOS 测试用例

.DESCRIPTION
    通过 hdc 在手机上远程执行测试用例：
    hdc shell "aa test -b com.example.template -m entry_test -s unittest OpenHarmonyTestRunner"

.PARAMETER BundleName
    应用包名。默认为 com.example.template。

.PARAMETER ModuleName
    测试模块名。默认为 entry_test。

.PARAMETER ShowLog
    执行测试后是否自动获取 hilog 日志。

.EXAMPLE
    .\run_tests.ps1
    .\run_tests.ps1 -ShowLog
    .\run_tests.ps1 -BundleName "com.example.myapp"
#>

param(
    [string]$BundleName = "com.example.template",
    [string]$ModuleName = "entry_test",
    [switch]$ShowLog
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

Write-Host ""
Write-ColorText "========================================" "Cyan"
Write-ColorText "  HarmonyOS Test Execution" "Cyan"
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
        Write-ColorText "         Please connect device and install app" "Yellow"
        exit 1
    }
    Write-ColorText " OK: $($targetList[0])" "Green"
} catch {
    Write-ColorText " ERROR" "Red"
    Write-ColorText "         Failed to check device: $($_.Exception.Message)" "Yellow"
    exit 1
}

# Step 3: Execute tests
Write-ColorText "[Step 3] Executing test cases..." "Cyan"
Write-ColorText "  Bundle: $BundleName" "Gray"
Write-ColorText "  Module: $ModuleName" "Gray"
Write-Host ""

$testCommand = "aa test -b $BundleName -m $ModuleName -s unittest OpenHarmonyTestRunner"
Write-ColorText "  Command: hdc shell `"$testCommand`"" "Gray"
Write-Host ""
Write-ColorText "--- Test Output ---" "Yellow"

try {
    $testOutput = & hdc shell $testCommand 2>&1
    $testStr = $testOutput -join "`n"

    Write-Host $testStr
} catch {
    Write-ColorText "--- Test Execution Failed ---" "Red"
    Write-ColorText $_.Exception.Message "Red"
    exit 1
}

Write-ColorText "--- End of Test Output ---" "Yellow"
Write-Host ""

# Step 4: Analyze test results
Write-ColorText "[Step 4] Analyzing test results..." "White"

$passCount = 0
$failCount = 0
$errorCount = 0
$totalTests = 0

# Parse test results - multiple formats supported
if ($testStr -match "Tests run:\s*(\d+)") {
    $totalTests = [int]$Matches[1]
}
if ($testStr -match "Passed:\s*(\d+)") {
    $passCount = [int]$Matches[1]
}
if ($testStr -match "Failed:\s*(\d+)") {
    $failCount = [int]$Matches[1]
}
if ($testStr -match "Error:\s*(\d+)") {
    $errorCount = [int]$Matches[1]
}

# Also try alternative formats
$hasPass = $testStr -match "\bPASS\b"
$hasFail = ($testStr -match "\bFAIL\b") -and ($testStr -notmatch "Failed:\s*0")

Write-Host ""
Write-ColorText "========================================" "Cyan"

if ($failCount -gt 0 -or $errorCount -gt 0 -or $hasFail) {
    Write-ColorText "  TEST FAILED" "Red"
    if ($totalTests -gt 0) {
        Write-ColorText "  Passed: $passCount | Failed: $failCount | Error: $errorCount | Total: $totalTests" "Yellow"
    } elseif ($passCount -or $failCount -or $errorCount) {
        Write-ColorText "  Passed: $passCount | Failed: $failCount | Error: $errorCount" "Yellow"
    }
    Write-Host ""
    Write-ColorText "  Please fix errors and rerun tests" "Yellow"
    Write-ColorText "  View detailed log: hdc hilog | Select-String 'test|FAIL|Error'" "Gray"

    # Get hilog error details
    if ($ShowLog) {
        Write-Host ""
        Write-ColorText "--- hilog Error Log ---" "Yellow"
        try {
            & hdc shell "hilog -t test 2>&1" | Select-Object -Last 50
        } catch {
            Write-ColorText "Failed to get hilog" "Red"
        }
        Write-ColorText "--- End of hilog ---" "Yellow"
    }

    exit 1
} elseif ($passCount -gt 0 -or $hasPass) {
    Write-ColorText "  ALL TESTS PASSED" "Green"
    if ($passCount -gt 0) {
        Write-ColorText "  Passed: $passCount test case(s)" "Gray"
    } elseif ($totalTests -gt 0) {
        Write-ColorText "  Total: $totalTests test case(s)" "Gray"
    }
} else {
    Write-ColorText "  CANNOT DETERMINE TEST RESULT" "Yellow"
    Write-ColorText "  Please check the output above" "Yellow"
}

Write-ColorText "========================================" "Cyan"
Write-Host ""
