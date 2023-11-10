package laurel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	// "os"
)

type Row struct {
	id       uint32
	username [COLUMN_USERNAME_SIZE]byte
	email    [COLUMN_EMAIL_SIZE]byte
}

func print_row(row *Row) {
	s := fmt.Sprintf("(%d, %s, %s)\n", row.id, bytes.TrimRight(row.username[:], "\x00"), bytes.TrimRight(row.email[:], "\x00"))
	// os.Stdout.WriteString(s)
	PrintMsgf(s)
}

func serialize_row(source *Row, destination []byte) {
	binary.LittleEndian.PutUint32(destination[ID_OFFSET:], source.id)
	copy(destination[USERNAME_OFFSET:], source.username[:])
	copy(destination[EMAIL_OFFSET:], source.email[:])
}

func deserialize_row(source []byte, destination *Row) {
	destination.id = binary.LittleEndian.Uint32(source[ID_OFFSET:])
	copy(destination.username[:], source[USERNAME_OFFSET:USERNAME_OFFSET+USERNAME_SIZE])
	copy(destination.email[:], source[EMAIL_OFFSET:EMAIL_OFFSET+EMAIL_SIZE])
}
