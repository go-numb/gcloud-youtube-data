package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/gocarina/gocsv"

	_ "github.com/joho/godotenv/autoload"
)

const (
	ISGCS      = true
	BUCKETNAME = "data-sdy"
	FILEPREFIX = "youtube-data"
	MAXRESULTS = 2
)

var (
	PORT       string
	APIKEY     = os.Getenv("APIKEY")
	PROJECTID  = os.Getenv("PROJECTID")
	FILEEXPIRE = time.Now().Add(24 * time.Hour)

	APICOUNT = 0
)

func init() {
	// OSによって環境変数を設定
	// 環境別の処理
	if runtime.GOOS == "linux" {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		PORT = fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))
		log.Debug().Msgf("Linuxでの処理, PORT: %s, API access: %d", PORT, APICOUNT)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		PORT = fmt.Sprintf("localhost:%s", os.Getenv("PORT"))
		log.Debug().Msgf("その他のOSでの処理, PORT: %s, API access: %d", PORT, APICOUNT)
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.CORS())

	api := e.Group("/api")

	api.GET("/youtube/channels", GetChannelsFromQuery)
	api.GET("/youtube/comments", GetCommentsFromQuery)

	log.Fatal().Err(e.Start(PORT)).Msgf("Server Stopped, API access: %d", APICOUNT)

}

// GetChannelsFromQuery は、クエリから動画のチャンネルを取得する
// 登録者数n以上、開設日m日以上のチャンネルを取得する
// 動画投稿日 動画のURL サムネイル画像 タイトル 高評価率 再生回数 チャンネル名 チャンネル登録者数 チャンネル開設日 何日前 一日当たり平均再生回数 動画の長さ
func GetChannelsFromQuery(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "query is required"})
	}

	// 登録者数の取得
	temp := c.QueryParam("subscribers_n")
	regiN, err := strconv.Atoi(temp)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "number as registors is required"})
	}

	// 開設日n日以上のチャンネルを取得
	temp = c.QueryParam("days")
	days, err := strconv.Atoi(temp)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "days as registors is required"})
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(APIKEY))
	if err != nil {
		log.Fatal().Msgf("Error creating YouTube client: %v, API access: %d", err, APICOUNT)
	}

	var (
		rows          []RowForChannel
		nextPageToken string
	)

	for { // チャンネル検索
		// 閲覧数が多い順にチャンネルを取得
		APICOUNT++
		searchResponse, err := service.Search.
			List([]string{"id", "snippet"}).
			Q(q).
			Type("channel").
			MaxResults(MAXRESULTS).
			Order("viewCount").
			PageToken(nextPageToken).Do()
		if err != nil {
			log.Fatal().Msgf("Error searching videos: %v, API access: %d", err, APICOUNT)
		}

		for _, item := range searchResponse.Items { // チャンネルの詳細情報取得
			channelID := item.Id.ChannelId
			log.Debug().Msgf("channelID: %s, API access: %d", channelID, APICOUNT)

			// チャンネル情報を取得
			APICOUNT++
			channelResponse, err := service.Channels.
				List([]string{"snippet", "statistics"}).
				Id(channelID).Do()
			if err != nil {
				log.Printf("error retrieving channel details for channel %s: %v, API access:%d", channelID, err, APICOUNT)
				continue
			}

			if len(channelResponse.Items) == 0 {
				continue
			}

			channel := channelResponse.Items[0]
			if channel.Statistics.SubscriberCount < uint64(regiN) {
				continue
			}
			// チャンネルの登録者数
			subscriberCount := channel.Statistics.SubscriberCount
			// チャンネルの動画数
			videoContents := channel.Statistics.VideoCount
			// チャンネル開設日の比較
			publishedAt, err := time.Parse(time.RFC3339, channel.Snippet.PublishedAt)
			if err != nil {
				log.Printf("error parsing publishedAt for channel id %s: %v, API access: %d", channelID, err, APICOUNT)
				continue
			}
			// チャンネル開設日から何日経過しているか
			daysAgo := uint64(math.Floor(time.Since(publishedAt).Hours() / 24))
			// 一日当たりの平均再生回数
			avgDailyContents := float64(videoContents) / float64(daysAgo)
			avgDailyViews := float64(subscriberCount) / float64(daysAgo)
			if uint64(days) > daysAgo {
				continue
			}

			// チャンネルの動画リストを取得
			APICOUNT++
			videosResponse, err := service.Search.
				List([]string{"id", "snippet", "statistics"}).
				ChannelId(channelID).
				Type("video").
				Order("videoCount").
				MaxResults(MAXRESULTS).Do()
			if err != nil {
				log.Printf("error retrieving videos for channel %s: %v, API access: %d", channelID, err, APICOUNT)
				continue
			}

			for _, videoItem := range videosResponse.Items {
				videoID := videoItem.Id.VideoId
				videoSnippet := videoItem.Snippet

				log.Debug().Msgf("--	videoID: %s, API access: %d", videoID, APICOUNT)

				// 動画の詳細情報を取得
				APICOUNT++
				videoResponse, err := service.Videos.
					List([]string{"snippet", "statistics", "contentDetails"}).
					Id(videoID).Do()
				if err != nil {
					log.Printf("error retrieving video details for video %s: %v, API access: %d", videoID, err, APICOUNT)
					continue
				}

				if len(videoResponse.Items) == 0 {
					continue
				}

				var (
					likeCount    uint64
					dislikeCount uint64
					viewCount    uint64
					likeRatio    float64
				)

				videoDetails := videoResponse.Items[0]
				if videoDetails.Statistics != nil {
					likeCount = videoDetails.Statistics.LikeCount
					dislikeCount = videoDetails.Statistics.DislikeCount
					viewCount = videoDetails.Statistics.ViewCount
					likeRatio = float64(likeCount) / float64(likeCount+dislikeCount)
				}

				row := RowForChannel{
					VideoId:      videoID,
					Url:          fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
					ThumbnailURL: videoSnippet.Thumbnails.Default.Url,
					LikeRate:     likeRatio,
					ViewCount:    viewCount,
					CommentCount: videoDetails.Statistics.CommentCount,

					ChannelId:        channelID,
					ChannelName:      channel.Snippet.Title,
					SubscriberCount:  subscriberCount,
					VideoContents:    videoContents,
					AvgDailyContents: avgDailyContents,
					AvgDailyViews:    avgDailyViews,
					DaysAgo:          daysAgo,
					ChannelCreatedAt: publishedAt,
					CreatedAt:        time.Now(),
				}

				rows = append(rows, row)

			}
		}

		nextPageToken = searchResponse.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	var m = echo.Map{
		"q": q,
	}

	if ISGCS {
		// GCSにファイルをアップロード
		// ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Fatal().Msgf("Error creating client: %v, API access: %d", err, APICOUNT)
		}
		defer client.Close()

		uniquename := uuid.New().String()
		objectname := fmt.Sprintf("%s-channel-%s-%s.csv", FILEPREFIX, q, uniquename)
		object := client.Bucket(BUCKETNAME).Object(objectname)

		w := object.NewWriter(ctx)
		defer w.Close()

		w.ContentType = "text/csv"

		// json to csv and write
		csvFile, err := gocsv.MarshalBytes(rows)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error unmarshaling bytes"})
		}

		// オブジェクトの有効期限を設定
		w.RetentionExpirationTime = FILEEXPIRE
		if _, err := w.Write(csvFile); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error writing file"})
		}

		for i := 0; i < 10; i++ {
			if err := w.Close(); err != nil {
				log.Error().Msgf("Error closing writer: %v, API access: %d", err, APICOUNT)
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}

		// create file link
		attrs, err := object.Attrs(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error getting object attrs"})
		}
		m["link"] = attrs.MediaLink
	}

	m["rows"] = rows

	return c.JSON(http.StatusOK, m)
}

// GetCommentsFromQuery は、クエリから動画のコメントを取得する
func GetCommentsFromQuery(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "query is required"})
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(APIKEY))
	if err != nil {
		log.Fatal().Msgf("Error creating YouTube client: %v, API access: %d", err, APICOUNT)
	}

	searchCall := service.Search.List([]string{"id"}).Q(q).MaxResults(MAXRESULTS)

	APICOUNT++
	searchResponse, err := searchCall.Do()
	if err != nil {
		log.Fatal().Msgf("Error searching videos: %v, API access: %d", err, APICOUNT)
	}

	var videoIds []string
	for _, item := range searchResponse.Items {
		if item.Id.Kind == "youtube#video" {
			videoIds = append(videoIds, item.Id.VideoId)
		}
	}

	uniquename := uuid.New().String()
	objectname := fmt.Sprintf("%s-comment-%s-%s.csv", FILEPREFIX, q, uniquename)

	// to Json
	var rows []RowForComment

	// 動画の情報とコメントの取得
	for _, videoId := range videoIds {
		// 動画情報の取得
		APICOUNT++
		videoResponse, err := service.Videos.
			List([]string{"snippet", "contentDetails", "statistics"}).
			Id(videoId).Do()
		if err != nil {
			log.Printf("Error retrieving video information: %v, API access: %d", err, APICOUNT)
			continue
		}

		// 動画情報の表示
		video := videoResponse.Items[0]

		// コメントの取得
		APICOUNT++
		commentsResponse, err := service.CommentThreads.
			List([]string{"snippet"}).
			VideoId(videoId).Do()
		if err != nil {
			log.Printf("Error retrieving comments: %v, API access: %d", err, APICOUNT)
			continue
		}

		id := video.Id
		var (
			like     uint64
			favo     uint64
			comments uint64
			replay   uint64
		)

		if video.Statistics != nil {
			like = video.Statistics.LikeCount
			comments = video.Statistics.CommentCount
		}

		// コメントの取得
		for _, commentThread := range commentsResponse.Items {
			comment := commentThread.Snippet.TopLevelComment.Snippet
			favo = uint64(comment.LikeCount)
			replay = uint64(commentThread.Snippet.TotalReplyCount)

			rows = append(rows, RowForComment{
				VideoId:  id,
				Like:     like,
				Comments: comments,

				Comment:   comment.TextDisplay,
				Name:      comment.AuthorDisplayName,
				NameId:    comment.AuthorChannelId.Value,
				Favo:      favo,
				Replay:    replay,
				CreatedAt: comment.PublishedAt,
			})
		}
	}

	var m = echo.Map{
		"q": q,
	}

	if ISGCS {
		// GCSにファイルをアップロード
		// ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Fatal().Msgf("Error creating client: %v, API access: %d", err, APICOUNT)
		}
		defer client.Close()

		object := client.Bucket(BUCKETNAME).Object(objectname)

		w := object.NewWriter(ctx)
		defer w.Close()

		w.ContentType = "text/csv"

		// json to csv and write
		csvFile, err := gocsv.MarshalBytes(rows)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error unmarshaling bytes"})
		}

		// オブジェクトの有効期限を設定
		w.RetentionExpirationTime = FILEEXPIRE
		if _, err := w.Write(csvFile); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error writing file"})
		}

		for i := 0; i < 10; i++ {
			if err := w.Close(); err != nil {
				log.Error().Msgf("Error closing writer: %v, API access: %d", err, APICOUNT)
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}

		// create file link
		attrs, err := object.Attrs(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error getting object attrs"})
		}
		m["link"] = attrs.MediaLink
	}

	m["rows"] = rows

	return c.JSON(http.StatusOK, m)
}

// 文字数を制限する
func Cut(s string, n int) string {
	if len([]rune(s)) > n {
		return string([]rune(s)[:n])
	}
	return s
}

// create struct
// comment,	display name,	name,	name id,	favo,	video id,	like,	replay,	comments,	created at
type RowForComment struct {
	VideoId  string `json:"video_id"`
	Like     uint64 `json:"like"`
	Comments uint64 `json:"comments"`

	Name      string `json:"name"`
	NameId    string `json:"name_id"`
	Comment   string `json:"comment"`
	Favo      uint64 `json:"favo"`
	Replay    uint64 `json:"replay"`
	CreatedAt string `json:"created_at"`
}

// 動画投稿日 動画のURL サムネイル画像 タイトル 高評価率 再生回数 チャンネル名 チャンネル登録者数 チャンネル開設日 何日前 一日当たり平均再生回数 動画の長さを取得する型を定義
type RowForChannel struct {
	VideoId      string  `json:"video_id"` // 動画ID
	Url          string  `json:"url"`
	ThumbnailURL string  `json:"thumbnail_url"`
	LikeRate     float64 `json:"like_rate"`
	ViewCount    uint64  `json:"view_count"`
	CommentCount uint64  `json:"comment_count"`

	ChannelId        string    `json:"channel_id"`
	ChannelName      string    `json:"channel_name"`
	SubscriberCount  uint64    `json:"subscriber_count "`
	VideoContents    uint64    `json:"video_contents"`
	AvgDailyContents float64   `json:"avg_daily_contents"`
	AvgDailyViews    float64   `json:"avg_daily_views "`
	DaysAgo          uint64    `json:"days_ago"`
	ChannelCreatedAt time.Time `json:"channel_created_at "`
	CreatedAt        time.Time `json:"created_at"`
}
