package novelai

import (
	"testing"

	"novelai/internal/testutil"
)

func TestExtractImagesFromZip(t *testing.T) {
	blob := testutil.BuildTestImageZip(t, map[string][]byte{"image_0.png": []byte("png")})
	images, err := ExtractImagesFromZip(blob)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 {
		t.Fatalf("len = %d", len(images))
	}
}
