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
	ErrInvalidHexID = errors.New("Invalid Id Hex")
)

const (
	databaseName            = "users"
	customersCollectionName = "customers"
	cardsCollectionName     = "cards"
	addressesCollectionName = "addresses"
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

func isID(id primitive.ObjectID) bson.M {
	return bson.M{"_id": id}
}

func inIDs(ids *[]primitive.ObjectID) bson.M {
	return bson.M{"_id": bson.M{"$in": *ids}}
}

// MongoUser is a wrapper for the users
type MongoUser struct {
	ID         primitive.ObjectID   `bson:"_id"`
	User       users.User           `bson:",inline"`
	AddressIDs []primitive.ObjectID `bson:"addresses"`
	CardIDs    []primitive.ObjectID `bson:"cards"`
}

// New Returns a new MongoUser
func NewMongoUser(user users.User) *MongoUser {
	return &MongoUser{
		ID:         primitive.NewObjectID(),
		User:       user,
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
		mu.User.Addresses = append(mu.User.Addresses, users.Address{ID: id.Hex()})
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

// MongoCard is a wrapper for Card
type MongoCard struct {
	users.Card `bson:",inline"`
	ID         primitive.ObjectID `bson:"_id"`
}

// CreateUser Insert user to MongoDB, including connected addresses and cards, update passed in user with Ids
func (m *Mongo) CreateUser(ctx context.Context, user *users.User) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		mongoUser := NewMongoUser(*user)
		mongoUser.User.UserID = mongoUser.ID.Hex()

		var carderr error
		mongoUser.CardIDs, carderr = m.createCards(s, &user.Cards)
		var addrerr error
		mongoUser.AddressIDs, addrerr = m.createAddresses(s, &user.Addresses)

		customers := s.Client().Database(databaseName).Collection(customersCollectionName)
		if _, err := customers.InsertOne(s, mongoUser); err != nil {
			addressesCol := s.Client().Database(databaseName).Collection(addressesCollectionName)
			if _, err := addressesCol.DeleteMany(s, inIDs(&mongoUser.AddressIDs)); err != nil {
				return err
			}
			cardsCol := s.Client().Database(databaseName).Collection(cardsCollectionName)
			if _, err := cardsCol.DeleteMany(s, inIDs(&mongoUser.CardIDs)); err != nil {
				return err
			}
			return err
		}

		// Cheap err for attributess
		if carderr != nil || addrerr != nil {
			return fmt.Errorf("%v %v", carderr, addrerr)
		}

		return nil
	})
}

func (m *Mongo) createCards(ctx context.Context, userCards *[]users.Card) ([]primitive.ObjectID, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	ids := make([]primitive.ObjectID, len(*userCards))
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		for i, card := range *userCards {
			cardObjectID := primitive.NewObjectID()
			card.ID = cardObjectID.Hex()
			mongoCard := MongoCard{Card: card, ID: cardObjectID}

			if _, err := s.Client().Database(databaseName).Collection(cardsCollectionName).InsertOne(s, mongoCard); err != nil {
				return err
			}

			ids[i] = cardObjectID
		}

		return nil
	})

	return ids, err
}

func (m *Mongo) createAddresses(ctx context.Context, userAddresses *[]users.Address) ([]primitive.ObjectID, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	ids := make([]primitive.ObjectID, len(*userAddresses))
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		for i, address := range *userAddresses {
			addressObjectID := primitive.NewObjectID()
			address.ID = addressObjectID.Hex()
			mongoAddress := MongoAddress{Address: address, ID: addressObjectID}

			addresses := s.Client().Database(databaseName).Collection(addressesCollectionName)
			if _, err := addresses.InsertOne(s, mongoAddress); err != nil {
				return err
			}

			ids[i] = addressObjectID
		}
		return nil
	})
	return ids, err
}

// GetUserByName Get user by their name
func (m *Mongo) GetUserByName(ctx context.Context, name string) (*users.User, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mongoUser := new(MongoUser)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customers := s.Client().Database(databaseName).Collection(customersCollectionName)
		return customers.FindOne(s, bson.M{"username": name}).Decode(mongoUser)
	}); err != nil {
		return nil, err
	}

	mongoUser.AddUserIDs()
	return &mongoUser.User, nil
}

// GetUser Get user by their object id
func (m *Mongo) GetUser(ctx context.Context, id string) (*users.User, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mongoUser := new(MongoUser)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		customers := s.Client().Database(databaseName).Collection(customersCollectionName)
		return customers.FindOne(s, _id).Decode(&mongoUser)
	}); err != nil {
		return nil, err
	}

	mongoUser.AddUserIDs()
	return &mongoUser.User, nil
}

// GetUsers Get all users
func (m *Mongo) GetUsers(ctx context.Context) (*[]*users.User, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mongoUsers := new([]MongoUser)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		// TODO: add paginations
		customers := s.Client().Database(databaseName).Collection(customersCollectionName)
		cursor, err := customers.Find(s, nil)
		if err != nil {
			return err
		}

		return cursor.All(s, &mongoUsers)
	}); err != nil {
		return nil, err
	}

	users := make([]*users.User, len(*mongoUsers))
	for _, mongoUser := range *mongoUsers {
		mongoUser.AddUserIDs()
		users = append(users, &mongoUser.User)
	}

	return &users, nil
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

		addresses := s.Client().Database(databaseName).Collection(addressesCollectionName)
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

		cards := s.Client().Database(databaseName).Collection(cardsCollectionName)
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
func (m *Mongo) GetCard(ctx context.Context, id string) (*users.Card, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mongoCard := new(MongoCard)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		cardObjectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		return s.Client().Database(databaseName).Collection(cardsCollectionName).
			FindOne(s, isID(cardObjectID)).Decode(mongoCard)
	}); err != nil {
		return nil, err
	}

	mongoCard.Card.ID = mongoCard.ID.Hex()
	return &mongoCard.Card, nil
}

// GetCards Gets all cards
func (m *Mongo) GetCards(ctx context.Context) (*[]*users.Card, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mongoCards := new([]MongoCard)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		// TODO: add pagination
		cursor, err := s.Client().Database(databaseName).Collection(cardsCollectionName).Find(s, nil)
		if err != nil {
			return err
		}
		return cursor.All(s, &mongoCards)
	}); err != nil {
		return nil, err
	}

	cards := make([]*users.Card, len(*mongoCards))
	for _, mc := range *mongoCards {
		cards = append(cards, &mc.Card)
	}

	return &cards, nil
}

// CreateCard adds card to MongoDB
func (m *Mongo) CreateCard(ctx context.Context, userCard *users.Card, userID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}

		cardObjectID := primitive.NewObjectID()
		userCard.ID = cardObjectID.Hex()
		mongoCard := MongoCard{Card: *userCard, ID: cardObjectID}
		if _, err := s.Client().Database(databaseName).Collection(cardsCollectionName).InsertOne(s, mongoCard); err != nil {
			return err
		}

		_, err = s.Client().Database(databaseName).Collection(customersCollectionName).
			UpdateOne(s, isID(userObjectID), bson.M{"cards": mongoCard.ID}, options.Update().SetUpsert(true))

		return err
	})
}

// GetAddress Gets an address by object Id
func (m *Mongo) GetAddress(ctx context.Context, id string) (*users.Address, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	userAddress := new(users.Address)
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		addressObjectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}

		mongoAddress := new(MongoAddress)
		if err := s.Client().Database(databaseName).Collection(addressesCollectionName).
			FindOne(s, isID(addressObjectID)).Decode(mongoAddress); err != nil {
			return err
		}

		userAddress = &mongoAddress.Address
		return nil
	})

	return userAddress, err
}

// GetAddresses gets all addresses
func (m *Mongo) GetAddresses(ctx context.Context) (*[]*users.Address, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	mongoAddresses := new([]MongoAddress)
	if err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		// TODO: add pagination
		addresses := s.Client().Database(databaseName).Collection(addressesCollectionName)
		cursor, err := addresses.Find(s, nil)
		if err != nil {
			return err
		}

		return cursor.All(s, &mongoAddresses)
	}); err != nil {
		return nil, err
	}

	userAddresses := make([]*users.Address, len(*mongoAddresses))
	for _, mongoAddress := range *mongoAddresses {
		userAddresses = append(userAddresses, &mongoAddress.Address)
	}

	return &userAddresses, nil
}

// CreateAddress Inserts Address into MongoDB
func (m *Mongo) CreateAddress(ctx context.Context, userAddress *users.Address, userID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}

		addressObjectID := primitive.NewObjectID()
		userAddress.ID = addressObjectID.Hex()
		mongoAddress := MongoAddress{Address: *userAddress, ID: addressObjectID}
		if _, err = s.Client().Database(databaseName).Collection(addressesCollectionName).InsertOne(s, mongoAddress); err != nil {
			return err
		}

		_, err = s.Client().Database(databaseName).Collection(customersCollectionName).
			UpdateOne(s, isID(userObjectID), bson.M{"cards": mongoAddress.ID}, options.Update().SetUpsert(true))
		return err
	})
}

// CreateAddress Inserts Address into MongoDB
func (m *Mongo) Delete(ctx context.Context, entity string, id string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}

		if entity == "customers" {
			user, err := m.GetUser(s, id)
			if err != nil {
				return err
			}

			addressObjectIDs := make([]primitive.ObjectID, len(user.Addresses))
			for _, address := range user.Addresses {
				addressObjectID, err := primitive.ObjectIDFromHex(address.ID)
				if err != nil {
					return err
				}
				addressObjectIDs = append(addressObjectIDs, addressObjectID)
			}

			cardObjectIDs := make([]primitive.ObjectID, len(user.Cards))
			for _, card := range user.Cards {
				cardID, err := primitive.ObjectIDFromHex(card.ID)
				if err != nil {
					return err
				}
				cardObjectIDs = append(cardObjectIDs, cardID)
			}

			if _, err := s.Client().Database(databaseName).Collection(addressesCollectionName).
				DeleteMany(s, inIDs(&addressObjectIDs)); err != nil {
				return err
			}

			if _, err := s.Client().Database(databaseName).Collection(cardsCollectionName).
				DeleteMany(s, inIDs(&cardObjectIDs)); err != nil {
				return err
			}

		} else {
			if _, err := s.Client().Database(databaseName).Collection(customersCollectionName).
				UpdateMany(s, nil, bson.M{"$pull": bson.M{entity: objectID}}); err != nil {
				return err
			}
		}

		collection := s.Client().Database(databaseName).Collection(entity)
		_, err = collection.DeleteMany(s, isID(objectID))
		return err
	})
}

func getURL() url.URL {
	ur := url.URL{
		Scheme: "mongodb",
		Host:   host,
		Path:   databaseName,
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
		customersCol := s.Client().Database(databaseName).Collection(customersCollectionName)
		index := mongo.IndexModel{Keys: []string{"username"}}
		index.Options = options.Index()
		index.Options.SetUnique(true).SetBackground(true).SetSparse(false)
		_, err := customersCol.Indexes().CreateOne(s, index)
		return err
	})
}

func (m *Mongo) Ping(ctx context.Context) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()
	return m.Client.Ping(_ctx, readpref.Primary())
}
