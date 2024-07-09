package ingress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"

	"github.com/m-horky/insights-client-next/internal/api"
	"github.com/m-horky/insights-client-next/internal/core"
)

var service = api.NewService("api/ingress/v1")

func UploadArchive(archive core.Archive) (*Uploaded, error) {
	formData := new(bytes.Buffer)
	form := multipart.NewWriter(formData)

	archiveHeader := make(textproto.MIMEHeader)
	archiveHeader.Set(
		"Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filepath.Base(archive.Path)),
	)
	archiveHeader.Set("Content-Type", archive.ContentType)

	archiveField, err := form.CreatePart(archiveHeader)
	if err != nil {
		slog.Error("could not create archive field", slog.Any("error", err))
		return nil, err
	}

	archiveDescriptor, err := os.Open(archive.Path)
	if err != nil {
		slog.Error("could not open archive file", slog.Any("error", err))
		return nil, err
	}
	defer archiveDescriptor.Close()

	_, err = io.Copy(archiveField, archiveDescriptor)
	if err != nil {
		slog.Error("could not load archive file", slog.Any("error", err))
		return nil, err
	}

	form.Close()

	params := url.Values{}
	headers := make(map[string][]string)
	headers["Content-Type"] = []string{form.FormDataContentType()}

	response, err := service.MakeRequest("POST", "upload", params, headers, formData)
	if err != nil {
		slog.Error("could not upload archive", slog.Any("error", err))
		return nil, err
	}

	if response.Code/100 != 2 {
		slog.Error("server rejected the archive", slog.Int("status code", response.Code), slog.Any("raw response", response.Data))
		return nil, fmt.Errorf("server returned %d: %s", response.Code, response.Data)
	}

	var uploaded Uploaded
	if err = json.Unmarshal(response.Data, &uploaded); err != nil {
		slog.Error("could not unmarshal response", slog.Any("error", err), slog.Any("raw response", response.Data))
		return nil, err
	}

	return &uploaded, nil
}
