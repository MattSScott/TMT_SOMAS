package infra

type ASPDecison int

const (
	SELF_SACRIFICE ASPDecison = iota
	NOT_SELF_SACRIFICE
	INACTION
)

type AttachmentType int

const (
	DISMISSIVE AttachmentType = iota
	FEARFUL
	PREOCCUPIED
	SECURE
)

var AllAttachmentTypes = []AttachmentType{DISMISSIVE, FEARFUL, PREOCCUPIED, SECURE}

const (
	// ASP weights
	W1 float32 = 0.25
	W2 float32 = 0.25
	W3 float32 = 0.25
	W4 float32 = 0.25

	W5 float32 = 0.33
	W6 float32 = 0.33
	W7 float32 = 0.34

	W8  float32 = 0.33
	W9  float32 = 0.33
	W10 float32 = 0.34
)
