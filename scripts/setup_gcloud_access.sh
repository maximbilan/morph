# Description: This script sets up the gcloud access for the service account

PROJECT_ID=$MORPH_PROJECT_ID
SERVICE_ACCOUNT=$MORPH_SERVICE_ACCOUNT
PROJECT_NUMBER=$MORPH_PROJECT_NUMBER
REGION=$MORPH_SERVER_REGION
POOL_ID="github-morph-pool"
POOL_NAME="GitHubCI"
LOCATION="global"
PROVIDER_ID="github-morth-provider"
PROVIDER_NAME="GitHub Provider"

gcloud services enable iam.googleapis.com

gcloud iam workload-identity-pools create $POOL_ID \
    --project=$PROJECT_ID \
	--location=$LOCATION \
	--display-name=$POOL_NAME

gcloud iam workload-identity-pools providers create-oidc $PROVIDER_ID \
    --project=$PROJECT_ID \
    --location=$LOCATION \
    --workload-identity-pool=$POOL_ID \
    --display-name="$PROVIDER_NAME" \
    --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.aud=assertion.aud,attribute.repository_owner=assertion.repository_owner" \
    --attribute-condition="attribute.repository == assertion.repository" \
    --issuer-uri="https://token.actions.githubusercontent.com"

gcloud iam workload-identity-pools providers describe $PROVIDER_ID \
    --project=$PROJECT_ID \
    --location=$LOCATION \
    --workload-identity-pool=$POOL_ID \
    --format="value(name)"

PRINCIPAL_SET="principalSet://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/$LOCATION/workloadIdentityPools/$POOL_ID/attribute.repository/maximbilan/morph"

gcloud iam service-accounts add-iam-policy-binding $SERVICE_ACCOUNT \
  --project=$PROJECT_ID \
  --role="roles/iam.workloadIdentityUser" \
  --member=$PRINCIPAL_SET

MEMBER_NAME="serviceAccount:$SERVICE_ACCOUNT"

gcloud run services add-iam-policy-binding handler \
    --region=$REGION \
    --member=$MEMBER_NAME \
    --role='roles/run.invoker'
