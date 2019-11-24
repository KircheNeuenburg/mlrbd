package ldap

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"log"
	"mlrbd/config"
)

var (
	l *ldap.Conn
	c *config.Config
)

func StartLDAP(conf *config.Config) error {
	if conf == nil {
		log.Fatal("No config loaded")
	}
	c = conf
	if l == nil {
		var err error
		l, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Ldap.Server, c.Ldap.Port))
		if err != nil {
			log.Fatal(err)
		}
		//l.Start()
		err = l.Bind(c.Ldap.BindDn, c.Ldap.BindPassword)
		if err != nil {
			log.Fatal(err)
		}
		return err
	}
	return nil
}

func StopLDAP() {
	if l != nil {
		log.Printf("Stopping LDAP")
		l.Close()
		log.Printf("LDAP stopped")
		l = nil
	}
}

func GetLDAPGroups() (lg []string, err error) {
	searchRequest := ldap.NewSearchRequest(
		c.Ldap.GroupBaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		c.Ldap.GroupFilter, []string{c.Ldap.GroupUniqueIdentifier}, nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range sr.Entries {
		lg = append(lg, e.GetAttributeValue(c.Ldap.GroupUniqueIdentifier))
	}
	return
}

func GetLdapGroupName(uuid string) (n string) {
	filter := "(&(" + c.Ldap.GroupUniqueIdentifier + "=" + uuid + ")(" + c.Ldap.GroupFilter + "))"
	searchRequest := ldap.NewSearchRequest(
		c.Ldap.GroupBaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter, []string{c.Ldap.GroupName}, nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}
	if len(sr.Entries) == 1 {
		n = sr.Entries[0].GetAttributeValue(c.Ldap.GroupName)
	}
	return
}

func GetLdapUserId(memberAttr string) (n string) {
	filter := "(&(" + c.Ldap.GroupMemberAttribute + "=" + memberAttr + ")" + c.Ldap.UserFilter + ")"
	searchRequest := ldap.NewSearchRequest(
		c.Ldap.UserBaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter, []string{c.Ldap.UserLoginAttribute}, nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}
	if len(sr.Entries) == 1 {
		n = sr.Entries[0].GetAttributeValue(c.Ldap.UserLoginAttribute)
	}
	return
}

func GetLdapUsers(lg string) (lu []string, err error) {
	filter := "(&(" + c.Ldap.GroupUniqueIdentifier + "=" + lg + ")" + c.Ldap.GroupFilter + ")"
	searchRequest := ldap.NewSearchRequest(
		c.Ldap.GroupBaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, 
		filter, []string{c.Ldap.GroupMemberAssociation}, nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}
	var users []string
	if len(sr.Entries) == 1 {
		users = sr.Entries[0].GetAttributeValues(c.Ldap.GroupMemberAssociation)
	}
	for _, u := range users {
		if id := GetLdapUserId(u); id != "" {
			lu = append(lu, id)
		}
	}
	return
}
