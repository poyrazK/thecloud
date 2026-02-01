// Package services implements core business workflows.
package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
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

func (s *GatewayService) CreateRoute(ctx context.Context, name, pattern, target string, strip bool, rateLimit int) (*domain.GatewayRoute, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	// Detect if it's a pattern or prefix
	patternType := "prefix"
	var paramNames []string
	if strings.Contains(pattern, "{") || strings.Contains(pattern, "*") {
		patternType = "pattern"
		matcher, err := routing.CompilePattern(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		paramNames = matcher.ParamNames
	}

	route := &domain.GatewayRoute{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		PathPrefix:  pattern, // Use pattern as prefix for backward compatibility where possible
		PathPattern: pattern,
		PatternType: patternType,
		ParamNames:  paramNames,
		TargetURL:   target,
		StripPrefix: strip,
		RateLimit:   rateLimit,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateRoute(ctx, route); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, route.UserID, "gateway.route_create", "gateway", route.ID.String(), map[string]interface{}{
		"name":    route.Name,
		"pattern": route.PathPattern,
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
		target, err := url.Parse(r.TargetURL)
		if err != nil {
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		// Create a local copy of route for closure
		route := r

		// Custom director to handle prefix stripping if needed
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			if route.StripPrefix {
				// For prefix routes, we use PathPrefix
				// For pattern routes, we currently don't have a specific strip logic
				// other than what worked for prefix. Let's keep it simple for now.
				prefix := route.PathPrefix
				if route.PatternType == "pattern" {
					prefix = route.PathPattern
				}
				req.URL.Path = strings.TrimPrefix(req.URL.Path, "/gw"+prefix)
				if !strings.HasPrefix(req.URL.Path, "/") {
					req.URL.Path = "/" + req.URL.Path
				}
			}
			req.Host = target.Host
		}

		newProxies[r.ID] = proxy
		if r.PatternType == "pattern" {
			matcher, err := routing.CompilePattern(r.PathPattern)
			if err == nil {
				newMatchers[r.ID] = matcher
			}
		}
	}

	s.proxyMu.Lock()
	s.proxies = newProxies
	s.routes = routes
	s.matchers = newMatchers
	s.proxyMu.Unlock()

	return nil
}

// ProxyHandler is handled in the API layer for now

func (s *GatewayService) GetProxy(path string) (*httputil.ReverseProxy, map[string]string, bool) {
	s.proxyMu.RLock()
	defer s.proxyMu.RUnlock()

	var bestMatch *domain.RouteMatch

	for _, route := range s.routes {
		var match *domain.RouteMatch

		if route.PatternType == "pattern" {
			matcher, ok := s.matchers[route.ID]
			if ok {
				if params, ok := matcher.Match(path); ok {
					match = &domain.RouteMatch{
						Route:      route,
						Params:     params,
						MatchScore: calculateMatchScore(route, path),
					}
				}
			}
		} else {
			// Legacy prefix matching
			if strings.HasPrefix(path, route.PathPrefix) {
				match = &domain.RouteMatch{
					Route:      route,
					Params:     nil,
					MatchScore: len(route.PathPrefix),
				}
			}
		}

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

func calculateMatchScore(route *domain.GatewayRoute, path string) int {
	// Simple scoring: more characters in pattern = more specific
	// We can refine this later (exact > parameterized > wildcard)
	score := len(route.PathPattern)
	if route.Priority > 0 {
		score += route.Priority * 1000 // Priority boost
	}
	return score
}
