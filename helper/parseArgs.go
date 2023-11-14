package helper

import "strings"

// Parse ssh command line arguments from string list into map
func ParseArgs(args []string) (map[string]string, error) {

	var parsedArgs = make(map[string]string, len(args))

	for i := 0; i < len(args); i++ {
		if strings.Contains(args[i], "=") {
			parsedArgs[args[i][:strings.Index(args[i], "=")]] = args[i][strings.Index(args[i], "=")+1:]
		} else {
			parsedArgs[args[i]] = ""
		}
	}

	return parsedArgs, nil

}
