#!/usr/bin/env bash

rm -rf svm/artifacts/
mkdir -p svm/artifacts/
touch svm/artifacts/.gitkeep

echo "Fetching the last successful SVM workflow run..."

# The workflow runs are formatted as a table. The first line is occupied by
# table headers, so we're only interested in the second line; more specifically,
# the third-to-last column of the second line.
LAST_SVM_WORKFLOW_RUN_ID=`gh run list --repo spacemeshos/svm --limit 1 | head -1 | rev | cut -f3 | rev`

echo "Done. Now downloading the artifacts. This might take up to a few minutes..."

cd svm/artifacts/
gh run download $LAST_SVM_WORKFLOW_RUN_ID --repo spacemeshos/svm

cp bins-Linux-release/svm.h ..
cp bins-Linux-release/libsvm.so .
cp bins-macOS-release/libsvm.dylib .
cp bins-Windows-release/svm.dll .

exit 0
