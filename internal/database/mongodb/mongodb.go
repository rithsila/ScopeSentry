package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/Autumn-27/ScopeSentry/internal/logger"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/runtime/testinit"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database
var IsConnected bool

func isOfflineSubcommand() bool {
	if len(os.Args) < 2 {
		return false
	}
	switch os.Args[1] {
	case "version", "--version", "-v", "pluginkey", "help", "--help", "-h":
		return true
	}
	return false
}

func init() {
	if testinit.Enabled {
		return
	}
	username := url.QueryEscape(config.GlobalConfig.MongoDB.Username)
	password := url.QueryEscape(config.GlobalConfig.MongoDB.Password)

	ip := config.GlobalConfig.MongoDB.IP
	if ip == "" {
		ip = "127.0.0.1"
	}
	port := config.GlobalConfig.MongoDB.Port
	if port == 0 {
		port = 27017
	}
	uri := ip + ":" + strconv.Itoa(port)
	if username != "" && password != "" {
		uri = "mongodb://" + username + ":" + password + "@" + uri
	} else {
		uri = "mongodb://" + uri
	}

	// 设置客户端选项
	clientOptions := options.Client().ApplyURI(uri)

	// 连接到 MongoDB（纯本地客户端结构体初始化，不发网络包）
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err == nil && client != nil {
		Client = client
		dbName := config.GlobalConfig.MongoDB.Database
		if dbName == "" {
			dbName = "ScopeSentry"
		}
		DB = client.Database(dbName)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect %v to MongoDB: %v", uri, err))
		return
	}

	// 如果是离线子命令，跳过 ping 网络探测，不阻塞也不输出报警日志
	if isOfflineSubcommand() {
		return
	}

	// 设置连接超时并检查连接
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err = client.Ping(ctx, nil); err != nil {
		logger.Error(fmt.Sprintf("failed to ping MongoDB: %v", err))
		return
	}

	IsConnected = true
	logger.Info("MongoDB connected successfully")
	return
}

func EnsureConnected() bool {
	if IsConnected && Client != nil {
		return true
	}
	if Client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := Client.Ping(ctx, nil); err == nil {
		IsConnected = true
		return true
	}
	return false
}

func Close() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := Client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect MongoDB: %v", err)
		}
	}
	return nil
}

// ================= 工具方法 =================

func Collection(name string) *mongo.Collection {
	return DB.Collection(name)
}

func defaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// ================= CRUD 封装 =================

// 插入一条
func InsertOne(col string, document interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).InsertOne(ctx, document)
}

// 插入多条
func InsertMany(col string, documents []interface{}) (*mongo.InsertManyResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).InsertMany(ctx, documents)
}

// 查询一条并解码
func FindOne(col string, filter interface{}, result interface{}) error {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).FindOne(ctx, filter).Decode(result)
}

// 查询多条
func FindMany(col string, filter interface{}, result interface{}, opts ...*options.FindOptions) error {
	ctx, cancel := defaultContext()
	defer cancel()

	cursor, err := Collection(col).Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, result)
}

func FindAll(collectionName string, query, selector, result interface{}) error {
	collection := Collection(collectionName)
	cur, err := collection.Find(context.Background(), query, options.Find().SetProjection(selector))
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	return cur.All(context.Background(), result)
}

// 更新一条
func UpdateOne(col string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).UpdateOne(ctx, filter, update)
}

// 更新多条
func UpdateMany(col string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).UpdateMany(ctx, filter, update)
}

// 删除一条
func DeleteOne(col string, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).DeleteOne(ctx, filter)
}

// 删除多条
func DeleteMany(col string, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).DeleteMany(ctx, filter)
}

// 统计数量
func Count(col string, filter interface{}) (int64, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).CountDocuments(ctx, filter)
}

// 分页查询
func FindPage(
	col string,
	filter interface{},
	result interface{},
	page int64,
	pageSize int64,
	sort bson.D,
) error {

	ctx, cancel := defaultContext()
	defer cancel()

	if page < 1 {
		page = 1
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(sort)

	cursor, err := Collection(col).Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, result)
}
