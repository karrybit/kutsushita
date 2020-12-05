package cart

import (
	"context"
	"flag"
	"net/url"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

var (
	name         string
	password     string
	host         string
	databaseName = "cart"
)

func init() {
	flag.StringVar(&name, "mongo-user", os.Getenv("MONGO_USER"), "Mongo user")
	flag.StringVar(&password, "mongo-password", os.Getenv("MONGO_PASS"), "Mongo password")
	flag.StringVar(&host, "mongo-host", os.Getenv("MONGO_HOST"), "Mongo host")
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

var db *Mongo

type Mongo struct {
	Client *mongo.Client
}

func InitDB() error {
	u := getURL()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	db = new(Mongo)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(u.Path))
	if err != nil {
		return err
	}
	db.Client = client
	return nil
}

type MongoCustomer struct {
	ID     primitive.ObjectID `bson:"_id"`
	CartID primitive.ObjectID `bson:"cart"`
}

type MongoCart struct {
	ID      primitive.ObjectID `bson:"_id"`
	Cart    `bson:",inline"`
	ItemIDs []primitive.ObjectID `bson:"items"`
}

type MongoItem struct {
	ID   primitive.ObjectID `bson:"_id"`
	Item Item               `bson:",inline"`
}

func (m *Mongo) GetCart(ctx context.Context, customerID string) (*Cart, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	cart := new(Cart)
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		id, err := primitive.ObjectIDFromHex(customerID)
		if err != nil {
			return err
		}
		customersCol := s.Client().Database(databaseName).Collection("customers")
		mongoCustomer := new(MongoCustomer)
		if err := customersCol.FindOne(s, bson.M{"_id": id}).Decode(mongoCustomer); err != nil {
			return err
		}

		cartsCol := s.Client().Database(databaseName).Collection("carts")
		mongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, bson.M{"_id": mongoCustomer.CartID}).Decode(mongoCart); err != nil {
			return err
		}

		return nil
	})
	return cart, err
}

func (m *Mongo) DeleteCart(ctx context.Context, customerID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		id, err := primitive.ObjectIDFromHex(customerID)
		if err != nil {
			return err
		}
		customersCol := s.Client().Database(databaseName).Collection("customers")
		mongoCustomer := new(MongoCustomer)
		if err := customersCol.FindOne(s, bson.M{"_id": id}).Decode(mongoCustomer); err != nil {
			return err
		}

		cartsCol := s.Client().Database(databaseName).Collection("carts")
		mongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, bson.M{"_id": id}).Decode(mongoCart); err != nil {
			return err
		}

		itemsCol := s.Client().Database(databaseName).Collection("items")
		if _, err = itemsCol.DeleteMany(s, bson.M{"_id": mongoCart.ItemIDs}); err != nil {
			return err
		}

		_, err = cartsCol.DeleteOne(s, bson.M{"_id": mongoCustomer.CartID})
		return err
	})
}

func (m *Mongo) MargeCart(ctx context.Context, customerID string, sessionID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		id, err := primitive.ObjectIDFromHex(customerID)
		if err != nil {
			return err
		}
		customersCol := s.Client().Database(databaseName).Collection("customers")
		mongoCustomer := new(MongoCustomer)
		if err := customersCol.FindOne(s, bson.M{"_id": id}).Decode(mongoCustomer); err != nil {
			return err
		}

		cartsCol := s.Client().Database(databaseName).Collection("carts")
		mongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, bson.M{"_id": mongoCustomer.CartID}).Decode(mongoCart); err != nil {
			return err
		}

		_sessionID, err := primitive.ObjectIDFromHex(sessionID)
		if err != nil {
			return err
		}

		sessionMongoCustomer := new(MongoCustomer)
		if err := customersCol.FindOne(s, bson.M{"_id": _sessionID}).Decode(sessionMongoCustomer); err != nil {
			return err
		}

		sessionMongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, bson.M{"_id": sessionMongoCustomer.CartID}).Decode(sessionMongoCart); err != nil {
			return err
		}

		mongoCart.ItemIDs = append(mongoCart.ItemIDs, sessionMongoCart.ItemIDs...)
		_, err = cartsCol.UpdateOne(s, bson.M{"_id": mongoCart.ID}, bson.M{"items": mongoCart.ItemIDs})
		return err
	})
}

func (m *Mongo) Ping(ctx context.Context) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()
	return m.Client.Ping(_ctx, readpref.Primary())
}
