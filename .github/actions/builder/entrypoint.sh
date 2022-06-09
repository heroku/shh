#!/usr/bin/env bash

set -euo pipefail

git config --global --add safe.directory /github/workspace

if [[ $INPUT_USE_BUILDPACK == "true" ]]; then
    export STACK=$INPUT_STACK
    export GO_INSTALL_PACKAGE_SPEC="./cmd/*"

    # Fetch buildpack
    curl -o buildpack.tgz "$INPUT_BUILDPACK_URL"
    mkdir -p build/pack
    mkdir -p build/cache
    tar -xf buildpack.tgz -C build/pack
    rm buildpack.tgz

    # Run buildpack
    ./build/pack/bin/compile . ./build/cache/
    mkdir -p vendor/cache
    mv ./build/cache ./vendor/cache
    rm -rf build/
fi

# Install AWS CLI
export AWS_REGION=$INPUT_AWS_REGION \
    AWS_ACCESS_KEY_ID=$INPUT_AWS_ACCESS_KEY_ID \
    AWS_SECRET_ACCESS_KEY=$INPUT_AWS_SECRET_ACCESS_KEY

curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip -qq awscliv2.zip
./aws/install
rm -rf aws awscliv2.zip

# Only build for branches
if [[ ! $INPUT_GITHUB_REF == "refs/heads"* ]];
then
    exit 0
fi

# Setup naming
SHORT_SHA=$(git rev-parse --short "$INPUT_GITHUB_SHA")
BRANCH_NAME=$(echo "$INPUT_GITHUB_REF" | sed "s/refs\/heads\///")
ARTIFACT_NAME="$BRANCH_NAME-artifact-sha-$SHORT_SHA"
echo "ARTIFACT_NAME -- $ARTIFACT_NAME"

# Write build artifact file
echo "$INPUT_APP_NAME:$ARTIFACT_NAME" > .build_artifact

# Create tmp tar file
tar -czf /tmp/"$ARTIFACT_NAME".tgz .

# Upload to s3
aws s3 cp /tmp/"$ARTIFACT_NAME".tgz s3://"$INPUT_S3_BUCKET"/app/"$INPUT_APP_NAME"/"$ARTIFACT_NAME".tgz

# Copy to `-latest`
aws s3 cp s3://"$INPUT_S3_BUCKET"/app/"$INPUT_APP_NAME"/"$ARTIFACT_NAME".tgz s3://"$INPUT_S3_BUCKET"/app/"$INPUT_APP_NAME"/"$BRANCH_NAME-latest".tgz
