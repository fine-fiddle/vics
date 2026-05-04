package core

type TagID string
type Lang string

const (
	LangEnglish  Lang = "en"
	LangJapanese Lang = "ja"
	LangRomaji   Lang = "ja-Latn"
)

type TagKind string

const (
	TagKindObject    TagKind = "object"
	TagKindEmotion   TagKind = "emotion"
	TagKindStyle     TagKind = "style"
	TagKindColor     TagKind = "color"
	TagKindAction    TagKind = "action"
	TagKindCharacter TagKind = "character"
	TagKindOther     TagKind = "other"
)

type Tag struct {
	ID        TagID
	Kind      TagKind
	CreatedAt int64
	UpdatedAt int64
}

type LabelSource string

const (
	LabelSourceOriginal    LabelSource = "original"
	LabelSourceTranslated  LabelSource = "translated"
	LabelSourceUserEntered LabelSource = "user_entered"
	LabelSourceImported    LabelSource = "imported"
)

type TagLabel struct {
	TagID      TagID
	Language   Lang
	Text       string
	Normalized string
	IsPrimary  bool
	Source     LabelSource
}

type TagAssignmentSource string

const (
	TagAssignmentUser     TagAssignmentSource = "user"
	TagAssignmentImported TagAssignmentSource = "imported"
	TagAssignmentAuto     TagAssignmentSource = "auto"
)

type AssetTag struct {
	AssetID    AssetID
	TagID      TagID
	Source     TagAssignmentSource
	Confidence float64
	CreatedAt  int64
}
