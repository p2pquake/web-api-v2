package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoDBURL string `json:"mongodb_url"`
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

var collection *mongo.Collection

func main() {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return
	}

	var config Config
	json.Unmarshal(file, &config)

	clientOptions := options.Client().ApplyURI(config.MongoDBURL)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return
	}
	defer client.Disconnect(ctx)

	collection = client.Database(config.Database).Collection(config.Collection)

	r := gin.Default()

	v2 := r.Group("/v2")
	{
		jma := v2.Group("/jma")
		{
			jma.GET("/quake", quakeEndpoint)
			jma.GET("/tsunami", tsunamiEndpoint)
		}
	}

	r.Run()
}

func quakeEndpoint(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	filter := bson.D{{"code", 551}}
	limit := int64(10)
	options := options.FindOptions{Limit: &limit}

	cur, err := collection.Find(ctx, filter, &options)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	var items []bson.M
	cur.All(ctx, &items)

	for _, item := range items {
		cleanJmaRecord(item)
	}

	c.JSON(200, items)
}

func tsunamiEndpoint(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	filter := bson.D{{"code", 552}}
	limit := int64(10)
	options := options.FindOptions{Limit: &limit}

	cur, err := collection.Find(ctx, filter, &options)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	var items []bson.M
	cur.All(ctx, &items)

	for _, item := range items {
		cleanJmaRecord(item)
	}

	c.JSON(200, items)
}

func cleanJmaRecord(m bson.M) {
	m["id"] = m["_id"]
	delete(m, "_id")
	delete(m, "expire")
}
