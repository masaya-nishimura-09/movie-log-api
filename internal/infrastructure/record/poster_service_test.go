package record

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

func newTestPoster(t *testing.T) record.Poster {
	t.Helper()

	poster, err := record.NewPoster([]byte("\xFF\xD8\xFF"))
	if err != nil {
		t.Fatalf("NewPoster() error = %v", err)
	}
	return poster
}

func objectExists(t *testing.T, key string) bool {
	t.Helper()

	_, err := testS3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true
	}

	var notFound *types.NotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("HeadObject(ctx, %q) error = %v", key, err)
	}
	return false
}

func TestPosterUpload(t *testing.T) {
	t.Run(
		"stores the poster and returns its public URL",
		func(t *testing.T) {
			ps := NewPosterService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)
			poster := newTestPoster(t)

			url, err := ps.Upload(ctx, userID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, url) })

			prefix := testBaseURL + "/1/"
			if !strings.HasPrefix(string(url), prefix) {
				t.Errorf("Upload(ctx, %d, poster) = %q, want prefix %q", userID, url, prefix)
			}
			if !strings.HasSuffix(string(url), ".jpg") {
				t.Errorf("Upload(ctx, %d, poster) = %q, want suffix %q", userID, url, ".jpg")
			}

			key := strings.TrimPrefix(string(url), testBaseURL+"/")
			if !objectExists(t, key) {
				t.Errorf("Upload(ctx, %d, poster) did not store %q in %q", userID, key, testBucket)
			}
		},
	)

	t.Run(
		"returns a different URL for each upload",
		func(t *testing.T) {
			ps := NewPosterService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)
			poster := newTestPoster(t)

			first, err := ps.Upload(ctx, userID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, first) })

			second, err := ps.Upload(ctx, userID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, second) })

			if first == second {
				t.Errorf("Upload(ctx, %d, poster) = %q, want a different URL from %q", userID, second, first)
			}
		},
	)
}

func TestPosterDelete(t *testing.T) {
	t.Run(
		"removes the object when the URL points to the bucket",
		func(t *testing.T) {
			ps := NewPosterService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)
			poster := newTestPoster(t)

			url, err := ps.Upload(ctx, userID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			key := strings.TrimPrefix(string(url), testBaseURL+"/")

			if err := ps.Delete(ctx, url); err != nil {
				t.Fatalf("Delete(ctx, %q) error = %v", url, err)
			}
			if objectExists(t, key) {
				t.Errorf("Delete(ctx, %q) did not remove %q from %q", url, key, testBucket)
			}
		},
	)

	t.Run(
		"does nothing when the URL is not in the bucket",
		func(t *testing.T) {
			ps := NewPosterService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			url := record.PosterURL("https://image.tmdb.org/t/p/w500/abc.jpg")

			if err := ps.Delete(ctx, url); err != nil {
				t.Errorf("Delete(ctx, %q) error = %v", url, err)
			}
		},
	)
}
