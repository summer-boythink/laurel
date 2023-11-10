package laurel

import "unsafe"

const (
	ID_SIZE              = uint32(unsafe.Sizeof(Row{}.id))
	USERNAME_SIZE        = uint32(unsafe.Sizeof(Row{}.username))
	EMAIL_SIZE           = uint32(unsafe.Sizeof(Row{}.email))
	ID_OFFSET            = 0
	USERNAME_OFFSET      = ID_OFFSET + ID_SIZE
	EMAIL_OFFSET         = USERNAME_OFFSET + USERNAME_SIZE
	ROW_SIZE             = ID_SIZE + USERNAME_SIZE + EMAIL_SIZE
	PAGE_SIZE            = 4096
	ROWS_PER_PAGE        = PAGE_SIZE / ROW_SIZE
	TABLE_MAX_ROWS       = ROWS_PER_PAGE * TABLE_MAX_PAGES
	COLUMN_USERNAME_SIZE = 32
	COLUMN_EMAIL_SIZE    = 255
	TABLE_MAX_PAGES      = 100
)
