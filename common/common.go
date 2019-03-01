package common

type CreateTokenInput struct {
	WriteOnly        bool   `json:"writeonly"`
	NumberOfArchives uint64 `json:"number_of_archives"`
}

type CreateUserInput struct {
	Email    string `binding:"required" json:"email"`
	Password string `binding:"required" json:"password"`
}

type CreateUserOutput struct {
	Secret  string `json:"secret"`
	TokenWO string `json:"token_wo"`
	TokenRW string `json:"token_rw"`
}

type QueryError struct {
	Error string `json:"error"`
}
