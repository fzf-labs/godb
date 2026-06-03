package gorediscache

import (
	"context"
	"testing"

	"github.com/bsm/redislock"
	"github.com/stretchr/testify/assert"
)

func TestClassifyObtainErr_NotObtained(t *testing.T) {
	err := classifyObtainErr(redislock.ErrNotObtained)
	assert.ErrorIs(t, err, NotObtained)
}

func TestClassifyObtainErr_PreservesUnexpectedErrors(t *testing.T) {
	err := classifyObtainErr(context.Canceled)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, NotObtained)
}
