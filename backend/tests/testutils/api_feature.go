package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	jwttoken "judge-backend/internal/app/identity_domain/auth/domain/valueobject/jwt_token"
	"judge-backend/internal/app/identity_domain/auth/infrastructure/repository"
	"judge-backend/internal/app/identity_domain/auth/infrastructure/service"
	userRole "judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/infrastructure/service/user/roleconverter"
	"judge-backend/internal/app/shared"
	"judge-backend/internal/app/shared/infrastructure/logger"
	hashservice "judge-backend/internal/app/shared/infrastructure/services/hash_service"
	zipservice "judge-backend/internal/app/shared/infrastructure/services/zip_service"
	"judge-backend/internal/app/shared/model"
	authmodel "judge-backend/internal/app/shared/model"
	imagemodel "judge-backend/internal/app/shared/model"
	"judge-backend/internal/config"
	"judge-backend/internal/projectpath"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type APIFeature struct {
	resp              *http.Response
	streamResp        *http.Response
	streamBuf         *bytes.Buffer
	config            *config.Config
	cookies           map[string]string
	streamMu          sync.RWMutex
	currentScenarioID string
	streamOwnerID     string
	streamDone        chan struct{}
}

func NewAPIFeature() *APIFeature {
	return &APIFeature{
		resp:      nil,
		cookies:   make(map[string]string),
		streamBuf: bytes.NewBuffer([]byte("")),
		config:    config.NewFeatureConfig(),
	}
}

// buildAPIURL turns a scenario's endpoint into an absolute URL.
//
// The query string is split off rather than left in Path: url.URL escapes Path,
// so a "?" inside it becomes %3F and the request goes to a path that literally
// contains those characters — which answers 404 rather than reaching the
// handler. Nothing caught this until the first endpoints with query parameters
// were added, because the spec had none before them.
func (a *APIFeature) buildAPIURL(endpoint string) string {
	path, rawQuery, _ := strings.Cut(endpoint, "?")
	u := url.URL{
		Scheme:   "http",
		Host:     "localhost:" + strconv.Itoa(a.config.Test.Server.Port),
		Path:     path,
		RawQuery: rawQuery,
	}
	return u.String()
}

func (a *APIFeature) ResetResponse(sc *godog.Scenario) {
	a.setCurrentScenario(sc)
	fmt.Printf("[feature][reset] scenario=%q id=%s\n", sc.Name, sc.Id)
	a.resp = nil
	a.cookies = make(map[string]string)
	a.resetStreamBuffer()
}

func (a *APIFeature) setCurrentScenario(sc *godog.Scenario) {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

	a.currentScenarioID = sc.Id
}

func (a *APIFeature) getCurrentScenarioID() string {
	a.streamMu.RLock()
	defer a.streamMu.RUnlock()

	return a.currentScenarioID
}

func (a *APIFeature) setActiveStream(ownerID string, resp *http.Response, streamDone chan struct{}) {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

	a.streamResp = resp
	a.streamOwnerID = ownerID
	a.streamDone = streamDone
}

func (a *APIFeature) clearActiveStream(ownerID string, streamDone chan struct{}) {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

	if a.streamOwnerID != ownerID || a.streamDone != streamDone {
		return
	}

	a.streamResp = nil
	a.streamOwnerID = ""
	a.streamDone = nil
}

func (a *APIFeature) CloseStreamForScenario(scenarioID string) {
	a.streamMu.Lock()
	ownerID := a.streamOwnerID
	hasStream := a.streamResp != nil
	if scenarioID == "" || ownerID != scenarioID || !hasStream {
		a.streamMu.Unlock()
		fmt.Printf("[feature][stream][close-skip] requested=%s owner=%s has_stream=%t\n", scenarioID, ownerID, hasStream)
		return
	}

	resp := a.streamResp
	streamDone := a.streamDone
	a.streamResp = nil
	a.streamOwnerID = ""
	a.streamDone = nil
	a.streamMu.Unlock()

	fmt.Printf("[feature][stream][close] owner=%s\n", scenarioID)

	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	if streamDone != nil {
		select {
		case <-streamDone:
			fmt.Printf("[feature][stream][closed] owner=%s\n", scenarioID)
		case <-time.After(2 * time.Second):
			fmt.Printf("[feature][stream][close-timeout] owner=%s\n", scenarioID)
		}
	}
}

func (a *APIFeature) resetStreamBuffer() {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

	a.streamBuf = bytes.NewBuffer([]byte(""))
}

func (a *APIFeature) appendStreamData(data []byte) {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

	a.streamBuf.Write(data)
}

func (a *APIFeature) streamContains(message string) bool {
	a.streamMu.RLock()
	defer a.streamMu.RUnlock()

	return bytes.Contains(a.streamBuf.Bytes(), []byte(message))
}

func (a *APIFeature) IHaveUsers(table *godog.Table) error {
	db := GetDB()
	logger := logger.NewMockLogger()
	hashService := hashservice.NewHashService(logger)
	roleConverter := roleconverter.NewRoleConverter(logger)
	roleService := service.NewUserRoleService(roleConverter, logger)

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		name := table.Rows[i].Cells[1].Value
		username := table.Rows[i].Cells[2].Value
		password := table.Rows[i].Cells[3].Value
		roles := strings.Split(table.Rows[i].Cells[4].Value, ",")
		rs := []*role.Role{}
		for _, r := range role.Roles {
			for _, ur := range roles {
				if r.Name() == strings.TrimSpace(ur) {
					rs = append(rs, r)
					break
				}
			}
		}
		user := user.MustNew("test", "test", "test")
		user.SetRoles(rs...)

		hashedPassword, err := hashService.Hash(password)
		if err != nil {
			return err
		}

		db.Create(&authmodel.User{
			ID:       id,
			Name:     name,
			Username: username,
			Password: hashedPassword,
			Roles:    roleService.GetRolesBit(user),
		})

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

// IHaveRefreshTokensForTheFollowingUsers plants refresh tokens straight into
// Redis, and has to leave behind the same state a real login would — that
// means the per-user index as well as the token itself. Writing only the
// token would still satisfy every scenario that looks a token up by value,
// while silently making the account's sessions un-revokable, which is exactly
// the behaviour some scenarios exist to check.
func (s *APIFeature) IHaveRefreshTokensForTheFollowingUsers(table *godog.Table) error {
	rdb := GetRedisClient()
	ctx := context.Background()

	for i := 1; i < len(table.Rows); i++ {
		refreshToken := table.Rows[i].Cells[0].Value
		userID := table.Rows[i].Cells[1].Value
		ttl, _ := strconv.Atoi(table.Rows[i].Cells[2].Value)
		expiry := time.Duration(ttl * int(time.Second))

		if err := rdb.Set(ctx, refreshToken, userID, expiry).Err(); err != nil {
			return err
		}

		// Mirrors RefreshTokenRepository.Save — keep the key format in step
		// with internal/app/identity_domain/auth/.../refresh_token_repository.go.
		indexKey := fmt.Sprintf("user_refresh_tokens:%s", userID)
		if err := rdb.SAdd(ctx, indexKey, refreshToken).Err(); err != nil {
			return err
		}
		if err := rdb.Expire(ctx, indexKey, expiry).Err(); err != nil {
			return err
		}
	}

	return nil
}

// ClearSubmissionCooldowns removes every cooldown reservation. It runs in the
// shared per-scenario reset: the database resets never touched Redis, so a
// cooldown seeded by one scenario's successful submit outlived it and 429'd
// the next scenario's submit whenever the gap was shorter than the window —
// green or red depended on machine speed. Keys are deleted by prefix rather
// than flushing the database, so the auth refresh tokens sharing this Redis
// never enter the argument.
func ClearSubmissionCooldowns() error {
	rdb := GetRedisClient()
	ctx := context.Background()

	keys, err := rdb.Keys(ctx, "submission_cooldown:*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	return rdb.Del(ctx, keys...).Err()
}

func (a *APIFeature) IHaveNoUsers() error {
	db := GetDB()
	return db.Delete(&authmodel.User{}, "1 = 1").Error
}

func (a *APIFeature) IHaveImages(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		tag := table.Rows[i].Cells[1].Value
		mountPath := table.Rows[i].Cells[2].Value
		description := table.Rows[i].Cells[3].Value
		isReady := table.Rows[i].Cells[4].Value == "1"
		createdAt := table.Rows[i].Cells[5].Value
		updatedAt := table.Rows[i].Cells[6].Value

		image := &imagemodel.Image{
			ID:          id,
			Tag:         tag,
			MountPath:   mountPath,
			IsReady:     isReady,
			Description: description,
		}
		image.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		image.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		db.Create(image)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveNoImages() error {
	db := GetDB()
	return db.Delete(&model.Image{}, "1 = 1").Error
}

func (a *APIFeature) IHaveBuilds(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		imageTag := table.Rows[i].Cells[1].Value
		batch, _ := strconv.Atoi(table.Rows[i].Cells[2].Value)
		status := table.Rows[i].Cells[3].Value
		createdAt := table.Rows[i].Cells[4].Value
		updatedAt := table.Rows[i].Cells[5].Value

		image := &imagemodel.Image{}
		if err := db.Where("tag = ?", imageTag).First(image).Error; err != nil {
			return err
		}

		s, err := buildstatus.NewByName(status)
		if err != nil {
			return err
		}
		build := &imagemodel.Build{
			ID:      id,
			ImageID: image.ID,
			Batch:   batch,
			Status:  s.Value(),
		}
		build.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		build.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		db.Create(build)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveBuildLogs(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		buildID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		level := table.Rows[i].Cells[2].Value
		message := table.Rows[i].Cells[3].Value
		createdAt := table.Rows[i].Cells[4].Value

		l, err := buildloglevel.NewByText(level)
		if err != nil {
			return err
		}
		buildLog := &imagemodel.BuildLog{
			ID:      id,
			BuildID: buildID,
			Level:   l.Value(),
			Message: message,
		}
		buildLog.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

		db.Create(buildLog)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveProjects(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		name := table.Rows[i].Cells[1].Value
		description := table.Rows[i].Cells[2].Value
		imageTag := table.Rows[i].Cells[3].Value
		env := table.Rows[i].Cells[4].Value
		testImageTag := table.Rows[i].Cells[5].Value
		testEnv := table.Rows[i].Cells[6].Value
		resultParser := table.Rows[i].Cells[7].Value
		createdAt := table.Rows[i].Cells[8].Value
		updatedAt := table.Rows[i].Cells[9].Value

		project := &model.Project{
			ID:           int32(id),
			Name:         name,
			Description:  description,
			ImageTag:     &imageTag,
			TestImageTag: &testImageTag,
			Env:          env,
			TestEnv:      testEnv,
			ResultParser: resultParser,
		}

		project.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		project.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(project)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveAssets(table *godog.Table) error {
	db := GetDB()
	cwd, _ := os.Getwd()
	os.Chdir(projectpath.Root)

	for i := 1; i < len(table.Rows); i++ {
		id := table.Rows[i].Cells[0].Value
		projectID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		name := table.Rows[i].Cells[2].Value
		description := table.Rows[i].Cells[3].Value
		size, _ := strconv.Atoi(table.Rows[i].Cells[4].Value)
		isEnabled := table.Rows[i].Cells[5].Value == "1"
		createdAt := table.Rows[i].Cells[6].Value
		updatedAt := table.Rows[i].Cells[7].Value

		asset := &model.Asset{
			ID:          id,
			ProjectID:   int32(projectID),
			Name:        name,
			Description: description,
			Size:        int64(size),
			IsEnabled:   isEnabled,
		}
		asset.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		asset.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(asset)

		if db.Error != nil {
			return db.Error
		}

		src := path.Join(Root, "/data/nginx.zip")
		srcFile, err := os.Open(src)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dst := path.Join(a.config.Server.StoragePath, "assets", id)
		dstFile, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
	}

	os.Chdir(cwd)
	return nil
}

func (a *APIFeature) IHaveDynamicFields(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		projectID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		t := table.Rows[i].Cells[2].Value
		name := table.Rows[i].Cells[3].Value
		label := table.Rows[i].Cells[4].Value
		defaultValue := table.Rows[i].Cells[5].Value
		isRequired := table.Rows[i].Cells[6].Value == "1"
		createdAt := table.Rows[i].Cells[7].Value
		updatedAt := table.Rows[i].Cells[8].Value

		dynamicField := &model.DynamicField{
			ID:           int32(id),
			ProjectID:    int32(projectID),
			Type:         t,
			Name:         name,
			Label:        label,
			DefaultValue: &defaultValue,
			IsRequired:   isRequired,
		}
		dynamicField.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		dynamicField.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(dynamicField)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveProjectsLifecycles(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		projectID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		name := table.Rows[i].Cells[2].Value
		description := table.Rows[i].Cells[3].Value
		t := table.Rows[i].Cells[4].Value
		runsOn := table.Rows[i].Cells[5].Value
		command := table.Rows[i].Cells[6].Value
		createdAt := table.Rows[i].Cells[7].Value
		updatedAt := table.Rows[i].Cells[8].Value

		lifecycle := &model.Lifecycle{
			ID:          int32(id),
			ProjectID:   int32(projectID),
			Name:        name,
			Description: description,
			Type:        t,
			RunsOn:      runsOn,
			Command:     command,
		}
		lifecycle.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		lifecycle.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(lifecycle)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveProjectsAssets(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id := table.Rows[i].Cells[0].Value
		projectID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		name := table.Rows[i].Cells[2].Value
		description := table.Rows[i].Cells[3].Value
		size, _ := strconv.Atoi(table.Rows[i].Cells[4].Value)
		isEnabled := table.Rows[i].Cells[5].Value == "1"
		createdAt := table.Rows[i].Cells[6].Value
		updatedAt := table.Rows[i].Cells[7].Value

		asset := &model.Asset{
			ID:          id,
			ProjectID:   int32(projectID),
			Name:        name,
			Description: description,
			Size:        int64(size),
			IsEnabled:   isEnabled,
		}
		asset.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		asset.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(asset)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveGradingItems(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		projectID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		name := table.Rows[i].Cells[2].Value
		description := table.Rows[i].Cells[3].Value
		score, _ := strconv.ParseFloat(table.Rows[i].Cells[4].Value, 64)
		order, _ := strconv.Atoi(table.Rows[i].Cells[5].Value)
		scoringStrategy := table.Rows[i].Cells[6].Value
		orderFormat := table.Rows[i].Cells[7].Value
		resultParser := table.Rows[i].Cells[8].Value
		label := table.Rows[i].Cells[9].Value
		createdAt := table.Rows[i].Cells[10].Value
		updatedAt := table.Rows[i].Cells[11].Value

		gradingItem := &model.GradingItem{
			ID:              int32(id),
			ProjectID:       int32(projectID),
			Name:            name,
			Description:     description,
			Score:           score,
			Order:           uint32(order),
			ScoringStrategy: scoringStrategy,
			OutputFormat:    orderFormat,
			ResultParser:    resultParser,
			Label:           label,
		}
		gradingItem.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		gradingItem.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(gradingItem)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveTestcases(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		gradingItemID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		name := table.Rows[i].Cells[2].Value
		command := table.Rows[i].Cells[3].Value
		score, _ := strconv.ParseFloat(table.Rows[i].Cells[4].Value, 64)
		isSample := table.Rows[i].Cells[5].Value == "1"
		createdAt := table.Rows[i].Cells[6].Value
		updatedAt := table.Rows[i].Cells[7].Value

		testcase := &model.Testcase{
			ID:            int32(id),
			GradingItemID: int32(gradingItemID),
			Name:          name,
			Command:       command,
			Score:         score,
			IsSample:      isSample,
		}
		testcase.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		testcase.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(testcase)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveGradingItemDepends(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		gradingItemID, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		testcaseID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)

		depend := &model.GradingItemDepend{
			GradingItemID: int32(gradingItemID),
			TestcaseID:    int32(testcaseID),
		}

		db.Create(depend)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveTestcaseDepends(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		testcaseID, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		dependID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)

		depend := &model.TestcaseDepend{
			TestcaseID: int32(testcaseID),
			DependID:   int32(dependID),
		}

		db.Create(depend)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveContests(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		name := table.Rows[i].Cells[1].Value
		description := table.Rows[i].Cells[2].Value
		isPublic := table.Rows[i].Cells[3].Value == "1"
		activateTime := table.Rows[i].Cells[4].Value
		deactivateTime := table.Rows[i].Cells[5].Value
		createdAt := table.Rows[i].Cells[6].Value
		updatedAt := table.Rows[i].Cells[7].Value

		contest := &model.Contest{
			ID:          int32(id),
			Name:        name,
			Description: description,
			IsPublic:    isPublic,
		}

		contest.ActivateTime, _ = time.Parse(time.DateTime, activateTime)
		contest.DeactivateTime, _ = time.Parse(time.DateTime, deactivateTime)
		contest.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		contest.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(contest)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveContestProjectAssociations(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		contestID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		projectID, _ := strconv.Atoi(table.Rows[i].Cells[2].Value)
		name := table.Rows[i].Cells[3].Value
		isPublic := table.Rows[i].Cells[4].Value == "1"
		judgePriorityStrategy := table.Rows[i].Cells[5].Value
		activeTime := table.Rows[i].Cells[6].Value
		deactivateTime := table.Rows[i].Cells[7].Value
		startTime := table.Rows[i].Cells[8].Value
		endTime := table.Rows[i].Cells[9].Value

		contestProject := &model.ContestProject{
			ID:                    int32(id),
			ContestID:             int32(contestID),
			ProjectID:             int32(projectID),
			Name:                  name,
			IsPublic:              isPublic,
			JudgePriorityStrategy: judgePriorityStrategy,
		}

		// Optional trailing column. Leaving it off means the contest project
		// has no cooldown of its own and inherits the system default, which
		// is what every table written before this column existed expects.
		if len(table.Rows[i].Cells) > 10 {
			if raw := strings.TrimSpace(table.Rows[i].Cells[10].Value); raw != "" {
				seconds, err := strconv.Atoi(raw)
				if err != nil {
					return fmt.Errorf("invalid submissionCooldownSeconds %q: %w", raw, err)
				}
				v := int32(seconds)
				contestProject.SubmissionCooldownSeconds = &v
			}
		}

		// A parse failure must be loud: a silently zero-valued window used to
		// slip through (nothing on the write side read these columns), but the
		// submission eligibility gate now rejects such a contest project, and
		// the resulting 400 looks nothing like the actual mistake — a table
		// cell written as "2021-09-01T00:00:00" instead of "2021-09-01 00:00:00".
		var err error
		if contestProject.ActivateTime, err = time.Parse(time.DateTime, activeTime); err != nil {
			return fmt.Errorf("invalid activateTime %q: %w", activeTime, err)
		}
		if contestProject.DeactivateTime, err = time.Parse(time.DateTime, deactivateTime); err != nil {
			return fmt.Errorf("invalid deactivateTime %q: %w", deactivateTime, err)
		}
		if contestProject.StartTime, err = time.Parse(time.DateTime, startTime); err != nil {
			return fmt.Errorf("invalid startTime %q: %w", startTime, err)
		}
		if contestProject.EndTime, err = time.Parse(time.DateTime, endTime); err != nil {
			return fmt.Errorf("invalid endTime %q: %w", endTime, err)
		}

		db.Create(contestProject)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IHaveClarifications(table *godog.Table) error {
	db := GetDB()

	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		contestProjectID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)
		subject := table.Rows[i].Cells[2].Value
		content := table.Rows[i].Cells[3].Value
		senderUserID, _ := strconv.Atoi(table.Rows[i].Cells[4].Value)
		receiverUserID, _ := strconv.Atoi(table.Rows[i].Cells[5].Value)
		sendToAll := table.Rows[i].Cells[6].Value == "1"
		isAnswered := table.Rows[i].Cells[7].Value == "1"
		replyClarificationID, _ := strconv.Atoi(table.Rows[i].Cells[8].Value)
		createdAt := table.Rows[i].Cells[9].Value
		updatedAt := table.Rows[i].Cells[10].Value

		clarification := &model.Clarification{
			ID:               int32(id),
			ContestProjectID: int32(contestProjectID),
			Subject:          subject,
			Content:          content,
			SendToAll:        sendToAll,
			IsAnswered:       isAnswered,
		}

		if senderUserID != 0 {
			clarification.SenderUserID = &senderUserID
		}

		if receiverUserID != 0 {
			clarification.ReceiverUserID = &receiverUserID
		}

		if replyClarificationID != 0 {
			var replyClarificationID int32 = int32(replyClarificationID)
			clarification.ReplyClarificationID = &replyClarificationID
		}

		clarification.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
		clarification.UpdatedAt, _ = time.Parse(time.DateTime, updatedAt)

		db.Create(clarification)

		if db.Error != nil {
			return db.Error
		}
	}

	return nil
}

func (a *APIFeature) IAmLoggedInAs(username string) {
	db := GetDB()
	logger := logger.NewMockLogger()
	roleConverter := roleconverter.NewRoleConverter(logger)
	userRepository := repository.NewUserRepository(db, service.NewUserRoleService(roleConverter, logger), logger)
	user, err := userRepository.FindByUsername(username)
	if err != nil {
		panic(err)
	}

	encryptService := service.NewEncryptService(a.config, logger)
	encryptedUserID, err := encryptService.Encrypt(strconv.Itoa(user.GetID()))
	if err != nil {
		panic(err)
	}

	jwtToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          a.config.Server.JwtKey,
		EncryptedUserID: encryptedUserID,
		ExpirationTime:  int64(a.config.Auth.AccessTokenTTLSeconds),
		Roles:           user.GetRoles(),
	})

	a.cookies["access_token"] = jwtToken.String()
}

// IAmLoggedInAsWithAnExpiredAccessToken mints a token whose exp is already in
// the past, so a protected endpoint has to fall through to the middleware's
// renewal path — the path production was failing on and that no scenario
// covered. NewJwtToken skips claims validation when signing, so an expiry in
// the past is minted happily; NewJwtTokenWithRaw, which the parse path uses,
// does validate and rejects it.
func (a *APIFeature) IAmLoggedInAsWithAnExpiredAccessToken(username string) {
	db := GetDB()
	logger := logger.NewMockLogger()
	roleConverter := roleconverter.NewRoleConverter(logger)
	userRepository := repository.NewUserRepository(db, service.NewUserRoleService(roleConverter, logger), logger)
	user, err := userRepository.FindByUsername(username)
	if err != nil {
		panic(err)
	}

	encryptService := service.NewEncryptService(a.config, logger)
	encryptedUserID, err := encryptService.Encrypt(strconv.Itoa(user.GetID()))
	if err != nil {
		panic(err)
	}

	jwtToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          a.config.Server.JwtKey,
		EncryptedUserID: encryptedUserID,
		ExpirationTime:  -1,
		Roles:           user.GetRoles(),
	})

	a.cookies["access_token"] = jwtToken.String()
}

// ILogInViaTheAPI performs a real login and keeps the cookies the server
// handed back. IAmLoggedInAs mints its own token, which is convenient but
// bypasses the server's configured lifetime — this is the step to use when
// the point of the scenario is what that lifetime actually does.
func (a *APIFeature) ILogInViaTheAPI(username, password string) error {
	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)

	resp, err := Request("POST", a.buildAPIURL("/api/v1/auth/login"), &body, nil, nil)
	if err != nil {
		return err
	}
	a.resp = resp

	if err := a.TheResponseCodeShouldBe(200); err != nil {
		return err
	}

	for _, ck := range resp.Cookies() {
		a.cookies[ck.Name] = ck.Value
	}

	return nil
}

// IWaitForSeconds sleeps. It is deliberately the only unconditional sleep in
// the suite: it exists for the short-TTL run, where the thing being waited
// for is a token expiring and there is no state to poll for instead.
func (a *APIFeature) IWaitForSeconds(seconds int) error {
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}

// TheResponseShouldSetCookie asserts the response carries a Set-Cookie for
// name with a non-empty value. Renewal is only useful if the browser is
// actually handed the new token, and nothing else in this suite looks at
// response cookies.
func (a *APIFeature) TheResponseShouldSetCookie(name string) error {
	if a.resp == nil {
		return errors.New("response is nil, please send a request first")
	}

	for _, cookie := range a.resp.Cookies() {
		if cookie.Name != name {
			continue
		}
		if cookie.Value == "" {
			return fmt.Errorf("response set cookie %s but its value is empty", name)
		}
		return nil
	}

	return fmt.Errorf("response did not set cookie %s", name)
}

func (a *APIFeature) TheUserShouldNotExist(username string) error {
	db := GetDB()
	user := &authmodel.User{}
	err := db.Where("username = ?", username).First(user).Error
	if err != nil {
		return nil
	}

	return fmt.Errorf("user with username %s exists", username)
}

func (a *APIFeature) TheUserShouldExistWithTheFollowingData(table *godog.Table) error {
	db := GetDB()
	logger := logger.NewMockLogger()
	roleConverter := roleconverter.NewRoleConverter(logger)

	name := table.Rows[1].Cells[0].Value
	username := table.Rows[1].Cells[1].Value
	roles := strings.Split(table.Rows[1].Cells[2].Value, ",")

	user := &model.User{}
	if err := db.Where("username = ?", username).First(user).Error; err != nil {
		return err
	}

	if user.Name != name {
		return errors.New("name is not correct")
	}

	rs := userRole.Roles{}
	for _, r := range userRole.GetRoles() {
		for _, ur := range roles {
			if r.Name() == strings.TrimSpace(ur) {
				rs = append(rs, r)
				break
			}
		}
	}

	rolesBit := roleConverter.ToBit(rs)
	if user.Roles != rolesBit {
		return errors.New("roles are not correct")
	}

	return nil
}

func (a *APIFeature) CanLoginWithTheUser(doc *godog.DocString) error {
	if err := a.ISendrequestToWithTheFollowingData("POST", "/api/v1/auth/login", doc); err != nil {
		return err
	}

	if err := a.TheResponseCodeShouldBe(200); err != nil {
		return err
	}

	return nil
}

func (a *APIFeature) TheImageShouldNotExist(tag string) error {
	db := GetDB()
	image := &model.Image{}
	err := db.Where("tag = ?", tag).First(image).Error
	if err != nil {
		return nil
	}

	return fmt.Errorf("image with tag %s exists", tag)
}

func (a *APIFeature) TheBuildStatusShouldBe(tag, status string, second int) error {
	db := GetDB()
	image := &model.Image{}
	if err := db.Where("tag = ?", tag).First(image).Error; err != nil {
		return err
	}

	ch := make(chan struct{})
	defer close(ch)
	go func() {
		for {
			build := &model.Build{}
			db.Where("image_id = ?", image.ID).Order("batch desc").First(build)
			s, _ := buildstatus.NewByName(status)
			if build.Status == s.Value() {
				ch <- struct{}{}
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-ch:
	case <-time.After(time.Duration(second) * time.Second):
		return fmt.Errorf("timeout waiting for image build to %s", status)
	}

	return nil
}

func (a *APIFeature) SetupProjectIntoImage(projectZip, tag string) error {
	file, err := os.Open(path.Join(Root, "/data/", projectZip))
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)

	path := path.Join(projectpath.Root, a.config.Server.StoragePath, "images", tag)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	logger := logger.NewMockLogger()
	// Same limits the running server enforces, so fixtures cannot pass here
	// while a real upload of the same archive would be rejected.
	zipService := zipservice.NewService(logger, shared.NewZipLimits(a.config))
	if err := zipService.UnzipToPath(buf, path, true); err != nil {
		return err
	}

	return nil
}

func (a *APIFeature) ISendrequestTo(method, endpoint string) (err error) {
	a.resp, err = Request(method, a.buildAPIURL(endpoint), nil, nil, a.cookies)
	if err != nil {
		return err
	}

	return nil
}

func (a *APIFeature) ISendrequestToWithTheFollowingData(method, endpoint string, body *godog.DocString) (err error) {
	a.resp, err = Request(method, a.buildAPIURL(endpoint), &body.Content, nil, a.cookies)
	if err != nil {
		return err
	}

	return nil
}

func (a *APIFeature) ISendrequestToWithTheFollowingFormData(method, endpoint string, form *godog.Table) (err error) {
	data := make(map[string]string)
	for i := 1; i < len(form.Rows); i++ {
		key := form.Rows[i].Cells[0].Value
		value := form.Rows[i].Cells[1].Value
		data[key] = value
	}

	a.resp, err = FormRequest(method, a.buildAPIURL(endpoint), data, nil, a.cookies)
	if err != nil {
		return err
	}

	return nil
}

func (a *APIFeature) ISendrequestToWithTheFollowingCookies(method, endpoint string, cookies *godog.Table) (err error) {
	cookie := make(map[string]string)
	for i := 1; i < len(cookies.Rows); i++ {
		key := cookies.Rows[i].Cells[0].Value
		value := cookies.Rows[i].Cells[1].Value
		cookie[key] = value
	}

	for k, v := range a.cookies {
		cookie[k] = v
	}

	a.resp, err = Request(method, a.buildAPIURL(endpoint), nil, nil, cookie)
	if err != nil {
		return err
	}

	return nil
}

func (a *APIFeature) StartStreamingViaRequestTo(endpoint string) error {
	scenarioID := a.getCurrentScenarioID()
	fmt.Printf("[feature][stream][open] scenario=%s endpoint=%s\n", scenarioID, endpoint)
	a.CloseStreamForScenario(scenarioID)
	a.resetStreamBuffer()

	err := a.ISendrequestTo("GET", endpoint)
	if err != nil {
		return err
	}

	if a.resp == nil {
		return errors.New("response is nil, please send a request first")
	}

	if a.resp.StatusCode != http.StatusOK {
		_ = a.resp.Body.Close()
		a.resp = nil
		time.Sleep(1 * time.Second)
		return a.StartStreamingViaRequestTo(endpoint)
	}

	streamResp := a.resp
	a.resp = nil

	incomingStream := make(chan struct{}, 1)
	streamDone := make(chan struct{})
	a.setActiveStream(scenarioID, streamResp, streamDone)
	fmt.Printf("[feature][stream][active] scenario=%s endpoint=%s\n", scenarioID, endpoint)

	go func(body io.ReadCloser, ownerID string, streamDone chan struct{}) {
		defer close(streamDone)
		defer a.clearActiveStream(ownerID, streamDone)
		defer fmt.Printf("[feature][stream][reader-exit] owner=%s endpoint=%s\n", ownerID, endpoint)

		for {
			buf := make([]byte, 1024)
			n, err := body.Read(buf)
			if n > 0 {
				if len(incomingStream) == 0 {
					incomingStream <- struct{}{}
				}
				a.appendStreamData(buf[:n])
			}

			if err != nil {
				return
			}
		}
	}(streamResp.Body, scenarioID, streamDone)

	select {
	case <-incomingStream:
		fmt.Printf("[feature][stream][ready] scenario=%s endpoint=%s\n", scenarioID, endpoint)
		return nil
	case <-time.After(30 * time.Second):
		fmt.Printf("[feature][stream][ready-timeout] scenario=%s endpoint=%s\n", scenarioID, endpoint)
		a.CloseStreamForScenario(scenarioID)
		return errors.New("timeout waiting for stream")
	}
}

func (a *APIFeature) WaitForSecondsForTheImageBuildToComplete(seconds int, tag string) error {
	db := GetDB()
	ch := make(chan struct{})
	go func() {
		for {
			image := &model.Image{}
			db.Where("tag = ?", tag).First(&image)
			build := &model.Build{}
			db.Where("image_id = ?", image.ID).Last(&build)
			if build != nil && build.Status == buildstatus.SUCCESS.Value() {
				close(ch)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-ch:
	case <-time.After(time.Duration(seconds) * time.Second):
		return fmt.Errorf("timeout waiting for image build to complete")
	}
	return nil
}

func (a *APIFeature) WaitForSecondsForTheImageBuildToError(seconds int, tag string) error {
	db := GetDB()
	ch := make(chan struct{})
	go func() {
		for {
			image := &model.Image{}
			db.Where("tag = ?", tag).First(&image)
			build := &model.Build{}
			db.Where("image_id = ?", image.ID).Last(&build)
			if build != nil && build.Status == buildstatus.ERROR.Value() {
				close(ch)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-ch:
	case <-time.After(time.Duration(seconds) * time.Second):
		return fmt.Errorf("timeout waiting for image build to error")
	}
	return nil
}

func (a *APIFeature) TheResponseCodeShouldBe(code int) error {
	return ResponseMatchStatusCode(a.resp, code)
}

func (a *APIFeature) TheResponseShouldMatchJSON(body *godog.DocString) (err error) {
	return ResponseMatchJSON(a.resp, body)
}

func (a *APIFeature) SetCookie(key string) error {
	if a.resp == nil {
		return errors.New("response is nil, please send a request first")
	}

	cookie := a.resp.Cookies()
	for _, c := range cookie {
		if c.Name == key {
			a.cookies[key] = c.Value
			return nil
		}
	}

	return fmt.Errorf("cookie with key %s not found", key)
}

func (a *APIFeature) TheStreamShouldHaveMessage(message string, second int) error {
	ch := make(chan struct{}, 1)
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			if a.streamContains(message) {
				select {
				case ch <- struct{}{}:
				case <-done:
				}
				return
			}

			select {
			case <-time.After(1 * time.Second):
			case <-done:
				return
			}
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(time.Duration(second) * time.Second):
		return errors.New("timeout waiting for message")
	}
}

// TheStreamShouldNotHaveMessage asserts the message is absent from everything
// the stream has delivered SO FAR. It proves something only after a positive
// TheStreamShouldHaveMessage on the same stream has shown that the frame which
// could have carried the message already arrived — absence before any frame
// would pass vacuously.
func (a *APIFeature) TheStreamShouldNotHaveMessage(message string) error {
	if a.streamContains(message) {
		return fmt.Errorf("expected the stream to not contain %q, but it does", message)
	}
	return nil
}

func (a *APIFeature) ICanDownloadFileWithRequestTo(url string) error {
	err := a.ISendrequestTo("GET", url)
	if err != nil {
		return err
	}

	if a.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response status code is %d", a.resp.StatusCode)
	}

	contentDisposition := a.resp.Header.Get("Content-Disposition")
	if contentDisposition != "" && (contentDisposition[:10] == "attachment" || contentDisposition[:9] == "inline") {
		return nil
	}

	return errors.New("response is not a file")
}

func (a *APIFeature) ICanDownloadAssetFileWithRequestTo(fileName, method, path string) error {
	db := GetDB()

	asset := &model.Asset{}
	if err := db.Where("name = ?", fileName).First(asset).Error; err != nil {
		return err
	}

	url := strings.Replace(path, "{assetId}", asset.ID, 1)

	return a.ICanDownloadFileWithRequestTo(url)
}

func (a *APIFeature) ICanDownloadGradingProjectWithRequestTo(method, url string) error {
	return a.ICanDownloadFileWithRequestTo(url)
}

func (a *APIFeature) IHaveGroups(table *godog.Table) error {
	db := GetDB()
	for i := 1; i < len(table.Rows); i++ {
		id, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		name := table.Rows[i].Cells[1].Value
		description := table.Rows[i].Cells[2].Value

		group := &model.Group{
			ID:          id,
			Name:        name,
			Description: description,
		}

		// Optional 4th column: deletedAt, for seeding soft-deleted groups.
		if len(table.Rows[i].Cells) > 3 && table.Rows[i].Cells[3].Value != "" {
			deletedAt, _ := time.Parse(time.DateTime, table.Rows[i].Cells[3].Value)
			group.DeletedAt = gorm.DeletedAt{Time: deletedAt, Valid: true}
		}

		db.Create(group)
		if db.Error != nil {
			return db.Error
		}
	}
	return nil
}

func (a *APIFeature) IHaveUserGroups(table *godog.Table) error {
	db := GetDB()
	for i := 1; i < len(table.Rows); i++ {
		userID, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		groupID, _ := strconv.Atoi(table.Rows[i].Cells[1].Value)

		db.Create(&model.UserGroup{UserID: userID, GroupID: groupID})
		if db.Error != nil {
			return db.Error
		}
	}
	return nil
}

func (a *APIFeature) IHaveContestParticipants(table *godog.Table) error {
	db := GetDB()
	for i := 1; i < len(table.Rows); i++ {
		contestID, _ := strconv.Atoi(table.Rows[i].Cells[0].Value)
		userIDStr := table.Rows[i].Cells[1].Value
		groupIDStr := table.Rows[i].Cells[2].Value

		p := &model.ContestParticipant{ContestID: int32(contestID)}
		if userIDStr != "" {
			uid, _ := strconv.Atoi(userIDStr)
			p.UserID = &uid
		}
		if groupIDStr != "" {
			gid, _ := strconv.Atoi(groupIDStr)
			p.GroupID = &gid
		}

		db.Create(p)
		if db.Error != nil {
			return db.Error
		}
	}
	return nil
}

// IHaveSubmissions inserts submission documents into MongoDB (judge database).
// Supported columns: id, userID, contestProjectID, isPartial (opt), ignored (opt)
func (a *APIFeature) IHaveSubmissions(table *godog.Table) error {
	mdb := GetMongoDB()
	ctx := context.Background()
	col := mdb.Database("judge").Collection("submissions")

	for _, row := range table.Rows[1:] {
		headers := table.Rows[0].Cells
		doc := bson.M{}
		for i, cell := range row.Cells {
			key := headers[i].Value
			val := cell.Value
			switch key {
			case "id":
				doc["_id"] = val
			case "userID", "contestProjectID":
				v, _ := strconv.Atoi(val)
				doc[key] = int32(v)
			case "isPartial":
				doc[key] = val == "true"
			case "ignored":
				doc[key] = val == "true"
			// Parsed the same way IHaveJudgments parses its time columns. A
			// string here would satisfy Mongo and then decode to a zero time
			// in every query that reads it — and the queue-wait analytics
			// subtracts this very field, so a fixture that cannot pin it can
			// only assert around wall-clock "now".
			case "createdAt":
				if ts, err := time.Parse(time.RFC3339, val); err == nil {
					doc[key] = ts
				} else if ts, err := time.Parse(time.DateTime, val); err == nil {
					doc[key] = ts
				} else {
					return fmt.Errorf("invalid submissions createdAt %q", val)
				}
			default:
				doc[key] = val
			}
		}
		if doc["gradingItemIds"] == nil {
			doc["gradingItemIds"] = bson.A{}
		}
		if doc["createdAt"] == nil {
			doc["createdAt"] = time.Now()
		}

		filter := bson.M{"_id": doc["_id"]}
		_, err := col.UpdateOne(ctx, filter, bson.M{"$set": doc}, options.Update().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("failed to insert submission %v: %w", doc["_id"], err)
		}
	}
	return nil
}

// gradingItemFixture and testcaseFixture mirror the grading item snapshot a
// judgment carries, so a scenario can describe one as JSON in the
// `gradingItems` column instead of only ever getting an empty array. Typed
// rather than a bare map because the ids have to reach Mongo as int32 to decode
// back into the int32 fields the queries read them into — encoding/json would
// otherwise hand every number over as a float64.
type testcaseFixture struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	TotalScore  float64 `json:"totalScore"`
	ParsedScore float64 `json:"parsedScore"`
	Status      string  `json:"status"`
	Command     string  `json:"command"`
	// Stdout/Stderr land in the snapshot's result document, so a scenario can
	// assert which testcases' run output reaches (or must not reach) a
	// contestant. Omitted, they stay the empty strings they always were.
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type gradingItemFixture struct {
	ID              int32             `json:"id"`
	Label           string            `json:"label"`
	Name            string            `json:"name"`
	Score           float64           `json:"score"`
	TotalScore      float64           `json:"totalScore"`
	Status          string            `json:"status"`
	ScoringStrategy string            `json:"scoringStrategy"`
	ResultParser    string            `json:"resultParser"`
	Testcases       []testcaseFixture `json:"testcases"`
}

func parseGradingItemsFixture(raw string) (bson.A, error) {
	if strings.TrimSpace(raw) == "" {
		return bson.A{}, nil
	}

	var fixtures []gradingItemFixture
	if err := json.Unmarshal([]byte(raw), &fixtures); err != nil {
		return nil, fmt.Errorf("failed to parse gradingItems fixture %q: %w", raw, err)
	}

	items := bson.A{}
	for _, fixture := range fixtures {
		testcases := bson.A{}
		for _, tc := range fixture.Testcases {
			testcases = append(testcases, bson.M{
				"id":          tc.ID,
				"name":        tc.Name,
				"score":       tc.Score,
				"totalScore":  tc.TotalScore,
				"parsedScore": tc.ParsedScore,
				"status":      tc.Status,
				"command":     tc.Command,
				// Readers dereference these without a nil check, so a fixture that
				// omitted them would panic the handler rather than fail an assertion.
				"lifecycle": bson.M{"beforeEach": bson.A{}, "afterEach": bson.A{}},
				"result":    bson.M{"exitCode": int32(0), "execTime": int64(0), "stdout": tc.Stdout, "stderr": tc.Stderr},
			})
		}

		items = append(items, bson.M{
			"id":              fixture.ID,
			"label":           fixture.Label,
			"name":            fixture.Name,
			"score":           fixture.Score,
			"totalScore":      fixture.TotalScore,
			"status":          fixture.Status,
			"scoringStrategy": fixture.ScoringStrategy,
			"resultParser":    fixture.ResultParser,
			"testcases":       testcases,
		})
	}

	return items, nil
}

// IHaveJudgments inserts judgment documents into MongoDB (judge database).
// Supported columns: id, submissionID, status, judgeType, score (opt),
// totalScore (opt), rejudgeID (opt), gradingItems (opt, JSON array)
func (a *APIFeature) IHaveJudgments(table *godog.Table) error {
	mdb := GetMongoDB()
	ctx := context.Background()
	col := mdb.Database("judge").Collection("judgments")

	for _, row := range table.Rows[1:] {
		headers := table.Rows[0].Cells
		doc := bson.M{}
		for i, cell := range row.Cells {
			key := headers[i].Value
			val := cell.Value
			switch key {
			case "id":
				doc["_id"] = val
			case "score", "totalScore":
				v, _ := strconv.ParseFloat(val, 64)
				doc[key] = v
			case "contestProjectID":
				// int32 to match what the repository writes, and what the
				// queries decode this field back into — as a string it makes
				// any such query fail to decode the whole document, which
				// surfaces as a 500 rather than as a wrong number. The
				// submission fixture alongside this one already does the same.
				v, err := strconv.ParseInt(val, 10, 32)
				if err != nil {
					return fmt.Errorf("failed to parse judgment contestProjectID %q: %w", val, err)
				}
				doc[key] = int32(v)
			case "gradingItems":
				items, err := parseGradingItemsFixture(val)
				if err != nil {
					return err
				}
				doc[key] = items
			case "createdAt", "judgingStartedAt", "judgingFinishedAt":
				if val == "" {
					continue
				}
				if ts, err := time.Parse(time.RFC3339, val); err == nil {
					doc[key] = ts
				} else if ts, err := time.Parse(time.DateTime, val); err == nil {
					doc[key] = ts
				} else {
					return fmt.Errorf("failed to parse judgment %s %q: %w", key, val, err)
				}
			default:
				doc[key] = val
			}
		}
		if doc["createdAt"] == nil {
			doc["createdAt"] = time.Now()
		}
		if doc["gradingItems"] == nil {
			doc["gradingItems"] = bson.A{}
		}
		if doc["project"] == nil {
			doc["project"] = bson.M{
				"id": int32(0), "imageTag": "", "testImageTag": "",
				"env": bson.A{}, "testEnv": bson.A{}, "mountPath": "",
			}
		}
		if doc["lifecycle"] == nil {
			doc["lifecycle"] = bson.M{"beforeAll": bson.A{}, "afterAll": bson.A{}}
		}
		if doc["totalScore"] == nil {
			doc["totalScore"] = 0.0
		}

		filter := bson.M{"_id": doc["_id"]}
		_, err := col.UpdateOne(ctx, filter, bson.M{"$set": doc}, options.Update().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("failed to insert judgment %v: %w", doc["_id"], err)
		}
	}
	return nil
}

// IHaveRejudges inserts rejudge documents into MongoDB (judge database).
// Supported columns: id, contestProjectID (opt), status, reason, judgments
// (comma-separated judgment IDs, opt)
func (a *APIFeature) IHaveRejudges(table *godog.Table) error {
	mdb := GetMongoDB()
	ctx := context.Background()
	col := mdb.Database("judge").Collection("rejudges")

	for _, row := range table.Rows[1:] {
		headers := table.Rows[0].Cells
		doc := bson.M{}
		for i, cell := range row.Cells {
			key := headers[i].Value
			val := cell.Value
			switch key {
			case "id":
				doc["_id"] = val
			case "contestProjectID":
				// int32 rather than the default string: the aggregation behind
				// GET /rejudges decodes this field into an int32, and a string
				// would surface as a zero the fixture never asked for.
				v, err := strconv.ParseInt(val, 10, 32)
				if err != nil {
					return fmt.Errorf("failed to parse rejudge contestProjectID %q: %w", val, err)
				}
				doc[key] = int32(v)
			case "judgments":
				if val == "" {
					doc["judgments"] = bson.A{}
				} else {
					arr := bson.A{}
					for _, id := range strings.Split(val, ",") {
						arr = append(arr, strings.TrimSpace(id))
					}
					doc["judgments"] = arr
				}
			case "createdAt":
				// Stored as a time rather than the default string: the listing
				// sorts on this field, and a string would both sort wrongly and
				// fail to decode back into a time.
				if val == "" {
					continue
				}
				if ts, err := time.Parse(time.RFC3339, val); err == nil {
					doc[key] = ts
				} else if ts, err := time.Parse(time.DateTime, val); err == nil {
					doc[key] = ts
				} else {
					return fmt.Errorf("failed to parse rejudge createdAt %q: %w", val, err)
				}
			default:
				doc[key] = val
			}
		}
		if doc["judgments"] == nil {
			doc["judgments"] = bson.A{}
		}
		if doc["createdAt"] == nil {
			doc["createdAt"] = time.Now()
		}

		filter := bson.M{"_id": doc["_id"]}
		_, err := col.UpdateOne(ctx, filter, bson.M{"$set": doc}, options.Update().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("failed to insert rejudge %v: %w", doc["_id"], err)
		}
	}
	return nil
}

// AllJudgeWorkersAreStopped stops every worker in the pool and waits for them
// to report Stopped.
//
// Queue-depth assertions need it: the pool has several workers racing to pop,
// so a judgment queued while they are running is usually gone again before a
// status request can observe it, and the scenario would be asserting on a
// timing accident either way. With the workers stopped, "queued" is a state the
// test controls rather than a window it has to catch.
//
// Driven through the same /judge-workers API an admin would use, so the seam
// under test is the real one.
func (a *APIFeature) AllJudgeWorkersAreStopped() error {
	ids, err := a.judgeWorkerIDs()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return errors.New("no judge workers exist, so stopping them proves nothing about queue depth")
	}

	for _, id := range ids {
		if err := a.ISendrequestTo("POST", fmt.Sprintf("/api/v1/judge-workers/%s/stop", id)); err != nil {
			return err
		}
		if a.resp == nil || a.resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to stop judge worker %s", id)
		}
	}

	// Stop is asynchronous: it signals the worker, which finishes whatever it
	// is holding before settling. Queueing work before every worker has
	// actually settled would let one last pop through.
	deadline := time.Now().Add(30 * time.Second)
	for {
		statuses, err := a.judgeWorkerStatuses()
		if err != nil {
			return err
		}
		running := []string{}
		for id, status := range statuses {
			if status != "STOPPED" {
				running = append(running, id+"="+status)
			}
		}
		if len(running) == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("judge workers did not stop within budget: %v", running)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// AllJudgeWorkersAreStarted is the mirror of AllJudgeWorkersAreStopped, for
// scenarios that stop the pool to observe a queue filling and then want to watch
// it drain again.
func (a *APIFeature) AllJudgeWorkersAreStarted() error {
	ids, err := a.judgeWorkerIDs()
	if err != nil {
		return err
	}

	for _, id := range ids {
		if err := a.ISendrequestTo("POST", fmt.Sprintf("/api/v1/judge-workers/%s/start", id)); err != nil {
			return err
		}
		if a.resp == nil || a.resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to start judge worker %s", id)
		}
	}
	return nil
}

// TheJudgeQueueShouldDrainWithin polls the status endpoint until nothing is
// queued. Draining is what the workers do on their own schedule, so this waits
// for it rather than asserting on one sample.
func (a *APIFeature) TheJudgeQueueShouldDrainWithin(seconds int) error {
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	for {
		depths, _, err := a.judgeQueueSnapshot()
		if err != nil {
			return err
		}

		total := 0
		for _, depth := range depths {
			total += depth
		}
		if total == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("judge queue still holds %d judgments after %ds: %v", total, seconds, depths)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// TheJudgeQueueDepthShouldReachWithin polls until the named judge type's queue
// holds exactly `want` judgments. Enqueueing happens on an async handler, so the
// depth is not observable the moment the triggering request returns.
func (a *APIFeature) TheJudgeQueueDepthShouldReachWithin(judgeType string, want int, seconds int) error {
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	for {
		depths, _, err := a.judgeQueueSnapshot()
		if err != nil {
			return err
		}
		if got, ok := depths[judgeType]; ok && got == want {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("judge queue depth for %s did not reach %d within %ds, last reading: %v",
				judgeType, want, seconds, depths)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// TheSubmissionRateShouldHavePointsTotalling checks the shape and the sum of
// the rate series rather than which bucket each submission landed in.
//
// The buckets are anchored to the wall clock, so a scenario that seeded a
// submission a few milliseconds before a minute boundary would otherwise be
// asserting on which side of that boundary the test happened to fall. The two
// properties that are actually specified — one point per minute in the window
// including empty ones, and every submission counted exactly once — do not
// depend on that timing.
func (a *APIFeature) TheSubmissionRateShouldHavePointsTotalling(minutes, wantPoints, wantTotal int) error {
	return a.assertSubmissionRate(fmt.Sprintf("/api/v1/submission-rate?minutes=%d", minutes), wantPoints, wantTotal)
}

func (a *APIFeature) TheSubmissionRateForContestProjectShouldHavePointsTotalling(minutes, contestProjectID, wantPoints, wantTotal int) error {
	endpoint := fmt.Sprintf("/api/v1/submission-rate?minutes=%d&contestProjectId=%d", minutes, contestProjectID)
	return a.assertSubmissionRate(endpoint, wantPoints, wantTotal)
}

func (a *APIFeature) assertSubmissionRate(endpoint string, wantPoints, wantTotal int) error {
	if err := a.ISendrequestTo("GET", endpoint); err != nil {
		return err
	}
	if a.resp == nil {
		return errors.New("no response from the submission rate endpoint")
	}
	defer a.resp.Body.Close()

	body, err := io.ReadAll(a.resp.Body)
	if err != nil {
		return err
	}
	if a.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("submission rate endpoint returned %d: %s", a.resp.StatusCode, string(body))
	}

	var payload struct {
		Data []struct {
			Minute string `json:"minute"`
			Count  int    `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to decode the submission rate: %w", err)
	}

	if len(payload.Data) != wantPoints {
		return fmt.Errorf("expected %d points (one per minute in the window, empty ones included), got %d: %s",
			wantPoints, len(payload.Data), string(body))
	}

	total := 0
	for _, point := range payload.Data {
		total += point.Count
	}
	if total != wantTotal {
		return fmt.Errorf("expected the buckets to total %d submissions, got %d: %s", wantTotal, total, string(body))
	}
	return nil
}

// TheStreamedSubmissionRateShouldTotalWithin waits for a streamed frame whose
// buckets sum to want.
//
// The rate payload is all numbers, so there is no distinctive substring to
// watch for the way there is with an error message — and the counts cannot be
// matched as raw text either, since the step syntax cannot carry the quotes a
// JSON key needs. Parsing the frames is what lets a scenario assert that a
// submission made after the connection opened shows up in a later push.
func (a *APIFeature) TheStreamedSubmissionRateShouldTotalWithin(want int, seconds int) error {
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	lastSeen := -1
	for {
		frames := a.streamedDataFrames()
		for _, frame := range frames {
			var payload struct {
				Data []struct {
					Count int `json:"count"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(frame), &payload); err != nil {
				// A frame can still be arriving in pieces; a later read sees it
				// whole.
				continue
			}
			total := 0
			for _, point := range payload.Data {
				total += point.Count
			}
			lastSeen = total
			if total == want {
				return nil
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("no streamed submission rate totalling %d within %ds (%d frames seen, last total %d)",
				want, seconds, len(frames), lastSeen)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// streamedDataFrames splits the buffered stream into the JSON bodies of its
// complete "data: ...\n\n" frames.
func (a *APIFeature) streamedDataFrames() []string {
	a.streamMu.RLock()
	buffered := a.streamBuf.String()
	a.streamMu.RUnlock()

	frames := []string{}
	for _, chunk := range strings.Split(buffered, "\n\n") {
		body, found := strings.CutPrefix(strings.TrimSpace(chunk), "data: ")
		if !found {
			continue
		}
		frames = append(frames, body)
	}
	return frames
}

// judgeQueueSnapshot reads the status endpoint, returning the depths keyed by
// the judge type names the API uses and the oldest wait (nil when idle).
func (a *APIFeature) judgeQueueSnapshot() (map[string]int, *float64, error) {
	if err := a.ISendrequestTo("GET", "/api/v1/judge-queue"); err != nil {
		return nil, nil, err
	}
	if a.resp == nil {
		return nil, nil, errors.New("no response from the judge queue endpoint")
	}
	defer a.resp.Body.Close()

	body, err := io.ReadAll(a.resp.Body)
	if err != nil {
		return nil, nil, err
	}
	if a.resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("judge queue endpoint returned %d: %s", a.resp.StatusCode, string(body))
	}

	var payload struct {
		Data struct {
			QueueDepths struct {
				Judge        int `json:"judge"`
				JudgePartial int `json:"judgePartial"`
				Rejudge      int `json:"rejudge"`
			} `json:"queueDepths"`
			OldestQueuedSeconds *float64 `json:"oldestQueuedSeconds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, nil, fmt.Errorf("failed to decode the judge queue status: %w", err)
	}

	return map[string]int{
		"judge":        payload.Data.QueueDepths.Judge,
		"judgePartial": payload.Data.QueueDepths.JudgePartial,
		"rejudge":      payload.Data.QueueDepths.Rejudge,
	}, payload.Data.OldestQueuedSeconds, nil
}

func (a *APIFeature) judgeWorkerIDs() ([]string, error) {
	statuses, err := a.judgeWorkerStatuses()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(statuses))
	for id := range statuses {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

func (a *APIFeature) judgeWorkerStatuses() (map[string]string, error) {
	if err := a.ISendrequestTo("GET", "/api/v1/judge-workers"); err != nil {
		return nil, err
	}
	if a.resp == nil {
		return nil, errors.New("no response from the judge worker listing")
	}
	defer a.resp.Body.Close()

	body, err := io.ReadAll(a.resp.Body)
	if err != nil {
		return nil, err
	}
	if a.resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("judge worker listing returned %d: %s", a.resp.StatusCode, string(body))
	}

	var payload struct {
		Data []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode the judge worker listing: %w", err)
	}

	statuses := make(map[string]string, len(payload.Data))
	for _, worker := range payload.Data {
		statuses[worker.ID] = worker.Status
	}
	return statuses, nil
}

// ARejudgeForContestProjectShouldExistWithin polls until a rejudge for the
// given contest project has been persisted.
//
// Requesting a rejudge is answered before the record exists: the request only
// publishes an event, and the judge domain creates the judgments and the
// rejudge on an async handler. A scenario that reads the rejudge back through
// the API therefore has to wait for that write rather than assume the 200 means
// it already landed.
func (a *APIFeature) ARejudgeForContestProjectShouldExistWithin(contestProjectID int, seconds int) error {
	ctx := context.Background()
	col := GetMongoDB().Database("judge").Collection("rejudges")
	filter := bson.M{"contestProjectID": int32(contestProjectID)}

	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	for {
		count, err := col.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		if time.Now().After(deadline) {
			total, err := col.CountDocuments(ctx, bson.M{})
			if err != nil {
				return err
			}
			return fmt.Errorf("no rejudge for contest project %d after %ds (%d rejudges exist in total)",
				contestProjectID, seconds, total)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
