package id3tag

import (
	"io"

	"github.com/bogem/id3v2/v2"
)

// CTOCFrame implements the ID3v2 Table Of Contents frame (CTOC)
// As defined in ID3v2 Chapter Frame Addendum (id3v2-chapters-1.0)
type CTOCFrame struct {
	ElementID  string           // Unique identifier for this CTOC frame
	IsTopLevel bool             // Whether this is a top-level table of contents
	IsOrdered  bool             // Whether chapters are in a specific order
	ChildIDs   []string         // IDs of child elements (usually CHAP frames)
	Title      *id3v2.TextFrame // Optional title
}

// Size returns the size of the frame
func (cf CTOCFrame) Size() int {
	// Size calculation:
	// - ElementID (null-terminated string)
	// - Flags (1 byte)
	// - Entry count (1 byte)
	// - Child element IDs (each null-terminated)
	// - Optional subframe (Title) (if present)

	size := len(cf.ElementID) + 1 // ElementID is null-terminated
	size += 1                     // Flags byte
	size += 1                     // Entry count byte

	// Add size of child element IDs (each null-terminated)
	for _, id := range cf.ChildIDs {
		size += len(id) + 1
	}

	// Add size of optional Title subframe if present
	if cf.Title != nil {
		// Frame ID (4 bytes) + Size (4 bytes) + Flags (2 bytes) + Frame content
		size += 10 + cf.Title.Size()
	}

	return size
}

// UniqueIdentifier returns "CTOC"
func (cf CTOCFrame) UniqueIdentifier() string {
	return "CTOC"
}

// WriteTo writes the frame to a writer
func (cf CTOCFrame) WriteTo(w io.Writer) (int64, error) {
	var n int64
	var written int

	// Write ElementID (null-terminated)
	written, err := w.Write(append([]byte(cf.ElementID), 0))
	n += int64(written)
	if err != nil {
		return n, err
	}

	// Prepare and write flags byte
	flags := byte(0)
	if cf.IsTopLevel {
		flags |= 1 // Set bit 0 if top-level
	}
	if cf.IsOrdered {
		flags |= 2 // Set bit 1 if ordered
	}

	written, err = w.Write([]byte{flags})
	n += int64(written)
	if err != nil {
		return n, err
	}

	// Write entry count (number of child elements)
	written, err = w.Write([]byte{byte(len(cf.ChildIDs))})
	n += int64(written)
	if err != nil {
		return n, err
	}

	// Write child element IDs (each null-terminated)
	for _, id := range cf.ChildIDs {
		written, err = w.Write(append([]byte(id), 0))
		n += int64(written)
		if err != nil {
			return n, err
		}
	}

	// Write optional Title subframe if present
	if cf.Title != nil {
		// Write frame ID (4 bytes: "TIT2")
		written, err = w.Write([]byte("TIT2"))
		n += int64(written)
		if err != nil {
			return n, err
		}

		// Write frame size (4 bytes)
		size := uint32(cf.Title.Size())
		written, err = w.Write([]byte{
			byte(size >> 24),
			byte(size >> 16),
			byte(size >> 8),
			byte(size),
		})
		n += int64(written)
		if err != nil {
			return n, err
		}

		// Write frame flags (2 bytes)
		written, err = w.Write([]byte{0, 0})
		n += int64(written)
		if err != nil {
			return n, err
		}

		// Write frame content
		writtenInt64, err := cf.Title.WriteTo(w)
		n += writtenInt64
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

// createCTOCFrame creates a new CTOC frame with the specified parameters
func createCTOCFrame(elementID string, isTopLevel, isOrdered bool, childIDs []string, title string) CTOCFrame {
	ctocFrame := CTOCFrame{
		ElementID:  elementID,
		IsTopLevel: isTopLevel,
		IsOrdered:  isOrdered,
		ChildIDs:   childIDs,
	}

	// Add title if present
	if title != "" {
		ctocFrame.Title = &id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF8,
			Text:     title,
		}
	}

	return ctocFrame
}
