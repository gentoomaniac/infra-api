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
	limit, _ := strconv.Atoi(r.FormValue("limit"))

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

func CreateBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(vars["region"])},
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed connecting to AWS")
		return
	}
	// Create S3 service client
	svc := s3.New(sess)

	bucket, err := getBucket(svc, vars["bucket"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed checking if bucket exists")
		return
	}
	if bucket != nil {
		http.Error(w, "bucket already exists", http.StatusConflict)
		return
	}

	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(vars["bucket"]),
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed creating bucket")
		return
	}

	svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(vars["bucket"]),
	})

	bucket, err = getBucket(svc, vars["bucket"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Error().Err(err).Msg("unable to find new bucket")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bucket)
}

func getBucket(svc *s3.S3, name string) (*Bucket, error) {
	result, err := svc.ListBuckets(nil)
	if err != nil {
		log.Error().Err(err).Msg("unable to list buckets")
		return nil, err
	}
	var bucket *Bucket
	for _, b := range result.Buckets {
		if aws.StringValue(b.Name) == name {
			bucket = &Bucket{
				Name:         aws.StringValue(b.Name),
				CreationDate: aws.TimeValue(b.CreationDate),
			}
			break
		}
	}
	return bucket, nil
}
