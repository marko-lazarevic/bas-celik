package widgets

var clipboardHook func(string) bool

// SetClipboard sets the clipboard hook function for copying text.
func SetClipboard(hook func(string) bool) {
	clipboardHook = hook
}

func copyToClipboard(str string) bool {
	if clipboardHook == nil {
		return false
	}

	return clipboardHook(str)
}
