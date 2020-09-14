package shared

// GraphQLErrorResolver graphql error with extensions
type GraphQLErrorResolver interface {
	Error() string
	Extensions() map[string]interface{}
}

type resolveErrorImpl struct {
	message    string
	extensions map[string]interface{}
}

func (r *resolveErrorImpl) Error() string {
	return r.message
}
func (r *resolveErrorImpl) Extensions() map[string]interface{} {
	return r.extensions
}

// NewGraphQLErrorResolver constructor
func NewGraphQLErrorResolver(errMesage string, extensions map[string]interface{}) GraphQLErrorResolver {
	return &resolveErrorImpl{
		message: errMesage, extensions: extensions,
	}
}
