// +build !appengine

package mongodb

import (
	"net/http"
	"time"

	"code.google.com/p/gorilla/securecookie"
	"code.google.com/p/gorilla/sessions"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
)

type MongoStore struct {
	Codecs       []securecookie.Codec
	Options      *sessions.Options // default configuration
	DBCollection *mgo.Collection
}

type MgSessionTbl struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	SessionID []byte
	Encoded   string
	Age       time.Time
}

func NewMongoStore(DBCollection *mgo.Collection, keyPairs ...[]byte) *MongoStore {
	index := mgo.Index{Unique: true, Key: []string{"sessionid"}}
	DBCollection.EnsureIndex(index)

	return &MongoStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		DBCollection: DBCollection,
	}
}

func (s *MongoStore) Get(r *http.Request, name string) (*sessions.Session,
	error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *MongoStore) New(r *http.Request, name string) (*sessions.Session,
	error) {
	session := sessions.NewSession(s, name)
	session.Options = &(*s.Options)
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID,
			s.Codecs...)
		if err == nil {
			err = s.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}

	// Remove older sessions
	s.DBCollection.RemoveAll(bson.M{
		"age": bson.M{
			"$lt": bson.Now().Add(time.Duration(-s.Options.MaxAge) * time.Second),
		},
	})
	return session, err
}

func (s *MongoStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	if session.ID == "" {
		session.ID = string(securecookie.GenerateRandomKey(32))
	}
	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w,
		sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (s *MongoStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(),
		session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}
	mg := &MgSessionTbl{
		Encoded:   encoded,
		SessionID: []byte(session.ID),
		Age:       bson.Now(),
	}
	_, err = s.DBCollection.Upsert(bson.M{"sessionid": session.ID}, mg)
	return err
}

func (s *MongoStore) load(session *sessions.Session) error {
	mg := &MgSessionTbl{SessionID: []byte(session.ID)}
	err := s.DBCollection.Find(bson.M{"sessionid": []byte(session.ID)}).One(mg)
	if err == nil {
		err = securecookie.DecodeMulti(session.Name(),
			string(mg.Encoded), &session.Values, s.Codecs...)
	}
	return err
}
