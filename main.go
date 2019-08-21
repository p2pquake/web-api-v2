package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
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

	// options
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	if limit <= 0 || limit > 100 {
		c.Status(400)
		return
	}
	options := options.FindOptions{Limit: &limit, Skip: &offset}

	// filters
	filters := bson.D{{"code", 551}}

	sinceDate := c.Query("since_date")
	untilDate := c.Query("until_date")
	dateRegexp := regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})$`)
	if matches := dateRegexp.FindStringSubmatch(sinceDate); matches != nil {
		filters = append(filters, bson.E{"earthquake.time", bson.D{{"$gte", matches[1] + "/" + matches[2] + "/" + matches[3] + " 00:00:00"}}})
	}
	if matches := dateRegexp.FindStringSubmatch(untilDate); matches != nil {
		filters = append(filters, bson.E{"earthquake.time", bson.D{{"$lte", matches[1] + "/" + matches[2] + "/" + matches[3] + " 23:59:59"}}})
	}

	if quakeType := c.Query("quake_type"); quakeType != "" {
		filters = append(filters, bson.E{"issue.type", quakeType})
	}
	if minMagnitude, err := strconv.ParseFloat(c.Query("min_magnitude"), 64); err == nil {
		filters = append(filters, bson.E{"earthquake.hypocenter.magnitude", bson.D{{"$gte", minMagnitude}}})
	}
	if maxMagnitude, err := strconv.ParseFloat(c.Query("max_magnitude"), 64); err == nil {
		filters = append(filters, bson.E{"earthquake.hypocenter.magnitude", bson.D{{"$lte", maxMagnitude}}})
		filters = append(filters, bson.E{"earthquake.hypocenter.magnitude", bson.D{{"$gte", 0.0}}})
	}
	if minScale, err := strconv.Atoi(c.Query("min_scale")); err == nil {
		filters = append(filters, bson.E{"earthquake.maxScale", bson.D{{"$gte", minScale}}})
	}
	if maxScale, err := strconv.Atoi(c.Query("max_scale")); err == nil {
		filters = append(filters, bson.E{"earthquake.maxScale", bson.D{{"$lte", maxScale}}})
		filters = append(filters, bson.E{"earthquake.maxScale", bson.D{{"$gte", 0}}})
	}

	prefectures, _ := c.GetQueryArray("prefectures[]")
	for _, prefecture := range prefectures {
		elements := strings.Split(prefecture, ",")
		prefectureName := elements[0]
		scale, _ := strconv.Atoi(elements[1])
		filters = append(filters, bson.E{"points", bson.D{{"$elemMatch", bson.D{{"pref", prefectureName}, {"scale", bson.D{{"$gte", scale}}}}}}})
	}

	cur, err := collection.Find(ctx, filters, &options)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	items := make([]bson.M, 0)
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
