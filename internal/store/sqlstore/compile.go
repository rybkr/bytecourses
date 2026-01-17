package sqlstore

import (
	"bytecourses/internal/store"
)

var _ store.UserStore = (*Store)(nil)
var _ store.ProposalStore = (*Store)(nil)
var _ store.CourseStore = (*Store)(nil)
var _ store.ModuleStore = (*Store)(nil)
var _ store.ContentStore = (*Store)(nil)
