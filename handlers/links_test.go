package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContentRangeHeader(t *testing.T) {
	assert.Equal(t, "links */0", getContentRangeHeader(10, 0, 0))
	assert.Equal(t, "links */8", getContentRangeHeader(10, 0, 8))
	assert.Equal(t, "links 2-2/44", getContentRangeHeader(2, 1, 44))
	assert.Equal(t, "links 10-48/100", getContentRangeHeader(10, 39, 100))
	assert.Equal(t, "links 0-0/1", getContentRangeHeader(0, 1, 1))
}
