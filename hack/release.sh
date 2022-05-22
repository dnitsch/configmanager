#!/usr/bin/env bash

id=$(curl -X POST -u $OWNER:$PAT -H "Accept: application/vnd.github.v3+json" https://api.github.com/repos/$OWNER/$NAME/releases -d "{\"tag_name\":\"$VERSION\",\"generate_release_notes\":true,\"prerelease\":false}" | jq -r .id)

upload_url=https://uploads.github.com/repos/$OWNER/$NAME/releases/$id/assets

for asset in dist/*; do \
    name=$(echo $asset | cut -c 6-)
    curl -u $OWNER:$PAT -H "Content-Type: application/x-binary" -X POST --data-binary "@$asset" "$upload_url?name=$name"
done
