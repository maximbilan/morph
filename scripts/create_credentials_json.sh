#!/bin/bash
# Download the credentials.json file

gcloud iam service-accounts keys create credentials.json \
    --iam-account $MORPH_SERVICE_ACCOUNT