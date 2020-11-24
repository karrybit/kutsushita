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
	ID      primitive.ObjectID   `bson:"_id"`
	CartIDs []primitive.ObjectID `bson:"carts"`
}

type MongoCart struct {
	ID   primitive.ObjectID `bson:"_id"`
	Cart `bson:",inline"`
}

func (m *Mongo) GetCart(ctx context.Context, customerID string) (*[]Cart, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()

	carts := new([]Cart)
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
		cursor, err := cartsCol.Find(s, bson.M{"_id": bson.M{"$in": mongoCustomer.CartIDs}})
		mongoCarts := new([]MongoCart)
		if err := cursor.All(s, mongoCarts); err != nil {
			return err
		}

		for _, mongoCart := range *mongoCarts {
			*carts = append(*carts, mongoCart.Cart)
		}
		return nil
	})
	return carts, err
}

func (m *Mongo) Ping(ctx context.Context) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()
	return m.Client.Ping(_ctx, readpref.Primary())
}
