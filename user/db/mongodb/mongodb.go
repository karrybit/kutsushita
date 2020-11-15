package mongodb

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"user/users"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	name            string
	password        string
	host            string
	db              = "users"
	ErrInvalidHexID = errors.New("Invalid Id Hex")
)

func init() {
	flag.StringVar(&name, "mongo-user", os.Getenv("MONGO_USER"), "Mongo user")
	flag.StringVar(&password, "mongo-password", os.Getenv("MONGO_PASS"), "Mongo password")
	flag.StringVar(&host, "mongo-host", os.Getenv("MONGO_HOST"), "Mongo host")
}

type Mongo struct {
	Client *mongo.Client
}

func (m *Mongo) Init() error {
	u := getURL()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	var err error
	if m.Client, err = mongo.Connect(ctx, options.Client().ApplyURI(u.Path)); err != nil {
		return err
	}

	return m.EnsureIndexes(ctx)
}

// MongoUser is a wrapper for the users
type MongoUser struct {
	// TODO: ,
	users.User `bson:",inline"`
	ID         primitive.ObjectID   `bson:"_id"`
	AddressIDs []primitive.ObjectID `bson:"addresses"`
	CardIDs    []primitive.ObjectID `bson:"cards"`
}

// New Returns a new MongoUser
func New() MongoUser {
	u := users.New()
	return MongoUser{
		User:       u,
		AddressIDs: make([]primitive.ObjectID, 0),
		CardIDs:    make([]primitive.ObjectID, 0),
	}
}

// AddUserIDs adds userID as string to user
func (mu *MongoUser) AddUserIDs() {
	if mu.User.Addresses == nil {
		mu.User.Addresses = make([]users.Address, 0)
	}
	for _, id := range mu.AddressIDs {
		mu.User.Addresses = append(mu.User.Addresses, users.Address{
			ID: id.Hex(),
		})
	}
	if mu.User.Cards == nil {
		mu.User.Cards = make([]users.Card, 0)
	}
	for _, id := range mu.CardIDs {
		mu.User.Cards = append(mu.User.Cards, users.Card{ID: id.Hex()})
	}
	mu.User.UserID = mu.ID.Hex()
}

// MongoAddress is a wrapper for Address
type MongoAddress struct {
	users.Address `bson:",inline"`
	ID            primitive.ObjectID `bson:"_id"`
}

// AddID ObjectID as string
func (m *MongoAddress) AddID() {
	m.Address.ID = m.ID.Hex()
}

// MongoCard is a wrapper for Card
type MongoCard struct {
	users.Card `bson:",inline"`
	ID         primitive.ObjectID `bson:"_id"`
}

// AddID ObjectID as string
func (m *MongoCard) AddID() {
	m.Card.ID = m.ID.Hex()
}

// CreateUser Insert user to MongoDB, including connected addresses and cards, update passed in user with Ids
func (m *Mongo) CreateUser(ctx context.Context, user *users.User) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		id := primitive.NewObjectID()
		mu := New()
		mu.User = *user
		mu.ID = id
		var carderr error
		var addrerr error
		mu.CardIDs, carderr = m.createCards(s, user.Cards)
		mu.AddressIDs, addrerr = m.createAddresses(s, user.Addresses)

		customers := s.Client().Database("").Collection("customers")
		if _, err := customers.InsertOne(s, mu); err != nil {
			// Gonna clean up if we can, ignore error
			// because the user save error takes precedence.
			m.cleanAttributes(s, mu)
			return err
		}

		mu.User.UserID = mu.ID.Hex()
		// Cheap err for attributess
		if carderr != nil || addrerr != nil {
			return fmt.Errorf("%v %v", carderr, addrerr)
		}
		*user = mu.User

		return nil
	})
}

func (m *Mongo) createCards(ctx context.Context, userCards []users.Card) ([]primitive.ObjectID, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	ids := make([]primitive.ObjectID, 0)
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		for k, ca := range userCards {
			id := primitive.NewObjectID()
			mc := MongoCard{Card: ca, ID: id}

			cards := s.Client().Database("").Collection("cards")
			if _, err := cards.InsertOne(s, mc); err != nil {
				return err
			}

			ids = append(ids, id)
			userCards[k].ID = id.Hex()
		}

		return nil
	})

	return ids, err
}

func (m *Mongo) createAddresses(ctx context.Context, userAddresses []users.Address) ([]primitive.ObjectID, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	ids := make([]primitive.ObjectID, 0)
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		for k, a := range userAddresses {
			id := primitive.NewObjectID()
			mongoAddress := MongoAddress{Address: a, ID: id}
			addresses := s.Client().Database("").Collection("addresses")
			if _, err := addresses.InsertOne(s, mongoAddress); err != nil {
				return err
			}
			ids = append(ids, id)
			userAddresses[k].ID = id.Hex()
		}
		return nil
	})
	return ids, err
}

func (m *Mongo) cleanAttributes(ctx context.Context, mu MongoUser) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		addresses := s.Client().Database("").Collection("addresses")
		if _, err := addresses.DeleteMany(s, bson.M{"_id": bson.M{"$in": mu.AddressIDs}}); err != nil {
			return err
		}
		cards := s.Client().Database("").Collection("cards")
		_, err := cards.DeleteMany(s, bson.M{"_id": bson.M{"$in": mu.CardIDs}})
		return err
	})
}

func (m *Mongo) appendAttributeId(ctx context.Context, attr string, id primitive.ObjectID, userID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customers := s.Client().Database("").Collection("customers")
		_id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}
		// TODO: addtoset option
		_, err = customers.UpdateOne(s, bson.M{"_id": _id}, bson.M{attr: id}, options.Update().SetUpsert(true))
		return err
	})
}

func (m *Mongo) removeAttributeId(ctx context.Context, attr string, id primitive.ObjectID, userID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customers := s.Client().Database("").Collection("customers")
		_id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}
		// TODO: pull option
		_, err = customers.UpdateOne(s, bson.M{"_id": _id}, bson.M{attr: id})
		return err
	})
}

// GetUserByName Get user by their name
func (m *Mongo) GetUserByName(ctx context.Context, name string) (users.User, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mu := New()
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customers := s.Client().Database("").Collection("customers")
		return customers.FindOne(s, bson.M{"username": name}).Decode(&mu)
	}); err != nil {
		return users.User{}, err
	}

	mu.AddUserIDs()
	return mu.User, nil
}

// GetUser Get user by their object id
func (m *Mongo) GetUser(ctx context.Context, id string) (users.User, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mu := New()
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		customers := s.Client().Database("").Collection("customers")
		return customers.FindOne(s, _id).Decode(&mu)
	}); err != nil {
		return users.User{}, err
	}

	mu.AddUserIDs()

	return mu.User, nil
}

// GetUsers Get all users
func (m *Mongo) GetUsers(ctx context.Context) ([]users.User, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	var mongoUsers []MongoUser
	users := make([]users.User, 0)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		// TODO: add paginations
		customers := s.Client().Database("").Collection("customers")
		cursor, err := customers.Find(s, nil)
		if err != nil {
			return err
		}

		return cursor.All(s, &mongoUsers)
	}); err != nil {
		return users, err
	}

	for _, mongoUser := range mongoUsers {
		mongoUser.AddUserIDs()
		users = append(users, mongoUser.User)
	}

	return users, nil
}

// GetUserAttributes given a user, load all cards and addresses connected to that user
func (m *Mongo) GetUserAttributes(ctx context.Context, user *users.User) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		ids := make([]primitive.ObjectID, 0)
		for _, address := range user.Addresses {
			_id, err := primitive.ObjectIDFromHex(address.ID)
			if err != nil {
				return err
			}
			ids = append(ids, _id)
		}

		addresses := s.Client().Database("").Collection("addresses")
		cursor, err := addresses.Find(s, bson.M{"_id": bson.M{"$in": ids}})
		if err != nil {
			return err
		}

		var mongoAddresses []MongoAddress
		if err := cursor.All(s, &mongoAddresses); err != nil {
			return err
		}

		userAddresses := make([]users.Address, 0)
		for _, address := range mongoAddresses {
			address.Address.ID = address.ID.Hex()
			userAddresses = append(userAddresses, address.Address)
		}
		user.Addresses = userAddresses

		ids = make([]primitive.ObjectID, 0)
		for _, card := range user.Cards {
			_id, err := primitive.ObjectIDFromHex(card.ID)
			if err != nil {
				return err
			}
			ids = append(ids, _id)
		}

		cards := s.Client().Database("").Collection("cards")
		cursor, err = cards.Find(s, bson.M{"_id": bson.M{"$in": ids}})
		if err != nil {
			return err
		}

		var mongoCards []MongoCard
		if err := cursor.All(s, &mongoCards); err != nil {
			return err
		}

		userCards := make([]users.Card, 0)
		for _, ca := range mongoCards {
			ca.Card.ID = ca.ID.Hex()
			userCards = append(userCards, ca.Card)
		}
		user.Cards = userCards

		return nil
	})
}

// GetCard Gets card by objects Id
func (m *Mongo) GetCard(ctx context.Context, id string) (users.Card, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mc := MongoCard{}
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		cards := s.Client().Database("").Collection("cards")
		return cards.FindOne(s, bson.M{"_id": _id}).Decode(&mc)
	}); err != nil {
		return mc.Card, err
	}

	mc.AddID()
	return mc.Card, nil
}

// GetCards Gets all cards
func (m *Mongo) GetCards(ctx context.Context) ([]users.Card, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	var mongoCards []MongoCard
	cards := make([]users.Card, 0)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		// TODO: add pagination
		cards := s.Client().Database("").Collection("cards")
		cursor, err := cards.Find(s, nil)
		if err != nil {
			return err
		}
		return cursor.All(s, &mongoCards)
	}); err != nil {
		return cards, err
	}

	for _, mongoCard := range mongoCards {
		mongoCard.AddID()
		cards = append(cards, mongoCard.Card)
	}

	return cards, nil
}

// CreateCard adds card to MongoDB
func (m *Mongo) CreateCard(ctx context.Context, userCard *users.Card, userID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		_, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}

		cards := s.Client().Database("").Collection("cards")

		id := primitive.NewObjectID()
		mongoCard := MongoCard{Card: *userCard, ID: id}
		if _, err := cards.InsertOne(s, mongoCard); err != nil {
			return err
		}

		if err := m.appendAttributeId(s, "cards", mongoCard.ID, userID); err != nil {
			return err
		}

		mongoCard.AddID()
		*userCard = mongoCard.Card

		return err
	})
}

// GetAddress Gets an address by object Id
func (m *Mongo) GetAddress(ctx context.Context, id string) (users.Address, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	var userAddress users.Address
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}

		addresses := s.Client().Database("").Collection("addresses")
		ma := MongoAddress{}
		if err := addresses.FindOne(s, _id).Decode(&ma); err != nil {
			return err
		}

		ma.AddID()
		userAddress = ma.Address
		return nil
	})

	return userAddress, err
}

// GetAddresses gets all addresses
func (m *Mongo) GetAddresses(ctx context.Context) ([]users.Address, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	var mongoAddresses []MongoAddress
	userAddresses := make([]users.Address, 0)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		// TODO: add pagination
		addresses := s.Client().Database("").Collection("addresses")
		cursor, err := addresses.Find(s, nil)
		if err != nil {
			return err
		}

		return cursor.All(s, &mongoAddresses)
	}); err != nil {
		return userAddresses, err
	}

	for _, mongoAddress := range mongoAddresses {
		mongoAddress.AddID()
		userAddresses = append(userAddresses, mongoAddress.Address)
	}

	return userAddresses, nil
}

// CreateAddress Inserts Address into MongoDB
func (m *Mongo) CreateAddress(ctx context.Context, userAddress *users.Address, userID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		_, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}

		addresses := s.Client().Database("").Collection("addresses")
		id := primitive.NewObjectID()
		mongoAddress := MongoAddress{Address: *userAddress, ID: id}
		_, err = addresses.InsertOne(s, mongoAddress)
		if err != nil {
			return err
		}

		// Address for anonymous user
		err = m.appendAttributeId(s, "addresses", mongoAddress.ID, userID)
		if err != nil {
			return err
		}

		mongoAddress.AddID()
		*userAddress = mongoAddress.Address

		return err
	})
}

// CreateAddress Inserts Address into MongoDB
func (m *Mongo) Delete(ctx context.Context, entity string, id string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}

		collection := s.Client().Database("").Collection(entity)

		if entity == "customers" {
			user, err := m.GetUser(s, id)
			if err != nil {
				return err
			}

			addressIDs := make([]primitive.ObjectID, 0)
			for _, address := range user.Addresses {
				addressID, err := primitive.ObjectIDFromHex(address.ID)
				if err != nil {
					return err
				}
				addressIDs = append(addressIDs, addressID)
			}

			cardIDs := make([]primitive.ObjectID, 0)
			for _, card := range user.Cards {
				cardID, err := primitive.ObjectIDFromHex(card.ID)
				if err != nil {
					return err
				}
				cardIDs = append(cardIDs, cardID)
			}

			address := s.Client().Database("").Collection("addresses")
			if _, err := address.DeleteMany(s, bson.M{"_id": bson.M{"$in": addressIDs}}); err != nil {
				return err
			}

			cards := s.Client().Database("").Collection("cards")
			if _, err := cards.DeleteMany(s, bson.M{"_id": bson.M{"$in": cardIDs}}); err != nil {
				return err
			}

		} else {
			customers := s.Client().Database("").Collection("customers")
			if _, err := customers.UpdateMany(s, nil, bson.M{"$pull": bson.M{entity: _id}}); err != nil {
				return err
			}
		}

		_, err = collection.DeleteMany(s, bson.M{"_id": _id})
		return err
	})
}

func getURL() url.URL {
	ur := url.URL{
		Scheme: "mongodb",
		Host:   host,
		Path:   db,
	}
	if name != "" {
		u := url.UserPassword(name, password)
		ur.User = u
	}
	return ur
}

// EnsureIndexes ensures username is unique
func (m *Mongo) EnsureIndexes(ctx context.Context) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customers := s.Client().Database("").Collection("customers")
		index := mongo.IndexModel{Keys: []string{"username"}}
		index.Options = options.Index()
		index.Options.SetUnique(true).SetBackground(true).SetSparse(false)
		_, err := customers.Indexes().CreateOne(s, index)
		return err
	})
}

func (m *Mongo) Ping(ctx context.Context) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()
	return m.Client.Ping(_ctx, readpref.Primary())
}
