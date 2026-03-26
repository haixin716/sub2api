package ent

import "entgo.io/ent/dialect"

// Driver returns the underlying dialect.Driver of the client.
// This extension method exists so that callers such as ProvideSQLDB
// can extract the underlying *sql.DB from the Ent client.
func (c *Client) Driver() dialect.Driver {
	return c.config.driver
}
