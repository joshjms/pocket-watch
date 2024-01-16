package isolate

import "strings"

type MetaFile struct {
	Pairs []string
}

func GetDetails(fileContent string) MetaFile {
	pairs := strings.Split(fileContent, "\n")
	return MetaFile{Pairs: pairs}
}

func (mf *MetaFile) Get(key string) string {
	for _, pair := range mf.Pairs {
		if strings.HasPrefix(pair, key) {
			return strings.Split(pair, ":")[1]
		}
	}
	return ""
}
