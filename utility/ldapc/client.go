package ldapc

// Package ldapc provides easy LDAP v3 authentication.
// Set LDAPC_DEBUG=yes to environment value then print debug log
import (
	"crypto/tls"
	"fmt"

	"github.com/michibiki-io/ldap-jwt-go/utility"
	"gopkg.in/ldap.v2"
)

// Protocol:  LDAP, LDAPS and START_TLS
type Protocol int

const (
	LDAP      Protocol = iota // No encrypted protocol
	LDAPS                     // SSL protocol
	START_TLS                 // TLS protocol
)

func searchFilter(filter, value string) string {
	utility.Log.Debug("Search: filter: %v, key: %v\n", filter, value)
	return fmt.Sprintf(filter, value)
}

func search(conn *ldap.Conn, baseDN string, filter string) ([]*ldap.Entry, error) {
	utility.Log.Debug("Search: dn: %v, baseDN: %v, filter: %v\n", baseDN, filter)
	request := ldap.NewSearchRequest(
		baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0,
		false, filter, nil, nil)

	result, err := conn.Search(request)
	if err != nil {
		return nil, utility.NewError(fmt.Sprintf("LDAP Search failed! (%v)", err), utility.InternalServerError)
	}

	return result.Entries, nil
}

func bind(conn *ldap.Conn, dn, password string) error {
	err := conn.Bind(dn, password)
	if err != nil {
		return utility.NewError(fmt.Sprintf("LDAP Bind error, %s:%v", dn, err), utility.InternalServerError)
	}
	return nil
}

// Client is a LDAP Client.
// Protocol, Host, Prot, Bind are required parameter.
// TLSConfig uses only Protocol is LDAP, LDAPS and START_TLS
type Client struct {
	Protocol  Protocol    // Security protocol. LDAP, LDAPS and START_TLS
	Host      string      // LDAP Server host
	Port      int         // Port number
	TLSConfig *tls.Config // TLSConfig used only LDAPS or START_TLS
	Bind      Bind        // Bind Information
}

func (c *Client) DoBind(dn, password string) error {
	conn, err := c.dial()
	if err != nil {
		return err
	}
	defer conn.Close()

	return bind(conn, dn, password)
}

func (c *Client) Search(filter, value string) ([]*ldap.Entry, error) {
	conn, err := c.dial()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	bind(conn, c.Bind.BindDN, c.Bind.BindPassword)

	return search(conn, c.Bind.BaseDN, searchFilter(filter, value))
}

func (c *Client) dial() (*ldap.Conn, error) {
	if c.Protocol == LDAPS {
		utility.Log.Debug("LDAP Auth : Start LDAPS Protocol\n")
		return ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), c.TLSConfig)
	}

	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return nil, fmt.Errorf("Dial: %v", err)
	}

	if c.Protocol == START_TLS {
		if err = conn.StartTLS(c.TLSConfig); err != nil {
			utility.Log.Debug("LDAP Auth : Start TLS Protocol\n")
			conn.Close()
			return nil, fmt.Errorf("StartTLS: %v", err)
		}
	}

	utility.Log.Debug("LDAP Auth : Start LDAP Protocol\n")

	return conn, nil
}
