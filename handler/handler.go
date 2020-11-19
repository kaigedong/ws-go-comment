package handler

import (
	comment "comment/comment_package"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type CommentHandler struct{}

// return all comments to explorer
func (h *CommentHandler) AllComments(c echo.Context) error {
	projectUniqID := c.Param("project_id")

	out, err := comment.AllComments(projectUniqID)
	if err != nil {
		c.Logger().Error(err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, out)
}

func (h *CommentHandler) ProjectComments(c echo.Context) error {
	projectUniqID := c.Param("project_id")

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	ch, id := comment.SubscribeNewComment()
	defer comment.UnsubscribeNewComment(id)

	for {
		select {
		case resp := <-ch:
			if resp.ProjectUniqID != projectUniqID {
				continue
			}

			bs, _ := json.Marshal(resp)
			if err := ws.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				c.Logger().Info("[websocket deadline]", err)
				return nil
			}
			if err := ws.WriteMessage(websocket.TextMessage, bs); err != nil {
				c.Logger().Info("[websocket write]", err)
				return nil
			}
		// Check ws connection every 10 seconds
		case <-ticker.C:
			// ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Logger().Error(err)
				return nil
			}
		}
	}
}

func (h *CommentHandler) AddComment(c echo.Context) error {

	userComment := new(comment.UserComment)
	if err := c.Bind(userComment); err != nil {
		return echo.ErrBadGateway
	}
	comment.UserComments <- userComment

	return c.JSON(http.StatusOK, nil)
}
