package records

type Profile struct {
	LexiconTypeID string   `json:"$type,const=world.bsky.demo.profile" cborgen:"$type,const=world.bsky.demo.profile"`
	CreatedAt     int64    `json:"createdAt" cborgen:"createdAt"`
	Links         []string `json:"links,omitempty" cborgen:"links,omitempty"`
	Text          string   `json:"text" cborgen:"text"`
}

type Comment struct {
	LexiconTypeID string `json:"$type,const=world.bsky.demo.comment" cborgen:"$type,const=world.bsky.demo.comment"`
	Profile       string `json:"profile" cborgen:"profile"`
	CreatedAt     int64  `json:"createdAt" cborgen:"createdAt"`
	Text          string `json:"text" cborgen:"text"`
}
