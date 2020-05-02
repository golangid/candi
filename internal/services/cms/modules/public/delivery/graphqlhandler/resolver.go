package graphqlhandler

import (
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

type HomepageResolver struct {
	Content string
	Skills  []string
	Footer  string
}

type VisitorListResolver struct {
	m      *shared.Meta
	events []*VisitorResolver
}

func (r *VisitorListResolver) Meta() *shared.MetaResolver {
	return r.m.ToResolver()
}
func (r *VisitorListResolver) Data() []*VisitorResolver {
	return r.events
}

type VisitorResolver struct {
	v domain.Visitor
}

func (r *VisitorResolver) ID() string {
	return r.v.ID.Hex()
}
func (r *VisitorResolver) IPAddress() string {
	return r.v.IPAddress
}
func (r *VisitorResolver) UserAgent() string {
	return r.v.UserAgent
}
func (r *VisitorResolver) CreatedAt() string {
	return r.v.CreatedAt.Format(time.RFC3339)
}
func (r *VisitorResolver) ModifiedAt() string {
	return r.v.ModifiedAt.Format(time.RFC3339)
}
