package metadata

import (
	"context"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

// ProjectID returns the project ID of the Cloud Run service. It fetches this
// information from the GCP metadata server.
func ProjectID() (string, error) {
	return metadata.ProjectID()
}

// Region returns the region of the Cloud Run service. It fetches this
// information from the GCP metadata server. The returned value is in the format
// of: projects/PROJECT_NUMBER/regions/REGION.
func Region() (string, error) {
	resp, err := metadata.Get("instance/region")
	if err != nil {
		return "", nil
	}

	return resp, nil
}

// IDToken returns a TokenSource that yields ID tokens. These tokens can be used
// to authenticate requests with the Token.SetAuthHeader method.
func IDToken(ctx context.Context, aud string) (oauth2.TokenSource, error) {
	idtoken.NewTokenSource(ctx, aud)
	return idtoken.NewTokenSource(ctx, aud)
}
