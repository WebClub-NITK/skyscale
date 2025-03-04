curl -Lo firectl https://firectl-release.s3.amazonaws.com/firectl-v0.1.0
curl -Lo firectl.sha256 https://firectl-release.s3.amazonaws.com/firectl-v0.1.0.sha256
sha256sum -c firectl.sha256
chmod +x firectl