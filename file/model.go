package file

const (
	Public  = 0x01
	Private = 0x02
	Read    = 0x04
	Write   = 0x08
	Share   = 0x16
	Owner   = 0x32
)

type Metadata map[string]string
type Permissions map[string]uint8

type File struct {
	id          string
	name        string
	metadata    Metadata
	permissions Permissions
	flags       uint8
	value       []byte
}
