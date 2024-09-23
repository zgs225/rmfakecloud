package fs

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func createContent(fileType string) string {
	fileType = strings.TrimPrefix(fileType, ".")
	str :=
		`
{
	"dummyDocument": false,
	"coverPageNumber": 0,
	"extraMetadata": {
		"LastPen": "Finelinerv2",
		"LastTool": "Finelinerv2",
		"ThicknessScale": "",
		"LastFinelinerv2Size": "1"
	},
	"fileType": "%s",
	"fontName": "",
	"lastOpenedPage": 0,
	"lineHeight": -1,
	"margins": 180,
	"orientation": "portrait",
	"pageCount": 0,
	"pages": [],
	"textScale": 1,
	"transform": {
		"m11": 1,
		"m12": 0,
		"m13": 0,
		"m21": 0,
		"m22": 1,
		"m23": 0,
		"m31": 0,
		"m32": 0,
		"m33": 1
	}
}
`
	return fmt.Sprintf(str, fileType)
}

func extractID(r io.Reader) (string, error) {
	return "", nil
}

// CreateDocument creates a new document
func (fs *FileSystemStorage) CreateDocument(uid, filename, parent string, stream io.Reader) (doc *storage.Document, err error) {
	ext := path.Ext(filename)
	switch ext {
	case storage.PdfFileExt:
		fallthrough
	case storage.EpubFileExt:
	default:
		return nil, errors.New("unsupported extension: " + ext)
	}

	var docid string

	var isZip = false
	if ext == storage.ZipFileExt {
		docid, err = extractID(stream)
		isZip = true
	} else {
		docid = uuid.New().String()
	}
	//create zip from pdf
	zipfile := fs.getPathFromUser(uid, docid+storage.ZipFileExt)
	file, err := os.Create(zipfile)
	if err != nil {
		return
	}
	defer file.Close()

	if !isZip {
		w := zip.NewWriter(file)
		defer w.Close()

		documentPath := docid + ext
		var entry io.Writer
		entry, err = w.Create(documentPath)
		if err != nil {
			return
		}

		_, err = io.Copy(entry, stream)
		if err != nil {
			return
		}

		entry, err = w.Create(docid + storage.PageFileExt)
		if err != nil {
			return
		}
		entry.Write([]byte{})

		entry, err = w.Create(docid + storage.ContentFileExt)
		if err != nil {
			return
		}

		content := createContent(ext)
		entry.Write([]byte(content))
	} else {
		logrus.Info("writing file")
		_, err = io.Copy(file, stream)
		if err != nil {
			return
		}
	}

	//create metadata
	name := strings.TrimSuffix(filename, ext)
	rawMetadata := createRawMedatadata(docid, name, parent, common.DocumentType)

	jsn, err := json.Marshal(rawMetadata)
	if err != nil {
		return
	}

	doc = &storage.Document{
		ID:      docid,
		Type:    rawMetadata.Type,
		Name:    name,
		Version: 1,
	}
	//save metadata
	metafilePath := fs.getPathFromUser(uid, docid+storage.MetadataFileExt)
	err = ioutil.WriteFile(metafilePath, jsn, 0600)
	return
}

func createRawMedatadata(id, name, parent string, entryType common.EntryType) *messages.RawMetadata {
	doc := messages.RawMetadata{
		ID:             id,
		VissibleName:   name,
		Version:        1,
		ModifiedClient: time.Now().UTC().Format(time.RFC3339Nano),
		CurrentPage:    0,
		Type:           entryType,
		Parent:         parent,
	}
	return &doc
}
