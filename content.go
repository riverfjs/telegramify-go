package telegramify

// ContentType represents the type of content.
type ContentType int

const (
	// ContentTypeText represents a text message.
	ContentTypeText ContentType = iota
	// ContentTypeFile represents a file attachment.
	ContentTypeFile
	// ContentTypePhoto represents a photo attachment.
	ContentTypePhoto
)

// String returns the string representation of ContentType.
func (ct ContentType) String() string {
	switch ct {
	case ContentTypeText:
		return "text"
	case ContentTypeFile:
		return "file"
	case ContentTypePhoto:
		return "photo"
	default:
		return "unknown"
	}
}

const (
	ContentTypeMermaid = "mermaid"
)

// ContentTrace tracks the source and metadata of content.
type ContentTrace struct {
	SourceType string
	Extra      map[string]interface{}
}

// Content represents a piece of content ready to be sent via Telegram.
type Content interface {
	GetContentType() ContentType
	GetContentTrace() ContentTrace
}

// Text represents a text message segment.
type Text struct {
	Text         string
	Entities     []MessageEntity
	ContentTrace ContentTrace
}

// GetContentType returns ContentTypeText.
func (t *Text) GetContentType() ContentType {
	return ContentTypeText
}

// GetContentTrace returns the content trace.
func (t *Text) GetContentTrace() ContentTrace {
	return t.ContentTrace
}

// File represents a file attachment.
type File struct {
	FileName        string
	FileData        []byte
	CaptionText     string
	CaptionEntities []MessageEntity
	ContentTrace    ContentTrace
}

// GetContentType returns ContentTypeFile.
func (f *File) GetContentType() ContentType {
	return ContentTypeFile
}

// GetContentTrace returns the content trace.
func (f *File) GetContentTrace() ContentTrace {
	return f.ContentTrace
}

// Photo represents a photo attachment.
type Photo struct {
	FileName        string
	FileData        []byte
	Caption         string
	CaptionText     string
	CaptionEntities []MessageEntity
	ContentTrace    ContentTrace
}

// GetContentType returns ContentTypePhoto.
func (p *Photo) GetContentType() ContentType {
	return ContentTypePhoto
}

// GetContentTrace returns the content trace.
func (p *Photo) GetContentTrace() ContentTrace {
	return p.ContentTrace
}

