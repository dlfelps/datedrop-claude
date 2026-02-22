package services

import "log"

// NotificationService is a mock that logs all notifications to console.
type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) NotifyNewDrop(userID, dropID string) {
	log.Printf("[NOTIFICATION] User %s: You have a new drop! (drop_id=%s)", userID, dropID)
}

func (s *NotificationService) NotifyMatch(user1ID, user2ID, dropID string) {
	log.Printf("[NOTIFICATION] Users %s and %s: It's a match! (drop_id=%s)", user1ID, user2ID, dropID)
}

func (s *NotificationService) NotifyMutualShot(user1ID, user2ID, dropID string) {
	log.Printf("[NOTIFICATION] Users %s and %s: Mutual shot! You both shot your shot! (drop_id=%s)", user1ID, user2ID, dropID)
}

func (s *NotificationService) NotifyCupidNomination(userID, nomID string) {
	log.Printf("[NOTIFICATION] User %s: You've been nominated by Cupid! (nomination_id=%s)", userID, nomID)
}

func (s *NotificationService) NotifyDropExpired(userID, dropID string) {
	log.Printf("[NOTIFICATION] User %s: Your drop has expired. (drop_id=%s)", userID, dropID)
}

func (s *NotificationService) NotifyBlock(blockerID, blockedID string) {
	log.Printf("[NOTIFICATION] User %s blocked user %s", blockerID, blockedID)
}

func (s *NotificationService) NotifyReport(reporterID, reportedID, category string) {
	log.Printf("[NOTIFICATION] User %s reported user %s for %s", reporterID, reportedID, category)
}
