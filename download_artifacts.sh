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

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    gh run download $LAST_SVM_WORKFLOW_RUN_ID --name bins-Linux-release --repo spacemeshos/svm
elif [[ "$OSTYPE" == "darwin"* ]]; then
    gh run download $LAST_SVM_WORKFLOW_RUN_ID --name bins-macOS-release --repo spacemeshos/svm
else
    gh run download $LAST_SVM_WORKFLOW_RUN_ID --name bins-Windows-release --repo spacemeshos/svm
fi

chmod a+x svm-cli

exit 0
