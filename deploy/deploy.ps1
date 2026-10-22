# Deploys the website to your server from this Windows computer.
#
#   First time:   .\deploy\deploy.ps1 -Server ubuntu@1.2.3.4 -Key $HOME\.ssh\oracle.key -Domain beyondexpeditions.in -Setup
#   Every update: .\deploy\deploy.ps1 -Server ubuntu@1.2.3.4 -Key $HOME\.ssh\oracle.key
#
# It rebuilds the frontend, builds the Linux program, uploads it and restarts the
# services. It never touches the database or photos uploaded on the server.
#
# -Arch: amd64 for Oracle's "AMD" / E2.1.Micro machines, arm64 for "Ampere" machines.
# -Photos: also upload the site photos (hero, gallery, logo). Do this the first time,
#          and whenever you add photos locally. Lead and review photos are never sent.
param(
  [Parameter(Mandatory = $true)][string]$Server,
  [Parameter(Mandatory = $true)][string]$Key,
  [string]$Domain = "",
  [ValidateSet("amd64", "arm64")][string]$Arch = "amd64",
  [switch]$Setup,
  [switch]$Photos,
  [switch]$SkipFrontend
)
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
$env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")

function Run($what, [scriptblock]$cmd) {
  Write-Host "==> $what" -ForegroundColor Green
  & $cmd
  if ($LASTEXITCODE -ne 0) { throw "Failed: $what" }
}
$sshOpts = @("-i", $Key, "-o", "StrictHostKeyChecking=accept-new")

if ($Setup) {
  if (-not $Domain) { throw "-Setup needs -Domain, e.g. -Domain beyondexpeditions.in" }
  Run "Upload server setup files" { ssh @sshOpts $Server "rm -rf /tmp/beyond-deploy"; scp @sshOpts -r "$root\deploy" "${Server}:/tmp/beyond-deploy" }
  Run "Run server setup (installs Caddy, firewall, services)" { ssh @sshOpts $Server "sudo DOMAIN=$Domain bash /tmp/beyond-deploy/server-setup.sh" }
}

if (-not $SkipFrontend) {
  Run "Build the frontend" { Push-Location "$root\frontend"; try { npm install --no-audit --no-fund; npm run build } finally { Pop-Location } }
}

Run "Build the Linux program ($Arch)" {
  $env:GOOS = "linux"; $env:GOARCH = $Arch; $env:CGO_ENABLED = "0"
  try { go build -trimpath -ldflags "-s -w" -o "$root\beyond-linux" . } finally { Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue }
}

if ($Photos) {
  $tgz = Join-Path $env:TEMP "beyond-photos.tgz"
  Run "Pack site photos (without lead/review uploads)" { tar -czf $tgz -C "$root\photos" --exclude=leads --exclude=reviews . }
  Run "Upload photos" { scp @sshOpts $tgz "${Server}:/tmp/beyond-photos.tgz" }
  Run "Unpack photos on the server" { ssh @sshOpts $Server "sudo tar -xzf /tmp/beyond-photos.tgz -C /opt/beyond/photos && sudo chown -R beyond:beyond /opt/beyond/photos && rm /tmp/beyond-photos.tgz" }
}

Run "Upload the program" { scp @sshOpts "$root\beyond-linux" "${Server}:/tmp/beyond.new" }
Run "Install and restart" {
  ssh @sshOpts $Server "sudo install -o beyond -g beyond -m 755 /tmp/beyond.new /opt/beyond/beyond && rm /tmp/beyond.new && sudo systemctl restart beyond-public beyond-admin && sleep 2 && systemctl is-active beyond-public beyond-admin"
}
Remove-Item "$root\beyond-linux" -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Done. If this was the first deploy, read the generated admin password with:" -ForegroundColor Yellow
Write-Host "  ssh -i $Key $Server `"sudo journalctl -u beyond-public --no-pager | grep -i password`""
