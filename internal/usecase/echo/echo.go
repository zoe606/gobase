package echo

type UseCase struct{}

func New() *UseCase { return &UseCase{} }
