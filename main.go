package main

import (
	"context"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/p2pquake/web-api-v2/userquake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoDBURL        string `envconfig:"mongodb_url"`
	Database          string `envconfig:"database"`
	JmaCollection     string `envconfig:"jma_collection"`
	HistoryCollection string `envconfig:"history_collection"`
}

type HumanReadableParam struct {
	Limit int64 `form:"limit" binding:"min=0,max=100"`
}

type QuakeParam struct {
	Offset       int64    `form:"offset" binding:"min=0"`
	Limit        int64    `form:"limit" binding:"min=0,max=100"`
	Order        int64    `form:"order" binding:"min=-1,max=1"`
	QuakeType    string   `form:"quake_type" binding:"omitempty,quaketype"`
	MinScale     int64    `form:"min_scale" binding:"omitempty,scale"`
	MaxScale     int64    `form:"max_scale" binding:"omitempty,scale"`
	MinMagnitude float64  `form:"min_magnitude" binding:"min=0.0"`
	MaxMagnitude float64  `form:"max_magnitude" binding:"min=0.0"`
	SinceDate    string   `form:"since_date" binding:"omitempty,numeric,len=8"`
	UntilDate    string   `form:"until_date" binding:"omitempty,numeric,len=8"`
	Prefectures  []string `form:"prefectures[]" binding:"omitempty,dive,contains=0x2C"`
}

type TsunamiParam struct {
	Offset    int64  `form:"offset" binding:"min=0"`
	Limit     int64  `form:"limit" binding:"min=0,max=100"`
	Order     int64  `form:"order" binding:"min=-1,max=1"`
	SinceDate string `form:"since_date" binding:"omitempty,numeric,len=8"`
	UntilDate string `form:"until_date" binding:"omitempty,numeric,len=8"`
}

type HistoryParam struct {
	Codes  []int64 `form:"codes" binding:"omitempty,dive,numeric"`
	Offset int64   `form:"offset" binding:"min=0"`
	Limit  int64   `form:"limit" binding:"min=0,max=100"`
}

var jmaCollection *mongo.Collection
var historyCollection *mongo.Collection

func validQuakeType(fl validator.FieldLevel) bool {
	if quakeType, ok := fl.Field().Interface().(string); ok {
		if quakeType == "ScalePrompt" || quakeType == "Destination" ||
			quakeType == "ScaleAndDestination" || quakeType == "DetailScale" ||
			quakeType == "Foreign" || quakeType == "Other" {
			return true
		}
	}
	return false
}

func validScale(fl validator.FieldLevel) bool {
	if scale, ok := fl.Field().Interface().(int64); ok {
		if scale == 10 || scale == 20 || scale == 30 || scale == 40 ||
			scale == 45 || scale == 50 || scale == 55 || scale == 60 || scale == 70 {
			return true
		}
	}
	return false
}

func main() {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("config parse error: %v", err)
	}

	clientOptions := options.Client().ApplyURI(config.MongoDBURL)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("mongo client create error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}
	defer client.Disconnect(ctx)

	jmaCollection = client.Database(config.Database).Collection(config.JmaCollection)
	historyCollection = client.Database(config.Database).Collection(config.HistoryCollection)

	r := gin.Default()
	r.Use(cors.Default())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("quaketype", validQuakeType)
		v.RegisterValidation("scale", validScale)
	}

	v1 := r.Group("/v1")
	{
		v1.GET("/human-readable", getHumanReadable)
	}

	v2 := r.Group("/v2")
	{
		jma := v2.Group("/jma")
		{
			jma.GET("/quake", searchQuake)
			jma.GET("/quake/:id", getQuake)
			jma.GET("/tsunami", searchTsunami)
			jma.GET("/tsunami/:id", getTsunami)
		}

		v2.GET("/history", getHistories)
	}

	r.Run()
}

func getHumanReadable(c *gin.Context) {
	var humanReadableParam HumanReadableParam
	if err := c.ShouldBindWith(&humanReadableParam, binding.Query); err != nil {
		c.Status(400)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	limit := humanReadableParam.Limit
	if limit == 0 {
		limit = 10
	}

	options := options.FindOptions{Limit: &limit, Sort: bson.D{{"$natural", -1}}}
	filters := bson.D{{"code", bson.M{"$in": bson.A{5510, 5520}}}}
	cur, err := historyCollection.Find(ctx, &filters, &options)
	if err != nil {
		c.Status(500)
		return
	}
	defer cur.Close(ctx)

	items := make([]bson.M, 0)
	cur.All(ctx, &items)

	for i, item := range items {
		code := item["code"].(int32)
		if code == 5510 {
			items[i]["code"] = 551
		}
		if code == 5520 {
			items[i]["code"] = 552
			items[i]["issue"].(primitive.M)["type"] = "Focus"
		}

		items[i]["_id"] = bson.M{"$oid": items[i]["_id"]}
		cleanToHumanReadable(item)
	}

	{
		// userquake と結合した後、 time で降順ソートして件数制限をかける.
		lastTime := items[len(items)-1]["time"]

		uqFilters := bson.D{
			{"code", 561},
			{"time", bson.M{"$gte": lastTime}},
		}

		uqCur, err := historyCollection.Find(ctx, &uqFilters)
		if err != nil {
			log.Printf("find error: %v\n", err)
			c.Status(500)
			return
		}
		defer uqCur.Close(ctx)

		var uqLastTime *time.Time = nil
		var uqLastRecords []primitive.M
		var uqAnalyzedData []primitive.M
		for uqCur.Next(ctx) {
			var result bson.M
			if err := uqCur.Decode(&result); err != nil {
				c.Status(500)
				return
			}

			t, err := time.Parse("2006/01/02 15:04:05", result["time"].(string)[:19])
			if err != nil {
				c.Status(500)
				return
			}

			if uqLastTime == nil || t.Sub(*uqLastTime) >= 30*time.Second {
				if len(uqLastRecords) >= 3 {
					uqAnalyzedData = append(uqAnalyzedData, analyzeCollection(uqLastRecords))
				}
				uqLastRecords = []primitive.M{}
			}
			uqLastRecords = append(uqLastRecords, result)
			uqLastTime = &t
		}

		if len(uqLastRecords) >= 3 {
			uqAnalyzedData = append(uqAnalyzedData, analyzeCollection(uqLastRecords))
		}

		items = append(items, uqAnalyzedData...)
	}

	sort.Slice(items, func(i, j int) bool { return items[i]["time"].(string) > items[j]["time"].(string) })
	items = items[0:limit]

	c.JSON(200, items)
}

func analyzeCollection(records []primitive.M) primitive.M {
	data := primitive.M{}
	data["time"] = records[0]["time"]
	data["code"] = 5610

	data["count"] = len(records)

	regions := map[string]int{}
	prefs := map[string]int{}
	areas := map[string]int{}

	for _, record := range records {
		area := int(record["area"].(int32))
		names, ok := userquake.GetAreaName(area)
		if !ok {
			continue
		}

		_, ok = regions[names[0]]
		if !ok {
			regions[names[0]] = 0
		}
		regions[names[0]] += 1

		_, ok = prefs[names[1]]
		if !ok {
			prefs[names[1]] = 0
		}
		prefs[names[1]] += 1

		_, ok = areas[names[2]]
		if !ok {
			areas[names[2]] = 0
		}
		areas[names[2]] += 1
	}

	data["regions"] = regions
	data["prefs"] = prefs
	data["areas"] = areas

	return data
}

func searchQuake(c *gin.Context) {
	var quakeParam QuakeParam
	if err := c.ShouldBindWith(&quakeParam, binding.Query); err != nil {
		c.Status(400)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// options
	offset := quakeParam.Offset
	limit := quakeParam.Limit
	if limit == 0 {
		limit = 10
	}
	order := quakeParam.Order
	if order == 0 {
		order = -1
	}
	options := options.FindOptions{Limit: &limit, Skip: &offset, Sort: bson.D{{"time", order}}}

	// filters
	filters := bson.D{{"code", 551}}

	dateRegexp := regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})$`)
	if matches := dateRegexp.FindStringSubmatch(quakeParam.SinceDate); matches != nil {
		filters = append(filters, bson.E{"earthquake.time", bson.D{{"$gte", matches[1] + "/" + matches[2] + "/" + matches[3] + " 00:00:00"}}})
	}
	if matches := dateRegexp.FindStringSubmatch(quakeParam.UntilDate); matches != nil {
		filters = append(filters, bson.E{"earthquake.time", bson.D{{"$lte", matches[1] + "/" + matches[2] + "/" + matches[3] + " 23:59:59"}}})
	}

	if quakeParam.QuakeType != "" {
		filters = append(filters, bson.E{"issue.type", quakeParam.QuakeType})
	}
	if quakeParam.MinMagnitude != 0.0 {
		filters = append(filters, bson.E{"earthquake.hypocenter.magnitude", bson.D{{"$gte", quakeParam.MinMagnitude}}})
	}
	if quakeParam.MaxMagnitude != 0.0 {
		filters = append(filters, bson.E{"earthquake.hypocenter.magnitude", bson.D{{"$lte", quakeParam.MaxMagnitude}}})
		filters = append(filters, bson.E{"earthquake.hypocenter.magnitude", bson.D{{"$gte", 0.0}}})
	}
	if quakeParam.MinScale != 0 {
		filters = append(filters, bson.E{"earthquake.maxScale", bson.D{{"$gte", quakeParam.MinScale}}})
	}
	if quakeParam.MaxScale != 0 {
		filters = append(filters, bson.E{"earthquake.maxScale", bson.D{{"$lte", quakeParam.MaxScale}}})
		filters = append(filters, bson.E{"earthquake.maxScale", bson.D{{"$gte", 0}}})
	}

	for _, prefecture := range quakeParam.Prefectures {
		elements := strings.Split(prefecture, ",")
		prefectureName := elements[0]
		scale, _ := strconv.Atoi(elements[1])
		filters = append(filters, bson.E{"points", bson.D{{"$elemMatch", bson.D{{"pref", prefectureName}, {"scale", bson.D{{"$gte", scale}}}}}}})
	}

	cur, err := jmaCollection.Find(ctx, filters, &options)
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

func searchTsunami(c *gin.Context) {
	var tsunamiParam TsunamiParam
	if err := c.ShouldBindWith(&tsunamiParam, binding.Query); err != nil {
		c.Status(400)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// options
	offset := tsunamiParam.Offset
	limit := tsunamiParam.Limit
	if limit == 0 {
		limit = 10
	}
	order := tsunamiParam.Order
	if order == 0 {
		order = -1
	}
	options := options.FindOptions{Limit: &limit, Skip: &offset, Sort: bson.D{{"time", order}}}

	// filters
	filters := bson.D{{"code", 552}}

	dateRegexp := regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})$`)
	if matches := dateRegexp.FindStringSubmatch(tsunamiParam.SinceDate); matches != nil {
		filters = append(filters, bson.E{"issue.time", bson.D{{"$gte", matches[1] + "/" + matches[2] + "/" + matches[3] + " 00:00:00"}}})
	}
	if matches := dateRegexp.FindStringSubmatch(tsunamiParam.UntilDate); matches != nil {
		filters = append(filters, bson.E{"issue.time", bson.D{{"$lte", matches[1] + "/" + matches[2] + "/" + matches[3] + " 23:59:59"}}})
	}

	cur, err := jmaCollection.Find(ctx, filters, &options)
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

func getQuake(c *gin.Context) {
	getItem(c, 551)
}

func getTsunami(c *gin.Context) {
	getItem(c, 552)
}

func getItem(c *gin.Context, code int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.Status(400)
		return
	}
	filters := bson.D{{"code", code}, {"_id", id}}

	var result bson.M
	err = jmaCollection.FindOne(ctx, filters).Decode(&result)
	if err != nil {
		c.Status(404)
		return
	}

	cleanJmaRecord(result)
	c.JSON(200, result)
}

func cleanJmaRecord(m bson.M) {
	m["id"] = m["_id"]
	delete(m, "_id")
	delete(m, "expire")
}

func cleanToHumanReadable(m bson.M) {
	delete(m, "expire")
	delete(m, "ver")
	delete(m, "hop")
	delete(m, "uid")
	delete(m, "user-agent")
}

func getHistories(c *gin.Context) {
	var historyParam HistoryParam
	if err := c.ShouldBindWith(&historyParam, binding.Query); err != nil {
		c.Status(400)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// options
	offset := historyParam.Offset
	limit := historyParam.Limit
	if limit == 0 {
		limit = 10
	}
	options := options.FindOptions{Limit: &limit, Skip: &offset, Sort: bson.D{{"$natural", -1}}}

	// filters
	filters := bson.D{{"code", bson.D{{"$nin", bson.A{5510, 5511}}}}}
	if len(historyParam.Codes) > 0 {
		filters = bson.D{{"$and", []bson.D{filters, bson.D{{"code", bson.D{{"$in", historyParam.Codes}}}}}}}
	}

	cur, err := historyCollection.Find(ctx, &filters, &options)
	if err != nil {
		c.Status(500)
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
