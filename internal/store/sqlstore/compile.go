package sqlstore

import (
    "bytecourses/internal/store"
)

var _ store.UserStore = (*Store)(nil)
