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
	name     string
	password string
	host     string
)

const (
	databaseName            = "cart"
	customersCollectionName = "customers"
	cartsCollectionName     = "carts"
	itemsCollectionName     = "items"
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
	ID      primitive.ObjectID   `bson:"_id"`
	ItemIDs []primitive.ObjectID `bson:"items"`
}

type MongoItem struct {
	ID    primitive.ObjectID `bson:"_id"`
	Value Item               `bson:",inline"`
}

func isID(id primitive.ObjectID) bson.M {
	return bson.M{"_id": id}
}

func inID(ids []primitive.ObjectID) bson.M {
	return bson.M{"_id": bson.M{"$in": ids}}
}

func (m *Mongo) GetCart(ctx context.Context, customerID string) (*Cart, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	cart := new(Cart)
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customerObjectID, err := primitive.ObjectIDFromHex(customerID)
		if err != nil {
			return err
		}
		customersCol := s.Client().Database(databaseName).Collection(customersCollectionName)
		mongoCustomer := new(MongoCustomer)
		if err := customersCol.FindOne(s, isID(customerObjectID)).Decode(mongoCustomer); err != nil {
			return err
		}

		cartsCol := s.Client().Database(databaseName).Collection(cartsCollectionName)
		mongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, isID(mongoCustomer.CartID)).Decode(mongoCart); err != nil {
			return err
		}

		itemsCol := s.Client().Database(databaseName).Collection(itemsCollectionName)
		mongoItems := new([]MongoItem)
		cursor, err := itemsCol.Find(s, inID(mongoCart.ItemIDs))
		if err != nil {
			return err
		}
		if err := cursor.All(s, mongoItems); err != nil {
			return err
		}

		items := make([]Item, len(*mongoItems))
		for _, mongoItem := range *mongoItems {
			items = append(items, mongoItem.Value)
		}

		cart = &Cart{
			ID:         mongoCart.ID.String(),
			CustomerID: mongoCustomer.ID.String(),
			Items:      items,
		}
		return nil
	})
	return cart, err
}

func (m *Mongo) DeleteCart(ctx context.Context, customerID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customerObjectID, err := primitive.ObjectIDFromHex(customerID)
		if err != nil {
			return err
		}
		customersCol := s.Client().Database(databaseName).Collection(customersCollectionName)
		mongoCustomer := new(MongoCustomer)
		if err := customersCol.FindOne(s, isID(customerObjectID)).Decode(mongoCustomer); err != nil {
			return err
		}

		cartsCol := s.Client().Database(databaseName).Collection(cartsCollectionName)
		mongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, isID(mongoCustomer.CartID)).Decode(mongoCart); err != nil {
			return err
		}

		itemsCol := s.Client().Database(databaseName).Collection(itemsCollectionName)
		if _, err = itemsCol.DeleteMany(s, inID(mongoCart.ItemIDs)); err != nil {
			return err
		}

		_, err = cartsCol.DeleteOne(s, isID(mongoCart.ID))
		return err
	})
}

func (m *Mongo) MargeCart(ctx context.Context, customerID string, sessionID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customerObjectID, err := primitive.ObjectIDFromHex(customerID)
		if err != nil {
			return err
		}
		customersCol := s.Client().Database(databaseName).Collection(customersCollectionName)
		mongoCustomer := new(MongoCustomer)
		if err := customersCol.FindOne(s, isID(customerObjectID)).Decode(mongoCustomer); err != nil {
			return err
		}

		cartsCol := s.Client().Database(databaseName).Collection(cartsCollectionName)
		mongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, isID(mongoCustomer.CartID)).Decode(mongoCart); err != nil {
			return err
		}

		sessionCustomerObjectID, err := primitive.ObjectIDFromHex(sessionID)
		if err != nil {
			return err
		}

		sessionMongoCustomer := new(MongoCustomer)
		if err := customersCol.FindOne(s, isID(sessionCustomerObjectID)).Decode(sessionMongoCustomer); err != nil {
			return err
		}

		sessionMongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, isID(sessionMongoCustomer.CartID)).Decode(sessionMongoCart); err != nil {
			return err
		}

		mongoCart.ItemIDs = append(mongoCart.ItemIDs, sessionMongoCart.ItemIDs...)
		_, err = cartsCol.UpdateOne(s, isID(mongoCart.ID), bson.M{"items": mongoCart.ItemIDs})
		return err
	})
}

func (m *Mongo) GetItem(ctx context.Context, customerID string, itemID string) (*Item, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	item := new(Item)
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		itemObjectID, err := primitive.ObjectIDFromHex(itemID)
		if err != nil {
			return err
		}
		itemsCol := s.Client().Database(databaseName).Collection(itemsCollectionName)
		mongoItem := new(MongoItem)
		if err := itemsCol.FindOne(s, isID(itemObjectID)).Decode(mongoItem); err != nil {
			return err
		}

		item = &mongoItem.Value
		return nil
	})

	return item, err
}

func (m *Mongo) GetItems(ctx context.Context, customerID string) (*[]Item, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	items := new([]Item)
	err := m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customerObjectID, err := primitive.ObjectIDFromHex(customerID)
		if err != nil {
			return err
		}

		customerCol := s.Client().Database(databaseName).Collection(customersCollectionName)
		mongoCustomer := new(MongoCustomer)
		if err := customerCol.FindOne(s, isID(customerObjectID)).Decode(mongoCustomer); err != nil {
			return err
		}

		cartsCol := s.Client().Database(databaseName).Collection(cartsCollectionName)
		mongoCart := new(MongoCart)
		if err := cartsCol.FindOne(s, isID(mongoCustomer.CartID)).Decode(mongoCart); err != nil {
			return err
		}

		itemsCol := s.Client().Database(databaseName).Collection(itemsCollectionName)
		cursor, err := itemsCol.Find(s, inID(mongoCart.ItemIDs))
		if err != nil {
			return err
		}

		mongoItems := new([]MongoItem)
		if err := cursor.All(s, mongoItems); err != nil {
			return err
		}

		for _, mongoItem := range *mongoItems {
			*items = append(*items, mongoItem.Value)
		}
		return nil
	})

	return items, err
}

func (m *Mongo) CreateItem(ctx context.Context, customerID string, item *Item) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		customerObjectID, err := primitive.ObjectIDFromHex(customerID)
		if err != nil {
			return err
		}

		mongoCustomer := new(MongoCustomer)
		if err := s.Client().Database(databaseName).Collection(customersCollectionName).FindOne(s, isID(customerObjectID)).Decode(mongoCustomer); err != nil {
			return err
		}
		mongoCart := new(MongoCart)
		if err := s.Client().Database(databaseName).Collection(cartsCollectionName).FindOne(s, isID(mongoCustomer.CartID)).Decode(mongoCart); err != nil {
			return err
		}

		itemObjectID, err := primitive.ObjectIDFromHex(customerID + mongoCart.ID.String())
		mongoItem := MongoItem{
			ID:    itemObjectID,
			Value: *item,
		}
		itemsCol := s.Client().Database(databaseName).Collection(itemsCollectionName)
		_, err = itemsCol.InsertOne(s, mongoItem)
		return err
	})
}

func (m *Mongo) DeleteItem(ctx context.Context, customerID string, itemID string) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		itemObjectID, err := primitive.ObjectIDFromHex(itemID)
		if err != nil {
			return err
		}

		_, err = s.Client().Database(databaseName).Collection(itemsCollectionName).DeleteOne(s, isID(itemObjectID))
		return err
	})
}

func (m *Mongo) UpdateItem(ctx context.Context, customerID string, item *Item) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	return m.Client.UseSession(_ctx, func(s mongo.SessionContext) error {
		itemObjectID, err := primitive.ObjectIDFromHex(item.ID)
		if err != nil {
			return err
		}

		itemsCol := s.Client().Database(databaseName).Collection(itemsCollectionName)

		mongoItem := new(MongoItem)
		if err := itemsCol.FindOne(s, isID(itemObjectID)).Decode(mongoItem); err != nil {
			return nil
		}

		mongoItem.Value = *item
		_, err = itemsCol.UpdateOne(s, isID(itemObjectID), mongoItem)
		return err
	})
}

func (m *Mongo) Ping(ctx context.Context) error {
	_ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()
	return m.Client.Ping(_ctx, readpref.Primary())
}
