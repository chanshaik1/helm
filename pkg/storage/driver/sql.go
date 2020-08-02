/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver // import "helm.sh/helm/v3/pkg/storage/driver"

import (
	"fmt"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"

	sq "github.com/Masterminds/squirrel"

	// Import pq for postgres dialect
	_ "github.com/lib/pq"

	rspb "helm.sh/helm/v3/pkg/release"
)

var _ Driver = (*SQL)(nil)

var LabelMap = map[string]struct{}{
	"modifiedAt": {},
	"createdAt":  {},
	"version":    {},
	"status":     {},
	"owner":      {},
	"name":       {},
}

const postgreSQLDialect = "postgres"

// SQLDriverName is the string name of this driver.
const SQLDriverName = "SQL"

const (
	sqlReleaseTableName = "releases_v1"
	sqlLabelTableName   = "Labels_v1"

	sqlReleaseTableKeyColumn        = "key"
	sqlReleaseTableTypeColumn       = "type"
	sqlReleaseTableBodyColumn       = "body"
	sqlReleaseTableNameColumn       = "name"
	sqlReleaseTableNamespaceColumn  = "namespace"
	sqlReleaseTableVersionColumn    = "version"
	sqlReleaseTableStatusColumn     = "status"
	sqlReleaseTableOwnerColumn      = "owner"
	sqlReleaseTableCreatedAtColumn  = "createdAt"
	sqlReleaseTableModifiedAtColumn = "modifiedAt"

	sqlLabelTableReleaseKeyColumn       = "releaseKey"
	sqlLabelTableReleaseNamespaceColumn = "releaseNamespace"
	sqlLabelTableKeyColumn              = "key"
	sqlLabelTableValueColumn            = "value"
)

const (
	sqlReleaseDefaultOwner = "helm"
	sqlReleaseDefaultType  = "helm.sh/release.v1"
)

// SQL is the sql storage driver implementation.
type SQL struct {
	db               *sqlx.DB
	namespace        string
	statementBuilder sq.StatementBuilderType

	Log func(string, ...interface{})
}

// Name returns the name of the driver.
func (s *SQL) Name() string {
	return SQLDriverName
}

func (s *SQL) ensureDBSetup() error {
	// Populate the database with the relations we need if they don't exist yet
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "init",
				Up: []string{
					fmt.Sprintf(`
						CREATE TABLE %s (
							%s VARCHAR(67),
							%s VARCHAR(64) NOT NULL,
							%s TEXT NOT NULL,
							%s VARCHAR(64) NOT NULL,
							%s VARCHAR(64) NOT NULL,
							%s INTEGER NOT NULL,
							%s TEXT NOT NULL,
							%s TEXT NOT NULL,
							%s INTEGER NOT NULL,
							%s INTEGER NOT NULL DEFAULT 0,
							PRIMARY KEY(%s, %s)
						);
						CREATE INDEX ON %s (%s, %s);
						CREATE INDEX ON %s (%s);
						CREATE INDEX ON %s (%s);
						CREATE INDEX ON %s (%s);
						CREATE INDEX ON %s (%s);
						CREATE INDEX ON %s (%s);
						
						GRANT ALL ON %s TO PUBLIC;

						ALTER TABLE %s ENABLE ROW LEVEL SECURITY;
					`,
						sqlReleaseTableName,
						sqlReleaseTableKeyColumn,
						sqlReleaseTableTypeColumn,
						sqlReleaseTableBodyColumn,
						sqlReleaseTableNameColumn,
						sqlReleaseTableNamespaceColumn,
						sqlReleaseTableVersionColumn,
						sqlReleaseTableStatusColumn,
						sqlReleaseTableOwnerColumn,
						sqlReleaseTableCreatedAtColumn,
						sqlReleaseTableModifiedAtColumn,
						sqlReleaseTableKeyColumn,
						sqlReleaseTableNamespaceColumn,
						sqlReleaseTableName,
						sqlReleaseTableKeyColumn,
						sqlReleaseTableNamespaceColumn,
						sqlReleaseTableName,
						sqlReleaseTableVersionColumn,
						sqlReleaseTableName,
						sqlReleaseTableStatusColumn,
						sqlReleaseTableName,
						sqlReleaseTableOwnerColumn,
						sqlReleaseTableName,
						sqlReleaseTableCreatedAtColumn,
						sqlReleaseTableName,
						sqlReleaseTableModifiedAtColumn,
						sqlReleaseTableName,
						sqlReleaseTableName,
					),
				},
				Down: []string{
					fmt.Sprintf(`
						DROP TABLE %s;
					`, sqlReleaseTableName),
				},
			},
			{
				Id: "Labels",
				Up: []string{
					fmt.Sprintf(`
						CREATE TABLE %s (
							%s VARCHAR(67),
							%s VARCHAR(67),
							%s VARCHAR(64), 
							%s VARCHAR(67)
						);
						CREATE INDEX ON %s (%s, %s);
						
						GRANT ALL ON %s TO PUBLIC;

						ALTER TABLE %s ENABLE ROW LEVEL SECURITY;
					`,
						sqlLabelTableName,
						sqlLabelTableReleaseKeyColumn,
						sqlLabelTableReleaseNamespaceColumn,
						sqlLabelTableKeyColumn,
						sqlLabelTableValueColumn,
						sqlLabelTableName,
						sqlLabelTableReleaseKeyColumn,
						sqlLabelTableReleaseNamespaceColumn,
						sqlLabelTableName,
						sqlLabelTableName,
					),
				},
				Down: []string{
					fmt.Sprintf(`
						DELETE TABLE %s;
					`, sqlLabelTableName),
				},
			},
		},
	}

	_, err := migrate.Exec(s.db.DB, postgreSQLDialect, migrations, migrate.Up)
	return err
}

// SQLReleaseWrapper describes how Helm releases are stored in an SQL database
type SQLReleaseWrapper struct {
	// The primary key, made of {release-name}.{release-version}
	Key string `db:"key"`

	// See https://github.com/helm/helm/blob/master/pkg/storage/driver/secrets.go#L236
	Type string `db:"type"`

	// The rspb.Release body, as a base64-encoded string
	Body string `db:"body"`

	// Release "Labels" that can be used as filters in the storage.Query(Labels map[string]string)
	// we implemented. Note that allowing Helm users to filter against new dimensions will require a
	// new migration to be added, and the Create and/or update functions to be updated accordingly.
	Name       string `db:"name"`
	Namespace  string `db:"namespace"`
	Version    int    `db:"version"`
	Status     string `db:"status"`
	Owner      string `db:"owner"`
	CreatedAt  int    `db:"createdAt"`
	ModifiedAt int    `db:"modifiedAt"`
}

type SQLReleaseLabelWrapper struct {
	ReleaseKey       string `db:"release_key"`
	ReleaseNamespace string `db:"release_namespace"`
	Key              string `db:"key"`
	Value            string `db:"value"`
}

// NewSQL initializes a new sql driver.
func NewSQL(connectionString string, logger func(string, ...interface{}), namespace string) (*SQL, error) {
	db, err := sqlx.Connect(postgreSQLDialect, connectionString)
	if err != nil {
		return nil, err
	}

	driver := &SQL{
		db:               db,
		Log:              logger,
		statementBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}

	if err := driver.ensureDBSetup(); err != nil {
		return nil, err
	}

	driver.namespace = namespace

	return driver, nil
}

// Get returns the release named by key.
func (s *SQL) Get(key string) (*rspb.Release, error) {
	var record SQLReleaseWrapper

	qb := s.statementBuilder.
		Select(sqlReleaseTableBodyColumn).
		From(sqlReleaseTableName).
		Where(sq.Eq{sqlReleaseTableKeyColumn: key}).
		Where(sq.Eq{sqlReleaseTableNamespaceColumn: s.namespace})

	query, args, err := qb.ToSql()
	if err != nil {
		s.Log("failed to build query: %v", err)
		return nil, err
	}

	// Get will return an error if the result is empty
	if err := s.db.Get(&record, query, args...); err != nil {
		s.Log("got SQL error when getting release %s: %v", key, err)
		return nil, ErrReleaseNotFound
	}

	release, err := decodeRelease(record.Body)
	if err != nil {
		s.Log("get: failed to decode data %q: %v", key, err)
		return nil, err
	}

	LabelsQuery, args, err := s.statementBuilder.
		Select(sqlLabelTableKeyColumn, sqlLabelTableValueColumn).
		From(sqlLabelTableName).
		Where(sq.Eq{sqlLabelTableReleaseKeyColumn: key,
			sqlLabelTableReleaseNamespaceColumn: s.namespace}).
		ToSql()
	if err != nil {
		s.Log("failed to build query: %v", err)
		return nil, err
	}

	var LabelsList = []SQLReleaseLabelWrapper{}
	if err := s.db.Select(&LabelsList, LabelsQuery, args...); err != nil {
		s.Log("get: failed to get release Labels: %v", err)
		return nil, err
	}

	LabelsMap := make(map[string]string)
	for _, i := range LabelsList {
		LabelsMap[i.Key] = i.Value
	}
	release.Labels = filterSystemLabels(LabelsMap)

	return release, nil
}

// List returns the list of all releases such that filter(release) == true
func (s *SQL) List(filter func(*rspb.Release) bool) ([]*rspb.Release, error) {
	sb := s.statementBuilder.
		Select(sqlReleaseTableBodyColumn).
		From(sqlReleaseTableName).
		Where(sq.Eq{sqlReleaseTableOwnerColumn: sqlReleaseDefaultOwner})

	// If a namespace was specified, we only list releases from that namespace
	if s.namespace != "" {
		sb = sb.Where(sq.Eq{sqlReleaseTableNamespaceColumn: s.namespace})
	}

	query, args, err := sb.ToSql()
	if err != nil {
		s.Log("failed to build query: %v", err)
		return nil, err
	}

	var records = []SQLReleaseWrapper{}
	if err := s.db.Select(&records, query, args...); err != nil {
		s.Log("list: failed to list: %v", err)
		return nil, err
	}

	var releases []*rspb.Release
	for _, record := range records {
		release, err := decodeRelease(record.Body)
		if err != nil {
			s.Log("list: failed to decode release: %v: %v", record, err)
			continue
		}
		if filter(release) {
			releases = append(releases, release)
		}
	}

	return releases, nil
}

// Query returns the set of releases that match the provided set of Labels.
func (s *SQL) Query(Labels map[string]string) ([]*rspb.Release, error) {
	sb := s.statementBuilder.
		Select(sqlReleaseTableBodyColumn).
		From(sqlReleaseTableName)

	keys := make([]string, 0, len(Labels))
	for key := range Labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, ok := LabelMap[key]; ok {
			sb = sb.Where(sq.Eq{key: Labels[key]})
		} else {
			s.Log("unknown Label %s", key)
			return nil, fmt.Errorf("unknow Label %s", key)
		}
	}

	// If a namespace was specified, we only list releases from that namespace
	if s.namespace != "" {
		sb = sb.Where(sq.Eq{sqlReleaseTableNamespaceColumn: s.namespace})
	}

	// Build our query
	query, args, err := sb.ToSql()
	if err != nil {
		s.Log("failed to build query: %v", err)
		return nil, err
	}

	var records = []SQLReleaseWrapper{}
	if err := s.db.Select(&records, query, args...); err != nil {
		s.Log("list: failed to query with Labels: %v", err)
		return nil, err
	}

	var releases []*rspb.Release
	for _, record := range records {
		release, err := decodeRelease(record.Body)
		if err != nil {
			s.Log("list: failed to decode release: %v: %v", record, err)
			continue
		}
		releases = append(releases, release)
	}

	if len(releases) == 0 {
		return nil, ErrReleaseNotFound
	}

	return releases, nil
}

// Create creates a new release.
func (s *SQL) Create(key string, rls *rspb.Release) error {
	namespace := rls.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	s.namespace = namespace

	body, err := encodeRelease(rls)
	if err != nil {
		s.Log("failed to encode release: %v", err)
		return err
	}

	transaction, err := s.db.Beginx()
	if err != nil {
		s.Log("failed to start SQL transaction: %v", err)
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	insertQuery, args, err := s.statementBuilder.
		Insert(sqlReleaseTableName).
		Columns(
			sqlReleaseTableKeyColumn,
			sqlReleaseTableTypeColumn,
			sqlReleaseTableBodyColumn,
			sqlReleaseTableNameColumn,
			sqlReleaseTableNamespaceColumn,
			sqlReleaseTableVersionColumn,
			sqlReleaseTableStatusColumn,
			sqlReleaseTableOwnerColumn,
			sqlReleaseTableCreatedAtColumn,
		).
		Values(
			key,
			sqlReleaseDefaultType,
			body,
			rls.Name,
			namespace,
			int(rls.Version),
			rls.Info.Status.String(),
			sqlReleaseDefaultOwner,
			int(time.Now().Unix()),
		).ToSql()
	if err != nil {
		s.Log("failed to build insert query: %v", err)
		return err
	}

	if _, err := transaction.Exec(insertQuery, args...); err != nil {
		defer transaction.Rollback()

		selectQuery, args, buildErr := s.statementBuilder.
			Select(sqlReleaseTableKeyColumn).
			From(sqlReleaseTableName).
			Where(sq.Eq{sqlReleaseTableKeyColumn: key}).
			Where(sq.Eq{sqlReleaseTableNamespaceColumn: s.namespace}).
			ToSql()
		if buildErr != nil {
			s.Log("failed to build select query: %v", buildErr)
			return err
		}

		var record SQLReleaseWrapper
		if err := transaction.Get(&record, selectQuery, args...); err == nil {
			s.Log("release %s already exists", key)
			return ErrReleaseExists
		}

		s.Log("failed to store release %s in SQL database: %v", key, err)
		return err
	}

	for lk, lv := range rls.Labels {
		insertLabelsQuery, args, err := s.statementBuilder.
			Insert(sqlLabelTableName).
			Columns(
				sqlLabelTableReleaseKeyColumn,
				sqlLabelTableReleaseNamespaceColumn,
				sqlLabelTableKeyColumn,
				sqlLabelTableValueColumn,
			).
			Values(
				key,
				namespace,
				lk,
				lv,
			).ToSql()

		if err != nil {
			defer transaction.Rollback()
			s.Log("failed to build insert query: %v", err)
			return err
		}

		if _, err := transaction.Exec(insertLabelsQuery, args...); err != nil {
			defer transaction.Rollback()
			s.Log("failed to write Labels: %v", err)
			return err
		}
	}

	defer transaction.Commit()

	return nil
}

// Update updates a release.
func (s *SQL) Update(key string, rls *rspb.Release) error {
	namespace := rls.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	s.namespace = namespace

	body, err := encodeRelease(rls)
	if err != nil {
		s.Log("failed to encode release: %v", err)
		return err
	}

	query, args, err := s.statementBuilder.
		Update(sqlReleaseTableName).
		Set(sqlReleaseTableBodyColumn, body).
		Set(sqlReleaseTableNameColumn, rls.Name).
		Set(sqlReleaseTableVersionColumn, int(rls.Version)).
		Set(sqlReleaseTableStatusColumn, rls.Info.Status.String()).
		Set(sqlReleaseTableOwnerColumn, sqlReleaseDefaultOwner).
		Set(sqlReleaseTableModifiedAtColumn, int(time.Now().Unix())).
		Where(sq.Eq{sqlReleaseTableKeyColumn: key}).
		Where(sq.Eq{sqlReleaseTableNamespaceColumn: namespace}).
		ToSql()

	if err != nil {
		s.Log("failed to build update query: %v", err)
		return err
	}

	if _, err := s.db.Exec(query, args...); err != nil {
		s.Log("failed to update release %s in SQL database: %v", key, err)
		return err
	}

	return nil
}

// Delete deletes a release or returns ErrReleaseNotFound.
func (s *SQL) Delete(key string) (*rspb.Release, error) {
	transaction, err := s.db.Beginx()
	if err != nil {
		s.Log("failed to start SQL transaction: %v", err)
		return nil, fmt.Errorf("error beginning transaction: %v", err)
	}

	selectQuery, args, err := s.statementBuilder.
		Select(sqlReleaseTableBodyColumn).
		From(sqlReleaseTableName).
		Where(sq.Eq{sqlReleaseTableKeyColumn: key}).
		Where(sq.Eq{sqlReleaseTableNamespaceColumn: s.namespace}).
		ToSql()
	if err != nil {
		s.Log("failed to build select query: %v", err)
		return nil, err
	}

	var record SQLReleaseWrapper
	err = transaction.Get(&record, selectQuery, args...)
	if err != nil {
		s.Log("release %s not found: %v", key, err)
		return nil, ErrReleaseNotFound
	}

	release, err := decodeRelease(record.Body)
	if err != nil {
		s.Log("failed to decode release %s: %v", key, err)
		transaction.Rollback()
		return nil, err
	}
	defer transaction.Commit()

	deleteQuery, args, err := s.statementBuilder.
		Delete(sqlReleaseTableName).
		Where(sq.Eq{sqlReleaseTableKeyColumn: key}).
		Where(sq.Eq{sqlReleaseTableNamespaceColumn: s.namespace}).
		ToSql()
	if err != nil {
		s.Log("failed to build delete query: %v", err)
		return nil, err
	}

	if _, err = transaction.Exec(deleteQuery, args...); err != nil {
		return release, err
	}

	deleteLabelsQuery, args, err := s.statementBuilder.
		Delete(sqlLabelTableName).
		Where(sq.Eq{sqlLabelTableReleaseKeyColumn: key}).
		Where(sq.Eq{sqlLabelTableReleaseNamespaceColumn: s.namespace}).
		ToSql()

	if err != nil {
		s.Log("failed to build delete Labels query: %v", err)
		return nil, err
	}

	_, err = transaction.Exec(deleteLabelsQuery, args...)
	return release, err
}
