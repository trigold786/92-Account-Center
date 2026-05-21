package model

type DataSourceConfig struct {
	PrimaryDSN  string `json:"primary_dsn"`
	ReplicaDSN  string `json:"replica_dsn,omitempty"`
	MaxOpenConn int    `json:"max_open_conn,omitempty"`
	MaxIdleConn int    `json:"max_idle_conn,omitempty"`
}
