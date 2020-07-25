package graphqlserver

// issue https://github.com/graph-gophers/graphql-go/issues/145
type rootResolver struct {
	rootQuery        interface{}
	rootMutation     interface{}
	rootSubscription interface{}
}

func (r *rootResolver) Query() interface{} {
	return r.rootQuery
}
func (r *rootResolver) Mutation() interface{} {
	return r.rootMutation
}
func (r *rootResolver) Subscription() interface{} {
	return r.rootSubscription
}

var root rootResolver

// SetRootSubscription public function
// this public method made because cannot create dynamic method for embedded struct (issue https://github.com/golang/go/issues/15924)
// and subscription in graphql can subscribe to at most one subscription at a time
func SetRootSubscription(subsRootResolver interface{}) {
	root.rootSubscription = subsRootResolver
}
