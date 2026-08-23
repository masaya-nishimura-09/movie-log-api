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

func newTestPoster(t *testing.T, data []byte) record.Poster {
	t.Helper()

	poster, err := record.NewPoster(data)
	if err != nil {
		t.Fatalf("NewPoster(len=%d) error = %v", len(data), err)
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
			poster := newTestPoster(t, []byte("\xFF\xD8\xFF"))

			url, err := ps.Upload(ctx, userID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, userID, url) })

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
			poster := newTestPoster(t, []byte("\xFF\xD8\xFF"))

			first, err := ps.Upload(ctx, userID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, userID, first) })

			second, err := ps.Upload(ctx, userID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, userID, second) })

			if first == second {
				t.Errorf("Upload(ctx, %d, poster) = %q, want a different URL from %q", userID, second, first)
			}
		},
	)

	t.Run(
		"uses the extension that matches the content type",
		func(t *testing.T) {
			ps := NewPosterService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)

			png, err := ps.Upload(ctx, userID, newTestPoster(t, []byte("\x89PNG\x0D\x0A\x1A\x0A")))
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, userID, png) })
			if !strings.HasSuffix(string(png), ".png") {
				t.Errorf("Upload(ctx, %d, poster) = %q, want suffix %q", userID, png, ".png")
			}

			webp, err := ps.Upload(ctx, userID, newTestPoster(t, []byte("RIFF____WEBPVP")))
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, userID, webp) })
			if !strings.HasSuffix(string(webp), ".webp") {
				t.Errorf("Upload(ctx, %d, poster) = %q, want suffix %q", userID, webp, ".webp")
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
			poster := newTestPoster(t, []byte("\xFF\xD8\xFF"))

			url, err := ps.Upload(ctx, userID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", userID, err)
			}
			key := strings.TrimPrefix(string(url), testBaseURL+"/")

			if err := ps.Delete(ctx, userID, url); err != nil {
				t.Fatalf("Delete(ctx, %d, %q) error = %v", userID, url, err)
			}
			if objectExists(t, key) {
				t.Errorf("Delete(ctx, %d, %q) did not remove %q from %q", userID, url, key, testBucket)
			}
		},
	)

	t.Run(
		"does nothing when the URL is not in the bucket",
		func(t *testing.T) {
			ps := NewPosterService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)
			url := record.PosterURL("https://image.tmdb.org/t/p/w500/abc.jpg")

			if err := ps.Delete(ctx, userID, url); err != nil {
				t.Errorf("Delete(ctx, %d, %q) error = %v", userID, url, err)
			}
		},
	)

	t.Run(
		"does nothing when the URL belongs to another user",
		func(t *testing.T) {
			ps := NewPosterService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			ownerID := user.ID(1)
			otherID := user.ID(2)
			poster := newTestPoster(t, []byte("\xFF\xD8\xFF"))

			url, err := ps.Upload(ctx, ownerID, poster)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, poster) error = %v", ownerID, err)
			}
			t.Cleanup(func() { _ = ps.Delete(ctx, ownerID, url) })
			key := strings.TrimPrefix(string(url), testBaseURL+"/")

			if err := ps.Delete(ctx, otherID, url); err != nil {
				t.Fatalf("Delete(ctx, %d, %q) error = %v", otherID, url, err)
			}
			if !objectExists(t, key) {
				t.Errorf("Delete(ctx, %d, %q) removed %q from %q", otherID, url, key, testBucket)
			}
		},
	)
}
