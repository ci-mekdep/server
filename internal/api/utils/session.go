package utils

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

type TokenClaim struct {
	userId   string
	schoolId string
	periodId string
	// regionId uint
	roleCode models.Role
}
type Session struct {
	token   string
	claim   TokenClaim
	ctx     context.Context
	model   models.Session
	user    models.User
	role    models.Role
	period  models.Period
	school  models.School
	schools []models.UserSchool
}

func InitSession(c *gin.Context) Session {
	token := c.GetString("token")
	model, err := SessionByToken(token)
	model, err = InitSessionTest(err, model, token)
	if err != nil {
		return Session{
			ctx: context.Background(),
		}
	}
	return Session{
		token: token,
		model: model,
		ctx:   c.Request.Context(),
		claim: TokenClaim{
			userId:   string(c.GetString("user_id")),        // PrepareSession
			roleCode: models.Role(c.GetString("role_code")), // PrepareSession
			schoolId: string(c.GetString("school_id")),      // PrepareSession
			periodId: string(c.GetString("period_id")),      // PrepareSession
			// regionId: uint(c.GetInt("region_id")),
		},
		school: models.School{ID: string(c.GetString("school_id"))},
		user:   model.User,
	}
}

func InitSessionTest(err error, model models.Session, token string) (models.Session, error) {
	if !config.Conf.AppEnvIsProd {
		// TODO: add to env
		dt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTI5NzUzMzQsInJvbGVfY29kZSI6ImFkbWluIiwic2Nob29sX2lkIjowLCJ1c2VyX2lkIjo2MTUxODZ9.1a7dEJdgH35vUzs8dLI6To97_igYwGpL5bdkddmVg-s"
		if err != nil && token == dt {
			return models.Session{
				ID:     "1",
				UserId: "615186",
				Token:  dt,
			}, nil
		}
	}
	return model, err
}

func PrepareSession(c *gin.Context, token string, userId string, role string, schoolId string, periodId string) {
	c.Set("token", token)
	c.Set("user_id", userId)
	c.Set("school_id", schoolId)
	c.Set("period_id", periodId)
	c.Set("role_code", role)
}

func (ses *Session) LoadSession() error {
	// user, err := store.Store().UsersFindById(ses.Context(), ses.claim.userId)
	// if err != nil {
	// 	return err
	// }
	user := ses.user
	user.Role = new(string)
	*user.Role = string(ses.claim.roleCode)
	user.SchoolId = &ses.claim.schoolId
	// user. = &ses.claim.periodId
	return ses.SetUser(&user)
}

func (ses *Session) Context() context.Context {
	if ses.ctx == nil {
		return context.Background()
	}
	return ses.ctx
}

func (ses *Session) SetContext(c context.Context) {
	ses.ctx = c
}

func (ses *Session) GetAvailableRole(r models.Role, sId *string) *models.UserSchool {
	if sId != nil && *sId == "" {
		sId = nil
	}
	for _, s := range ses.schools {
		if s.RoleCode == r {
			if sId != nil && s.School != nil && s.School.ID == *sId {
				return &s
			} else if sId == nil {
				return &s
			}
		}

	}
	return nil
}

func (ses *Session) GetSessionId() *string {
	if ses.model.ID == "" {
		return nil
	}
	return &ses.model.ID
}

// Set session user
// Set available schools for user (permission control)
// check available roles for user (permission control
func (ses *Session) SetUser(user *models.User) error {
	if user != nil {
		ses.user = *user                                   //
		ses.role = ses.claim.roleCode                      // selectedRole
		ses.period = models.Period{ID: ses.claim.periodId} // selectedRole
		ses.school = models.School{ID: ses.claim.schoolId} // selectedSchool
		ses.schools = []models.UserSchool{}                // availableSchools // userSchool haysy mekdepde haysy role da otyrn
		schoolId := ""
		for _, userSchool := range user.Schools {
			ses.schools = append(ses.schools, *userSchool)
			if userSchool.RoleCode == ses.role {
				if userSchool.SchoolUid != nil {
					schoolId = *userSchool.SchoolUid
				}
			}
		}

		// set availableSchools for dynamic roles like admin, organization
		if ses.claim.periodId != "" {
			per, _ := store.Store().PeriodsFindById(ses.Context(), ses.claim.periodId)
			if per != nil {
				ses.period = *per
			}
		}
		if ses.role == models.RoleOrganization {
			//ses.schools = []models.UserSchool{}
			parentUids := []string{} // ses.school.ID
			for _, userSchool := range user.Schools {
				if userSchool.RoleCode != models.RoleOrganization {
					continue
				}
				// optional code here because we could just dump error "no region set for organization"
				if userSchool.School == nil {
					return errors.New("organization has no school")
				}

				parentUids = append(parentUids, userSchool.School.ID)
				for province, regions := range models.Regions {
					if province == *userSchool.School.Code {
						regionModels, _, err := store.Store().SchoolsFindBy(ses.Context(), models.SchoolFilterRequest{
							Codes: &regions,
						})
						if err != nil {
							return err
						}
						for _, v := range regionModels {
							parentUids = append(parentUids, v.ID)
						}
					}
					for _, region := range regions {
						if region == *userSchool.School.Code {
							parentUids = append(parentUids, userSchool.School.ID)
						}
					}
				}
			}
			getSchoolsDto := models.SchoolFilterRequest{
				ParentUids: &parentUids,
			}
			getSchoolsDto.Limit = new(int)
			*getSchoolsDto.Limit = 1000
			schools, _, err := store.Store().SchoolsFindBy(ses.Context(), getSchoolsDto)
			if err != nil {
				return err
			}
			for _, school := range schools {
				ses.schools = append(ses.schools, models.UserSchool{
					SchoolUid: &school.ID,
					School:    school,
					User:      user,
					RoleCode:  ses.role,
				})
			}
		} else if ses.role == models.RoleAdmin || ses.role == models.RoleOperator {
			getSchoolsDto := models.SchoolFilterRequest{}
			getSchoolsDto.Limit = new(int)
			*getSchoolsDto.Limit = 1000
			if schoolId != "" {
				getSchoolsDto.ID = new(string)
				*getSchoolsDto.ID = schoolId
			}
			// getSchoolsDto.IsSecondarySchool = new(bool)
			// *getSchoolsDto.IsSecondarySchool = true
			schools, _, err := store.Store().SchoolsFindBy(ses.Context(), getSchoolsDto)
			if err != nil {
				return err
			}
			for _, item := range schools {
				// if parentId is not null then eq selectedSchoolId (etrap saylanan)
				// else if parentId is null then ok (bu etrap)
				// else if selectedSchoolId is null then ok (ahlisi)
				// else if selectedSchoolParentId is null then ok (mekdep saylandy)
				if item.ParentUid != nil && *item.ParentUid == ses.school.ID || // etrap saylamak
					ses.school.ParentUid == nil || // etrap saylanan
					item.ParentUid == nil || // hemme etrap ok
					ses.school.ID == "" { // ahlisi
					ses.schools = append(ses.schools, models.UserSchool{
						SchoolUid: &item.ID,
						School:    item,
						User:      user,
						RoleCode:  ses.role,
					})
				}
			}
		}
		// filter access school-role
		if len(ses.schools) > 0 {
			if userSchool := ses.GetAvailableRole(ses.claim.roleCode, &ses.claim.schoolId); userSchool != nil {
				if userSchool.School != nil {
					ses.school = *userSchool.School
				}
				ses.role = userSchool.RoleCode
			} else {
				return errors.New("no result in GetAvailableRole: " + string(ses.claim.roleCode) + " - " + (ses.claim.schoolId))
			}
		} else {
			return errors.New("user does not have any role id: " + user.ID)
		}
	}
	return nil
}

func (ses Session) GetUser() *models.User {
	if ses.user.ID != "" {
		return &ses.user
	}
	return nil
}

func (ses Session) GetSchool() *models.School {
	if ses.school.ID != "" && ses.school.ParentUid != nil {
		return &ses.school
	}
	return nil
}

func (ses Session) GetPeriod() *models.Period {
	if ses.period.ID != "" {
		return &ses.period
	}
	return nil
}

func (ses Session) GetRole() *models.Role {
	if ses.role != "" {
		return &ses.role
	}
	if ses.claim.roleCode != "" {
		return &ses.claim.roleCode
	}
	return nil
}

func (ses Session) GetSchools() []models.UserSchool {
	l := []models.UserSchool{}
	for _, v := range ses.schools {
		if v.School == nil || *v.School.IsSecondarySchool {
			l = append(l, v)
		}
	}
	return l
}

func (ses Session) GetSchoolIds() []string {
	s := ses.GetSchool()
	if s == nil {
		return ses.SchoolAllIds()
	}
	return []string{s.ID}
}

func (ses Session) GetSchoolId() *string {
	s := ses.GetSchool()
	if s != nil {
		return &s.ID
	}
	return nil
}

func (ses Session) GetPeriodId() *string {
	s := ses.GetPeriod()
	if s != nil {
		return &s.ID
	}
	return nil
}

func (ses Session) GetSchoolIdByFilter(id *string) *string {
	if ses.role == models.RoleAdmin {
		return id
	}
	l := ses.SchoolAllIds()
	for _, v := range l {
		if id != nil && v == *id {
			return id
		}
	}
	return ses.GetSchoolId()
}

func (ses Session) SchoolAllIds() []string {
	sl := ses.GetSchools()
	if sl != nil {
		ids := []string{}
		for _, v := range sl {
			if v.School != nil {
				ids = append(ids, v.School.ID)
			}
		}
		return ids
	}
	return nil
}

func (ses Session) SchoolsWithCentersIds() []string {
	sl := ses.schools
	if sl != nil {
		ids := []string{}
		for _, v := range sl {
			if v.School != nil {
				ids = append(ids, v.School.ID)
			}
		}
		return ids
	}
	return nil
}

func (ses Session) GetToken() string {
	return ses.token
}

func (ses *Session) GetSchoolsByAdminRoles() []string {
	sl := ses.GetSchools()
	if sl != nil {
		ids := []string{}
		for _, v := range sl {
			if v.RoleCode == models.RoleAdmin || v.RoleCode == models.RoleOperator || v.RoleCode == models.RoleOrganization || v.RoleCode == models.RolePrincipal {
				if v.School != nil {
					ids = append(ids, v.School.ID)
				}
			}
		}
		return ids
	}
	return nil
}
