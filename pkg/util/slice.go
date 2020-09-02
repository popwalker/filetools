package util

func CheckRepeat(s []string) (repeated []string) {
	m := make(map[string]struct{})

	for _, v := range s {
		if _, ok := m[v]; ok {
			repeated = append(repeated, v)
		} else {
			m[v] = struct{}{}
		}
	}
	return
}

// DivideSliceIntoGroup 将一个slice中的元素，均匀的分散到N组中
func DivideSliceIntoGroup(s []string, groupCount int) [][]string {
	count := len(s)
	if groupCount > count {
		groupCount = count
	}
	groups := make([][]string, groupCount)
	remainder := count % groupCount

	for i := 0; i < count; i += groupCount {
		step := i + groupCount
		var sg []string
		if remainder != 0 && step > count {
			sg = s[i:]
		} else {
			sg = s[i:step]
		}

		for i, link := range sg {
			groups[i] = append(groups[i], []string{link}...)
		}
	}

	return groups
}
