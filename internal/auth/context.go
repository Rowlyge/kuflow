package auth

import authcache "github.com/Rowlyge/kuflow/internal/auth/cache"

func APIKeyFromContext(v any) (*authcache.APIKey, bool) {
	key, ok := v.(*authcache.APIKey)
	return key, ok
}
