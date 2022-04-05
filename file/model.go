package file

const (
	Public  = 0x01
	Private = 0x02
	Read    = 0x04
	Write   = 0x08
	Share   = 0x16
	Owner   = 0x32
)

type Flags uint8

type Permissions map[int32]uint8

type File struct {
	Name        string
	Metadata    map[string]string
	Permissions Permissions
	Flags       Flags
	Value       []byte
}
