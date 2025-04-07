package rueidiscache

import (
	"context"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
)

type Locker struct {
	client rueidis.Client
}

func NewLocker(client rueidis.Client) *Locker {
	return &Locker{client: client}
}

func (l *Locker) AutoLock(ctx context.Context, key string, do func() error) error {
	locker, err := rueidislock.NewLocker(rueidislock.LockerOption{
		ClientBuilder: func(_ rueidis.ClientOption) (rueidis.Client, error) {
			return l.client, nil
		},
		KeyMajority:    1,    // Use KeyMajority=1 if you have only one Redis instance. Also make sure that all your `Locker`s share the same KeyMajority.
		NoLoopTracking: true, // Enable this to have better performance if all your Redis are >= 7.0.5.
	})
	if err != nil {
		return err
	}
	defer locker.Close()
	_, cancel, err := locker.WithContext(ctx, key)
	if err != nil {
		return err
	}
	defer cancel()
	return do()
}
