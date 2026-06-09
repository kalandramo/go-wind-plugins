git tag cache/v0.0.1 --force
git tag circuitbreaker/v0.0.1 --force
git tag config/v0.0.1 --force
git tag encoding/v0.0.1 --force
git tag log/v0.0.1 --force
git tag metrics/v0.0.1 --force
git tag ratelimit/v0.0.1 --force
# OSS modules have no shared interface; each sub-module is self-contained
# git tag oss/v0.0.1 --force
# git tag oss/minio/v0.0.1 --force
# git tag oss/s3/v0.0.1 --force
git tag registry/v0.0.1 --force
git tag tracer/v0.0.1 --force
git tag workflow/v0.0.1 --force

git push origin --tags