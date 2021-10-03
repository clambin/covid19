package fetcher

type StringSet struct {
	set map[string]struct{}
}

func (ss *StringSet) Set(name string) (found bool) {
	found = ss.IsSet(name)
	ss.set[name] = struct{}{}
	return
}

func (ss *StringSet) IsSet(name string) (found bool) {
	if ss.set == nil {
		ss.set = make(map[string]struct{})
	}
	_, found = ss.set[name]
	return
}
