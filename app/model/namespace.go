package model

// Namespace represents an Okteto namespace
type Namespace struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
