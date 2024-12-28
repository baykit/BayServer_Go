package groups

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/bcf/impl"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/md5password"
	"strings"
)

/****************************************/
/* Struct Member                        */
/****************************************/

type Member struct {
	Name   string
	Digest string
}

func NewMember(name string, digest string) *Member {
	return &Member{
		Name:   name,
		Digest: digest,
	}
}

func (m *Member) validate(password string) bool {
	if password == "" {
		return false
	}
	dig := md5password.Encode(password)
	return m.Digest == dig
}

/****************************************/
/* Struct Group                         */
/****************************************/

type Group struct {
	Name    string
	Members []string
}

func NewGroup(name string) *Group {
	return &Group{
		Name:    name,
		Members: []string{},
	}
}

func (g *Group) Add(name string) {
	g.Members = append(g.Members, name)
}

func (g *Group) Validate(mName string, pass string) bool {
	if !arrayutil.Contains(g.Members, mName) {
		return false
	}

	m := allMembers[mName]
	if m == nil {
		return false
	}

	return m.validate(pass)
}

/****************************************/
/* Public functions                     */
/****************************************/

var allGroups = map[string]*Group{}
var allMembers = map[string]*Member{}

func GroupsInit(conf string) exception.ConfigException {
	p := impl.NewBcfParser()
	doc, err := p.Parse(conf)
	if err != nil {
		return err
	}

	for _, o := range doc.ContentList {
		if elm, ok := o.(*bcf.BcfElement); ok {
			if strings.ToLower(elm.Name) == "group" {
				initGroups(elm)

			} else if strings.ToLower(elm.Name) == "member" {
				initMembers(elm)
			}
		}
	}
	return nil
}

func GetGroup(name string) *Group {
	return allGroups[name]
}

func initGroups(elm *bcf.BcfElement) {
	for _, o := range elm.ContentList {
		if kv, ok := o.(*bcf.BcfKeyVal); ok {
			grp := NewGroup(kv.Key)
			allGroups[grp.Name] = grp
			members := strings.Fields(kv.Value)
			for _, member := range members {
				grp.Add(member)
			}
		}
	}
}

func initMembers(elm *bcf.BcfElement) {
	for _, o := range elm.ContentList {
		if kv, ok := o.(*bcf.BcfKeyVal); ok {
			m := NewMember(kv.Key, kv.Value)
			allMembers[m.Name] = m
		}
	}
}
