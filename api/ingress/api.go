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

	"github.com/m-horky/insights-client-next/api"
	"github.com/m-horky/insights-client-next/app"
)

var service api.Service

// Init has to be called to set up the API configuration for the service.
func Init(s *api.Service) {
	service = *s
}

// UploadArchive loads an archive from filesystem and uploads it to Ingress.
func UploadArchive(archive Archive) (*Uploaded, app.HumanError) {
	slog.Debug(
		"uploading archive",
		slog.String("path", archive.Path),
		slog.String("content-type", archive.ContentType),
	)

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
		slog.Error("could not create archive field", slog.String("error", err.Error()))
		return nil, api.NewError(
			ErrArchive,
			err,
			nil,
			"Could not prepare data archive.",
		)
	}

	archiveDescriptor, err := os.Open(archive.Path)
	if err != nil {
		slog.Error("could not open archive", slog.String("error", err.Error()))
		return nil, api.NewError(
			ErrArchive,
			err,
			nil,
			"Could not prepare data archive.",
		)
	}
	defer archiveDescriptor.Close()

	_, err = io.Copy(archiveField, archiveDescriptor)
	if err != nil {
		slog.Error("could not read archive file", slog.String("error", err.Error()))
		return nil, api.NewError(
			ErrArchive,
			err,
			nil,
			"Could not prepare data archive.",
		)
	}

	form.Close()

	headers := make(map[string][]string)
	headers["Content-Type"] = []string{form.FormDataContentType()}

	response, err := service.MakeRequest("POST", "upload", url.Values{}, headers, formData)
	if err != nil {
		slog.Error("could not upload archive", slog.String("error", err.Error()))
		return nil, api.NewError(
			api.ErrServiceUnreachable,
			err,
			response,
			"Upload service could not be contacted.",
		)
	}

	if response.Code/100 != 2 {
		slog.Error(
			"server rejected the archive",
			slog.Int("status code", response.Code),
			slog.Any("raw response", response.Data),
		)
		return nil, api.NewError(
			api.ErrBadResponse,
			err,
			response,
			"Upload service rejected the archive.",
		)
	}

	var uploaded Uploaded
	if err = json.Unmarshal(response.Data, &uploaded); err != nil {
		slog.Error(
			"could not unmarshal response",
			slog.String("error", err.Error()),
			slog.Any("raw response", response.Data),
		)
		return nil, api.NewError(
			api.ErrUnparseable,
			err,
			response,
			"Upload service returned bad response.",
		)
	}

	return &uploaded, nil
}
