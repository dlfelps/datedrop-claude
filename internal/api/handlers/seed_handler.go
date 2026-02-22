package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"datedrop/internal/domain/entities"
	"datedrop/internal/repository/memory"
	"datedrop/pkg/utils"
)

type SeedHandler struct {
	userRepo     *memory.UserRepository
	questionRepo *memory.QuestionRepository
	responseRepo *memory.ResponseRepository
}

func NewSeedHandler(
	userRepo *memory.UserRepository,
	questionRepo *memory.QuestionRepository,
	responseRepo *memory.ResponseRepository,
) *SeedHandler {
	return &SeedHandler{
		userRepo:     userRepo,
		questionRepo: questionRepo,
		responseRepo: responseRepo,
	}
}

func (h *SeedHandler) Seed(c *gin.Context) {
	ctx := c.Request.Context()

	users := h.seedUsers(ctx)
	questions := h.seedQuestions(ctx)
	responseCount := h.seedResponses(ctx, users[:16], questions)

	c.JSON(http.StatusOK, gin.H{
		"message":         "seed data loaded",
		"users_created":   len(users),
		"questions_created": len(questions),
		"responses_created": responseCount,
	})
}

func (h *SeedHandler) seedUsers(ctx context.Context) []*entities.User {
	type seedUser struct {
		name         string
		email        string
		gender       entities.Gender
		orientations []entities.Orientation
		age          int
	}

	seeds := []seedUser{
		{"Alice Chen", "alice@stanford.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationStraight}, 22},
		{"Bob Smith", "bob@mit.edu", entities.GenderMale, []entities.Orientation{entities.OrientationStraight}, 23},
		{"Carol Davis", "carol@harvard.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationBisexual}, 21},
		{"David Kim", "david@berkeley.edu", entities.GenderMale, []entities.Orientation{entities.OrientationStraight}, 24},
		{"Emma Wilson", "emma@yale.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationStraight}, 22},
		{"Frank Lopez", "frank@columbia.edu", entities.GenderMale, []entities.Orientation{entities.OrientationGay}, 25},
		{"Grace Patel", "grace@princeton.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationStraight}, 23},
		{"Henry Zhang", "henry@caltech.edu", entities.GenderMale, []entities.Orientation{entities.OrientationStraight}, 22},
		{"Iris Martinez", "iris@upenn.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationBisexual}, 21},
		{"Jack Thompson", "jack@duke.edu", entities.GenderMale, []entities.Orientation{entities.OrientationStraight}, 24},
		{"Kate Brown", "kate@northwestern.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationStraight}, 23},
		{"Leo Garcia", "leo@uchicago.edu", entities.GenderMale, []entities.Orientation{entities.OrientationGay}, 22},
		{"Mia Johnson", "mia@cornell.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationStraight}, 25},
		{"Noah Williams", "noah@nyu.edu", entities.GenderMale, []entities.Orientation{entities.OrientationStraight}, 23},
		{"Olivia Lee", "olivia@rice.edu", entities.GenderNonBinary, []entities.Orientation{entities.OrientationBisexual}, 22},
		{"Paul Anderson", "paul@georgetown.edu", entities.GenderMale, []entities.Orientation{entities.OrientationBisexual}, 24},
		{"Quinn Taylor", "quinn@vanderbilt.edu", entities.GenderNonBinary, []entities.Orientation{entities.OrientationBisexual}, 21},
		{"Rachel Moore", "rachel@emory.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationStraight}, 23},
		{"Sam Jackson", "sam@tulane.edu", entities.GenderMale, []entities.Orientation{entities.OrientationStraight}, 22},
		{"Tina Nguyen", "tina@washu.edu", entities.GenderFemale, []entities.Orientation{entities.OrientationStraight}, 24},
	}

	var users []*entities.User
	for _, s := range seeds {
		dob := time.Now().AddDate(-s.age, 0, 0)
		user := entities.NewUser(utils.GenerateID(), s.email, s.name, dob, s.gender, s.orientations)
		user.Bio = fmt.Sprintf("Hi, I'm %s! Looking for meaningful connections.", s.name)
		h.userRepo.Create(ctx, user)
		users = append(users, user)
	}
	return users
}

func (h *SeedHandler) seedQuestions(ctx context.Context) []*entities.Question {
	var questions []*entities.Question
	idx := 0

	// Lifestyle domain (22 questions)
	lifestyleQuestions := []struct {
		text         string
		responseType entities.ResponseType
		options      []string
	}{
		{"How important is staying physically active?", entities.ResponseTypeScale5, nil},
		{"How often do you cook at home?", entities.ResponseTypeScale5, nil},
		{"What's your ideal weekend activity?", entities.ResponseTypeMultipleChoice, []string{"Outdoor adventure", "Netflix and chill", "Socializing", "Creative projects"}},
		{"Are you a morning person?", entities.ResponseTypeBoolean, nil},
		{"How important is travel to you?", entities.ResponseTypeScale7, nil},
		{"Do you enjoy large social gatherings?", entities.ResponseTypeScale5, nil},
		{"How adventurous are you with food?", entities.ResponseTypeScale5, nil},
		{"Do you prefer city or countryside living?", entities.ResponseTypeMultipleChoice, []string{"City", "Suburbs", "Countryside", "No preference"}},
		{"How important is a consistent daily routine?", entities.ResponseTypeScale5, nil},
		{"Do you enjoy reading books regularly?", entities.ResponseTypeBoolean, nil},
		{"How much do you value personal space?", entities.ResponseTypeScale7, nil},
		{"What's your preferred exercise type?", entities.ResponseTypeMultipleChoice, []string{"Gym", "Team sports", "Yoga/meditation", "Outdoor activities"}},
		{"How important is music in your daily life?", entities.ResponseTypeScale5, nil},
		{"Do you prefer planned vacations or spontaneous trips?", entities.ResponseTypeMultipleChoice, []string{"Planned", "Spontaneous", "Mix of both", "Don't travel much"}},
		{"How often do you go out on weeknights?", entities.ResponseTypeScale5, nil},
		{"Are you comfortable with long-distance relationships?", entities.ResponseTypeBoolean, nil},
		{"How important is financial stability to you?", entities.ResponseTypeScale7, nil},
		{"Do you enjoy pets?", entities.ResponseTypeBoolean, nil},
		{"How clean and organized do you keep your space?", entities.ResponseTypeScale5, nil},
		{"What's your relationship with social media?", entities.ResponseTypeMultipleChoice, []string{"Heavy user", "Moderate user", "Minimal user", "Don't use it"}},
		{"How important is alone time to you?", entities.ResponseTypeScale5, nil},
		{"Do you prefer deep conversations or light banter?", entities.ResponseTypeMultipleChoice, []string{"Deep conversations", "Light banter", "Both equally", "Depends on mood"}},
	}

	for _, q := range lifestyleQuestions {
		question := &entities.Question{
			ID:           utils.GenerateID(),
			Text:         q.text,
			Domain:       entities.DomainLifestyle,
			ResponseType: q.responseType,
			Options:      q.options,
			Version:      1,
			OrderIndex:   idx,
		}
		h.questionRepo.Create(ctx, question)
		questions = append(questions, question)
		idx++
	}

	// Values domain (22 questions)
	valuesQuestions := []struct {
		text         string
		responseType entities.ResponseType
		options      []string
	}{
		{"How important is family in your life?", entities.ResponseTypeScale7, nil},
		{"Do you want children in the future?", entities.ResponseTypeMultipleChoice, []string{"Definitely yes", "Probably yes", "Unsure", "No"}},
		{"How important is religious or spiritual practice?", entities.ResponseTypeScale7, nil},
		{"Do you believe in marriage?", entities.ResponseTypeBoolean, nil},
		{"How important is career ambition in a partner?", entities.ResponseTypeScale5, nil},
		{"How do you handle conflict in relationships?", entities.ResponseTypeMultipleChoice, []string{"Direct confrontation", "Calm discussion", "Need space first", "Avoid conflict"}},
		{"How important is honesty, even when it hurts?", entities.ResponseTypeScale7, nil},
		{"Do you value tradition or innovation more?", entities.ResponseTypeMultipleChoice, []string{"Tradition", "Innovation", "Balance of both", "Neither"}},
		{"How important is education level in a partner?", entities.ResponseTypeScale5, nil},
		{"Do you believe people can fundamentally change?", entities.ResponseTypeBoolean, nil},
		{"How important is shared sense of humor?", entities.ResponseTypeScale7, nil},
		{"What role should gender play in relationships?", entities.ResponseTypeMultipleChoice, []string{"Traditional roles", "Equal partnership", "Flexible roles", "No defined roles"}},
		{"How do you view work-life balance?", entities.ResponseTypeScale5, nil},
		{"Is physical appearance important in a partner?", entities.ResponseTypeScale5, nil},
		{"How important is emotional intelligence?", entities.ResponseTypeScale7, nil},
		{"Do you value independence or togetherness more?", entities.ResponseTypeMultipleChoice, []string{"Independence", "Togetherness", "Balance", "Depends on context"}},
		{"How important is loyalty to you?", entities.ResponseTypeScale7, nil},
		{"Do you believe in soulmates?", entities.ResponseTypeBoolean, nil},
		{"How important is cultural compatibility?", entities.ResponseTypeScale5, nil},
		{"How do you show love?", entities.ResponseTypeMultipleChoice, []string{"Words", "Actions", "Gifts", "Quality time"}},
		{"How important is intellectual stimulation?", entities.ResponseTypeScale5, nil},
		{"Do you think vulnerability is a strength?", entities.ResponseTypeBoolean, nil},
	}

	for _, q := range valuesQuestions {
		question := &entities.Question{
			ID:           utils.GenerateID(),
			Text:         q.text,
			Domain:       entities.DomainValues,
			ResponseType: q.responseType,
			Options:      q.options,
			Version:      1,
			OrderIndex:   idx,
		}
		h.questionRepo.Create(ctx, question)
		questions = append(questions, question)
		idx++
	}

	// Politics domain (22 questions)
	politicsQuestions := []struct {
		text         string
		responseType entities.ResponseType
		options      []string
	}{
		{"How important is it that your partner shares your political views?", entities.ResponseTypeScale7, nil},
		{"Where do you fall on the political spectrum?", entities.ResponseTypeMultipleChoice, []string{"Progressive", "Moderate", "Conservative", "Libertarian"}},
		{"How important is environmental activism?", entities.ResponseTypeScale5, nil},
		{"Do you believe in universal healthcare?", entities.ResponseTypeBoolean, nil},
		{"How important is social justice to you?", entities.ResponseTypeScale7, nil},
		{"What's your view on gun control?", entities.ResponseTypeMultipleChoice, []string{"Stricter laws needed", "Current laws fine", "Fewer restrictions", "No opinion"}},
		{"How important is freedom of speech, even for offensive views?", entities.ResponseTypeScale5, nil},
		{"Do you support immigration reform?", entities.ResponseTypeBoolean, nil},
		{"How important are LGBTQ+ rights to you?", entities.ResponseTypeScale7, nil},
		{"What role should government play in the economy?", entities.ResponseTypeMultipleChoice, []string{"More regulation", "Less regulation", "Balanced approach", "No opinion"}},
		{"How important is racial equity to you?", entities.ResponseTypeScale7, nil},
		{"Do you believe in climate change action?", entities.ResponseTypeBoolean, nil},
		{"How do you feel about taxes for social programs?", entities.ResponseTypeScale5, nil},
		{"What's your view on drug legalization?", entities.ResponseTypeMultipleChoice, []string{"Full legalization", "Medical only", "Decriminalization", "Against"}},
		{"How important is separation of church and state?", entities.ResponseTypeScale7, nil},
		{"Do you support labor unions?", entities.ResponseTypeBoolean, nil},
		{"How important is criminal justice reform?", entities.ResponseTypeScale5, nil},
		{"What's your stance on foreign policy?", entities.ResponseTypeMultipleChoice, []string{"Interventionist", "Isolationist", "Diplomatic focus", "No strong opinion"}},
		{"How important is wealth inequality as an issue?", entities.ResponseTypeScale7, nil},
		{"Do you believe in affirmative action?", entities.ResponseTypeBoolean, nil},
		{"How politically engaged are you?", entities.ResponseTypeScale5, nil},
		{"Can you date someone with very different political views?", entities.ResponseTypeBoolean, nil},
	}

	for _, q := range politicsQuestions {
		question := &entities.Question{
			ID:           utils.GenerateID(),
			Text:         q.text,
			Domain:       entities.DomainPolitics,
			ResponseType: q.responseType,
			Options:      q.options,
			Version:      1,
			OrderIndex:   idx,
		}
		h.questionRepo.Create(ctx, question)
		questions = append(questions, question)
		idx++
	}

	return questions
}

func (h *SeedHandler) seedResponses(ctx context.Context, users []*entities.User, questions []*entities.Question) int {
	rng := rand.New(rand.NewSource(42))
	count := 0

	for _, user := range users {
		for _, q := range questions {
			resp := &entities.QuizResponse{
				ID:              utils.GenerateID(),
				UserID:          user.ID,
				QuestionID:      q.ID,
				ImportanceScore: rng.Intn(5) + 1,
				CreatedAt:       time.Now(),
			}

			switch q.ResponseType {
			case entities.ResponseTypeScale5:
				v := rng.Intn(5) + 1
				resp.ScaleValue = &v
			case entities.ResponseTypeScale7:
				v := rng.Intn(7) + 1
				resp.ScaleValue = &v
			case entities.ResponseTypeMultipleChoice:
				resp.ChoiceValue = q.Options[rng.Intn(len(q.Options))]
			case entities.ResponseTypeBoolean:
				v := rng.Intn(2) == 1
				resp.BooleanValue = &v
			}

			h.responseRepo.Create(ctx, resp)
			count++
		}

		// Mark user as quiz complete
		user.QuizCompleted = true
		user.UpdatedAt = time.Now()
		h.userRepo.Update(ctx, user)
	}

	return count
}
