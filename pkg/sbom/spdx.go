package sbom

import (
	"bytes"
	"fmt"
	"io"

	spdx_json "github.com/spdx/tools-golang/json"
	"github.com/spdx/tools-golang/spdx"
)

// NewDocumentFromSpdxJSON creates a new SBOM Document from SPDX raw JSON data.
func NewDocumentFromSpdxJSON(data []byte) (*Document, error) {
	reader := bytes.NewReader(data)
	doc, err := spdx_json.Read(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read SPDX JSON: %w", err)
	}
	return newDocument(doc)
}

// ToSpdxJSON writes the SBOM Document to the writer in SPDX JSON format.
func (d *Document) ToSpdxJSON(w io.Writer) error {
	switch d.docType {
	case StandardTypeSpdx:
		var opt []spdx_json.WriteOption
		opt = append(opt, spdx_json.Indent("  "))
		opt = append(opt, spdx_json.EscapeHTML(true))
		return spdx_json.Write(d.document.(*spdx.Document), w, opt...)
	default:
		return fmt.Errorf("conversion to SPDX JSON not supported for document type: %s", d.docType)
	}
}

// ToSpdxDocument converts the SBOM Document to SPDX Document.
func (d *Document) ToSpdxDocument() (*spdx.Document, error) {
	switch d.docType {
	case StandardTypeSpdx:
		return d.document.(*spdx.Document), nil
	default:
		return nil, fmt.Errorf("conversion to SPDX document not supported for document type: %s", d.docType)
	}
}
