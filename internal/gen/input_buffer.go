// Copyright (C) 2022-2023, Roslan Amir. All rights reserved.
// Created on: 22-Jul-2023
//
// Struct to hold the source lines.

package gen

import (
	"fmt"
	"strings"

	"github.com/roslamir/epubgen/internal/fileutil"
)

// SectionData holds the attributes for a section.
// Each generated HTML is considered a section and each section metadata is kept here.
type SectionData struct {
	ID       string // section id is used as the name of the section file and also used as the id in the package manifest
	EpubType string // used as the value for "epub-type" attribute for the HTML <section> tag
	Heading  string // used as the section heading to be displayed in the table of contents (TOC)
}

// ImageData holds the file name, the media type and optionally the caption for an image file.
type ImageData struct {
	FileName  string // image file name with extension
	MediaType string // the media type (png/jpeg) based on extension
	Caption   string // the caption for the image (optional)
}

// InputBuffer contains the input lines and other artifacts derived from the input lines.
type InputBuffer struct {
	CurrLine   string            // holds the string representing the current line
	lineIndex  int               // index into the 'lines' slice', points to the current line
	lines      []string          // holds the list of all lines from the source HTML file
	attributes map[string]string // contains all the metadata attibutes
	coverImage ImageData         // holds the file name and extension for the cover image
	// images        []ImageData       // holds the list of all image files (other than the cover image) used in the book
	images        map[string]ImageData // holds the maps of all image files (other than the cover image) used in the book
	sections      []SectionData        // used to generated TOC and MANIFEST files
	guides        []SectionData        // used in the Guides section of the manifest
	currSectionNo int                  // Holds the current section counter
}

func NewInputBuffer(sourceFileSpec string) *InputBuffer {
	b := InputBuffer{}
	b.lines = fileutil.ReadLines(sourceFileSpec)
	b.attributes = make(map[string]string)
	b.sections = make([]SectionData, 0, 50)
	b.guides = make([]SectionData, 0, 10)
	return &b
}

// NewSectionData creates a new instance of SectionData and adds it to the 'sections' list.
// It uses a running number to generate the section ID in the format "sectionNNN".
func (b *InputBuffer) NewSectionData(epubType, heading string) SectionData {
	b.currSectionNo++
	return SectionData{
		ID:       fmt.Sprintf("section%03d", b.currSectionNo),
		EpubType: epubType,
		Heading:  heading,
	}
}

// NumLines returns the number of lines in the buffer.
func (b *InputBuffer) NumLines() int {
	return len(b.lines)
}

// NextLine returns the next source line.
func (b *InputBuffer) NextLine() {
	b.lineIndex++
	if b.lineIndex == len(b.lines) {
		panic("epubgen: unexpected end of input file")
	}
	b.CurrLine = b.lines[b.lineIndex]
}

// LoadAttributes scans the metadata lines from the input file and extract the attributes.
func (b *InputBuffer) LoadAttributes() {
	for {
		b.NextLine()
		if b.CurrLine == "</head>" {
			break
		}
		if strings.HasPrefix(b.CurrLine, "<meta") {
			index := strings.Index(b.CurrLine, "name=")
			if index != -1 {
				name := b.CurrLine[index+len("name=")+1:] // skip past 'name="'
				index = strings.Index(name, "\"")
				if index == -1 {
					panic("Invalid 'meta' HTML line: " + b.CurrLine)
				}
				name = name[:index]

				index = strings.Index(b.CurrLine, "content=")
				if index == -1 {
					panic("Invalid 'meta' HTML line: " + b.CurrLine)
				}
				content := b.CurrLine[index+len("content=")+1:] // skip past 'content="'
				index = strings.Index(content, "\"")
				if index == -1 {
					panic("Invalid 'meta' HTML line: " + b.CurrLine)
				}
				content = content[:index]
				if name != "" {
					b.attributes[name] = content
				}
			}
		}
	}
}

// GetAttribute returns the attribute value or the empty string if the atrribute with the given key does not exist.
func (b *InputBuffer) GetAttribute(key string) string {
	return b.attributes[key]
}

// SetAttribute sets or adds an attribute with the given key/value pair.
func (b *InputBuffer) SetAttribute(key, value string) {
	b.attributes[key] = value
}

// CheckCoverImage checks for the presence of the attribute "cover-image".
// The value must be the name of the cover image file with extension of either ".jpeg" or ".png".
// To make life easier, assume all JPEG files have extension ".jpeg" instead of ".jpg".
func (b *InputBuffer) CheckCoverImage() {
	imageFile := b.attributes["cover-image"]
	if imageFile == "" {
		panic("epubgen: attribute 'cover-image' required")
	}
	_, mediaType, _ := strings.Cut(imageFile, ".")
	if mediaType != "png" && mediaType != "jpeg" {
		panic("epubgen: only image files with extension 'png' or 'jpeg' are accepted")
	}
	b.coverImage = ImageData{
		FileName:  imageFile,
		MediaType: mediaType,
	}
}

// CheckImageFiles checks for the presence of the optional attribute "images".
// The value must be the comma-separated image file names with extension of either ".jpeg" or ".png".
// To make life easier, assume all JPEG files have extension ".jpeg" instead of ".jpg".
func (b *InputBuffer) CheckImageFiles() {
	value := b.attributes["images"]
	if value == "" {
		return
	}
	// b.images = make([]ImageData, 0, 5)
	b.images = make(map[string]ImageData)
	files := strings.Split(value, ",")
	for _, imageFile := range files {
		_, mediaType, _ := strings.Cut(imageFile, ".")
		if mediaType != "png" && mediaType != "jpeg" {
			panic("epubgen: only image files with extension 'png' or 'jpeg' are accepted")
		}
		image := ImageData{
			FileName:  imageFile,
			MediaType: mediaType,
		}
		// b.images = append(b.images, image)
		b.images[imageFile] = image
	}
}

// AddSection adds the given section to the list of sections.
func (b *InputBuffer) AddSection(section SectionData) {
	b.sections = append(b.sections, section)
}

// AddGuide adds the given section to the list of guides.
func (b *InputBuffer) AddGuide(section SectionData) {
	b.guides = append(b.guides, section)
}
