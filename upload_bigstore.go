// Check the content of source bucket, and move them to destination bucket.

package main

import (
  "flag"
  "fmt"
  "os"

  "golang.org/x/net/context"
  "golang.org/x/oauth2/google"
  storage "google.golang.org/api/storage/v1"
)

const (
  // This scope allows the application full control over resources in Google Cloud Storage
  scope = storage.DevstorageFullControlScope
)

var (
  ProjectID  = flag.String("project", "receiver-1470423672370", "The cloud project ID.")
  DestBucketName = flag.String("dest_bucket", "open_pipeline_bucket", "The name of destination bucket within your project.")
  SourceBucketName = flag.String("source_bucket", "source_bucket_pipeline", "The name of bucket for source files.")
)

func main() {
  flag.Parse()
  if *DestBucketName == "" {
    fmt.Printf("Destination Bucket argument is required. See --help.\n")
    os.Exit(1)
  }
  if *SourceBucketName == "" {
    fmt.Printf("Source Bucket argument is required. See --help.\n")
    os.Exit(1)
  }
  if *ProjectID == "" {
    fmt.Printf("Project argument is required. See --help.\n")
    os.Exit(1)
  }

  // Authentication is provided by the gcloud tool when running locally, and
  // by the associated service account when running on Compute Engine.
  client, err := google.DefaultClient(context.Background(), scope)
  if err != nil {
    fmt.Printf("Unable to get default client: %v \n", err)
    os.Exit(1)
  }
  service, err := storage.New(client)
  if err != nil {
    fmt.Printf("Unable to create storage service: %v\n", err)
    os.Exit(1)
  }

  // Check whether the destination bucket already exists. If not, create one.
  if _, err := service.Buckets.Get(*DestBucketName).Do(); err == nil {
    fmt.Printf("Bucket %s already exists.\n", *DestBucketName)
  } else {
    // Create a bucket.
    if res, err := service.Buckets.Insert(*ProjectID, &storage.Bucket{Name: *DestBucketName}).Do(); err == nil {
      fmt.Printf("Created bucket %v at location %v\n", res.Name, res.SelfLink)
    } else {
      fmt.Printf("Failed creating bucket %s: %v\n", *DestBucketName, err)
      os.Exit(1)
    }
  }

  // Get list all objects in source bucket.
  source_files, err := service.Objects.List(*SourceBucketName).Do()
  if err != nil {
    fmt.Printf("Objects.List failed: %v\n", err)
    os.Exit(1)
  }
  for _, OneItem := range source_files.Items {
    if file_content, err := service.Objects.Get(*SourceBucketName, OneItem.Name).Download(); err == nil {
      // Insert the object into destination bucket.
      object := &storage.Object{Name: OneItem.Name}
      if res, err := service.Objects.Insert(*DestBucketName, object).Media(file_content.Body).Do(); err == nil {
        fmt.Printf("Created object %v at location %v\n", res.Name, res.SelfLink)
      } else {
        fmt.Printf("Objects.Insert failed: %v\n", err)
        os.Exit(1)
      }
    }
  }      
}
