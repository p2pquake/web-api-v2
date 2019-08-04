package main

import (
	"encoding/json"
	"context"
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

func main() {
	file, err := ioutil.ReadFile("config.json")
	if err != nil { return }

	var config Config
	json.Unmarshal(file, &config)

	clientOptions := options.Client().ApplyURI(config.MongoDBURL)
	client, err := mongo.NewClient(clientOptions)
	if err != nil { return }

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil { return }
	defer client.Disconnect(ctx)

	collection := client.Database(config.Database).Collection(config.Collection)

	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "こんにちは世界",
		})
	})

	r.GET("/count", func(c *gin.Context) {
		count, err := collection.CountDocuments(ctx, bson.D{{}})
		if err != nil { count = -1 }
		c.JSON(200, gin.H{
			"count": count,
		})
	})

	r.Run()
}
