// Package services implements core business workflows.
package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/routing"
)

// GatewayService manages API gateway routes and reverse proxies.
type GatewayService struct {
	repo     ports.GatewayRepository
	proxyMu  sync.RWMutex
	proxies  map[uuid.UUID]*httputil.ReverseProxy
	routes   []*domain.GatewayRoute
	matchers map[uuid.UUID]*routing.PatternMatcher
	auditSvc ports.AuditService
}

// NewGatewayService constructs a GatewayService and loads existing routes.
func NewGatewayService(repo ports.GatewayRepository, auditSvc ports.AuditService) *GatewayService {
	s := &GatewayService{
		repo:     repo,
		proxies:  make(map[uuid.UUID]*httputil.ReverseProxy),
		routes:   make([]*domain.GatewayRoute, 0),
		matchers: make(map[uuid.UUID]*routing.PatternMatcher),
		auditSvc: auditSvc,
	}
	// Initial load
	_ = s.RefreshRoutes(context.Background())
	return s
}

func (s *GatewayService) CreateRoute(ctx context.Context, params ports.CreateRouteParams) (*domain.GatewayRoute, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	// Detect if it's a pattern or prefix
	patternType := "prefix"
	var paramNames []string
	if strings.Contains(params.Pattern, "{") || strings.Contains(params.Pattern, "*") {
		patternType = "pattern"
		matcher, err := routing.CompilePattern(params.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		paramNames = matcher.ParamNames
	}

	route := &domain.GatewayRoute{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        params.Name,
		PathPrefix:  params.Pattern, // Use pattern as prefix for backward compatibility where possible
		PathPattern: params.Pattern,
		PatternType: patternType,
		ParamNames:  paramNames,
		TargetURL:   params.Target,
		Methods:     params.Methods,
		StripPrefix: params.StripPrefix,
		RateLimit:   params.RateLimit,
		Priority:    params.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateRoute(ctx, route); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, route.UserID, "gateway.route_create", "gateway", route.ID.String(), map[string]interface{}{
		"name":    route.Name,
		"pattern": route.PathPattern,
		"methods": route.Methods,
	})

	_ = s.RefreshRoutes(ctx)
	return route, nil
}

func (s *GatewayService) ListRoutes(ctx context.Context) ([]*domain.GatewayRoute, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}
	return s.repo.ListRoutes(ctx, userID)
}

func (s *GatewayService) DeleteRoute(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	route, err := s.repo.GetRouteByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteRoute(ctx, id); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, route.UserID, "gateway.route_delete", "gateway", route.ID.String(), map[string]interface{}{
		"name": route.Name,
	})

	return s.RefreshRoutes(ctx)
}

func (s *GatewayService) RefreshRoutes(ctx context.Context) error {
	routes, err := s.repo.GetAllActiveRoutes(ctx)
	if err != nil {
		return err
	}

	newProxies := make(map[uuid.UUID]*httputil.ReverseProxy)
	newMatchers := make(map[uuid.UUID]*routing.PatternMatcher)

	for _, r := range routes {
		proxy, err := s.createReverseProxy(r)
		if err != nil {
			continue
		}

		newProxies[r.ID] = proxy
		if r.PatternType == "pattern" {
			matcher, err := routing.CompilePattern(r.PathPattern)
			if err == nil {
				newMatchers[r.ID] = matcher
			}
		}
	}

	s.sortRoutes(routes)

	s.proxyMu.Lock()
	s.proxies = newProxies
	s.routes = routes
	s.matchers = newMatchers
	s.proxyMu.Unlock()

	return nil
}

func (s *GatewayService) createReverseProxy(route *domain.GatewayRoute) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(route.TargetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		if route.StripPrefix {
			prefix := route.PathPrefix
			if route.PatternType == "pattern" {
				prefix = routing.GetLiteralPrefix(route.PathPattern)
			}
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/gw"+prefix)
			if !strings.HasPrefix(req.URL.Path, "/") {
				req.URL.Path = "/" + req.URL.Path
			}
		}
		originalDirector(req)
		req.Host = target.Host
	}

	return proxy, nil
}

func (s *GatewayService) sortRoutes(routes []*domain.GatewayRoute) {
	// Sort routes by specificity (longer literal prefixes and higher priority first)
	sort.Slice(routes, func(i, j int) bool {
		scoreI := calculateMatchScore(routes[i], "")
		scoreJ := calculateMatchScore(routes[j], "")
		return scoreI > scoreJ // Descending order
	})
}

// ProxyHandler is handled in the API layer for now

func (s *GatewayService) GetProxy(method, path string) (*httputil.ReverseProxy, map[string]string, bool) {
	s.proxyMu.RLock()
	defer s.proxyMu.RUnlock()

	var bestMatch *domain.RouteMatch

	for _, route := range s.routes {
		match := s.checkRouteMatch(route, method, path)
		if match != nil {
			if bestMatch == nil || match.MatchScore > bestMatch.MatchScore {
				bestMatch = match
			}
		}
	}

	if bestMatch != nil {
		return s.proxies[bestMatch.Route.ID], bestMatch.Params, true
	}

	return nil, nil, false
}

func (s *GatewayService) checkRouteMatch(route *domain.GatewayRoute, method, path string) *domain.RouteMatch {
	// 1. Method filter
	if !s.isMethodAllowed(route, method) {
		return nil
	}

	// 2. Path matching
	if route.PatternType == "pattern" {
		matcher, ok := s.matchers[route.ID]
		if ok {
			if params, ok := matcher.Match(path); ok {
				return &domain.RouteMatch{
					Route:      route,
					Params:     params,
					MatchScore: calculateMatchScore(route, path),
				}
			}
		}
	} else if strings.HasPrefix(path, route.PathPrefix) {
		return &domain.RouteMatch{
			Route:      route,
			Params:     nil,
			MatchScore: calculateMatchScore(route, path),
		}
	}

	return nil
}

func (s *GatewayService) isMethodAllowed(route *domain.GatewayRoute, method string) bool {
	if len(route.Methods) == 0 {
		return true
	}
	for _, m := range route.Methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func calculateMatchScore(route *domain.GatewayRoute, path string) int {
	// 1. Literal prefix length is a good indicator of specificity
	score := len(routing.GetLiteralPrefix(route.PathPattern))

	// 2. Bonus for exact matches (no parameters or wildcards)
	if !strings.ContainsAny(route.PathPattern, "{*") {
		score += 100
	}

	// 3. Priority is the strongest signal if provided
	if route.Priority > 0 {
		score += route.Priority * 1000
	}

	return score
}
