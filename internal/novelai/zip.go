package novelai

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
)

type ImageData struct {
	Filename string
	Bytes    []byte
	MIMEType string
}

func ExtractImagesFromZip(blob []byte) ([]ImageData, error) {
	zr, err := zip.NewReader(bytes.NewReader(blob), int64(len(blob)))
	if err != nil {
		return nil, err
	}
	var out []ImageData
	for _, file := range zr.File {
		if !strings.HasPrefix(file.Name, "image") {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		out = append(out, ImageData{Filename: file.Name, Bytes: data, MIMEType: "image/png"})
	}
	return out, nil
}
