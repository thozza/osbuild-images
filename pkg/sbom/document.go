package sbom

import (
	"fmt"

	"github.com/spdx/tools-golang/spdx"
)

type StandardType uint64

const (
	StandardTypeNone StandardType = iota
	StandardTypeSpdx
)

func (t StandardType) String() string {
	switch t {
	case StandardTypeNone:
		return "none"
	case StandardTypeSpdx:
		return "spdx"
	default:
		panic("invalid standard type")
	}
}

type Document struct {
	// type of the document standard
	docType StandardType

	// document in a specific standard format
	document interface{}
}

func newDocument(d interface{}) (*Document, error) {
	var docType StandardType

	switch d.(type) {
	case *spdx.Document:
		docType = StandardTypeSpdx
	default:
		return nil, fmt.Errorf("unsupported SBOM document type: %T", d)
	}

	return &Document{
		docType:  docType,
		document: d,
	}, nil
}
