package models

import "encoding/json"

type Message struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

func NoteToMessage(n Note) (Message, error) {
	v, err := json.Marshal(n)
	if err != nil {
		return Message{}, err
	}
	k, err := json.Marshal(n.ID)
	if err != nil {
		return Message{}, err
	}
	return Message{
		Key:   k,
		Value: v,
	}, nil
}
