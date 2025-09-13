param(
  [string]$Output = ("project_snapshot_{0}.txt" -f (Get-Date).ToString("yyyy-MM-dd_HH-mm-ss")),
  [int]$MaxFileKB = 1024,         # cap per file; bump if needed
  [switch]$SkipEnv                # use -SkipEnv to exclude .env files
)

$ErrorActionPreference = "SilentlyContinue"
$ProgressPreference = "SilentlyContinue"
$root = (Get-Location).Path

# Exclusions (paths)
$excludePathRegexes = @(
  '\\\.git\\', '\\node_modules\\', '\\dist\\', '\\build\\', '\\coverage\\',
  '\\tmp\\', '\\\.vscode\\', '\\\.idea\\', '\\\.cache\\', '\\\.pnpm-store\\',
  '\\pgdata\\', '\\web\\\.vite\\', '\\server\\tmp\\'
)

# Exclusions (filenames)
$excludeNames = @(
  'export-project.ps1',           # this script
  '.DS_Store', 'Thumbs.db', 'desktop.ini',
  '.keep',
  'pnpm-lock.yaml','package-lock.json','yarn.lock',
  'go.sum','go.work.sum'
)

# Env files: include by default; skip only if -SkipEnv is passed
$envNames = @('.env', '.env.local', '.env.production', '.env.development')

# Binary-ish extensions
$binaryExt = @(
  '.png','.jpg','.jpeg','.gif','.ico','.exe','.dll','.so','.class',
  '.jar','.pdf','.zip','.gz','.7z','.tar','.woff','.woff2','.ttf','.eot',
  '.mp3','.mp4','.avi','.mov','.psd','.ai'
)

function Is-ExcludedPath([string]$path) {
  $p = $path -replace '/', '\'
  foreach ($re in $excludePathRegexes) { if ($p -imatch $re) { return $true } }
  return $false
}
function Rel($full) { return $full.Substring($root.Length + 1) }

# Gather candidate files
$files = Get-ChildItem -Recurse -File |
  Where-Object {
    -not (Is-ExcludedPath $_.FullName) -and
    -not ($binaryExt -contains $_.Extension.ToLower()) -and
    ($_.Length -lt ($MaxFileKB * 1KB)) -and
    -not ($excludeNames -contains $_.Name) -and
    ($SkipEnv -or $envNames -contains $_.Name -or -not ($envNames -contains $_.Name))
  } |
  Sort-Object FullName

# Ensure env inclusion (even if found later)
if (-not $SkipEnv) {
  $envFiles = Get-ChildItem -Recurse -File | Where-Object {
    -not (Is-ExcludedPath $_.FullName) -and ($envNames -contains $_.Name)
  }
  $files = @($files + $envFiles) | Sort-Object FullName -Unique
}

# Header
"### PROJECT SNAPSHOT" | Out-File $Output -Encoding UTF8
("Generated: {0}" -f (Get-Date).ToString("yyyy-MM-dd HH:mm:ss zzz")) | Out-File $Output -Append -Encoding UTF8
("Root: {0}" -f $root) | Out-File $Output -Append -Encoding UTF8

# Structure (filtered)
"`r`n--- PROJECT STRUCTURE (filtered) ---`r`n" | Out-File $Output -Append -Encoding UTF8
$dirs = $files | ForEach-Object { Split-Path $_.FullName -Parent } | Sort-Object -Unique
$dirs = @($root) + ($dirs | Where-Object { $_ -ne $root })
foreach ($d in $dirs) {
  $relDir = if ($d -eq $root) { "." } else { Rel $d }
  ("[DIR] {0}" -f $relDir) | Out-File $Output -Append -Encoding UTF8
  Get-ChildItem $d -File | Where-Object { $files.FullName -contains $_.FullName } |
    Sort-Object Name | ForEach-Object {
      ("  - {0}" -f $_.Name) | Out-File $Output -Append -Encoding UTF8
    }
}

# Contents
"`r`n--- FILE CONTENTS (filtered) ---`r`n" | Out-File $Output -Append -Encoding UTF8
foreach ($f in $files) {
  $rel = Rel $f.FullName
  "`r`n===== BEGIN FILE: $rel =====`r`n" | Out-File $Output -Append -Encoding UTF8
  try { Get-Content -Path $f.FullName -Raw -Encoding UTF8 | Out-File $Output -Append -Encoding UTF8 }
  catch { "[[Skipped: unreadable or binary]]" | Out-File $Output -Append -Encoding UTF8 }
  "`r`n===== END FILE =====`r`n" | Out-File $Output -Append -Encoding UTF8
}

"Done. Wrote: $Output" | Write-Host
