package population

import "errors"

type FakeStore struct {
	Content map[string]int64
	Fail    bool
}

func (f *FakeStore) List() (map[string]int64, error) {
	if f.Fail {
		return nil, errors.New("db error")
	}
	return f.Content, nil
}

func (f *FakeStore) Add(s string, i int64) error {
	if f.Fail {
		return errors.New("db error")
	}
	if f.Content == nil {
		f.Content = make(map[string]int64)
	}
	f.Content[s] = i
	return nil
}
