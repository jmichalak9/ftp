package ftp

// This represents a basic file system. Values can be either files or
// directory names.
var files = Directory(map[string]interface{}{
	"test": File("Test file"),
	"asdf": File("asdf file"),
	"dir": Directory(map[string]interface{}{
		"file1": File("file1"),
		"file2": File("file2"),
	}),
})
