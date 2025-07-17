package menu



func MaxLengthLabel(labels []Item) int {
	if len(labels) == 0 {
		return 0
	}

	maxLen := len(labels[0].Label.Display)
	for _, l := range labels[1:] {
		curLen := len(l.Label.Display)

		if l.IconData != nil || l.IconName != "" {
			curLen += 2
		}
		if curLen > maxLen {
			maxLen = curLen
		}
	}

	return maxLen
}
