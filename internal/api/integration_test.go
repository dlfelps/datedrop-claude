package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"datedrop/internal/api"
	"datedrop/internal/api/handlers"
	"datedrop/internal/config"
	"datedrop/internal/repository/memory"
	"datedrop/internal/services"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := config.NewDefaultConfig()

	userRepo := memory.NewUserRepository()
	questionRepo := memory.NewQuestionRepository()
	responseRepo := memory.NewResponseRepository()
	dropRepo := memory.NewDropRepository()
	shotRepo := memory.NewShotRepository()
	cupidRepo := memory.NewCupidRepository()
	moderationRepo := memory.NewModerationRepository()

	notifService := services.NewNotificationService()
	userService := services.NewUserService(userRepo)
	quizService := services.NewQuizService(questionRepo, responseRepo, userRepo, cfg.Quiz.TotalQuestions)
	matchingService := services.NewMatchingService(cfg, userRepo, questionRepo, responseRepo, dropRepo, moderationRepo, notifService)
	dropService := services.NewDropService(dropRepo, notifService)
	socialService := services.NewSocialService(cfg, shotRepo, cupidRepo, userRepo, dropRepo, moderationRepo, matchingService, notifService)
	moderationService := services.NewModerationService(moderationRepo, notifService)

	userHandler := handlers.NewUserHandler(userService)
	quizHandler := handlers.NewQuizHandler(quizService)
	matchingHandler := handlers.NewMatchingHandler(matchingService)
	dropHandler := handlers.NewDropHandler(dropService)
	socialHandler := handlers.NewSocialHandler(socialService)
	moderationHandler := handlers.NewModerationHandler(moderationService)
	seedHandler := handlers.NewSeedHandler(userRepo, questionRepo, responseRepo)

	router := api.NewRouter(userHandler, quizHandler, matchingHandler, dropHandler, socialHandler, moderationHandler, seedHandler)

	engine := gin.Default()
	router.Setup(engine)
	return engine
}

func TestHealthCheck(t *testing.T) {
	engine := setupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateUser(t *testing.T) {
	engine := setupRouter(t)

	body := `{"email":"test@stanford.edu","name":"Test User","date_of_birth":"2000-01-15","gender":"female","orientations":["straight"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var user map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &user)
	if user["email"] != "test@stanford.edu" {
		t.Errorf("expected email test@stanford.edu, got %v", user["email"])
	}
}

func TestRejectNonEduEmail(t *testing.T) {
	engine := setupRouter(t)

	body := `{"email":"test@gmail.com","name":"Bad User","date_of_birth":"2000-01-15","gender":"male","orientations":["straight"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRejectUnderage(t *testing.T) {
	engine := setupRouter(t)

	body := `{"email":"young@stanford.edu","name":"Young User","date_of_birth":"2015-01-15","gender":"male","orientations":["straight"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLoginFlow(t *testing.T) {
	engine := setupRouter(t)

	// Create user
	createBody := `{"email":"login@stanford.edu","name":"Login User","date_of_birth":"2000-01-15","gender":"female","orientations":["straight"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	// Login
	loginBody := `{"email":"login@stanford.edu"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["token"] == nil || result["token"] == "" {
		t.Error("expected token in login response")
	}
}

func TestAuthRequired(t *testing.T) {
	engine := setupRouter(t)

	// No auth header
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/quiz/questions", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestSeedAndMatch(t *testing.T) {
	engine := setupRouter(t)

	// Seed
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/debug/seed", nil)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("seed failed: %d %s", w.Code, w.Body.String())
	}

	var seedResult map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &seedResult)
	t.Logf("Seed: %v users, %v questions, %v responses",
		seedResult["users_created"], seedResult["questions_created"], seedResult["responses_created"])

	// Get a user token (create one)
	createBody := `{"email":"matcher@stanford.edu","name":"Matcher","date_of_birth":"2000-01-15","gender":"female","orientations":["straight"]}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	var user map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &user)
	token := user["id"].(string)

	// Run matching
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/matching/run", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("matching failed: %d %s", w.Code, w.Body.String())
	}

	var matchResult map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &matchResult)
	matchCount := matchResult["matches_created"].(float64)
	t.Logf("Matches created: %v", matchCount)

	if matchCount == 0 {
		t.Error("expected at least one match to be created")
	}
}

func TestShootYourShotMutual(t *testing.T) {
	engine := setupRouter(t)

	// Create two users
	bodyA := `{"email":"shootera@stanford.edu","name":"Shooter A","date_of_birth":"2000-01-15","gender":"female","orientations":["straight"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(bodyA))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	var userA map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &userA)
	tokenA := userA["id"].(string)

	bodyB := `{"email":"shooterb@mit.edu","name":"Shooter B","date_of_birth":"2000-01-15","gender":"male","orientations":["straight"]}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(bodyB))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	var userB map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &userB)
	tokenB := userB["id"].(string)

	// A shoots at B
	shotBody := `{"target_id":"` + tokenB + `"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/social/shoot", bytes.NewBufferString(shotBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenA)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("shot A->B failed: %d %s", w.Code, w.Body.String())
	}
	var shotResult map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &shotResult)
	if shotResult["mutual"].(bool) != false {
		t.Error("first shot should not be mutual")
	}

	// B shoots at A (should be mutual)
	shotBody = `{"target_id":"` + tokenA + `"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/social/shoot", bytes.NewBufferString(shotBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenB)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("shot B->A failed: %d %s", w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &shotResult)
	if shotResult["mutual"].(bool) != true {
		t.Error("reverse shot should be mutual")
	}
}

func TestBlockAndModeration(t *testing.T) {
	engine := setupRouter(t)

	// Create two users
	bodyA := `{"email":"blocker@stanford.edu","name":"Blocker","date_of_birth":"2000-01-15","gender":"female","orientations":["straight"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(bodyA))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	var userA map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &userA)
	tokenA := userA["id"].(string)

	bodyB := `{"email":"blocked@mit.edu","name":"Blocked","date_of_birth":"2000-01-15","gender":"male","orientations":["straight"]}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(bodyB))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	var userB map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &userB)
	tokenB := userB["id"].(string)

	// Block user B
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/moderation/block/"+tokenB, nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("block failed: %d %s", w.Code, w.Body.String())
	}

	// Unblock user B
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/moderation/block/"+tokenB, nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("unblock failed: %d %s", w.Code, w.Body.String())
	}

	// Report user B
	reportBody := `{"reported_id":"` + tokenB + `","category":"harassment","details":"Test report"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/moderation/report", bytes.NewBufferString(reportBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenA)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("report failed: %d %s", w.Code, w.Body.String())
	}
}

func TestFullWeeklyLifecycle(t *testing.T) {
	engine := setupRouter(t)

	// Seed data
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/debug/seed", nil)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("seed failed: %d", w.Code)
	}

	// Create a user to trigger matching
	createBody := `{"email":"trigger@stanford.edu","name":"Trigger","date_of_birth":"2000-01-15","gender":"female","orientations":["straight"]}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	var user map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &user)
	token := user["id"].(string)

	// Run matching
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/matching/run", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("matching failed: %d", w.Code)
	}

	var matchResult map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &matchResult)
	drops := matchResult["drops"].([]interface{})
	if len(drops) == 0 {
		t.Fatal("no drops created")
	}

	// Get first drop and its users
	firstDrop := drops[0].(map[string]interface{})
	dropID := firstDrop["id"].(string)
	user1ID := firstDrop["user1_id"].(string)
	user2ID := firstDrop["user2_id"].(string)

	t.Logf("Drop %s: user1=%s, user2=%s, score=%.2f",
		dropID, user1ID, user2ID, firstDrop["compatibility_score"].(float64))

	// User 1 gets current drop
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/drops/current", nil)
	req.Header.Set("Authorization", "Bearer "+user1ID)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("get drop failed for user1: %d %s", w.Code, w.Body.String())
	}

	// User 1 accepts
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/drops/"+dropID+"/accept", nil)
	req.Header.Set("Authorization", "Bearer "+user1ID)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("accept failed for user1: %d %s", w.Code, w.Body.String())
	}

	var dropAfterAccept1 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &dropAfterAccept1)
	if dropAfterAccept1["status"] != "pending_mutual" {
		t.Errorf("expected pending_mutual after first accept, got %s", dropAfterAccept1["status"])
	}

	// User 2 accepts (should become matched)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/drops/"+dropID+"/accept", nil)
	req.Header.Set("Authorization", "Bearer "+user2ID)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("accept failed for user2: %d %s", w.Code, w.Body.String())
	}

	var dropAfterAccept2 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &dropAfterAccept2)
	if dropAfterAccept2["status"] != "matched" {
		t.Errorf("expected matched after both accept, got %s", dropAfterAccept2["status"])
	}

	// Verify history
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/drops/history", nil)
	req.Header.Set("Authorization", "Bearer "+user1ID)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("history failed: %d %s", w.Code, w.Body.String())
	}
}
