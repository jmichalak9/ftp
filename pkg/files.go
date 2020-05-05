package ftp

// This represents a basic file system. Values can be either files or
// directory names.
var files = map[string]interface{}{
	"test": "Test file",
	"asdf": "asdf file",
	"dir": map[string]interface{}{
		"file1": "file1",
		"file2": "file2",
	},
}
