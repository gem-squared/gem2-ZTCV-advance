package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/caller"
	"github.com/gem-squared/gem2-ZTCV/internal/chain"
	"github.com/gem-squared/gem2-ZTCV/internal/didregistry"
	"github.com/gem-squared/gem2-ZTCV/internal/identity"
	"github.com/gem-squared/gem2-ZTCV/internal/intent"
	"github.com/gem-squared/gem2-ZTCV/internal/llm"
	"github.com/gem-squared/gem2-ZTCV/internal/passport"
	"github.com/gem-squared/gem2-ZTCV/internal/risk"
	"github.com/gem-squared/gem2-ZTCV/internal/risk/layer1"
	"github.com/gem-squared/gem2-ZTCV/internal/risk/layer2"
	"github.com/gem-squared/gem2-ZTCV/internal/server"
	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// Service bundles all collaborators session-svc owns. This is what
// main.go constructs and wires routes onto.
type Service struct {
	Repo     *Repo
	DIDs     *didregistry.Repo
	Identity identity.IdentityProvider
	Chain    chain.ChainAnchor
	Layer2   *layer2.Mock
	Intent   intent.Generator // WP-ZTCV-06 — Step 7 Predictive Disclosure
	BlockBus *EventBus        // SSE event broadcaster
}

// NewService composes the service from already-constructed deps.
//
// The Intent generator is wired with an LLM client derived from env
// (ANTHROPIC_API_KEY / LLM_PROVIDER / LLM_MODEL / LLM_TIMEOUT_MS). If
// no API key is configured, the deterministic fallback path is used
// unconditionally — the demo never blocks on LLM availability.
func NewService(repo *Repo, dids *didregistry.Repo, idp identity.IdentityProvider, anchor chain.ChainAnchor) *Service {
	llmClient := llm.NewClient(llm.LoadFactoryConfig())
	return &Service{
		Repo:     repo,
		DIDs:     dids,
		Identity: idp,
		Chain:    anchor,
		Layer2:   layer2.NewMock(),
		Intent:   intent.New(llmClient),
		BlockBus: NewEventBus(),
	}
}

// RegisterRoutes wires all session-svc HTTP endpoints on the builder.
func (s *Service) RegisterRoutes(b *server.Builder) {
	b.Route("/api/session/create", s.handleCreate)
	b.Route("/api/session/", s.handleSessionDispatch)
	b.Route("/api/scenarios/run", s.handleRunScenario)
	b.Route("/api/did/", s.handleDIDDispatch)
}

// ─── Handlers ───

// handleCreate creates a fresh session.
func (s *Service) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		server.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
		return
	}
	now := time.Now().UTC()
	id := "sess_" + randHex(8)
	nonce := randHex(16)
	sess := Create(id, nonce, now)
	if err := s.Repo.Save(sess); err != nil {
		server.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	server.WriteJSON(w, http.StatusCreated, sess)
}

// handleSessionDispatch routes /api/session/{id}/* sub-paths.
func (s *Service) handleSessionDispatch(w http.ResponseWriter, r *http.Request) {
	// Strip prefix
	path := strings.TrimPrefix(r.URL.Path, "/api/session/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		server.WriteError(w, http.StatusNotFound, "not_found", "session id missing")
		return
	}
	sessionID := parts[0]
	sub := ""
	if len(parts) == 2 {
		sub = parts[1]
	}
	switch {
	case sub == "" && r.Method == http.MethodGet:
		s.handleGetSession(w, r, sessionID)
	case sub == "caller-proof" && r.Method == http.MethodPost:
		s.handleCallerProof(w, r, sessionID)
	case sub == "customer-proof" && r.Method == http.MethodPost:
		s.handleCustomerProof(w, r, sessionID)
	case sub == "passport" && r.Method == http.MethodGet:
		s.handleGetPassport(w, r, sessionID)
	case sub == "events" && r.Method == http.MethodGet:
		s.handleSSE(w, r, sessionID)
	default:
		server.WriteError(w, http.StatusNotFound, "not_found", "no route for "+r.Method+" "+r.URL.Path)
	}
}

// handleDIDDispatch routes /api/did/{did}.
func (s *Service) handleDIDDispatch(w http.ResponseWriter, r *http.Request) {
	did := strings.TrimPrefix(r.URL.Path, "/api/did/")
	if did == "" {
		server.WriteError(w, http.StatusNotFound, "not_found", "did missing")
		return
	}
	doc, err := s.DIDs.Resolve(did)
	if err != nil {
		server.WriteError(w, http.StatusNotFound, "unknown_did", err.Error())
		return
	}
	server.WriteJSON(w, http.StatusOK, doc)
}

func (s *Service) handleGetSession(w http.ResponseWriter, r *http.Request, id string) {
	sess, err := s.Repo.Load(id)
	if err != nil {
		server.WriteError(w, http.StatusNotFound, "session_not_found", err.Error())
		return
	}
	server.WriteJSON(w, http.StatusOK, sess)
}

func (s *Service) handleCallerProof(w http.ResponseWriter, r *http.Request, id string) {
	sess, err := s.Repo.Load(id)
	if err != nil {
		server.WriteError(w, http.StatusNotFound, "session_not_found", err.Error())
		return
	}
	var proof types.CallerProof
	if err := json.NewDecoder(r.Body).Decode(&proof); err != nil {
		server.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	now := time.Now().UTC()
	res := caller.Verify(s.DIDs, &proof, id, sess.Nonce, now)
	if !res.OK {
		// Block the session — write a BLOCK passport anchored.
		_, _ = s.runRiskAndAnchor(sess, &proof, nil, "", string(res.Reason), now)
		server.WriteError(w, http.StatusForbidden, string(res.Reason), explainReason(res))
		return
	}
	newSess, err := SubmitCallerProof(sess, &proof, now)
	if err != nil {
		server.WriteError(w, http.StatusConflict, "invalid_transition", err.Error())
		return
	}
	if err := s.Repo.Save(newSess); err != nil {
		server.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	s.BlockBus.Publish(id, sseEvent{Event: "state", Data: jsonString(newSess)})
	server.WriteJSON(w, http.StatusOK, newSess)
}

func (s *Service) handleCustomerProof(w http.ResponseWriter, r *http.Request, id string) {
	sess, err := s.Repo.Load(id)
	if err != nil {
		server.WriteError(w, http.StatusNotFound, "session_not_found", err.Error())
		return
	}
	var body struct {
		OACXToken string `json:"oacx_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		server.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	claims, err := s.Identity.VerifyToken(body.OACXToken)
	if err != nil {
		server.WriteError(w, http.StatusUnauthorized, "invalid_token", err.Error())
		return
	}
	now := time.Now().UTC()
	cp := &types.CustomerProof{
		OACXToken:  body.OACXToken,
		Claims:     claims,
		SessionID:  id,
		ReceivedAt: now,
	}
	newSess, err := SubmitCustomerProof(sess, cp, now)
	if err != nil {
		server.WriteError(w, http.StatusConflict, "invalid_transition", err.Error())
		return
	}
	if err := s.Repo.Save(newSess); err != nil {
		server.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	s.BlockBus.Publish(id, sseEvent{Event: "state", Data: jsonString(newSess)})

	// After both proofs are in, run risk + anchor inline.
	final, ferr := s.runRiskAndAnchor(newSess, newSess.CallerProof, newSess.CustomerProof, "", "", now)
	if ferr != nil {
		server.WriteError(w, http.StatusInternalServerError, "orchestration_error", ferr.Error())
		return
	}
	server.WriteJSON(w, http.StatusOK, final)
}

func (s *Service) handleGetPassport(w http.ResponseWriter, r *http.Request, id string) {
	sess, err := s.Repo.Load(id)
	if err != nil {
		server.WriteError(w, http.StatusNotFound, "session_not_found", err.Error())
		return
	}
	trace, _ := s.Repo.LoadTrace(id)
	manifest, manifestHash := s.runIntentHandshake(r.Context(), sess)
	cp := passport.Build(passport.Input{
		Session:            sess,
		Trace:              trace,
		ReceiptTxHash:      sess.TxHash,
		ExplorerURL:        explorerURL(sess.TxHash),
		CallerOrgName:      callerOrgName(sess),
		BlockReason:        sess.BlockReason,
		IntentManifest:     manifest,
		IntentManifestHash: manifestHash,
	})
	server.WriteJSON(w, http.StatusOK, cp)
}

// runIntentHandshake executes Step 7 of the 9-step pipeline. Always
// returns a usable (manifest, hash) pair — the manifest's Source field
// distinguishes "live" (LLM call succeeded) from "fallback"
// (deterministic per-scenario script). Never returns an error so a
// caller can ignore failure modes; the demo never breaks on LLM
// unavailability.
func (s *Service) runIntentHandshake(ctx context.Context, sess *types.CallSession) (*types.IntentManifest, string) {
	if s.Intent == nil || sess == nil {
		return nil, ""
	}
	in := intent.Input{
		OrgDisplayName: callerOrgName(sess),
		Outcome:        deriveIntentOutcome(sess),
	}
	if sess.CallerProof != nil {
		in.CallerDID = sess.CallerProof.CallerDID
		in.Purpose = sess.CallerProof.Purpose
	}
	if sess.RiskVerdict != nil && sess.RiskVerdict.Layer2.RiskScore != 0 {
		in.RiskScore = sess.RiskVerdict.Layer2.RiskScore
	}
	m := s.Intent.Generate(ctx, in)
	return &m, intent.Hash(m)
}

// deriveIntentOutcome maps the terminal session state + block reason
// onto the intent.Outcome enum that selects the deterministic
// fallback. The mapping is conservative: any non-verified, non-mapped
// state degrades to the unknown-DID fallback (which warns the receiver
// most strongly).
func deriveIntentOutcome(sess *types.CallSession) intent.Outcome {
	switch sess.State {
	case types.StateVerified:
		return intent.OutcomeSafe
	case types.StateBlocked:
		br := strings.ToLower(sess.BlockReason)
		switch {
		case strings.Contains(br, "unauthorized_purpose") || strings.Contains(br, "권한 외"):
			return intent.OutcomeUnauthorizedScope
		case strings.Contains(br, "unknown_did") || strings.Contains(br, "did 없음"):
			return intent.OutcomeUnknownDID
		default:
			return intent.OutcomeUnknownDID
		}
	default:
		return intent.OutcomeUnknownDID
	}
}

// handleSSE streams state-change events for a session.
func (s *Service) handleSSE(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		server.WriteError(w, http.StatusInternalServerError, "no_flusher", "streaming unsupported")
		return
	}
	ch := s.BlockBus.Subscribe(id)
	defer s.BlockBus.Unsubscribe(id, ch)
	// Send one snapshot immediately so the client renders right away.
	if sess, err := s.Repo.Load(id); err == nil {
		writeSSE(w, "state", jsonString(sess))
		flusher.Flush()
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			writeSSE(w, ev.Event, ev.Data)
			flusher.Flush()
		}
	}
}

// ─── Risk + Chain orchestration (in-process for simulation) ───

// runRiskAndAnchor produces the composed verdict + anchors a receipt
// (either SAFE or BLOCK). Returns the (potentially terminal) updated
// session. blockOverride is used when caller-proof fails and we want
// to anchor a BLOCK with a specific reason without a customer proof.
func (s *Service) runRiskAndAnchor(sess *types.CallSession, callerP *types.CallerProof, custP *types.CustomerProof, callScript, blockOverride string, now time.Time) (*types.CallSession, error) {
	// Determine purpose-scope status for Layer 1 Rule 4
	var authorizedPurposes []string
	isAuth := false
	if callerP != nil {
		if ok, auth, _ := s.DIDs.IsAuthorized(callerP.CallerDID, callerP.Purpose, now); auth != nil {
			authorizedPurposes = auth.AllowedPurposes
			isAuth = ok
		}
	}
	purpose := ""
	callerDID := ""
	orgDIDClaim := ""
	if callerP != nil {
		purpose = callerP.Purpose
		callerDID = callerP.CallerDID
		orgDIDClaim = callerP.OrgDID
	}

	// Layer 1 (deterministic)
	l1 := layer1.Classify(layer1.Input{
		CallerDID:         callerDID,
		OrgDIDClaim:       orgDIDClaim,
		Purpose:           purpose,
		AuthorizedPurpose: authorizedPurposes,
		IsAuthorized:      isAuth,
		CallScript:        callScript,
	})
	// Layer 2 (mock LLM)
	l2 := s.Layer2.Evaluate(layer2.Input{
		CallerDID:         callerDID,
		OrgDID:            orgDIDClaim,
		Purpose:           purpose,
		IsAuthorized:      isAuth,
		AuthorizedPurpose: authorizedPurposes,
		CallScript:        callScript,
		MobileIDVerified:  custP != nil,
	})
	composed := risk.Compose(l1, l2)

	// Apply explicit blockOverride (e.g., unknown_did from caller-proof)
	finalIsSafe := composed.Final == types.RiskLOW
	blockReason := blockOverride
	if blockReason == "" && !finalIsSafe {
		blockReason = "composed verdict=" + string(composed.Final)
	}

	// Always attach composed verdict to the session (regardless of
	// state-transition eligibility). This ensures the passport builder
	// can see Layer1 triggered_rules even on early-block paths where
	// state hasn't reached customer_proved.
	{
		cp := *sess
		cp.RiskVerdict = composed
		sess = &cp
		_ = s.Repo.Save(sess)
	}

	// Risk-checked transition (if eligible — happy path)
	if sess.State == types.StateCustomerProved {
		updated, err := MarkRiskChecked(sess, composed, now)
		if err == nil {
			sess = updated
			_ = s.Repo.Save(sess)
		}
	}

	// Build OnChain receipt + anchor
	onChain := types.ReceiptOnChain{
		Timestamp:     now.Unix(),
		IsSafe:        finalIsSafe,
		PolicyVersion: composed.PolicyVersion,
	}
	sessionHash, receiptHash := chain.MakeHashes(sess.ID, sess.Nonce, onChain)
	onChain.SessionHash = sessionHash
	onChain.ReceiptHash = receiptHash
	txHash, explorerURL, err := s.Chain.AnchorReceipt(sessionHash, receiptHash, finalIsSafe, composed.PolicyVersion)
	if err != nil {
		return nil, err
	}
	receipt := &types.Receipt{
		OnChain: onChain,
		OffChain: types.ReceiptOffChain{
			CallerDID:        callerDID,
			OrgDID:           orgDIDClaim,
			MobileIDVerified: custP != nil,
			CallerVerified:   callerP != nil && blockOverride == "",
			AIRiskVerdict:    composed.Final,
			FinalDecision:    map[bool]string{true: "SAFE", false: "BLOCK"}[finalIsSafe],
			BlockReason:      blockReason,
			ComposedVerdict:  composed,
			AnchoredTxHash:   txHash,
			ExplorerURL:      explorerURL,
			GeneratedAt:      now,
		},
	}

	// Transition to anchored → verified or blocked
	if sess.State == types.StateRiskChecked {
		anchored, err := MarkAnchored(sess, receipt, txHash, now)
		if err == nil {
			sess = anchored
			_ = s.Repo.Save(sess)
		}
	}
	if finalIsSafe && sess.State == types.StateAnchored {
		v, err := MarkVerified(sess, now)
		if err == nil {
			sess = v
		}
	} else {
		// BLOCK path — may bypass risk_checked when called for caller-proof failure
		b, _ := MarkBlocked(sess, blockReason, receipt, txHash, now)
		if b != nil {
			sess = b
		}
	}
	_ = s.Repo.Save(sess)

	// Build + persist a simple TrustTrace for the SSE timeline
	trace := buildTrustTrace(sess, composed, blockReason)
	_ = s.Repo.SaveTrace(trace)
	s.BlockBus.Publish(sess.ID, sseEvent{Event: "state", Data: jsonString(sess)})

	return sess, nil
}

func buildTrustTrace(sess *types.CallSession, composed *types.ComposedVerdict, blockReason string) *types.TrustTrace {
	now := time.Now().UTC()
	t := &types.TrustTrace{
		SessionID: sess.ID,
		StartedAt: sess.CreatedAt,
		EndedAt:   now,
	}
	addGate := func(name string, v types.GateVerdict, reasons ...string) {
		t.Gates = append(t.Gates, types.GateResult{
			Gate:        name,
			Verdict:     v,
			Reasons:     reasons,
			StartedAt:   sess.CreatedAt,
			CompletedAt: now,
		})
	}
	addGate("L0", types.GatePASS, "input shape OK")
	if sess.CallerProof != nil {
		if blockReason == string(caller.ReasonUnknownDID) || blockReason == string(caller.ReasonUnauthorizedPurpose) {
			addGate("L1", types.GateFAIL, blockReason)
		} else {
			addGate("L1", types.GatePASS, "DID resolved + purpose authorized")
		}
	} else {
		addGate("L1", types.GateFAIL, "no caller proof")
	}
	if composed != nil {
		v := types.GatePASS
		if composed.Final == types.RiskBLOCK {
			v = types.GateFAIL
		}
		reasons := []string{"L1=" + string(composed.Layer1.Verdict), "L2=" + string(composed.Layer2.Verdict)}
		addGate("F", v, reasons...)
		addGate("L2", v, "policy="+composed.PolicyVersion)
	}
	addGate("L3", types.GatePASS, "explanation scrubbed")
	if sess.State == types.StateVerified {
		t.Final = types.GatePASS
	} else {
		t.Final = types.GateFAIL
	}
	return t
}

// ─── Scenario runner ───

// handleRunScenario is the one-call demo runner. POST with ?n=1|2|3.
// Returns the final session + CallPassport in one response.
func (s *Service) handleRunScenario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		server.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
		return
	}
	n := r.URL.Query().Get("n")
	now := time.Now().UTC()
	sess := Create("sess_demo_"+n+"_"+randHex(4), randHex(16), now)
	if err := s.Repo.Save(sess); err != nil {
		server.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	var (
		callerP    *types.CallerProof
		callScript string
		blockSeed  string
	)
	switch n {
	case "1": // Verified KakaoBank AI loan counselor → SAFE
		callerP = caller.MakeProof(
			"did:opendid:agent:kakaobank-ai-loan-counselor-001",
			"did:opendid:org:kakaobank",
			"loan_consultation",
			sess.ID, sess.Nonce,
		)
		callScript = "고객님 안녕하세요, KakaoBank 대출 상담사입니다. 신청하신 대출 한도 안내드리겠습니다."

	case "2": // Fake prosecutor / unknown DID + transfer demand → BLOCK
		callerP = caller.MakeProof(
			"did:opendid:agent:unknown-prosecutor-fake",
			"did:opendid:org:fake-prosecutor",
			"investigation",
			sess.ID, sess.Nonce,
		)
		callScript = "검찰청 수사관입니다. 안전계좌로 즉시 송금 안 하면 체포 됩니다."
		blockSeed = string(caller.ReasonUnknownDID)

	case "3": // security_alert-only agent attempts loan_consultation → BLOCK
		callerP = caller.MakeProof(
			"did:opendid:agent:kakaobank-ai-security-alert-007",
			"did:opendid:org:kakaobank",
			"loan_consultation",
			sess.ID, sess.Nonce,
		)
		callScript = "안녕하세요, 카카오뱅크 보안 알림 상담사입니다. 신규 대출 상품 안내드리고자 합니다."
		blockSeed = string(caller.ReasonUnauthorizedPurpose)

	default:
		server.WriteError(w, http.StatusBadRequest, "bad_request", "n must be 1, 2, or 3")
		return
	}

	// Submit caller-proof (this will resolve+verify; on failure, runRiskAndAnchor with blockSeed)
	verResult := caller.Verify(s.DIDs, callerP, sess.ID, sess.Nonce, now)
	if !verResult.OK {
		final, _ := s.runRiskAndAnchor(sess, callerP, nil, callScript, string(verResult.Reason), now)
		serveScenarioResult(w, r, s, final)
		return
	}
	// caller_proved
	updated, _ := SubmitCallerProof(sess, callerP, now)
	_ = s.Repo.Save(updated)
	sess = updated

	// Submit customer-proof (mock OACX token for customer-00n where n matches scenario)
	customerID := "customer-00" + n
	if n == "2" {
		customerID = "customer-001"
	}
	token := identity.MockToken(customerID, sess.ID)
	claims, _ := s.Identity.VerifyToken(token)
	cp := &types.CustomerProof{OACXToken: token, Claims: claims, SessionID: sess.ID, ReceivedAt: now}
	updated, _ = SubmitCustomerProof(sess, cp, now)
	_ = s.Repo.Save(updated)
	sess = updated

	// Risk + anchor (use returned final session — local `sess` is stale)
	final, ferr := s.runRiskAndAnchor(sess, callerP, cp, callScript, blockSeed, now)
	if ferr != nil {
		server.WriteError(w, http.StatusInternalServerError, "orchestration_error", ferr.Error())
		return
	}
	serveScenarioResult(w, r, s, final)
}

func serveScenarioResult(w http.ResponseWriter, r *http.Request, s *Service, sess *types.CallSession) {
	// Step 7 of the 9-step pipeline — Intent Handshake. The Generate
	// call always returns a Manifest (live or fallback); never errors.
	manifest, manifestHash := s.runIntentHandshake(r.Context(), sess)
	// Persist the off-chain receipt hash too, so the SQLite audit
	// trail records what the passport surfaced (best-effort).
	if sess.Receipt != nil && manifestHash != "" {
		sess.Receipt.OffChain.IntentManifestHash = manifestHash
		_ = s.Repo.Save(sess)
	}
	cp := passport.Build(passport.Input{
		Session:            sess,
		ReceiptTxHash:      sess.TxHash,
		ExplorerURL:        explorerURL(sess.TxHash),
		CallerOrgName:      callerOrgName(sess),
		BlockReason:        sess.BlockReason,
		IntentManifest:     manifest,
		IntentManifestHash: manifestHash,
	})
	server.WriteJSON(w, http.StatusOK, map[string]any{
		"session":  sess,
		"passport": cp,
	})
}

// ─── Helpers ───

func randHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func jsonString(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func writeSSE(w http.ResponseWriter, event, data string) {
	_, _ = w.Write([]byte("event: " + event + "\n"))
	for _, line := range strings.Split(data, "\n") {
		_, _ = w.Write([]byte("data: " + line + "\n"))
	}
	_, _ = w.Write([]byte("\n"))
}

func explorerURL(txHash string) string {
	if txHash == "" {
		return ""
	}
	return "https://sepolia.etherscan.io/tx/" + txHash
}

func callerOrgName(sess *types.CallSession) string {
	if sess.CallerProof == nil {
		return ""
	}
	// Phase 1 mapping — could be looked up from DID Document if needed
	switch sess.CallerProof.OrgDID {
	case "did:opendid:org:kakaobank":
		return "카카오뱅크"
	}
	return sess.CallerProof.OrgDID
}

func explainReason(res caller.Result) string {
	switch res.Reason {
	case caller.ReasonUnknownDID:
		return "발신자 DID가 시스템에 등록되어 있지 않습니다."
	case caller.ReasonInvalidSignature:
		return "발신자 서명을 검증할 수 없습니다."
	case caller.ReasonExpiredNonce:
		return "세션 nonce가 일치하지 않거나 만료되었습니다."
	case caller.ReasonUnauthorizedPurpose:
		return "발신자가 요청한 통화 목적에 대한 권한이 없습니다. (보유 권한: " + strings.Join(res.AuthScope, ",") + ")"
	}
	return "알 수 없는 사유로 검증에 실패했습니다."
}

// EventBus is a tiny per-session SSE broadcaster.
type EventBus struct {
	mu   sync.Mutex
	subs map[string]map[chan sseEvent]struct{}
}

type sseEvent struct {
	Event string
	Data  string
}

func NewEventBus() *EventBus {
	return &EventBus{subs: map[string]map[chan sseEvent]struct{}{}}
}

func (b *EventBus) Subscribe(id string) chan sseEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan sseEvent, 8)
	if b.subs[id] == nil {
		b.subs[id] = map[chan sseEvent]struct{}{}
	}
	b.subs[id][ch] = struct{}{}
	return ch
}

func (b *EventBus) Unsubscribe(id string, ch chan sseEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if m, ok := b.subs[id]; ok {
		delete(m, ch)
		close(ch)
		if len(m) == 0 {
			delete(b.subs, id)
		}
	}
}

func (b *EventBus) Publish(id string, ev sseEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs[id] {
		select {
		case ch <- ev:
		default:
		}
	}
}
