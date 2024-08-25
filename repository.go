package gohorse

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataList[T interface{}] struct {
	Data  []T
	Total int64
}

type MongoDbRepository[T interface{}] struct {
	db         *mongo.Database
	collection *mongo.Collection
	dataList   *DataList[T]
}

func NewMongoDbRepository[T interface{}](
	db *mongo.Database,
	v *viper.Viper,
) IRepository[T] {
	var r T
	reg := regexp.MustCompile(`\[.*`)
	coll := db.Collection(reg.ReplaceAllString(strings.ToLower(reflect.TypeOf(r).Name()), ""))

	return &MongoDbRepository[T]{
		db:         db,
		collection: coll,
		dataList:   &DataList[T]{},
	}
}

func (r *MongoDbRepository[T]) ChangeCollection(collectionName string) {
	r.collection = r.db.Collection(collectionName)
}

func (r *MongoDbRepository[T]) GetAll(
	ctx context.Context,
	filter map[string]interface{},
	optsFind ...*options.FindOptions) *[]T {

	cur, err := r.collection.Find(ctx, filter, optsFind...)
	if err != nil {
		panic(err)
	}
	result := []T{}
	for cur.Next(ctx) {
		var el T
		err = cur.Decode(&el)
		if err != nil {
			panic(err)
		}
		result = append(result, el)
	}

	return &result
}

func (r *MongoDbRepository[T]) GetAllSkipTake(
	ctx context.Context,
	filter map[string]interface{},
	skip int64,
	take int64,
	optsFind ...*options.FindOptions) *DataList[T] {

	result := &DataList[T]{}

	opts := make([]*options.FindOptions, 0)

	op := options.Find()
	op.SetSkip(skip)
	op.SetLimit(take)

	opts = append(opts, op)
	opts = append(opts, optsFind...)

	if os.Getenv("env") == "local" {
		_, obj, err := bson.MarshalValue(filter)
		fmt.Print(bson.Raw(obj), err)
	}

	result.Total, _ = r.collection.CountDocuments(ctx, filter)
	if result.Total > 0 {

		cur, err := r.collection.Find(ctx, filter, opts...)

		if err != nil {
			panic(err)
		}
		for cur.Next(ctx) {
			var el T
			err = cur.Decode(&el)
			if err != nil {
				panic(err)
			}
			result.Data = append(result.Data, el)
		}
	}

	return result
}

func (r *MongoDbRepository[T]) GetFirst(
	ctx context.Context,
	filter map[string]interface{}) *T {
	var el T

	if os.Getenv("env") == "local" {
		_, obj, err := bson.MarshalValue(filter)
		fmt.Print(bson.Raw(obj), err)
	}

	err := r.collection.FindOne(ctx, filter).Decode(&el)

	if err == mongo.ErrNoDocuments {
		return nil
	}

	if err != nil {
		panic(err)
	}

	return &el
}

func (r *MongoDbRepository[T]) Insert(
	ctx context.Context,
	entity *T) error {

	opt := options.InsertOne()
	opt.SetBypassDocumentValidation(true)

	_, err := r.collection.InsertOne(ctx, entity, opt)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoDbRepository[T]) InsertAll(
	ctx context.Context,
	entities *[]T) error {

	uis := []interface{}{entities}

	_, err := r.collection.InsertMany(ctx, uis)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoDbRepository[T]) Replace(
	ctx context.Context,
	filter map[string]interface{},
	entity *T) error {

	if os.Getenv("env") == "local" {
		_, obj, err := bson.MarshalValue(filter)
		fmt.Print(bson.Raw(obj), err)
	}

	var el bson.M
	err := r.collection.FindOne(ctx, filter).Decode(&el)

	if err == mongo.ErrNoDocuments {
		return r.Insert(ctx, entity)
	}

	_, err = r.collection.ReplaceOne(ctx, filter, entity, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoDbRepository[T]) Update(
	ctx context.Context,
	filter map[string]interface{},
	fields interface{}) error {

	re, err := r.collection.UpdateOne(ctx, filter, map[string]interface{}{"$set": fields})

	if err != nil {
		return err
	}

	if re.MatchedCount == 0 {
		return fmt.Errorf("MatchedCountZero")
	}

	return nil
}

func (r *MongoDbRepository[T]) Delete(
	ctx context.Context,
	filter map[string]interface{}) error {

	re, err := r.collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if re.DeletedCount == 0 {
		return fmt.Errorf("MatchedCountZero")
	}

	return nil
}

func (r *MongoDbRepository[T]) DeleteMany(
	ctx context.Context,
	filter map[string]interface{}) error {

	re, err := r.collection.DeleteMany(ctx, filter)

	if err != nil {
		return err
	}

	if re.DeletedCount == 0 {
		return fmt.Errorf("MatchedCountZero")
	}

	return nil
}

func (r *MongoDbRepository[T]) Aggregate(ctx context.Context, pipeline []interface{}) (*mongo.Cursor, error) {
	return r.collection.Aggregate(ctx, pipeline)
}

func (r *MongoDbRepository[T]) Count(ctx context.Context,
	filter map[string]interface{}, optsFind ...*options.CountOptions) int64 {

	count, err := r.collection.CountDocuments(ctx, filter, optsFind...)
	if err != nil {
		panic(err)
	}

	return count
}
