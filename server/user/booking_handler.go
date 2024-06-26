package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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
	eventTypes, err := h.service.EventType(ctx, user.ID, user.Username)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	user.EventTypes = eventTypes
	ctx.JSON(http.StatusOK, user)
}

func (h bookingHandler) schedule(ctx *gin.Context) {
	bid := ctx.Param("id")
	num, err := strconv.Atoi(bid)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	booking, err := h.service.Booking(ctx, num)
	if err != nil {
		ctx.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, booking)
}

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
	eventTypes, err := h.service.EventType(ctx, user.ID, user.Username)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": err.Error()})
		return
	}
	// booking
	var summary string
	var event interface{}
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
	if body.MeetingLocation == "google" && user.GoogleToken.Valid {
		cfg := hof.GetGoogleOAuthConfig()
		tok := &oauth2.Token{}
		if err = json.Unmarshal([]byte(user.GoogleToken.String), tok); err != nil {
			ctx.JSON(http.StatusUnprocessableEntity,
				gin.H{"error": err.Error()})
			return
		}
		calendarService := hof.GetGoogleCalendarService(ctx, tok, cfg)
		_, email, err := hof.GetGoogleUserData(ctx, tok, cfg)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity,
				gin.H{"error": err.Error()})
			return
		}
		summary = fmt.Sprintf("%s between %s and %s", title, user.Username, body.Name)
		description := fmt.Sprintf("maybe notes? %s", body.Notes)
		event, err = hof.SetGoogleNewMeeting(calendarService, summary, description, timezone,
			email, body.Email, body.Date, body.Time, duration)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.MeetingLocation == "microsoft" && user.MicrosoftToken.Valid {
		tok := &oauth2.Token{}
		if err = json.Unmarshal([]byte(user.MicrosoftToken.String), tok); err != nil {
			ctx.JSON(http.StatusUnprocessableEntity,
				gin.H{"error": err.Error()})
			return
		}
		summary = fmt.Sprintf("%s between %s and %s", title, user.Username, body.Name)
		eventData := hof.ComposeMSMeetingData(timezone, summary, body.Date, body.Time,
			duration, body.Name, body.Email)
		// limitation: Only Work for Business Account (personal account not supported)
		//meeting, err := hof.SetMicrosoftNewMeeting(eventData.Start.DateTime, eventData.End.DateTime,
		//	eventData.Subject, tok.AccessToken)
		//if err != nil {
		//	ctx.JSON(http.StatusBadRequest, err.Error())
		//	return
		//}
		//meetingURL := meeting["joinWebUrl"].(string)
		//eventData.Body = hof.MSBody{
		//	ContentType: "HTML",
		//	Content:     fmt.Sprintf("Does next month work for you? <br><a href=\"%s\">Join the meeting</a>", meetingURL),
		//}
		event, err = hof.SetMicrosoftNewCalendarEvent(eventData, tok.AccessToken)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}
		return
	}
	// insert booking data
	id, err := h.service.NewBooking(ctx, user.ID, summary, &body, event)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	}
	ctx.JSON(http.StatusCreated, gin.H{"id": id})
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
