package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	_ "github.com/joho/godotenv/autoload"
)

const (
	ISGCS      = true
	BUCKETNAME = "data-sdy"
	FILEPREFIX = "youtube-data"
	MAXRESULTS = 10
)

var (
	PORT       string
	APIKEY     = os.Getenv("APIKEY")
	PROJECTID  = os.Getenv("PROJECTID")
	FILEEXPIRE = time.Now().Add(24 * time.Hour)
)

func init() {
	// OSによって環境変数を設定
	// 環境別の処理
	if runtime.GOOS == "linux" {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		PORT = fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))
		log.Debug().Msgf("Linuxでの処理, PORT: %s", PORT)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		PORT = fmt.Sprintf("localhost:%s", os.Getenv("PORT"))
		log.Debug().Msgf("その他のOSでの処理, PORT: %s", PORT)
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.CORS())

	api := e.Group("/api")
	api.GET("/youtube", GetMovie)

	log.Fatal().Err(e.Start(PORT)).Msg("Server Stopped")
}

func GetMovie(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "query is required"})
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(APIKEY))
	if err != nil {
		log.Fatal().Msgf("Error creating YouTube client: %v", err)
	}

	searchCall := service.Search.List([]string{"id"}).Q(q).MaxResults(MAXRESULTS)

	searchResponse, err := searchCall.Do()
	if err != nil {
		log.Fatal().Msgf("Error searching videos: %v", err)
	}

	var videoIds []string
	for _, item := range searchResponse.Items {
		if item.Id.Kind == "youtube#video" {
			videoIds = append(videoIds, item.Id.VideoId)
		}
	}

	uniquename := uuid.New().String()
	objectname := fmt.Sprintf("%s-%s-%s.csv", FILEPREFIX, q, uniquename)

	header := "video_id,like,comments,comment,auth_channel_id,auth_display_name,favo,replay,created_at"
	bytedata := header + "\n"

	// to Json
	var rows []Row

	// 動画の情報とコメントの取得
	for _, videoId := range videoIds {
		// 動画情報の取得
		videoResponse, err := service.Videos.List([]string{"snippet", "contentDetails", "statistics"}).Id(videoId).Do()
		if err != nil {
			log.Printf("Error retrieving video information: %v", err)
			continue
		}

		// 動画情報の表示
		video := videoResponse.Items[0]

		// コメントの取得
		commentsCall := service.CommentThreads.List([]string{"snippet"}).VideoId(videoId)
		commentsResponse, err := commentsCall.Do()
		if err != nil {
			log.Printf("Error retrieving comments: %v", err)
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

		// コメントの表示
		// fmt.Println("video_id	like	comments	comment	auth_channel_id	auth_display_name	favo	replay	created_at")
		for _, commentThread := range commentsResponse.Items {
			comment := commentThread.Snippet.TopLevelComment.Snippet
			favo = uint64(comment.LikeCount)
			replay = uint64(commentThread.Snippet.TotalReplyCount)

			// "video_id,like,comments,comment,auth_channel_id,auth_display_name,favo,replay,created_at"
			row := fmt.Sprintf("%s,%d,%d,%s,%s,%s,%d,%d,%s",
				id, like, comments, comment.TextDisplay, comment.AuthorChannelId.Value, comment.AuthorDisplayName, favo, replay, comment.PublishedAt)

			bytedata += row + "\n"

			rows = append(rows, Row{
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
			log.Fatal().Msgf("Error creating client: %v", err)
		}
		defer client.Close()

		object := client.Bucket(BUCKETNAME).Object(objectname)

		w := object.NewWriter(ctx)
		defer w.Close()

		w.ContentType = "text/csv"

		// オブジェクトの有効期限を設定
		w.RetentionExpirationTime = FILEEXPIRE
		if _, err := w.Write([]byte(bytedata)); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error writing file"})
		}

		for {
			if err := w.Close(); err != nil {
				log.Error().Msgf("Error closing writer: %v", err)
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
type Row struct {
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
