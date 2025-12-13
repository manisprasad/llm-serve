package types

type Chat struct {
	ModelName string `validate:"required"`
	UserId    string `validate:"required"`
	Message   string `validare:"required"`
	Stream    bool
}
