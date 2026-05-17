
# Git Push Retry Script
# 每隔5分钟尝试git push，直到20:00

$repoPath = "C:\Users\sunxi\Documents\opencode\92-Account-Center"
$endTime = [DateTime]::Today.AddHours(20)

Write-Host "Git Push Retry Script Started at $(Get-Date -Format 'HH:mm:ss')" -ForegroundColor Green
Write-Host "Will retry every 5 minutes until $($endTime.ToString('HH:mm:ss'))" -ForegroundColor Cyan

while ((Get-Date) -lt $endTime) {
    $currentTime = Get-Date -Format "HH:mm:ss"
    Write-Host "[$currentTime] Attempting git push..." -ForegroundColor Yellow

    try {
        Push-Location $repoPath
        $result = git push 2&gt;&amp;1

        if ($LASTEXITCODE -eq 0) {
            Write-Host "[$currentTime] SUCCESS! Git push completed successfully!" -ForegroundColor Green
            Write-Host $result
            Pop-Location
            exit 0
        } else {
            Write-Host "[$currentTime] FAILED: $result" -ForegroundColor Red
        }
    } catch {
        Write-Host "[$currentTime] ERROR: $_" -ForegroundColor Red
    } finally {
        Pop-Location
    }

    # Check if we should wait more
    $now = Get-Date
    if ($now -lt $endTime) {
        $nextAttempt = $now.AddMinutes(5)
        if ($nextAttempt -gt $endTime) {
            $nextAttempt = $endTime
        }
        $waitSeconds = ($nextAttempt - $now).TotalSeconds
        Write-Host "[$currentTime] Waiting $([Math]::Round($waitSeconds, 0)) seconds until next attempt at $($nextAttempt.ToString('HH:mm:ss'))..." -ForegroundColor Cyan
        Start-Sleep -Seconds $waitSeconds
    }
}

Write-Host "[$(Get-Date -Format 'HH:mm:ss')] Reached end time (20:00). Git push not successful. Skipping." -ForegroundColor Magenta
exit 1
