package auth

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"rgt-server/log"
	"rgt-server/option"
	"rgt-server/security"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type ldapAttributes struct {
	userGroups    string
	userStatus    string
	username      string
	userType      string
	accountStatus string
}

type ldapValues struct {
	usersBase           string
	userType            string
	userStatusActive    string
	groupsBase          string
	accountStatusActive string
	groups              []string
}

type LDAPAuthenticator struct {
	address       *url.URL
	queryUser     string
	queryPassword string
	attribute     ldapAttributes
	value         ldapValues
}

type LDAPAuthenticatorFactory struct{}

func init() {
	AddAuthenticator("ldap", &LDAPAuthenticatorFactory{})
}

func (f *LDAPAuthenticatorFactory) Create(prefix string, conf map[string]option.Option) UserAuthenticator {
	ldapAuth := &LDAPAuthenticator{
		address:       ldapUrl(getProperty(conf, prefix, "address")),
		queryUser:     getProperty(conf, prefix, "query.user"),
		queryPassword: getProperty(conf, prefix, "query.pswd"),
		attribute: ldapAttributes{userGroups: getProperty(conf, prefix, "user.groups.attribute"),
			userStatus:    getProperty(conf, prefix, "user.status.attribute"),
			accountStatus: getProperty(conf, prefix, "account.status.attribute"),
			username:      getProperty(conf, prefix, "user.name.attribute"),
			userType:      getProperty(conf, prefix, "user.type.attribute")},
		value: ldapValues{usersBase: getProperty(conf, prefix, "users.base"),
			userType:            getProperty(conf, prefix, "user.type"),
			groupsBase:          getProperty(conf, prefix, "groups.base"),
			userStatusActive:    getProperty(conf, prefix, "user.status.active"),
			accountStatusActive: getProperty(conf, prefix, "account.status.active"),
			groups:              strings.Split(getProperty(conf, prefix, "groups"), ";")}}
	return ldapAuth
}

func getProperty(conf map[string]option.Option, prefix string, propName string) string {
	value, found := conf[prefix+".ldap."+propName]
	if found {
		return value.GetString()
	}
	return ""
}

func ldapUrl(address string) *url.URL {
	if !strings.HasPrefix(address, "ldap://") && !strings.HasPrefix(address, "ldaps://") {
		address = "ldap://" + address
	}
	addr, err := url.Parse(address)
	if err != nil {
		log.Errorf("Error parsing LDAP server address: %s", err)
		return nil
	}
	if addr.Port() == "" {
		if addr.Scheme == "ldaps" {
			addr.Host = addr.Host + ":636"
		} else {
			addr.Host = addr.Host + ":389"
		}
	}
	return addr
}

func (p *LDAPAuthenticator) connect(username, password string) (*ldap.Conn, error) {
	conn, err := ldap.DialURL(p.address.String())
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(password) == "" {
		err = conn.UnauthenticatedBind(username)
	} else {
		err = conn.Bind(username, password)
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (p *LDAPAuthenticator) searchUser(conn *ldap.Conn, username string) *ldap.Entry {
	var filter string
	if p.attribute.userType != "" {
		filter = fmt.Sprintf("(&(%s=%s)(%s=%s))", p.attribute.userType, p.value.userType, p.attribute.username, username)
	} else {
		filter = fmt.Sprintf("(&(%s=%s))", p.attribute.username, username)
	}

	searchReq := ldap.NewSearchRequest(
		p.value.usersBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{},
		nil)
	result, err := conn.Search(searchReq)
	if err != nil {
		log.Errorf("LDAP Search user Error: %s", err)
		return nil
	}

	if len(result.Entries) == 0 {
		log.Debugf("LDAPAuthenticator.searchUser(). Entry not found for user %s ", username)
		return nil
	}
	return result.Entries[0]
}

func contaisGroup(groups []string, groupSearch string) bool {
	for _, group := range groups {
		if group == groupSearch {
			return true
		}
	}
	return false
}

func (p *LDAPAuthenticator) authorizedUser(username, password string, userEntry *ldap.Entry) bool {
	conn, err := p.connect(userEntry.DN, password)
	if err != nil {
		log.Errorf("Error authorazing user %s: %v", username, err)
		return false
	}
	conn.Close()
	// Any aunthenticated user
	if len(p.value.groups) == 0 {
		return true
	}
	for _, group := range p.value.groups {
		if strings.HasSuffix(userEntry.DN, group) {
			return true
		}
	}
	for _, attribute := range userEntry.Attributes {
		if strings.EqualFold(p.attribute.userGroups, attribute.Name) {
			for _, group := range attribute.Values {
				if contaisGroup(p.value.groups, group) {
					return true
				}
			}
		}
	}
	return false
}

func decryptPassword(password string) string {
	if strings.HasPrefix(password, "{") {
		start := strings.IndexRune(password, '{')
		end := strings.IndexRune(password, '}')
		if end-start > 1 {
			cipher := security.GetCipher(password[start+1 : end])
			if cipher != nil {
				pswd, err := hex.DecodeString(password[end+1:])
				if err == nil {
					return string(cipher.Decrypt(pswd))
				}
			}
		}
	}
	return password
}

func (p *LDAPAuthenticator) Authenticate(username, password string) bool {
	conn, err := p.connect(p.queryUser, p.queryPassword)
	if err != nil {
		log.Errorf("Error connecting to LDAP: %v", err)
		return false
	}
	defer conn.Close()
	userEntry := p.searchUser(conn, username)
	if userEntry == nil {
		return false
	}
	return p.authorizedUser(username, decryptPassword(password), userEntry)
}
