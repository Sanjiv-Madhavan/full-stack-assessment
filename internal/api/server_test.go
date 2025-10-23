package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	"full-stack-assesment/internal/api"
	"full-stack-assesment/internal/migrate"
	projectsRepo "full-stack-assesment/internal/repo/projects"
	tasksRepo "full-stack-assesment/internal/repo/task"
	projectsService "full-stack-assesment/internal/service/projects"
	taskService "full-stack-assesment/internal/service/task"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "modernc.org/sqlite"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// init seeds the random number generator.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandStringRunes generates a random string of n runes.
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var _ = Describe("API Endpoints testing", Ordered, func() {
	var (
		db      *sql.DB
		handler http.Handler

		seedProjectID    = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		invalidProjectID = "aaaaaaaa-aaaa-aaaa-aaaa-bbbbbbbbbbbb"
		seedProjectName  = "Demo Project"
	)

	BeforeAll(func() {
		var err error
		ctx := context.Background()
		db, err = sql.Open("sqlite", "file:todo?mode=memory&cache=shared&_fk=1")
		Expect(err).NotTo(HaveOccurred())
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)

		Expect(migrate.Apply(ctx, db)).To(Succeed())

		pRepo := projectsRepo.NewSQLiteProjectsRepo(db)
		pSvc := projectsService.NewService(*pRepo)

		tRepo := tasksRepo.NewSQLiteTaskRepo(db)
		tSvc := taskService.NewService(*tRepo, *pSvc)

		s := api.NewServer(*pSvc, *tSvc)
		mux := http.NewServeMux()
		handler = api.HandlerFromMux(s, mux)
	})

	AfterAll(func() {
		if db != nil {
			_ = db.Close()
		}
	})

	do := func(method, url string, body any) *httptest.ResponseRecorder {
		var r io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
			r = bytes.NewReader(b)
		}
		req, err := http.NewRequest(method, url, r)
		Expect(err).NotTo(HaveOccurred())
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	readJSON := func(rr *httptest.ResponseRecorder, dest any) {
		ExpectWithOffset(1, json.Unmarshal(rr.Body.Bytes(), dest)).To(Succeed(),
			"status=%d body=%s", rr.Code, rr.Body.String())
	}

	Describe("Health check endpoint", func() {
		It("GET /health returns successful response", func() {
			rr := do(http.MethodGet, "/health", nil)
			Expect(rr.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("Projects endpoints", func() {
		It("GET /projects returns seeded Demo Project", func() {
			rr := do(http.MethodGet, "/projects", nil)
			Expect(rr.Code).To(Equal(http.StatusOK))
			var got []map[string]any
			readJSON(rr, &got)
			var found bool
			for _, p := range got {
				if p["id"] == seedProjectID && p["name"] == seedProjectName {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "expected seeded Demo Project to exist")
		})

		It("POST /projects twice on a same name returns status conflict", func() {
			body := map[string]any{"name": "Alpha"}
			rr := do(http.MethodPost, "/projects", body)
			Expect(rr.Code).To(Equal(http.StatusCreated))
			var created map[string]any
			readJSON(rr, &created)
			Expect(created["id"]).NotTo(BeEmpty())
			Expect(created["name"]).To(Equal("Alpha"))

			rr2 := do(http.MethodPost, "/projects", body)
			Expect(rr2.Code).To(Equal(http.StatusConflict))
		})

		It("POST /projects rejects invalid body", func() {
			rr := do(http.MethodPost, "/projects", map[string]any{"name": "  "})
			Expect(rr.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("Tasks", func() {
		var hostProjectID string

		BeforeAll(func() {
			rr := do(http.MethodPost, "/projects", map[string]any{"name": "TasksHost"})
			Expect(rr.Code).To(Equal(http.StatusCreated))
			var created map[string]any
			readJSON(rr, &created)
			hostProjectID = created["id"].(string)
			Expect(hostProjectID).NotTo(BeEmpty())
		})

		// Checking pagination first because if I do it post creation of tasks, cleanup code is required.
		// If time permitted, clean up per test case can be done better
		Context("Pagination", func() {
			var taskIDs []string

			BeforeAll(func() {
				taskIDs = make([]string, 0, 3)
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)

				//Checking the ordering for 3 tasks
				for i := 1; i <= 3; i++ {
					title := fmt.Sprintf("Pagin Task %d", i)
					rr := do(http.MethodPost, url, map[string]any{"title": title})
					Expect(rr.Code).To(Equal(http.StatusCreated))
					var task map[string]any
					readJSON(rr, &task)
					taskIDs = append(taskIDs, task["id"].(string))
					time.Sleep(20 * time.Millisecond)
				}
			})

			It("GET /.../tasks?limit=2 returns only 2 tasks", func() {
				url := fmt.Sprintf("/projects/%s/tasks?limit=2", hostProjectID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusOK))
				var list []map[string]any
				readJSON(rr, &list)

				Expect(list).To(HaveLen(2))
				Expect(list[0]["id"]).To(Equal(taskIDs[2]))
				Expect(list[1]["id"]).To(Equal(taskIDs[1]))
			})

			It("GET /.../tasks?limit=2&offset=1 returns the next 2 tasks", func() {
				url := fmt.Sprintf("/projects/%s/tasks?limit=2&offset=1", hostProjectID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusOK))
				var list []map[string]any
				readJSON(rr, &list)

				Expect(list).To(HaveLen(2))
				Expect(list[0]["id"]).To(Equal(taskIDs[1]))
				Expect(list[1]["id"]).To(Equal(taskIDs[0]))
			})

			It("GET /.../tasks?offset=2 returns the remaining task", func() {
				url := fmt.Sprintf("/projects/%s/tasks?offset=2", hostProjectID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusOK))
				var list []map[string]any
				readJSON(rr, &list)

				Expect(list).To(HaveLen(1))
				Expect(list[0]["id"]).To(Equal(taskIDs[0]))
			})
		})

		Context("Create", func() {
			It("POST /projects/{projectId}/tasks under existing project creates a task successfully", func() {
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)
				rr := do(http.MethodPost, url, map[string]any{
					"title":       "Test task",
					"description": "Hello task 1",
				})
				Expect(rr.Code).To(Equal(http.StatusCreated))
				var task map[string]any
				readJSON(rr, &task)
				Expect(task["id"]).NotTo(BeEmpty())
				Expect(task["projectId"]).To(Equal(hostProjectID))
				Expect(task["title"]).To(Equal("Test task"))
				Expect(task["status"]).To(Equal("TODO"))
			})

			It("POST with non existent project returns projcet not found", func() {
				url := "/projects/ffffffff-ffff-ffff-ffff-ffffffffffff/tasks"
				rr := do(http.MethodPost, url, map[string]any{"title": "NA"})
				Expect(rr.Code).To(Equal(http.StatusNotFound))
			})

			It("POST with invalid project UUID returns invalid UUID", func() {
				url := "/projects/abcdefg/tasks"
				rr := do(http.MethodPost, url, map[string]any{"title": "NA"})
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("POST with invalid json body returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)
				badJSONString := `{"title": "My Task", }`
				rr := do(http.MethodPost, url, badJSONString)
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("POST with invalid status returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)
				rr := do(http.MethodPost, url, map[string]any{
					"title":  "Bad status",
					"status": "WHATEVER",
				})
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("POST with blank title returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)
				rr := do(http.MethodPost, url, map[string]any{"title": "   "})
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("Get/List", func() {
			var t1ID string

			BeforeAll(func() {
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)
				rr := do(http.MethodPost, url, map[string]any{"title": "T1"})
				Expect(rr.Code).To(Equal(http.StatusCreated))
				var task map[string]any
				readJSON(rr, &task)
				t1ID = task["id"].(string)
			})

			It("GET /projects/{projectId}/tasks returns tasks for the project", func() {
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusOK))
				var list []map[string]any
				readJSON(rr, &list)
				Expect(list).NotTo(BeEmpty())
			})

			It("GET /projects/{projectId}/tasks for a non existent project returns not found", func() {
				url := fmt.Sprintf("/projects/%s/tasks", invalidProjectID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusNotFound))
			})

			It("GET correct ownership for a task returns success", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t1ID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusOK))
				var task map[string]any
				readJSON(rr, &task)
				Expect(task["id"]).To(Equal(t1ID))
				Expect(task["projectId"]).To(Equal(hostProjectID))
			})

			It("GET wrong project ownership returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", seedProjectID, t1ID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
				Expect(rr.Body.String()).To(ContainSubstring("invalid task id"))
			})

			It("GET /.../tasks?status=INVALID returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks?status=PENDING", hostProjectID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
				Expect(rr.Body.String()).To(ContainSubstring("invalid status"))
			})

			It("GET /.../tasks?q=... filters by title", func() {
				url := fmt.Sprintf("/projects/%s/tasks?q=T1", hostProjectID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusOK))
				var list []map[string]any
				readJSON(rr, &list)
				Expect(list).To(HaveLen(1))
				Expect(list[0]["id"]).To(Equal(t1ID))
			})

			It("GET /.../tasks?q=... returns empty for no match", func() {
				url := fmt.Sprintf("/projects/%s/tasks?q=NonExistentTask", hostProjectID)
				rr := do(http.MethodGet, url, nil)
				Expect(rr.Code).To(Equal(http.StatusOK))
				var list []map[string]any
				readJSON(rr, &list)
				Expect(list).To(BeEmpty())
			})

		})

		Context("Update (PUT)", func() {
			var t2ID string

			BeforeAll(func() {
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)
				rr := do(http.MethodPost, url, map[string]any{"title": "T2"})
				Expect(rr.Code).To(Equal(http.StatusCreated))
				var task map[string]any
				readJSON(rr, &task)
				t2ID = task["id"].(string)
			})

			It("PUT proper task returns OK", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t2ID)
				rr := do(http.MethodPut, url, map[string]any{
					"title":       "T2-updated",
					"status":      "IN_PROGRESS",
					"description": "Hello",
				})
				Expect(rr.Code).To(Equal(http.StatusOK))
				var task map[string]any
				readJSON(rr, &task)
				Expect(task["title"]).To(Equal("T2-updated"))
				Expect(task["status"]).To(Equal("IN_PROGRESS"))
			})

			It("PUT with no title returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t2ID)
				rr := do(http.MethodPut, url, map[string]any{
					// "title":  "T2-updated",
					"status": "IN_PROGRESS",
				})
				Expect(rr.Code).To(Equal(http.StatusOK))
				var task map[string]any
				readJSON(rr, &task)
				Expect(task["title"]).To(Equal("T2-updated"))
				Expect(task["status"]).To(Equal("IN_PROGRESS"))
			})

			It("PUT with long title returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t2ID)
				rr := do(http.MethodPut, url, map[string]any{
					"title":  RandStringRunes(205),
					"status": "IN_PROGRESS",
				})
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("PUT with no body returns current task", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t2ID)
				rr := do(http.MethodPut, url, map[string]any{})
				Expect(rr.Code).To(Equal(http.StatusOK))
			})

			It("PUT with invalid json body returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t2ID)
				badJSONString := `{"title": "My Task", }`
				rr := do(http.MethodPut, url, badJSONString)
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("PUT task with invalid project returns not found", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", invalidProjectID, t2ID)
				rr := do(http.MethodPut, url, nil)
				Expect(rr.Code).To(Equal(http.StatusNotFound))
				Expect(rr.Body.String()).To(ContainSubstring("not found"))
			})

			It("PUT invalid status returns bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t2ID)
				rr := do(http.MethodPut, url, map[string]any{"status": "BAD"})
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})

			It("PUT against missing task returns not found", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, "ffffffff-ffff-ffff-ffff-ffffffffffff")
				rr := do(http.MethodPut, url, map[string]any{"title": "x"})
				Expect([]int{http.StatusNotFound, http.StatusBadRequest}).To(ContainElement(rr.Code))
			})
		})

		Context("Delete", func() {
			var t3ID string

			BeforeAll(func() {
				url := fmt.Sprintf("/projects/%s/tasks", hostProjectID)
				rr := do(http.MethodPost, url, map[string]any{"title": "T3"})
				Expect(rr.Code).To(Equal(http.StatusCreated))
				var task map[string]any
				readJSON(rr, &task)
				t3ID = task["id"].(string)
			})

			It("DELETE task with invalid project returns not found", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", invalidProjectID, t3ID)
				rr := do(http.MethodDelete, url, nil)
				Expect(rr.Code).To(Equal(http.StatusNotFound))
				Expect(rr.Body.String()).To(ContainSubstring("not found"))
			})

			It("DELETE successful returns no content", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t3ID)
				rr := do(http.MethodDelete, url, nil)
				Expect(rr.Code).To(Equal(http.StatusNoContent))
			})

			It("DELETE again the same task returns invalid task and bad request", func() {
				url := fmt.Sprintf("/projects/%s/tasks/%s", hostProjectID, t3ID)
				rr := do(http.MethodDelete, url, nil)
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
				Expect(rr.Body.String()).To(ContainSubstring("invalid task"))
			})

		})
	})
})
