package users

// Role is a named, admin-manageable grouping of permissions. Roles are data,
// not a fixed enum — admins create/rename/delete roles and edit each role's
// permission set at runtime.
type Role struct {
	ID   int64
	Name string
}

// Permission is one control-plane action, owned by a registered service and
// addressed by the dotted name service_name.feature_group.feature.action.
type Permission struct {
	ID           int64
	ServiceID    int64
	ServiceKey   string
	FeatureGroup string
	Feature      string
	Action       string
}

// DottedName returns the full service_name.feature_group.feature.action
// string clients send/receive.
func (p Permission) DottedName() string {
	return p.ServiceKey + "." + p.FeatureGroup + "." + p.Feature + "." + p.Action
}
