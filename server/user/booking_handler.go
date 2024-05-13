package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0xForked/goca/server/hof"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type bookingHandler struct {
	service IUserService
}

func (h bookingHandler) host(ctx *gin.Context) {
	uname := ctx.Param("username")
	user, err := h.service.Profile(ctx, uname, false)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	eventTypes, err := h.service.EventType(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	user.EventTypes = eventTypes
	ctx.JSON(http.StatusOK, user)
}

func (h bookingHandler) schedule(ctx *gin.Context) {}

func (h bookingHandler) add(ctx *gin.Context) {
	var body BookingForm
	if err := ctx.ShouldBind(&body); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	if err := body.Validate(); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err})
		return
	}
	// get user data
	user, err := h.service.Profile(ctx, body.Username, false)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	eventTypes, err := h.service.EventType(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	// booking
	cfg := hof.GetOAuthConfig()
	tok := &oauth2.Token{}
	if err = json.Unmarshal([]byte(user.Token.String), tok); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	calendarService := hof.GetCalendarService(ctx, tok, cfg)
	var timezone string
	var title string
	var duration int
	for _, et := range eventTypes {
		if et.ID == body.EventTypeID {
			title = et.Title
			timezone = et.Availability.Timezone
			duration = et.Duration
			break
		}
	}
	email, err := hof.GetUserEmail(ctx, tok, cfg)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	summary := fmt.Sprintf("%s between %s and %s", title, user.Username, body.Name)
	description := fmt.Sprintf("maybe notes? %s", body.Notes)
	event, err := hof.SetNewMeeting(calendarService, summary, description, timezone,
		email, body.Email, body.Date, body.Time, duration)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	}
	ctx.JSON(http.StatusOK, event)
	// TODO STORE database title, date with time (start & end), attendees, google meet link, notes
}

func newBookingHandler(
	service IUserService,
	router *gin.RouterGroup,
) {
	h := &bookingHandler{service: service}
	router.GET("/booking/:username", h.host)
	router.POST("/booking", h.add)
	router.GET("/schedule/:id", h.schedule)
}