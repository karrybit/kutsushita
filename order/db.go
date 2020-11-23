package order

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
)

var (
	name           string
	password       string
	host           string
	databaseName   = "order"
	collectionName = "orders"
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

func (m *Mongo) CreateOrder(ctx context.Context, customerOrder CustomerOrder) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()
	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		order := struct {
			ID            primitive.ObjectID `bson:"_id"`
			CustomerOrder `bson:",inline"`
		}{
			ID:            primitive.NewObjectID(),
			CustomerOrder: customerOrder,
		}
		col := s.Client().Database(databaseName).Collection(collectionName)
		_, err := col.InsertOne(s, order)
		return err
	})
}

func (m *Mongo) Ping(ctx context.Context) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()
	return m.Client.Ping(_ctx, readpref.Primary())
}
