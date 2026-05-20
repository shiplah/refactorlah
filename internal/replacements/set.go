package replacements

import "sort"

type ByStartDescending []Replacement

func (r ByStartDescending) Len() int           { return len(r) }
func (r ByStartDescending) Less(i, j int) bool { return r[i].Start > r[j].Start }
func (r ByStartDescending) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

func SortDescending(items []Replacement) {
	sort.Sort(ByStartDescending(items))
}
