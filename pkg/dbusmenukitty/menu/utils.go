package menu

func MaxLengthLabel(labels []Item) int {
	if len(labels) == 0 {
		return 0
	}
  
  maxLen := 0
	for _, l := range labels {
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
