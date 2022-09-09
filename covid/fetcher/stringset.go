package fetcher

// StringSet implements a simple Set of strings
type StringSet struct {
	set map[string]struct{}
}

// Set adds the provided string to the Set.  Returns false if the string was not already in the Set
func (ss *StringSet) Set(name string) (found bool) {
	found = ss.IsSet(name)
	ss.set[name] = struct{}{}
	return
}

// IsSet returns true if the provided string is in the Set
func (ss *StringSet) IsSet(name string) (found bool) {
	if ss.set == nil {
		ss.set = make(map[string]struct{})
	}
	_, found = ss.set[name]
	return
}
