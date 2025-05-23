package rueidiscache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

func TestLocker_AutoLock(t *testing.T) {
	client, err := NewRueidisClient(&rueidis.ClientOption{
		Username:    "",
		Password:    "123456",
		InitAddress: []string{"127.0.0.1:6379"},
		SelectDB:    0,
	})
	if err != nil {
		return
	}
	locker := NewLocker(NewDefaultLockerOption(client))
	ctx := context.Background()
	err = locker.LockOnce(ctx, "test_lock", 10*time.Second, func() error {
		fmt.Println("test_lock do ")
		return nil
	})
	if err != nil {
		return
	}
	assert.Equal(t, nil, err)
}
