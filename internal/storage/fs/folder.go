package fs

import (
	"archive/zip"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/google/uuid"
)

func (fs *FileSystemStorage) CreateFolder(uid, name, parentID string) (*storage.Document, error) {
	docID := uuid.New().String()

	// Create zip file
	zipFilePath := fs.getPathFromUser(uid, docID+storage.ZipFileExt)

	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	contentEntry, err := zipWriter.Create(docID + storage.ContentFileExt)
	if err != nil {
		return nil, err
	}
	contentEntry.Write([]byte(`{"tags":[]}`))

	// Create metadata file
	mdFilePath := fs.getPathFromUser(uid, docID+storage.MetadataFileExt)

	mdFile, err := os.Create(mdFilePath)
	if err != nil {
		return nil, err
	}
	defer mdFile.Close()

	md := messages.RawMetadata{
		ID:             docID,
		VissibleName:   strings.TrimSpace(name),
		Parent:         parentID,
		Version:        1,
		ModifiedClient: time.Now().UTC().Format(time.RFC3339Nano),
		Type:           common.CollectionType,
	}

	if err = json.NewEncoder(mdFile).Encode(md); err != nil {
		return nil, err
	}

	doc := &storage.Document{
		ID:      md.ID,
		Type:    md.Type,
		Name:    md.VissibleName,
		Version: md.Version,
	}

	return doc, nil
}
