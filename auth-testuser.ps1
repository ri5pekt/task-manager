param(
  [string]$Email = "test@example.com",
  [string]$Password = "test123"
)

# Build login JSON
$login = @{ email=$Email; password=$Password } | ConvertTo-Json -Compress

# Save cookies
curl.exe -s -i -c cookies.txt -b cookies.txt `
  -X POST http://localhost:5173/api/login `
  -H "Content-Type: application/json" `
  --data-binary "$login" | Out-Null

# Verify
$me = curl.exe -s -b cookies.txt http://localhost:5173/api/me | ConvertFrom-Json
Write-Host "Logged in as $($me.name) <$($me.email)>, user_id=$($me.id)"
