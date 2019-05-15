package cache

type Cache interface {
	Get(string) []byte
	Set(string, []byte) (string, []byte)
	Del(string) []byte
}
