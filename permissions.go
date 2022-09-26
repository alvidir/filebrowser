package filebrowser

const (
	Read  Permission = 0x01
	Write Permission = 0x02
	Owner Permission = 0x04
)

type Permission uint8
