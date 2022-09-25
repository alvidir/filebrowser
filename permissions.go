package filebrowser

const (
	Read  Permissions = 0x01
	Write Permissions = 0x02
	Owner Permissions = 0x04
)

type Permissions uint8
