package sqlstore

import (
	"bytecourses/internal/store"
)

var _ store.UserStore = (*Store)(nil)
var _ store.ProposalStore = (*Store)(nil)
