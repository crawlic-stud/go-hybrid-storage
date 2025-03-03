package utils

import (
	"fmt"
	"strings"
	"sync"
)

type SQLQueryType string

const (
	SQLite     SQLQueryType = "sqlite"
	PostgreSQL SQLQueryType = "postgresql"
)

var (
	cache     map[string]interface{} = make(map[string]any)
	cacheLock sync.RWMutex
)

type Query struct {
	Type SQLQueryType
}

func replaceQueryPlaceholdersPostgreSQL(query string) string {
	maxIterations := strings.Count(query, "?")
	currentIteration := 1

	for {
		if currentIteration > maxIterations {
			break
		}
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", currentIteration), 1)
		currentIteration++
	}

	return query
}

func (q Query) transformQueryByType(query string) string {
	switch q.Type {
	case SQLite:
		return query
	case PostgreSQL:
		query = replaceQueryPlaceholdersPostgreSQL(query)
		return query
	default:
		panic(fmt.Errorf("unknown query type: %s", q.Type))
	}
}

func (q Query) GetCachedQuery(query string) string {
	cacheKey := fmt.Sprintf("%s:%s", q.Type, query)

	cacheLock.RLock()
	result, ok := cache[cacheKey]
	cacheLock.RUnlock()
	if !ok {
		result = q.transformQueryByType(query)
		cacheLock.Lock()
		cache[cacheKey] = result
		cacheLock.Unlock()
	}

	return result.(string)
}
