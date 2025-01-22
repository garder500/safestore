package utils

import "fmt"

func GeneratePaths(data interface{}, currentPath string, paths *[]map[string]interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newPath := currentPath
			if newPath != "" {
				newPath += "."
			}
			newPath += key
			// if the value is a map of interface we need to call the function recursively to get all the paths, and we don't add it to the paths
			if _, ok := value.(map[string]interface{}); !ok {
				*paths = append(*paths, map[string]interface{}{"path": newPath, "value": value})
			} else {
				GeneratePaths(value, newPath, paths)
			}
		}
	case []interface{}:
		for i, value := range v {
			newPath := fmt.Sprintf("%s[%d]", currentPath, i)
			*paths = append(*paths, map[string]interface{}{"path": newPath, "value": value})
			GeneratePaths(value, newPath, paths)
		}
	}
}
