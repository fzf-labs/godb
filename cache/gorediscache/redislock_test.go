package gorediscache

import (
	"context"
	"testing"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewLocker(t *testing.T) {
	client, _ := redismock.NewClientMock()
	locker := NewLocker(client)
	assert.NotNil(t, locker)
	assert.NotNil(t, locker.locker)
}

func TestClassifyObtainErrNil(t *testing.T) {
	assert.NoError(t, classifyObtainErr(nil))
}

func TestClassifyObtainErr_NotObtained(t *testing.T) {
	err := classifyObtainErr(redislock.ErrNotObtained)
	assert.ErrorIs(t, err, NotObtained)
}

func TestClassifyObtainErr_PreservesUnexpectedErrors(t *testing.T) {
	err := classifyObtainErr(context.Canceled)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, NotObtained)
}

func TestLockerLockMethodsWithMockedLua(t *testing.T) {
	tests := []struct {
		name         string
		run          func(*Locker, context.Context, func() error) error
		wantReleases int
	}{
		{
			name: "lock once",
			run: func(locker *Locker, ctx context.Context, fn func() error) error {
				return locker.LockOnce(ctx, "lock:key", time.Minute, fn)
			},
			wantReleases: 1,
		},
		{
			name: "lock retry",
			run: func(locker *Locker, ctx context.Context, fn func() error) error {
				return locker.LockRetry(ctx, "lock:key", time.Minute, fn)
			},
			wantReleases: 1,
		},
		{
			name: "lock custom",
			run: func(locker *Locker, ctx context.Context, fn func() error) error {
				return locker.LockWithCustom(ctx, "lock:key", time.Minute, time.Millisecond, 1, fn)
			},
			wantReleases: 1,
		},
		{
			name: "lock once not release",
			run: func(locker *Locker, ctx context.Context, fn func() error) error {
				return locker.LockOnceNotRelease(ctx, "lock:key", time.Minute, fn)
			},
			wantReleases: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			locker := NewLocker(client)
			ctx := context.Background()
			called := false

			mock.Regexp().ExpectEvalSha("[0-9a-f]+", []string{"lock:key"}, ".+", 22, "60000").SetVal("OK")
			for i := 0; i < tt.wantReleases; i++ {
				mock.Regexp().ExpectEvalSha("[0-9a-f]+", []string{"lock:key"}, ".+").SetVal(int64(1))
			}

			err := tt.run(locker, ctx, func() error {
				called = true
				return nil
			})

			assert.NoError(t, err)
			assert.True(t, called)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLockerLockMethodsReturnObtainErrors(t *testing.T) {
	tests := []struct {
		name string
		run  func(*Locker, context.Context, func() error) error
	}{
		{
			name: "lock once",
			run: func(locker *Locker, ctx context.Context, fn func() error) error {
				return locker.LockOnce(ctx, "lock:key", time.Minute, fn)
			},
		},
		{
			name: "lock retry",
			run: func(locker *Locker, ctx context.Context, fn func() error) error {
				return locker.LockRetry(ctx, "lock:key", time.Minute, fn)
			},
		},
		{
			name: "lock custom",
			run: func(locker *Locker, ctx context.Context, fn func() error) error {
				return locker.LockWithCustom(ctx, "lock:key", time.Minute, time.Millisecond, 0, fn)
			},
		},
		{
			name: "lock once not release",
			run: func(locker *Locker, ctx context.Context, fn func() error) error {
				return locker.LockOnceNotRelease(ctx, "lock:key", time.Minute, fn)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			locker := NewLocker(client)
			called := false

			mock.Regexp().ExpectEvalSha("[0-9a-f]+", []string{"lock:key"}, ".+", 22, "60000").SetErr(context.Canceled)

			err := tt.run(locker, context.Background(), func() error {
				called = true
				return nil
			})

			assert.ErrorIs(t, err, context.Canceled)
			assert.False(t, called)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLockerLockMethodsRejectNilCallback(t *testing.T) {
	client, mock := redismock.NewClientMock()
	locker := NewLocker(client)

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "lock once", run: func() error { return locker.LockOnce(context.Background(), "lock:key", time.Minute, nil) }},
		{name: "lock retry", run: func() error { return locker.LockRetry(context.Background(), "lock:key", time.Minute, nil) }},
		{name: "lock custom", run: func() error {
			return locker.LockWithCustom(context.Background(), "lock:key", time.Minute, time.Millisecond, 1, nil)
		}},
		{name: "lock once not release", run: func() error { return locker.LockOnceNotRelease(context.Background(), "lock:key", time.Minute, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			assert.ErrorContains(t, err, "lock callback cannot be nil")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
