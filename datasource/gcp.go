package datasource

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	cloudspanner "cloud.google.com/go/spanner"
	"github.com/HMasataka/tbls/drivers/bq"
	"github.com/HMasataka/tbls/drivers/spanner"
	"github.com/HMasataka/tbls/schema"
	"google.golang.org/api/option"
)

// AnalyzeBigquery analyze `bq://`
func AnalyzeBigquery(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	ctx := context.Background()
	client, projectID, datasetID, err := NewBigqueryClient(ctx, urlstr)
	if err != nil {
		return s, err
	}
	defer client.Close()

	s.Name = fmt.Sprintf("%s:%s", projectID, datasetID)
	driver, err := bq.New(ctx, client, datasetID)
	if err != nil {
		return s, err
	}
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// NewBigqueryClient returns new bigquery.Client
func NewBigqueryClient(ctx context.Context, urlstr string) (*bigquery.Client, string, string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, "", "", err
	}

	splitted := strings.Split(u.Path, "/")
	projectID := u.Host
	datasetID := splitted[1]

	values := u.Query()
	if err := setEnvGoogleApplicationCredentials(values); err != nil {
		return nil, "", "", err
	}
	var client *bigquery.Client
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON") != "" {
		client, err = bigquery.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"))))
	} else {
		client, err = bigquery.NewClient(ctx, projectID)
	}
	return client, projectID, datasetID, err
}

// AnalyzeSpanner analyze `spanner://`
func AnalyzeSpanner(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	ctx := context.Background()
	client, db, err := NewSpannerClient(ctx, urlstr)
	defer client.Close()

	s.Name = db
	driver, err := spanner.New(ctx, client)
	if err != nil {
		return s, err
	}
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// NewSpannerClient returns new cloudspanner.Client
func NewSpannerClient(ctx context.Context, urlstr string) (*cloudspanner.Client, string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, "", err
	}

	splitted := strings.Split(u.Path, "/")
	projectID := u.Host
	instanceID := splitted[1]
	databaseID := splitted[2]
	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)

	values := u.Query()
	if err := setEnvGoogleApplicationCredentials(values); err != nil {
		return nil, "", err
	}
	var client *cloudspanner.Client
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON") != "" {
		client, err = cloudspanner.NewClient(ctx, db, option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"))))
	} else {
		client, err = cloudspanner.NewClient(ctx, db)
	}
	if err != nil {
		return nil, "", err
	}
	return client, db, nil
}

func setEnvGoogleApplicationCredentials(values url.Values) error {
	keys := []string{
		"google_application_credentials",
		"credentials",
		"creds",
	}
	for _, k := range keys {
		if values.Get(k) != "" {
			return os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", values.Get(k))
		}
	}
	return nil
}
