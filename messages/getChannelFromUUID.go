package messages

import (
	"github.com/gocql/gocql"
	"sort"
	"strings"
)

func CombineUUIDs(uuid1, uuid2 gocql.UUID) string {
	// Convert UUIDs to strings
	uuidStr1 := uuid1.String()
	uuidStr2 := uuid2.String()

	// Store UUID strings in a slice
	uuids := []string{uuidStr1, uuidStr2}

	// Sort the UUID strings lexicographically
	sort.Strings(uuids)

	// Concatenate the sorted UUIDs with an underscore
	return strings.Join(uuids, "_")
}
