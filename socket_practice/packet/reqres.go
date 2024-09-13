package packet

type RegisterRequest struct {
	Email           string `msgpack: "email"`
	NickName        string `msgpack: "nickname"`
	Password        string `msgpack: "password"`
	ConfirmPassword string `msgpack: "confirm_password"`
}

type RegisterRespone struct {
	ErrorCode int64 `msgpack: "error_code"`
}

type LoginRequest struct {
	Email    string `msgpack: "email"`
	Password string `msgpack: "password"`
}

type LoginRespone struct {
	ErrorCode int64  `msgpack: "error_code"`
	Token     string `msgpack: "token"` // 인증된 사용자인지 확인
}

type UserInfoRequest struct {
	Token string `msgpack: "token"`
}

type UserInfoReponse struct {
	NickName string `msgpack: "nickname"`
	Email    string `msgpack: "email"`
}

func CreateRegisterRequest() {

}

func ParseRegisterRequest() {

}

func CreateRegisterReponse() {

}

func ParseRegisterResponse() {

}

func CreateLoginRequest() {

}

func ParseLoginRequest() {

}

func CreateLoginResponse() {

}

func ParseLoginRespone() {

}
