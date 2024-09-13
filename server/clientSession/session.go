package clientsession

type ClientSession struct {
	SessionUniqueID uint64
	SessionID       int32
	UserID          []byte
	UserIDLength    int8
	IsAuth          bool
}
