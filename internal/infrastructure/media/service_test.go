package media

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

func newTestMedia(t *testing.T, value []byte) media.Media {
	t.Helper()

	data, err := media.NewData(value)
	if err != nil {
		t.Fatalf("NewData(len=%d) error = %v", len(value), err)
	}
	m, err := media.NewMedia(data)
	if err != nil {
		t.Fatalf("NewMedia(len=%d) error = %v", len(data), err)
	}
	return m
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

func TestUpload(t *testing.T) {
	t.Run(
		"stores the media and returns its public URL",
		func(t *testing.T) {
			s := NewService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)
			m := newTestMedia(t, []byte("\xFF\xD8\xFF"))

			url, err := s.Upload(ctx, userID, m)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, media) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = s.Delete(ctx, userID, url) })

			prefix := testBaseURL + "/1/"
			if !strings.HasPrefix(string(url), prefix) {
				t.Errorf("Upload(ctx, %d, media) = %q, want prefix %q", userID, url, prefix)
			}
			if !strings.HasSuffix(string(url), ".jpg") {
				t.Errorf("Upload(ctx, %d, media) = %q, want suffix %q", userID, url, ".jpg")
			}

			key := strings.TrimPrefix(string(url), testBaseURL+"/")
			if !objectExists(t, key) {
				t.Errorf("Upload(ctx, %d, media) did not store %q in %q", userID, key, testBucket)
			}
		},
	)

	t.Run(
		"returns a different URL for each upload",
		func(t *testing.T) {
			s := NewService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)
			m := newTestMedia(t, []byte("\xFF\xD8\xFF"))

			first, err := s.Upload(ctx, userID, m)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, media) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = s.Delete(ctx, userID, first) })

			second, err := s.Upload(ctx, userID, m)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, media) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = s.Delete(ctx, userID, second) })

			if first == second {
				t.Errorf("Upload(ctx, %d, media) = %q, want a different URL from %q", userID, second, first)
			}
		},
	)

	t.Run(
		"uses the extension that matches the content type",
		func(t *testing.T) {
			s := NewService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)

			png, err := s.Upload(ctx, userID, newTestMedia(t, []byte("\x89PNG\x0D\x0A\x1A\x0A")))
			if err != nil {
				t.Fatalf("Upload(ctx, %d, media) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = s.Delete(ctx, userID, png) })
			if !strings.HasSuffix(string(png), ".png") {
				t.Errorf("Upload(ctx, %d, media) = %q, want suffix %q", userID, png, ".png")
			}

			webp, err := s.Upload(ctx, userID, newTestMedia(t, []byte("RIFF____WEBPVP")))
			if err != nil {
				t.Fatalf("Upload(ctx, %d, media) error = %v", userID, err)
			}
			t.Cleanup(func() { _ = s.Delete(ctx, userID, webp) })
			if !strings.HasSuffix(string(webp), ".webp") {
				t.Errorf("Upload(ctx, %d, media) = %q, want suffix %q", userID, webp, ".webp")
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			s := NewService(testS3Client, testBucket, testBaseURL)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			userID := user.ID(1)
			m := newTestMedia(t, []byte("\xFF\xD8\xFF"))

			got, err := s.Upload(ctx, userID, m)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Upload(ctx, %d, media) (media.URL, error) = %v, %v, want %v",
					userID, got, err, context.Canceled,
				)
			}
			if got != "" {
				t.Errorf(
					"Upload(ctx, %d, media) (media.URL, error) = %v, %v, want %q",
					userID, got, err, "",
				)
			}
		},
	)
}

func TestDelete(t *testing.T) {
	t.Run(
		"removes the object when the URL points to the bucket",
		func(t *testing.T) {
			s := NewService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)
			m := newTestMedia(t, []byte("\xFF\xD8\xFF"))

			url, err := s.Upload(ctx, userID, m)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, media) error = %v", userID, err)
			}
			key := strings.TrimPrefix(string(url), testBaseURL+"/")

			if err := s.Delete(ctx, userID, url); err != nil {
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
			s := NewService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			userID := user.ID(1)
			url := media.URL("https://image.tmdb.org/t/p/w500/abc.jpg")

			if err := s.Delete(ctx, userID, url); err != nil {
				t.Errorf("Delete(ctx, %d, %q) error = %v", userID, url, err)
			}
		},
	)

	t.Run(
		"does nothing when the URL belongs to another user",
		func(t *testing.T) {
			s := NewService(testS3Client, testBucket, testBaseURL)

			ctx := context.Background()
			ownerID := user.ID(1)
			otherID := user.ID(2)
			m := newTestMedia(t, []byte("\xFF\xD8\xFF"))

			url, err := s.Upload(ctx, ownerID, m)
			if err != nil {
				t.Fatalf("Upload(ctx, %d, media) error = %v", ownerID, err)
			}
			t.Cleanup(func() { _ = s.Delete(ctx, ownerID, url) })
			key := strings.TrimPrefix(string(url), testBaseURL+"/")

			if err := s.Delete(ctx, otherID, url); err != nil {
				t.Fatalf("Delete(ctx, %d, %q) error = %v", otherID, url, err)
			}
			if !objectExists(t, key) {
				t.Errorf("Delete(ctx, %d, %q) removed %q from %q", otherID, url, key, testBucket)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			s := NewService(testS3Client, testBucket, testBaseURL)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			userID := user.ID(1)
			url := media.URL(testBaseURL + "/1/00000000-0000-0000-0000-000000000000.jpg")

			err := s.Delete(ctx, userID, url)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Delete(ctx, %d, %q) error = %v, want %v",
					userID, url, err, context.Canceled,
				)
			}
		},
	)
}
