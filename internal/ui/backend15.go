package ui

import (
	"errors"
	"io"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/google/uuid"
)

type backend15 struct {
	blobHandler blobHandler
	h           *hub.Hub
}

func (b *backend15) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	hashTree, err := b.blobHandler.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}

	return viewmodel.DocTreeFromHashTree(hashTree), nil
}
func (b *backend15) Export(uid, docid, exporttype string, opt storage.ExportOption) (r io.ReadCloser, err error) {
	r, err = b.blobHandler.Export(uid, docid)
	return
}

func (b *backend15) CreateDocument(uid, filename, parent string, stream io.Reader) (doc *storage.Document, err error) {
	doc, err = b.blobHandler.CreateBlobDocument(uid, filename, parent, stream)
	return
}

func (b *backend15) UpdateDocument(uid, docID, name, parent string) (err error) {
	return b.blobHandler.UpdateBlobDocument(uid, docID, name, parent)
}
func (b *backend15) CreateFolder(uid, name, parent string) (doc *storage.Document, err error) {
	return b.blobHandler.CreateBlobFolder(uid, name, parent)
}

func (b *backend15) DeleteDocument(uid, docID string) (err error) {
	return b.blobHandler.DeleteBlobDocument(uid, docID)
}

func (b *backend15) Sync(uid string) {
	b.h.NotifySync(uid, uuid.NewString())
}

// RenameDocument rename file and folder, the bool type returns value indicates
// whether updated or not
func (d *backend15) RenameDocument(uid, docId, newName string) (bool, error) {
	metadata, err := d.blobHandler.GetBlobMetadata(uid, docId)

	if err != nil {
		return false, err
	}

	if newName == metadata.DocumentName {
		return false, nil
	}

	metadata.DocumentName = newName

	if err = d.blobHandler.UpdateBlobMetadata(uid, docId, metadata); err != nil {
		return false, err
	}

	return true, nil
}

// MoveDocument move document to a new parent
func (d *backend15) MoveDocument(uid, docId, newParent string) (bool, error) {
	// Check parent
	parentMD, err := d.blobHandler.GetBlobMetadata(uid, newParent)

	if err != nil {
		return false, err
	}

	if parentMD.CollectionType != common.CollectionType {
		return false, errors.New("Parent is not a folder")
	}

	metadata, err := d.blobHandler.GetBlobMetadata(uid, docId)

	if err != nil {
		return false, err
	}

	if metadata.Parent == newParent {
		return false, nil
	}

	metadata.Parent = newParent

	if err = d.blobHandler.UpdateBlobMetadata(uid, docId, metadata); err != nil {
		return false, err
	}

	return true, nil
}

func (d *backend15) GetMetadata(uid, docId string) (*common.MetadataFile, error) {
	return d.blobHandler.GetBlobMetadata(uid, docId)
}
