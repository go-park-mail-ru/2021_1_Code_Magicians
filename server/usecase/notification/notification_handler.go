package notification

import (
	"net/http"
	"pinterest/delivery"
	"pinterest/domain/entity"
	"strconv"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

type NotificationInfo struct {
	notificationApp delivery.NotificationAppInterface
	logger          *zap.Logger
}

func NewNotificationInfo(notificationApp delivery.NotificationAppInterface, logger *zap.Logger) *NotificationInfo {
	return &NotificationInfo{
		notificationApp: notificationApp,
		logger:          logger,
	}
}

func (notificationInfo *NotificationInfo) HandleReadNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notificationIDStr := vars[string(entity.IDKey)]
	notificationID, _ := strconv.Atoi(notificationIDStr)
	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	err := notificationInfo.notificationApp.ReadNotification(userID, notificationID)

	if err != nil {
		notificationInfo.logger.Info(err.Error(), zap.Int("for user", userID))
		switch err {
		case entity.NotificationNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		case entity.NotificationAlreadyReadError:
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}