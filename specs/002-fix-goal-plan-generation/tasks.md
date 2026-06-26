---
description: "Task list for 修复学习目标生成学习计划功能"
---

# Tasks: 修复学习目标生成学习计划功能

**Input**: Design documents from `/specs/002-fix-goal-plan-generation/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/api.md

**Tests**: Include validation tasks for P1 user stories and measurable success criteria.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database migration and shared model definitions

- [x] T001 Run database migration to create `generation_sessions` and `generated_tasks` tables in server/migrations/002_add_generation_sessions.sql
- [x] T002 [P] Define `GenerationSession` model in server/internal/model/generation_session.go
- [x] T003 [P] Define `GeneratedTask` model in server/internal/model/generated_task.go
- [x] T004 [P] Create `GenerationSessionRepository` in server/internal/repository/generation_session.go
- [x] T005 [P] Create `GeneratedTaskRepository` in server/internal/repository/generated_task.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core AI response parsing and SSE infrastructure that both user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 Implement multi-layer JSON parser (pure JSON → markdown code block → JSON array extraction) in server/internal/ai/parser.go
- [x] T007 [P] Implement SSE stream helper functions in server/internal/ai/stream.go
- [x] T008 [P] Implement exponential backoff retry logic (1s/2s/4s, max 3 retries) in server/internal/ai/retry.go
- [x] T009 Add `GenerationSessionService` to manage session lifecycle in server/internal/service/generation_session.go
- [x] T010 [P] Create Zustand store for generation state management in web/src/stores/goal-store.ts
- [x] T011 [P] Create `useGoalStream` hook for SSE connection and state updates in web/src/hooks/use-goal-stream.ts

**Checkpoint**: Foundation ready - AI parsing, SSE, session management, and frontend state management are in place

---

## Phase 3: User Story 1 - 生成学习计划的即时反馈 (Priority: P1) 🎯 MVP

**Goal**: 用户点击"生成学习计划"后，系统提供完整的状态反馈流程（loading → 流式生成 → 预览确认 → 保存）

**Independent Test**: 创建学习目标后，观察从点击按钮到计划生成完成的完整状态反馈流程，验证每个阶段都有正确的 UI 反馈

### Tests for User Story 1

- [x] T012 [P] [US1] Contract test for `POST /api/learning-goals` endpoint in server/tests/integration/test_learning_goal_create.go
- [x] T013 [P] [US1] Contract test for `GET /api/learning-goals/:id/generate-stream` SSE endpoint in server/tests/integration/test_learning_goal_stream.go
- [x] T014 [P] [US1] Contract test for `POST /api/learning-goals/:id/tasks/confirm` endpoint in server/tests/integration/test_learning_goal_confirm.go
- [x] T015 [P] [US1] Integration test for complete generation flow in web/tests/integration/test-goal-generation.spec.tsx

### Implementation for User Story 1

**Backend**:
- [x] T016 [US1] Update `POST /api/learning-goals` handler to create session and return session_id in server/internal/handler/learning_goal.go
- [x] T017 [US1] Implement `GET /api/learning-goals/:id/generate-stream` SSE endpoint in server/internal/handler/learning_goal.go
- [x] T018 [US1] Update `LearningGoalService.GeneratePlan` to use session-based flow in server/internal/service/learning_goal.go
- [x] T019 [US1] Implement `POST /api/learning-goals/:id/tasks/confirm` endpoint in server/internal/handler/learning_goal.go
- [x] T020 [US1] Add task count progress events (event: progress) to SSE stream in server/internal/ai/stream.go

**Frontend**:
- [x] T021 [US1] Update `goals/page.tsx` to use `useGoalStream` hook and display all generation phases in web/src/app/(dashboard)/goals/page.tsx
- [x] T022 [US1] Refactor `StreamingPlanViewer` to support preview mode with confirm/regenerate buttons in web/src/components/goal/streaming-plan-viewer.tsx
- [x] T023 [US1] Add loading indicators and phase transitions (connecting → streaming → preview → done) in web/src/components/goal/streaming-plan-viewer.tsx
- [x] T024 [US1] Implement "Confirm Save" button that calls confirm API and shows success state in web/src/components/goal/streaming-plan-viewer.tsx
- [x] T025 [US1] Implement "Regenerate" button that clears current results and restarts generation in web/src/components/goal/streaming-plan-viewer.tsx
- [x] T026 [US1] Add task count display ("已生成 N 个任务...") during streaming phase in web/src/components/goal/streaming-plan-viewer.tsx
- [x] T027 [US1] Add smooth animations for task appearance during streaming in web/src/components/goal/streaming-plan-viewer.tsx

**Checkpoint**: User Story 1 complete - full generation flow with loading, streaming, preview, confirm, and error states

---

## Phase 4: User Story 2 - 修复 AI 响应解析错误 (Priority: P1)

**Goal**: 系统能够正确解析 AI 返回的各种格式（纯 JSON、markdown 代码块、带解释文本的 JSON），解析错误率 < 1%

**Independent Test**: 使用标准学习目标描述调用 API，验证能成功解析 markdown 代码块格式的 AI 响应

### Tests for User Story 2

- [x] T028 [P] [US2] Unit test for multi-layer JSON parser with various formats in server/tests/unit/test_ai_parser.go
- [ ] T029 [P] [US2] Integration test for AI response parsing with real AI service in server/tests/integration/test_learning_goal_parse.go
- [x] T030 [P] [US2] Unit test for error handling and retry logic in server/tests/unit/test_ai_retry.go

### Implementation for User Story 2

**Backend**:
- [x] T031 [US2] Enhance `parseAIResponse` function with robust multi-layer parsing in server/internal/ai/parser.go
- [x] T032 [US2] Implement `extractMarkdownCodeBlock` to handle ```json ... ``` format in server/internal/ai/parser.go
- [x] T033 [US2] Implement `findJSONArray` to extract JSON from mixed content in server/internal/ai/parser.go
- [x] T034 [US2] Add detailed error messages for each parsing failure type in server/internal/ai/parser.go
- [x] T035 [US2] Integrate retry logic with AI service calls in server/internal/service/learning_goal.go
- [x] T036 [US2] Add logging for AI response parsing attempts and failures in server/internal/ai/parser.go

**Frontend**:
- [x] T037 [US2] Display user-friendly error messages based on error codes (40003, 50001, 50002) in web/src/components/goal/streaming-plan-viewer.tsx
- [x] T038 [US2] Add retry button that allows users to retry generation without re-entering description in web/src/components/goal/streaming-plan-viewer.tsx

**Checkpoint**: User Story 2 complete - AI response parsing handles all common formats with < 1% error rate

---

## Phase 5: User Story 3 - 网络中断恢复 (Priority: P2)

**Goal**: 用户在生成过程中离开页面后返回，系统保留已生成内容并提供继续/重新生成选项

**Independent Test**: 在生成过程中刷新页面，验证系统显示已生成的部分任务并提供继续生成选项

### Tests for User Story 3

- [x] T039 [P] [US3] Unit test for localStorage save/load logic in web/tests/unit/test-goal-progress.ts
- [x] T040 [P] [US3] Integration test for session resumption flow in web/tests/integration/test-goal-resume.spec.tsx

### Implementation for User Story 3

**Frontend**:
- [x] T041 [US3] Implement `saveProgress` function to persist tasks to localStorage in web/src/lib/goal-progress.ts
- [x] T042 [US3] Implement `loadProgress` function to restore tasks from localStorage in web/src/lib/goal-progress.ts
- [x] T043 [US3] Add 24-hour expiration check for saved progress in web/src/lib/goal-progress.ts
- [x] T044 [US3] Update `goals/page.tsx` to check for saved progress on mount and display resume options in web/src/app/(dashboard)/goals/page.tsx
- [x] T045 [US3] Add "Continue Generation" button that reconnects to existing session in web/src/hooks/use-goal-stream.ts
- [x] T046 [US3] Add "Regenerate" button that clears saved progress and starts fresh in web/src/components/goal/streaming-plan-viewer.tsx
- [x] T047 [US3] Clear localStorage on successful confirm save in web/src/hooks/use-goal-stream.ts

**Backend**:
- [x] T048 [US3] Implement `POST /api/learning-goals/:id/regenerate` endpoint in server/internal/handler/learning_goal.go
- [x] T049 [US3] Add logic to clean up old sessions when regenerating in server/internal/service/generation_session.go

**Checkpoint**: User Story 3 complete - users can resume generation after page refresh or navigation

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T050 [P] Add comprehensive error logging throughout the generation flow in server/internal/handler/learning_goal.go
- [x] T051 [P] Add frontend error boundary for generation components in web/src/components/goal/error-boundary.tsx
- [x] T052 Implement scheduled job to clean up expired generation sessions (runs every hour) in server/cmd/cleanup/main.go
- [x] T053 [P] Add unit tests for Zustand store actions in web/tests/unit/test-goal-store.ts
- [x] T054 Update quickstart.md with actual test results and any new findings in specs/002-fix-goal-plan-generation/quickstart.md
- [x] T055 Run quickstart.md validation and document any issues in specs/002-fix-goal-plan-generation/quickstart.md

---

## Phase 7: Operability & Observability (Constitution Requirement)

**Purpose**: Health checks, monitoring, and observability mandated by Operability By Default principle

- [x] T056 [P] Implement health check endpoint for learning goal service (includes AI service status) in server/internal/handler/health.go
- [x] T057 [P] Add structured logging for AI service calls, response parsing, and retry attempts in server/internal/ai/client.go
- [x] T058 [P] Add generation flow monitoring: track session creation, task generation, completion rates in server/internal/service/generation_session.go
- [x] T059 [P] Add Redis caching for generation session state to support recovery in server/internal/repository/generation_session.go

---

## Phase 8: Performance & Quality Validation

**Purpose**: Performance tests and error scenario coverage mandated by success criteria

- [x] T060 [P] [US1] Performance test for first task display time (< 3 seconds target) in server/tests/performance/test_learning_goal_perf.go
- [x] T061 [P] [US2] Error code validation test for all error scenarios (40001, 40002, 40003, 50001, 50002) in server/tests/integration/test_learning_goal_errors.go
- [x] T062 [P] [US2] Error scenario coverage test (100% of error cases display user-friendly messages) in web/tests/integration/test-goal-error-scenarios.spec.tsx
- [x] T063 [P] Load test for concurrent generation sessions (target: 10 concurrent users) in server/tests/performance/test_learning_goal_load.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - US1 and US2 are both P1 and can proceed in parallel after Phase 2
  - US3 (P2) can proceed in parallel with US1/US2 or sequentially
- **Operability (Phase 7)**: Depends on Phase 2 completion, can run in parallel with User Stories
- **Performance & Quality (Phase 8)**: Depends on US1 and US2 completion
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - depends on session management, SSE, frontend state
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - depends on AI parser and retry logic
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - depends on session management and localStorage

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Backend endpoints before frontend integration
- Core implementation before error handling
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: T002-T005 can all run in parallel (different model files)
- **Phase 2**: T006-T011 can run in parallel (parser, SSE, retry, session service, frontend state)
- **Phase 3 (US1)**: 
  - T012-T015 (tests) can run in parallel
  - T016-T020 (backend) can run in parallel after tests
  - T021-T027 (frontend) can run in parallel after backend
- **Phase 4 (US2)**:
  - T028-T030 (tests) can run in parallel
  - T031-T036 (backend) can run in parallel after tests
  - T037-T038 (frontend) can run in parallel after backend
- **Phase 5 (US3)**:
  - T039-T040 (tests) can run in parallel
  - T041-T047 (frontend) can run in parallel after tests
  - T048-T049 (backend) can run in parallel with frontend
- **Phase 6**: T050-T055 can run in parallel where marked [P]
- **Phase 7 (Operability)**: T056-T059 can all run in parallel (health, logging, monitoring, caching)
- **Phase 8 (Performance)**: T060-T063 can run in parallel after US1/US2 complete

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Contract test for POST /api/learning-goals in server/tests/integration/test_learning_goal_create.go"
Task: "Contract test for GET /api/learning-goals/:id/generate-stream in server/tests/integration/test_learning_goal_stream.go"
Task: "Contract test for POST /api/learning-goals/:id/tasks/confirm in server/tests/integration/test_learning_goal_confirm.go"
Task: "Integration test for complete generation flow in web/tests/integration/test-goal-generation.spec.tsx"

# After tests, launch backend tasks in parallel:
Task: "Update POST /api/learning-goals handler in server/internal/handler/learning_goal.go"
Task: "Implement GET /api/learning-goals/:id/generate-stream in server/internal/handler/learning_goal.go"
Task: "Update LearningGoalService.GeneratePlan in server/internal/service/learning_goal.go"
Task: "Implement POST /api/learning-goals/:id/tasks/confirm in server/internal/handler/learning_goal.go"
Task: "Add task count progress events in server/internal/ai/stream.go"

# After backend, launch frontend tasks in parallel:
Task: "Update goals/page.tsx to use useGoalStream hook in web/src/app/(dashboard)/goals/page.tsx"
Task: "Refactor StreamingPlanViewer for preview mode in web/src/components/goal/streaming-plan-viewer.tsx"
Task: "Add loading indicators and phase transitions in web/src/components/goal/streaming-plan-viewer.tsx"
Task: "Implement Confirm Save button in web/src/components/goal/streaming-plan-viewer.tsx"
Task: "Implement Regenerate button in web/src/components/goal/streaming-plan-viewer.tsx"
Task: "Add task count display in web/src/components/goal/streaming-plan-viewer.tsx"
Task: "Add smooth animations in web/src/components/goal/streaming-plan-viewer.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 2 Only)

1. Complete Phase 1: Setup (database migration)
2. Complete Phase 2: Foundational (AI parser, SSE, session management, frontend state)
3. Complete Phase 3: User Story 1 (full generation flow)
4. Complete Phase 4: User Story 2 (AI parsing fixes)
5. Complete Phase 7: Operability (health checks, monitoring, logging)
6. **STOP and VALIDATE**: Test complete generation flow end-to-end
7. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test generation flow → Deploy/Demo (MVP!)
3. Add User Story 2 → Test AI parsing → Deploy/Demo
4. Add Operability → Health checks and monitoring in place → Deploy/Demo
5. Add User Story 3 → Test resume flow → Deploy/Demo
6. Add Performance & Quality → Validate success criteria → Final release
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (frontend + backend)
   - Developer B: User Story 2 (AI parsing + error handling)
   - Developer C: Operability (health checks, monitoring, logging)
3. After US1 and US2 complete:
   - Developer A or B: User Story 3 (resume flow)
   - Developer C: Performance & Quality validation
4. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- US1 and US2 are both P1 - prioritize US1 for immediate user value
- US3 is P2 - can be deferred if time constraints exist
- Phase 7 (Operability) is required by Constitution - must complete before production deployment
- Phase 8 (Performance) validates success criteria SC-001, SC-003, SC-004
