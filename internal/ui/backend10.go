package ui

import (
	"errors"
	"io"
	"time"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	log "github.com/sirupsen/logrus"
)

const webDevice = "web"

type backend10 struct {
	documentHandler documentHandler
	metadataStore   storage.MetadataStorer
	hub             *hub.Hub
}

func (d *backend10) Sync(uid string) {
	//nop
}

func (d *backend10) CreateFolder(uid, filename, parent string) (doc *storage.Document, err error) {
	if len(parent) > 0 {
		md, err := d.metadataStore.GetMetadata(uid, parent)
		if err != nil {
			return nil, err
		}
		if md.Type != common.CollectionType {
			return nil, errors.New("Parent is not a folder")
		}
	}

	doc, err = d.documentHandler.CreateFolder(uid, filename, parent)
	if err != nil {
		return
	}

	ntf := hub.DocumentNotification{
		ID:      doc.ID,
		Type:    common.CollectionType,
		Version: 1,
		Parent:  parent,
		Name:    doc.Name,
	}
	log.Info(uiLogger, "created folder", doc.ID)
	d.hub.Notify(uid, webDevice, ntf, messages.DocAddedEvent)
	return
}
func (d *backend10) CreateDocument(uid, filename, parent string, stream io.Reader) (doc *storage.Document, err error) {
	doc, err = d.documentHandler.CreateDocument(uid, filename, parent, stream)
	if err != nil {
		return
	}

	ntf := hub.DocumentNotification{
		ID:      doc.ID,
		Type:    common.DocumentType,
		Version: 1,
		Parent:  parent,
		Name:    doc.Name,
	}
	log.Info(uiLogger, ui10, "Uploaded document id", doc.ID)
	d.hub.Notify(uid, webDevice, ntf, messages.DocAddedEvent)
	return
}

func (d *backend10) DeleteDocument(uid, docid string) error {
	meta, err := d.metadataStore.GetMetadata(uid, docid)

	if err != nil {
		return err
	}

	// Check if document is folder, it must be empty
	if meta.Type == common.CollectionType {
		tree, err := d.GetDocumentTree(uid)
		if err != nil {
			return err
		}

		for _, entry := range tree.Entries {
			dir, ok := entry.(*viewmodel.Directory)
			if !ok {
				continue
			}
			if dir.ID == meta.ID {
				if len(dir.Entries) > 0 {
					return errors.New("can't remove non-empty folder")
				}
			}
		}
	}

	if err = d.documentHandler.RemoveDocument(uid, docid); err != nil {
		return err
	}

	ntf := hub.DocumentNotification{
		ID:      meta.ID,
		Type:    meta.Type,
		Version: meta.Version,
		Parent:  meta.Parent,
		Name:    meta.VissibleName,
	}
	log.Info(uiLogger, "Document deleted: id=", meta.ID, " name=", meta.VissibleName)
	d.hub.Notify(uid, "web", ntf, messages.DocDeletedEvent)
	return nil
}

func (d *backend10) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	documents, err := d.documentHandler.GetAllMetadata(uid)
	if err != nil {
		return nil, err
	}
	docs := make([]*viewmodel.InternalDoc, 0)
	for _, d := range documents {
		lastMod, err := time.Parse(time.RFC3339Nano, d.ModifiedClient)
		if err != nil {
			log.Warn("incorrect time for: ", d.VissibleName, " value: ", lastMod)
		}
		docs = append(docs, &viewmodel.InternalDoc{
			ID:           d.ID,
			Parent:       d.Parent,
			Name:         d.VissibleName,
			Type:         d.Type,
			FileType:     "TODO",
			LastModified: lastMod,
		})

	}
	return viewmodel.DocTreeFromRawMetadata(docs), nil
}
func (d *backend10) Export(uid, docID, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error) {
	r, err := d.documentHandler.ExportDocument(uid, docID, exporttype, opt)
	if err != nil {
		return nil, err
	}
	log.Info(uiLogger, ui10, "Exported document id: ", docID)
	return r, nil
}

func (d *backend10) UpdateDocument(uid, docID, name, parent string) (err error) {
	metadata, err := d.documentHandler.GetMetadata(uid, docID)
	if err != nil {
		return err
	}
	metadata.VissibleName = name
	metadata.Parent = parent
	metadata.Version++

	err = d.documentHandler.UpdateMetadata(uid, metadata)
	if err != nil {
		return err
	}
	ntf := hub.DocumentNotification{
		ID:      docID,
		Type:    common.DocumentType,
		Version: metadata.Version,
		Parent:  parent,
		Name:    metadata.VissibleName,
	}

	log.Info(uiLogger, "Updated document id: ", docID)
	d.hub.Notify(uid, webDevice, ntf, messages.DocAddedEvent)
	return nil

}

// RenameDocument rename file and folder, the bool type returns value indicates
// whether updated or not
func (d *backend10) RenameDocument(uid, docId, newName string) (bool, error) {
	metadata, err := d.metadataStore.GetMetadata(uid, docId)

	if err != nil {
		return false, err
	}

	if newName == metadata.VissibleName {
		return false, nil
	}

	metadata.VissibleName = newName
	metadata.Version += 1

	if err = d.metadataStore.UpdateMetadata(uid, metadata); err != nil {
		return false, err
	}

	return true, nil
}

// MoveDocument move document to a new parent
func (d *backend10) MoveDocument(uid, docId, newParent string) (bool, error) {
	// Check parent
	parentMD, err := d.metadataStore.GetMetadata(uid, newParent)

	if err != nil {
		return false, err
	}

	if parentMD.Type != common.CollectionType {
		return false, errors.New("Parent is not a folder")
	}

	metadata, err := d.metadataStore.GetMetadata(uid, docId)

	if err != nil {
		return false, err
	}

	if metadata.Parent == newParent {
		return false, nil
	}

	metadata.Parent = newParent
	metadata.Version += 1

	if err = d.metadataStore.UpdateMetadata(uid, metadata); err != nil {
		return false, err
	}

	return true, nil
}

func (d *backend10) GetMetadata(uid, docId string) (*common.MetadataFile, error) {
	rawMd, err := d.metadataStore.GetMetadata(uid, docId)
	if err != nil {
		return nil, err
	}
	md := &common.MetadataFile{
		DocumentName: rawMd.VissibleName,
		CollectionType: rawMd.Type,
		Parent: rawMd.Parent,
		Version: rawMd.Version,
	}
	return md, nil
}
