package main

type uploader interface {
	save([]byte) error
}

type uploaderFactory func(map[string]string) (uploader, error)

var (
	uploadDrivers = map[string]uploaderFactory{
		"google": googleStorageFromParameters,
	}
)
