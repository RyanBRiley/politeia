// Copyright (c) 2017-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package cockroachdb

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/url"
	"sync"

	"github.com/decred/politeia/politeiawww/user"
	"github.com/decred/politeia/util"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/marcopeereboom/sbox"
)

const (
	databaseID             = "users"
	databaseVersion uint32 = 1

	// Database table names
	tableKeyValue   = "key_value"
	tableUsers      = "users"
	tableIdentities = "identities"

	// Database user (read/write access)
	userPoliteiawww = "politeiawww"

	// Key-value store keys
	keyVersion             = "version"
	keyPaywallAddressIndex = "paywalladdressindex"
)

// cockroachdb implements the user database interface.
type cockroachdb struct {
	sync.RWMutex

	shutdown       bool                            // Backend is shutdown
	encryptionKey  *[32]byte                       // Data at rest encryption key
	userDB         *gorm.DB                        // Database context
	pluginSettings map[string][]user.PluginSetting // [pluginID][]PluginSettings
}

// isShutdown returns whether the backend has been shutdown.
func (c *cockroachdb) isShutdown() bool {
	c.RLock()
	defer c.RUnlock()

	return c.shutdown
}

// encrypt encrypts the provided data with the cockroachdb encryption key. The
// encrypted blob is prefixed with an sbox header which encodes the provided
// version. The read lock is taken despite the encryption key being a static
// value because the encryption key is zeroed out on shutdown, which causes
// race conditions to be reported when the golang race detector is used.
//
// This function must be called without the lock held.
func (c *cockroachdb) encrypt(version uint32, b []byte) ([]byte, error) {
	c.RLock()
	defer c.RUnlock()

	return sbox.Encrypt(version, c.encryptionKey, b)
}

// decrypt decrypts the provided packed blob using the cockroachdb encryption
// key. The read lock is taken despite the encryption key being a static value
// because the encryption key is zeroed out on shutdown, which causes race
// conditions to be reported when the golang race detector is used.
//
// This function must be called without the lock held.
func (c *cockroachdb) decrypt(b []byte) ([]byte, uint32, error) {
	c.RLock()
	defer c.RUnlock()

	return sbox.Decrypt(c.encryptionKey, b)
}

// setPaywallAddressIndex updates the paywall address index record in the
// key-value store.
//
// This function can be called using a transaction when necessary.
func setPaywallAddressIndex(db *gorm.DB, index uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, index)
	kv := KeyValue{
		Key:   keyPaywallAddressIndex,
		Value: b,
	}
	return db.Save(&kv).Error
}

// SetPaywallAddressIndex updates the paywall address index record in the
// key-value database table.
func (c *cockroachdb) SetPaywallAddressIndex(index uint64) error {
	log.Tracef("SetPaywallAddressIndex: %v", index)

	if c.isShutdown() {
		return user.ErrShutdown
	}

	return setPaywallAddressIndex(c.userDB, index)
}

// userNew creates a new user the database.  The userID and paywall address
// index are set before the user record is inserted into the database.
//
// This function must be called using a transaction.
func (c *cockroachdb) userNew(tx *gorm.DB, u user.User) (*uuid.UUID, error) {
	// Set user paywall address index
	var index uint64
	kv := KeyValue{
		Key: keyPaywallAddressIndex,
	}
	err := tx.Find(&kv).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("find paywall index: %v", err)
		}
	} else {
		index = binary.LittleEndian.Uint64(kv.Value) + 1
	}

	u.PaywallAddressIndex = index

	// Set user ID
	u.ID = uuid.New()

	// Create user record
	ub, err := user.EncodeUser(u)
	if err != nil {
		return nil, err
	}

	eb, err := c.encrypt(user.VersionUser, ub)
	if err != nil {
		return nil, err
	}

	ur := convertUserFromUser(u, eb)
	err = tx.Create(&ur).Error
	if err != nil {
		return nil, fmt.Errorf("create user: %v", err)
	}

	// Update paywall address index
	err = setPaywallAddressIndex(tx, index)
	if err != nil {
		return nil, fmt.Errorf("set paywall index: %v", err)
	}

	return &u.ID, nil
}

// UserNew creates a new user record in the database.
func (c *cockroachdb) UserNew(u user.User) error {
	log.Tracef("UserNew: %v", u.Username)

	if c.isShutdown() {
		return user.ErrShutdown
	}

	// Create new user with a transaction
	tx := c.userDB.Begin()
	_, err := c.userNew(tx, u)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UserGetByUsername returns a user record given its username, if found in the
// database.
func (c *cockroachdb) UserGetByUsername(username string) (*user.User, error) {
	log.Tracef("UserGetByUsername: %v", username)

	if c.isShutdown() {
		return nil, user.ErrShutdown
	}

	var u User
	err := c.userDB.
		Where("username = ?", username).
		Find(&u).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = user.ErrUserNotFound
		}
		return nil, err
	}

	b, _, err := c.decrypt(u.Blob)
	if err != nil {
		return nil, err
	}

	usr, err := user.DecodeUser(b)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// UserGetByUsername returns a user record given its UUID, if found in the
// database.
func (c *cockroachdb) UserGetById(id uuid.UUID) (*user.User, error) {
	log.Tracef("UserGetById: %v", id)

	if c.isShutdown() {
		return nil, user.ErrShutdown
	}

	var u User
	err := c.userDB.
		Where("id = ?", id).
		Find(&u).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = user.ErrUserNotFound
		}
		return nil, err
	}

	b, _, err := c.decrypt(u.Blob)
	if err != nil {
		return nil, err
	}

	usr, err := user.DecodeUser(b)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// UserGetByPubKey returns a user record given its public key. The public key
// can be any of the public keys in the user's identity history.
func (c *cockroachdb) UserGetByPubKey(pubKey string) (*user.User, error) {
	log.Tracef("UserGetByPubKey: %v", pubKey)

	if c.isShutdown() {
		return nil, user.ErrShutdown
	}

	var u User
	q := `SELECT *
        FROM users
        INNER JOIN identities
          ON users.id = identities.user_id
          WHERE identities.public_key = ?`
	err := c.userDB.Raw(q, pubKey).Scan(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = user.ErrUserNotFound
		}
		return nil, err
	}

	b, _, err := c.decrypt(u.Blob)
	if err != nil {
		return nil, err
	}
	usr, err := user.DecodeUser(b)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// UsersGetByPubKey, given a list of public keys, returns a map where the keys
// are a public key and the value is a user record. Public keys can be any of
// the public keys in the user's identity history.
//
// UsersGetByPubKey satisfies the Database interface.
func (c *cockroachdb) UsersGetByPubKey(pubKeys []string) (map[string]user.User, error) {

	log.Tracef("UserGetByPubKey: %v", pubKeys)

	if c.isShutdown() {
		return nil, user.ErrShutdown
	}

	query := `SELECT * FROM users INNER JOIN identities
                ON users.id = identities.user_id
                WHERE identities.public_key IN (?)`

	rows, err := c.userDB.Raw(query, pubKeys).Rows()
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	users := make(map[string]user.User)
	pubKeyLookup := make(map[string]bool)
	for _, pk := range pubKeys {
		pubKeyLookup[pk] = true
	}

	for rows.Next() {
		var u User
		err := c.userDB.ScanRows(rows, &u)
		if err != nil {
			return nil, err
		}

		b, _, err := c.decrypt(u.Blob)
		if err != nil {
			return nil, err
		}

		usr, err := user.DecodeUser(b)
		if err != nil {
			return nil, err
		}

		for _, id := range usr.Identities {
			pk := id.String()
			if _, ok := pubKeyLookup[pk]; ok {
				users[pk] = *usr
			}
		}

	}

	return users, nil
}

// UserUpdate updates an existing user record in the database.
func (c *cockroachdb) UserUpdate(u user.User) error {
	log.Tracef("UserUpdate: %v", u.Username)

	if c.isShutdown() {
		return user.ErrShutdown
	}

	b, err := user.EncodeUser(u)
	if err != nil {
		return err
	}

	eb, err := c.encrypt(user.VersionUser, b)
	if err != nil {
		return err
	}

	ur := convertUserFromUser(u, eb)
	return c.userDB.Save(ur).Error
}

// AllUsers iterates over every user in the database, invoking the given
// callback function on each user.
func (c *cockroachdb) AllUsers(callback func(u *user.User)) error {
	log.Tracef("AllUsers")

	if c.isShutdown() {
		return user.ErrShutdown
	}

	// Lookup all users
	var users []User
	err := c.userDB.Find(&users).Error
	if err != nil {
		return err
	}

	// Invoke callback on each user
	for _, v := range users {
		b, _, err := c.decrypt(v.Blob)
		if err != nil {
			return err
		}

		u, err := user.DecodeUser(b)
		if err != nil {
			return err
		}

		callback(u)
	}

	return nil
}

// rotateKeys rotates the existing database encryption key with the given new
// key.
//
// This function must be called using a transaction.
func rotateKeys(tx *gorm.DB, oldKey *[32]byte, newKey *[32]byte) error {
	// Lookup all users
	var users []User
	err := tx.Find(&users).Error
	if err != nil {
		return err
	}

	// Rotate keys
	for _, v := range users {
		b, _, err := sbox.Decrypt(oldKey, v.Blob)
		if err != nil {
			return fmt.Errorf("decrypt user '%v': %v",
				v.ID, err)
		}

		eb, err := sbox.Encrypt(user.VersionUser, newKey, b)
		if err != nil {
			return fmt.Errorf("encrypt user '%v': %v",
				v.ID, err)
		}

		v.Blob = eb
		err = tx.Save(&v).Error
		if err != nil {
			return fmt.Errorf("save user '%v': %v",
				v.ID, err)
		}
	}

	return nil
}

// RotateKeys rotates the existing database encryption key with the given new
// key.
func (c *cockroachdb) RotateKeys(newKeyPath string) error {
	log.Tracef("RotateKeys: %v", newKeyPath)

	if c.isShutdown() {
		return user.ErrShutdown
	}

	// Load and validate new encryption key
	newKey, err := loadEncryptionKey(newKeyPath)
	if err != nil {
		return fmt.Errorf("load encryption key '%v': %v",
			newKeyPath, err)
	}

	if bytes.Equal(newKey[:], c.encryptionKey[:]) {
		return fmt.Errorf("keys are the same")
	}

	log.Infof("Rotating encryption keys")

	// Rotate keys using a transaction
	tx := c.userDB.Begin()
	err = rotateKeys(tx, c.encryptionKey, newKey)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("commit tx: %v", err)
	}

	// Update context
	c.encryptionKey = newKey

	return nil
}

// InsertUser inserts a user record into the database. The record must be a
// complete user record and the user must not already exist. This function is
// intended to be used for migrations between databases.
func (c *cockroachdb) InsertUser(u user.User) error {
	log.Tracef("InsertUser: %v", u.ID)

	if c.isShutdown() {
		return user.ErrShutdown
	}

	ub, err := user.EncodeUser(u)
	if err != nil {
		return err
	}

	eb, err := c.encrypt(user.VersionUser, ub)
	if err != nil {
		return err
	}

	ur := convertUserFromUser(u, eb)
	return c.userDB.Create(&ur).Error
}

// PluginExec executes the provided plugin command.
func (c *cockroachdb) PluginExec(pc user.PluginCommand) (*user.PluginCommandReply, error) {
	log.Tracef("PluginExec: %v %v", pc.ID, pc.Command)

	if c.isShutdown() {
		return nil, user.ErrShutdown
	}

	var payload string
	var err error
	switch pc.ID {
	case user.CMSPluginID:
		payload, err = c.cmsPluginExec(pc.Command, pc.Payload)
	default:
		return nil, user.ErrInvalidPlugin
	}
	if err != nil {
		return nil, err
	}

	return &user.PluginCommandReply{
		ID:      pc.ID,
		Command: pc.Command,
		Payload: payload,
	}, nil
}

// RegisterPlugin registers a plugin with the user database.
func (c *cockroachdb) RegisterPlugin(p user.Plugin) error {
	log.Tracef("RegisterPlugin: %v %v", p.ID, p.Version)

	if c.isShutdown() {
		return user.ErrShutdown
	}

	// Setup plugin tables
	var err error
	switch p.ID {
	case user.CMSPluginID:
		err = c.cmsPluginSetup()
	default:
		return user.ErrInvalidPlugin
	}
	if err != nil {
		return err
	}

	// Save plugin settings
	c.Lock()
	defer c.Unlock()

	c.pluginSettings[p.ID] = p.Settings

	return nil
}

// Close shuts down the database.  All interface functions must return with
// errShutdown if the backend is shutting down.
func (c *cockroachdb) Close() error {
	log.Tracef("Close")

	c.Lock()
	defer c.Unlock()

	// Zero out encryption key
	util.Zero(c.encryptionKey[:])
	c.encryptionKey = nil

	c.shutdown = true
	return c.userDB.Close()
}

func (c *cockroachdb) createTables(tx *gorm.DB) error {
	if !tx.HasTable(tableKeyValue) {
		err := tx.CreateTable(&KeyValue{}).Error
		if err != nil {
			return err
		}
	}
	if !tx.HasTable(tableUsers) {
		err := tx.CreateTable(&User{}).Error
		if err != nil {
			return err
		}
	}
	if !tx.HasTable(tableIdentities) {
		err := tx.CreateTable(&Identity{}).Error
		if err != nil {
			return err
		}
	}

	// Insert version record
	kv := KeyValue{
		Key: keyVersion,
	}
	err := tx.Find(&kv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint32(b, databaseVersion)
			kv.Value = b
			err = tx.Save(&kv).Error
		}
	}

	return err
}

func loadEncryptionKey(filepath string) (*[32]byte, error) {
	log.Tracef("loadEncryptionKey: %v", filepath)

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("load encryption key %v: %v",
			filepath, err)
	}

	if hex.DecodedLen(len(b)) != 32 {
		return nil, fmt.Errorf("invalid key length %v",
			filepath)
	}

	k := make([]byte, 32)
	_, err = hex.Decode(k, b)
	if err != nil {
		return nil, fmt.Errorf("decode hex %v: %v",
			filepath, err)
	}

	var key [32]byte
	copy(key[:], k)
	util.Zero(k)

	return &key, nil
}

// New opens a connection to the CockroachDB user database and returns a new
// cockroachdb context. sslRootCert, sslCert, sslKey, and encryptionKey are
// file paths.
func New(host, network, sslRootCert, sslCert, sslKey, encryptionKey string) (*cockroachdb, error) {
	log.Tracef("New: %v %v %v %v %v %v", host, network, sslRootCert,
		sslCert, sslKey, encryptionKey)

	// Build url
	dbName := databaseID + "_" + network
	h := "postgresql://" + userPoliteiawww + "@" + host + "/" + dbName
	u, err := url.Parse(h)
	if err != nil {
		return nil, fmt.Errorf("parse url '%v': %v",
			h, err)
	}

	q := u.Query()
	q.Add("sslmode", "require")
	q.Add("sslrootcert", sslRootCert)
	q.Add("sslcert", sslCert)
	q.Add("sslkey", sslKey)
	u.RawQuery = q.Encode()

	// Connect to database
	db, err := gorm.Open("postgres", u.String())
	if err != nil {
		return nil, fmt.Errorf("connect to database '%v': %v",
			u.String(), err)
	}

	log.Infof("UserDB host: %v", h)

	// Load encryption key
	key, err := loadEncryptionKey(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create context
	c := &cockroachdb{
		encryptionKey:  key,
		userDB:         db,
		pluginSettings: make(map[string][]user.PluginSetting),
	}

	// Disable gorm logging. This prevents duplicate errors
	// from being printed since we handle errors manually.
	c.userDB.LogMode(false)

	// Disable automatic table name pluralization.
	// We set table names manually.
	c.userDB.SingularTable(true)

	// Setup database tables
	tx := c.userDB.Begin()
	err = c.createTables(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	// Check version record
	kv := KeyValue{
		Key: keyVersion,
	}
	err = c.userDB.Find(&kv).Error
	if err != nil {
		return nil, fmt.Errorf("find version: %v", err)
	}

	// XXX A version mismatch will need to trigger a db
	// migration, but just return an error for now.
	version := binary.LittleEndian.Uint32(kv.Value)
	if version != databaseVersion {
		return nil, fmt.Errorf("version mismatch: got %v, want %v",
			version, databaseVersion)
	}

	return c, err
}
