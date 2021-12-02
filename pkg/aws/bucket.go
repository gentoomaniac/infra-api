package aws

import (
	"encoding/json"
	"net/http"
	"time"

	"cuelang.org/go/pkg/strconv"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Bucket struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creationDate"`
}

func ListBuckets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	region := vars["region"]
	limit, err := strconv.Atoi(r.FormValue("limit"))

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed connecting to AWS")
		return
	}

	// Create S3 service client
	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Error().Err(err).Msg("unable to list buckets")
	}

	var buckets []Bucket
	for i, b := range result.Buckets {
		buckets = append(buckets, Bucket{
			Name:         aws.StringValue(b.Name),
			CreationDate: aws.TimeValue(b.CreationDate),
		})
		if limit != 0 && limit > i {
			break
		}
	}
	json.NewEncoder(w).Encode(buckets)
}
