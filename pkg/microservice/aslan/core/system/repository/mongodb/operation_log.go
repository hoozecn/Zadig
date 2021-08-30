package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	models2 "github.com/koderover/zadig/pkg/microservice/aslan/core/system/repository/models"
	mongotool "github.com/koderover/zadig/pkg/tool/mongo"
)

type OperationLogArgs struct {
	Username    string `json:"username"`
	ProductName string `json:"product_name"`
	Function    string `json:"function"`
	Status      int    `json:"status"`
	PerPage     int    `json:"per_page"`
	Page        int    `json:"page"`
}

type OperationLogColl struct {
	*mongo.Collection

	coll string
}

func NewOperationLogColl() *OperationLogColl {
	name := models2.OperationLog{}.TableName()
	return &OperationLogColl{
		Collection: mongotool.Database(config.MongoDatabase()).Collection(name),
		coll:       name,
	}
}

func (c *OperationLogColl) GetCollectionName() string {
	return c.coll
}

func (c *OperationLogColl) EnsureIndex(_ context.Context) error {
	return nil
}

func (c *OperationLogColl) Insert(args *models2.OperationLog) error {
	if args == nil {
		return errors.New("nil operation_log args")
	}

	res, err := c.InsertOne(context.TODO(), args)
	if err != nil || res == nil {
		return err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		args.ID = oid
	}

	return nil
}

func (c *OperationLogColl) Update(id string, status int) error {
	if id == "" {
		return errors.New("nil operation_log args")
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	change := bson.M{"$set": bson.M{
		"status": status,
	}}
	_, err = c.UpdateByID(context.TODO(), oid, change, options.Update().SetUpsert(true))

	return err
}

func (c *OperationLogColl) Find(args *OperationLogArgs) ([]*models2.OperationLog, int, error) {
	var res []*models2.OperationLog
	query := bson.M{}
	if args.ProductName != "" {
		query["product_name"] = bson.M{"$regex": args.ProductName}
	}
	if args.Username != "" {
		query["username"] = bson.M{"$regex": args.Username}
	}
	if args.Function != "" {
		query["function"] = bson.M{"$regex": args.Function}
	}
	if args.Status != 0 {
		query["status"] = args.Status
	}

	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})
	if args.Page > 0 && args.PerPage > 0 {
		opts.SetSkip(int64(args.PerPage * (args.Page - 1))).SetLimit(int64(args.PerPage))
	}
	cursor, err := c.Collection.Find(context.TODO(), query, opts)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(context.TODO(), &res)
	if err != nil {
		return nil, 0, err
	}

	count, err := c.CountDocuments(context.TODO(), query)
	if err != nil {
		return nil, 0, err
	}

	return res, int(count), err
}