package db

import "time"

type User struct {
	UserId         int
	Username       string
	Email		   string
	CreationSource UserCreationSource
	CreationDate   time.Time
	UserRole       UserRoles
	UserPrivileges UserPrivileges
}

type UserRoles int

const (
	VoterRole          UserRoles = iota // 0
	OneSubmissionRole                   // 1
	CuratorRole                         // 5
	TrustedCuratorRole                  // 15
	UnlimitedRole                       // unlimited
)

var userRolesMap = map[int]string{
	VoterRole.EnumIndex():          "Voter",
	OneSubmissionRole.EnumIndex():  "OneSubmission",
	CuratorRole.EnumIndex():        "Curator",
	TrustedCuratorRole.EnumIndex(): "TrustedCurator",
	UnlimitedRole.EnumIndex():      "Unlimited",
}

func (ur UserRoles) String() string {
	return userRolesMap[ur.EnumIndex()]
}
func (ur UserRoles) EnumIndex() int {
	return int(ur)
}
func StringToUserRoles(s string) UserRoles {
	switch s {
	case userRolesMap[VoterRole.EnumIndex()]:
		return VoterRole
	default:
		return VoterRole
	}
}

type UserPrivileges int

const (
	NoPrivileges UserPrivileges = iota
	ModeratorPrivileges
	AdminPrivileges
	OwnerPrivileges
)

var userPrivilegesMap = map[int]string{
	NoPrivileges.EnumIndex():        "None",
	ModeratorPrivileges.EnumIndex(): "Moderator",
	AdminPrivileges.EnumIndex():     "Admin",
	OwnerPrivileges.EnumIndex():     "Owner",
}

func StringToUserPrivileges(s string) UserPrivileges {
	switch s {
	case userPrivilegesMap[NoPrivileges.EnumIndex()]:
		return NoPrivileges
	default:
		return NoPrivileges
	}
}

func (up UserPrivileges) String() string {
	return userPrivilegesMap[up.EnumIndex()]
}
func (up UserPrivileges) EnumIndex() int {
	return int(up)
}

type UserCreationSource int

var userCreationSourceMap = map[int]string{
	LocalUserCreationSource.EnumIndex(): "Local",
}

const (
	LocalUserCreationSource UserCreationSource = iota
)

func StringToUserCreationSource(s string) UserCreationSource {
	switch s {
	default:
		return LocalUserCreationSource
	}
}

func (up UserCreationSource) String() string {
	return userCreationSourceMap[up.EnumIndex()]
}
func (up UserCreationSource) EnumIndex() int {
	return int(up)
}
