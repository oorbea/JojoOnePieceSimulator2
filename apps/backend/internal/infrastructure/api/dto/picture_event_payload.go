package dto

// PictureEventPayload is the SSE `picture` event's data shape, sent by
// EventsEndpoints when a Power/Stage/User picture moves through the async
// compression pipeline (PENDING -> READY/FAILED). Kind and Status are the
// wire strings of enums.PictureSubjectKind and enums.PictureStatus.
type PictureEventPayload struct {
	Kind      string `json:"kind" ts:"PictureSubjectKind"`
	SubjectID string `json:"subjectId"`
	Status    string `json:"status" ts:"PictureStatus"`
}
